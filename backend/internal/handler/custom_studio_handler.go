package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp"
)

const studioFileLimit = 30 << 20

// 投稿额度：只统计仍保留素材的投稿（待审核、已公开、已下架）。退回只留记录、删除素材；撤回删除记录和素材。
const (
	studioImageLimit       = 12 << 20
	studioStorageLimit     = 512 << 20
	studioRecordLimit      = 1000
	studioOwnedLimit       = 20 // 每人保留素材的投稿＋退回记录
	studioPendingLimit     = 5
	studioUploadConcurrent = 4 // 同时处理的上传，限制临时文件和内存峰值
)

// 视频成片与前端一致暂未开放，后端同样拒绝，不能靠隐藏选项绕过。
const studioVideoSubmissionsEnabled = false

var studioIDPattern = regexp.MustCompile("^[a-f0-9]{32}$")
var studioCategories = strings.Fields("年轻人像 时尚肖像 Cosplay 动漫二次元 科技机甲 动物自然 奇幻风景 电商产品 场景插画 海报社媒 空间设计")
var studioStyles = strings.Fields("写实摄影 电影感 二次元 3D手作 概念设计 平面海报 水彩 水墨 像素 美漫 剪纸 复古未来")

// 清理失败保留原业务错误或已发送的响应；记录服务端错误，不记录投稿正文。
func studioCleanup(operation string, cleanup func() error) {
	if err := cleanup(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("studio %s cleanup failed: %v", operation, err)
	}
}

type StudioPublicEntry struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Author     string    `json:"author"`
	Category   string    `json:"category"`
	Style      string    `json:"style"`
	Media      string    `json:"media"`
	Prompt     string    `json:"prompt"`
	PromptKind string    `json:"prompt_kind"`
	Model      string    `json:"model"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	MediaURL   string    `json:"media_url"`
}
type StudioEntry struct {
	StudioPublicEntry
	OwnerID    int64     `json:"owner_id"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason"`
	MIME       string    `json:"mime"`
	Size       int64     `json:"size"`
	ReviewedAt time.Time `json:"reviewed_at"`
	ReviewerID int64     `json:"reviewer_id"`
}
type StudioHandler struct {
	root       string
	allowVideo bool
	mu         sync.RWMutex

	uploadOnce sync.Once
	uploads    chan struct{}
	uploadMu   sync.Mutex
	uploading  map[int64]bool
}

func RegisterStudioRoutes(v1 *gin.RouterGroup, dataDir string, jwtAuth, adminAuth, auditLog gin.HandlerFunc, settingService *service.SettingService, limiter *middleware.PanelRateLimiter) {
	handler := &StudioHandler{root: filepath.Join(dataDir, "studio-submissions"), allowVideo: studioVideoSubmissionsEnabled}
	// 匿名可访问：按用户计数的 Global 没有登录主体时直接放行，这里必须按公网 IP 限流。
	public := v1.Group("/studio/gallery")
	public.GET("", limiter.PublicIPScoped("studio-gallery"), handler.PublicList)
	public.GET("/:id/media", limiter.PublicIPScoped("studio-media"), handler.PublicMedia)
	user := v1.Group("/studio/submissions")
	user.Use(jwtAuth, middleware.BackendModeUserGuard(settingService), limiter.Global(), auditLog)
	user.GET("", handler.Mine)
	user.POST("", limiter.Heavy(), handler.Submit)
	user.GET("/:id/media", handler.PrivateMedia)
	user.POST("/:id/withdraw", handler.Withdraw)
	admin := v1.Group("/admin/studio/submissions")
	admin.Use(adminAuth, limiter.Global(), auditLog, middleware.AdminComplianceGuard(settingService))
	admin.GET("", handler.ReviewList)
	admin.GET("/:id/media", handler.PrivateMedia)
	admin.POST("/:id/review", handler.Review)
	admin.DELETE("/:id", handler.AdminDelete)
}

// 同一用户同时只处理一份上传，全站同时最多 studioUploadConcurrent 份；返回释放函数。
func (handler *StudioHandler) acquireUpload(userID int64) (func(), bool) {
	handler.uploadOnce.Do(func() {
		handler.uploads = make(chan struct{}, studioUploadConcurrent)
		handler.uploading = map[int64]bool{}
	})
	handler.uploadMu.Lock()
	if handler.uploading[userID] {
		handler.uploadMu.Unlock()
		return nil, false
	}
	select {
	case handler.uploads <- struct{}{}:
	default:
		handler.uploadMu.Unlock()
		return nil, false
	}
	handler.uploading[userID] = true
	handler.uploadMu.Unlock()
	return func() {
		handler.uploadMu.Lock()
		delete(handler.uploading, userID)
		handler.uploadMu.Unlock()
		<-handler.uploads
	}, true
}

// 待审核、已公开、已下架仍保留素材，占用存储；退回和撤回不保留素材。
func studioKeepsMedia(status string) bool {
	return status == "pending" || status == "published" || status == "unpublished"
}

// studioQuota 检查额度并返回具体原因；addSize 为本次上传大小，预检时传 1。
func studioQuota(entries []StudioEntry, userID, addSize int64) string {
	var size int64
	records, owned, pending := 0, 0, 0
	for _, existing := range entries {
		if studioKeepsMedia(existing.Status) {
			size += existing.Size
			records++
		}
		if existing.OwnerID == userID && existing.Status != "withdrawn" {
			owned++
			if existing.Status == "pending" {
				pending++
			}
		}
	}
	switch {
	case pending >= studioPendingLimit:
		return "你已有5份投稿在等待审核，请等审核完成后再提交"
	case owned >= studioOwnedLimit:
		return "你已有20份投稿或退回记录，请先撤回不需要的投稿再提交"
	case records >= studioRecordLimit || size+addSize > studioStorageLimit:
		return "全站投稿存储已满，请稍后再试或联系管理员清理"
	}
	return ""
}

func (handler *StudioHandler) removeMedia(id string) {
	studioCleanup("media file", func() error { return os.Remove(filepath.Join(handler.root, id+".media")) })
}
func studioContains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
func studioText(value string, minimum, maximum int) bool {
	length := utf8.RuneCountInString(value)
	return utf8.ValidString(value) && length >= minimum && length <= maximum
}
func (handler *StudioHandler) entries() ([]StudioEntry, error) {
	files, err := os.ReadDir(handler.root)
	if errors.Is(err, os.ErrNotExist) {
		return []StudioEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	entries := make([]StudioEntry, 0)
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(file.Name(), ".json")
		if !studioIDPattern.MatchString(id) {
			continue
		}
		entry, err := handler.read(id)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(first, second int) bool { return entries[first].CreatedAt.After(entries[second].CreatedAt) })
	return entries, nil
}
func (handler *StudioHandler) read(id string) (StudioEntry, error) {
	var entry StudioEntry
	if !studioIDPattern.MatchString(id) {
		return entry, os.ErrNotExist
	}
	content, err := os.ReadFile(filepath.Join(handler.root, id+".json"))
	if err != nil {
		return entry, err
	}
	err = json.Unmarshal(content, &entry)
	if err == nil && entry.ID != id {
		err = errors.New("invalid studio record")
	}
	return entry, err
}
func (handler *StudioHandler) save(entry StudioEntry) error {
	if err := os.MkdirAll(handler.root, 0700); err != nil {
		return err
	}
	content, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(handler.root, ".record-")
	if err != nil {
		return err
	}
	name := file.Name()
	defer studioCleanup("temporary record", func() error { return os.Remove(name) })
	if _, err = file.Write(content); err != nil {
		studioCleanup("record file", file.Close)
		return err
	}
	if err = file.Sync(); err != nil {
		studioCleanup("record file", file.Close)
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(name, filepath.Join(handler.root, entry.ID+".json"))
}
func (handler *StudioHandler) Submit(ctx *gin.Context) {
	subject, authenticated := middleware.GetAuthSubjectFromContext(ctx)
	if !authenticated {
		response.Unauthorized(ctx, "请先登录再投稿")
		return
	}
	release, ok := handler.acquireUpload(subject.UserID)
	if !ok {
		response.Error(ctx, http.StatusTooManyRequests, "上传通道繁忙或你已有上传在进行中，请稍后再试")
		return
	}
	defer release()
	// 读取正文前先查额度：额度已满时不接收30MB正文，也不写临时文件。
	handler.mu.RLock()
	existing, err := handler.entries()
	handler.mu.RUnlock()
	if err != nil {
		response.InternalError(ctx, "投稿存储不可用，请稍后重试")
		return
	}
	if reason := studioQuota(existing, subject.UserID, 1); reason != "" {
		response.Error(ctx, http.StatusTooManyRequests, reason)
		return
	}
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, studioFileLimit+(64<<10))
	if err := ctx.Request.ParseMultipartForm(1 << 20); err != nil {
		if ctx.Request.MultipartForm != nil {
			defer studioCleanup("multipart files", ctx.Request.MultipartForm.RemoveAll)
		}
		response.Error(ctx, http.StatusRequestEntityTooLarge, "文件或表单过大，视频最多30MB、图片最多12MB")
		return
	}
	defer studioCleanup("multipart files", ctx.Request.MultipartForm.RemoveAll)
	entry := StudioEntry{StudioPublicEntry: StudioPublicEntry{Title: strings.TrimSpace(ctx.PostForm("title")), Author: strings.TrimSpace(ctx.PostForm("author")),
		Category: ctx.PostForm("category"), Style: ctx.PostForm("style"), Media: ctx.PostForm("media"),
		Prompt: strings.TrimSpace(ctx.PostForm("prompt")), PromptKind: ctx.PostForm("prompt_kind"),
		Model: strings.TrimSpace(ctx.PostForm("model")), Notes: strings.TrimSpace(ctx.PostForm("notes")), CreatedAt: time.Now().UTC()},
		OwnerID: subject.UserID, Status: "pending"}
	if !studioText(entry.Title, 2, 80) || !studioText(entry.Author, 1, 40) || !studioText(entry.Prompt, 10, 12000) ||
		!studioText(entry.Model, 1, 80) || !studioText(entry.Notes, 0, 1000) ||
		!studioContains(studioCategories, entry.Category) || !studioContains(studioStyles, entry.Style) ||
		!studioContains([]string{"image", "video"}, entry.Media) || !studioContains([]string{"actual", "reference"}, entry.PromptKind) ||
		ctx.PostForm("rights_confirmed") != "true" || ctx.PostForm("publish_consent") != "true" {
		response.BadRequest(ctx, "请完整填写题材、风格、提示词与模型，并确认素材权利和审核后公开授权")
		return
	}
	if entry.Media == "video" && !handler.allowVideo {
		response.BadRequest(ctx, "视频投稿暂未开放，请先分享图片作品")
		return
	}
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		response.BadRequest(ctx, "请选择图片或视频文件")
		return
	}
	defer studioCleanup("upload file", file.Close)
	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		response.BadRequest(ctx, "无法读取上传文件")
		return
	}
	entry.MIME = http.DetectContentType(head[:n])
	limit := int64(studioFileLimit)
	if entry.Media == "image" {
		limit = studioImageLimit
		if !studioContains([]string{"image/png", "image/jpeg", "image/webp"}, entry.MIME) {
			response.BadRequest(ctx, "图片仅支持PNG、JPEG、WebP，最多12MB")
			return
		}
	} else if entry.MIME != "video/mp4" {
		response.BadRequest(ctx, "视频仅支持MP4，请先确认可完整播放")
		return
	}
	if err := os.MkdirAll(handler.root, 0700); err != nil {
		response.InternalError(ctx, "投稿目录不可写")
		return
	}
	// 边读边写临时文件，不把整份素材放进内存；校验全部通过后再改名为正式素材。
	temp, err := os.CreateTemp(handler.root, ".upload-")
	if err != nil {
		response.InternalError(ctx, "保存素材失败")
		return
	}
	tempName, tempClosed := temp.Name(), false
	defer func() {
		if !tempClosed {
			studioCleanup("upload temp file", temp.Close)
		}
		studioCleanup("upload temp file", func() error { return os.Remove(tempName) })
	}()
	written, err := io.Copy(temp, io.LimitReader(io.MultiReader(bytes.NewReader(head[:n]), file), limit+1))
	if err != nil {
		response.Error(ctx, http.StatusRequestEntityTooLarge, "文件或表单过大，视频最多30MB、图片最多12MB")
		return
	}
	if written > limit {
		response.Error(ctx, http.StatusRequestEntityTooLarge, map[bool]string{true: "图片最多12MB", false: "视频文件不能超过30MB"}[entry.Media == "image"])
		return
	}
	if entry.Media == "image" {
		if _, err := temp.Seek(0, io.SeekStart); err != nil {
			response.InternalError(ctx, "保存素材失败")
			return
		}
		config, _, err := image.DecodeConfig(temp)
		if err != nil || config.Width < 32 || config.Height < 32 || int64(config.Width)*int64(config.Height) > 24000000 {
			response.BadRequest(ctx, "图片无法解析或尺寸不符：至少32像素，最多2400万像素")
			return
		}
	}
	if err := temp.Sync(); err != nil {
		response.InternalError(ctx, "保存素材失败")
		return
	}
	tempClosed = true
	if err := temp.Close(); err != nil {
		response.InternalError(ctx, "保存素材失败")
		return
	}
	entry.Size = written
	handler.mu.Lock()
	defer handler.mu.Unlock()
	entries, err := handler.entries()
	if err != nil {
		response.InternalError(ctx, "投稿存储不可用，请稍后重试")
		return
	}
	if reason := studioQuota(entries, subject.UserID, entry.Size); reason != "" {
		response.Error(ctx, http.StatusTooManyRequests, reason)
		return
	}
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		response.InternalError(ctx, "无法创建投稿编号")
		return
	}
	entry.ID = hex.EncodeToString(token)
	entry.MediaURL = "/api/v1/studio/gallery/" + entry.ID + "/media"
	mediaPath := filepath.Join(handler.root, entry.ID+".media")
	if err := os.Rename(tempName, mediaPath); err != nil {
		response.InternalError(ctx, "保存素材失败")
		return
	}
	if err := handler.save(entry); err != nil {
		studioCleanup("orphan media", func() error { return os.Remove(mediaPath) })
		response.InternalError(ctx, "保存投稿记录失败")
		return
	}
	response.Created(ctx, entry)
}
func (handler *StudioHandler) list(ctx *gin.Context, accept func(StudioEntry) bool) {
	handler.mu.RLock()
	defer handler.mu.RUnlock()
	entries, err := handler.entries()
	if err != nil {
		response.InternalError(ctx, "作品记录暂时不可用")
		return
	}
	result := make([]StudioEntry, 0)
	for _, entry := range entries {
		if accept(entry) {
			result = append(result, entry)
		}
	}
	ctx.Header("Cache-Control", "no-store")
	response.Success(ctx, result)
}
func (handler *StudioHandler) Mine(ctx *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(ctx)
	if !ok {
		response.Unauthorized(ctx, "请先登录")
		return
	}
	handler.list(ctx, func(entry StudioEntry) bool { return entry.OwnerID == subject.UserID })
}
func (handler *StudioHandler) ReviewList(ctx *gin.Context) {
	role, ok := middleware.GetUserRoleFromContext(ctx)
	if !ok || role != service.RoleAdmin {
		response.Forbidden(ctx, "只有管理员可以审核")
		return
	}
	handler.list(ctx, func(entry StudioEntry) bool { return true })
}
func (handler *StudioHandler) PublicList(ctx *gin.Context) {
	handler.mu.RLock()
	defer handler.mu.RUnlock()
	entries, err := handler.entries()
	if err != nil {
		response.InternalError(ctx, "作品库暂时不可用")
		return
	}
	result := make([]StudioPublicEntry, 0)
	for _, entry := range entries {
		if entry.Status == "published" {
			result = append(result, entry.StudioPublicEntry)
		}
	}
	ctx.Header("Cache-Control", "no-store")
	response.Success(ctx, result)
}
func (handler *StudioHandler) media(ctx *gin.Context, public bool) {
	handler.mu.RLock()
	defer handler.mu.RUnlock()
	entry, err := handler.read(ctx.Param("id"))
	if err != nil {
		response.NotFound(ctx, "作品不存在")
		return
	}
	if public {
		if entry.Status != "published" {
			response.NotFound(ctx, "作品不存在")
			return
		}
	} else {
		subject, authenticated := middleware.GetAuthSubjectFromContext(ctx)
		role, _ := middleware.GetUserRoleFromContext(ctx)
		if !authenticated || (entry.OwnerID != subject.UserID && role != service.RoleAdmin) {
			response.NotFound(ctx, "作品不存在")
			return
		}
	}
	file, err := os.Open(filepath.Join(handler.root, entry.ID+".media"))
	if err != nil {
		response.NotFound(ctx, "素材不存在")
		return
	}
	defer studioCleanup("media file", file.Close)
	ctx.Header("Content-Type", entry.MIME)
	ctx.Header("Cache-Control", "no-store")
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	http.ServeContent(ctx.Writer, ctx.Request, "studio-media", entry.CreatedAt, file)
}
func (handler *StudioHandler) PublicMedia(ctx *gin.Context)  { handler.media(ctx, true) }
func (handler *StudioHandler) PrivateMedia(ctx *gin.Context) { handler.media(ctx, false) }
func (handler *StudioHandler) Review(ctx *gin.Context) {
	role, ok := middleware.GetUserRoleFromContext(ctx)
	subject, authenticated := middleware.GetAuthSubjectFromContext(ctx)
	if !ok || role != service.RoleAdmin || !authenticated {
		response.Forbidden(ctx, "只有管理员可以审核")
		return
	}
	var input struct {
		Action    string `json:"action"`
		Expected  string `json:"expected_status"`
		Reason    string `json:"reason"`
		Confirmed bool   `json:"review_confirmed"`
	}
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 8192)
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.BadRequest(ctx, "审核参数无效")
		return
	}
	handler.mu.Lock()
	defer handler.mu.Unlock()
	entry, err := handler.read(ctx.Param("id"))
	if err != nil {
		response.NotFound(ctx, "投稿不存在")
		return
	}
	if entry.Status != input.Expected {
		response.Error(ctx, http.StatusConflict, "状态已变化，请重新打开作品再审核")
		return
	}
	input.Reason = strings.TrimSpace(input.Reason)
	switch input.Action {
	case "approve":
		if (entry.Status != "pending" && entry.Status != "unpublished") || !input.Confirmed {
			response.BadRequest(ctx, "请先检查素材、人物年龄、权利与提示词，再明确确认公开")
			return
		}
		entry.Status, entry.Reason = "published", ""
	case "reject", "unpublish":
		if (input.Action == "reject" && entry.Status != "pending") || (input.Action == "unpublish" && entry.Status != "published") || !studioText(input.Reason, 3, 500) {
			response.BadRequest(ctx, "当前状态不支持此操作，退回或下架需填写3–500字原因")
			return
		}
		entry.Status = map[string]string{"reject": "rejected", "unpublish": "unpublished"}[input.Action]
		entry.Reason = input.Reason
	default:
		response.BadRequest(ctx, "不支持的审核操作")
		return
	}
	entry.ReviewerID, entry.ReviewedAt = subject.UserID, time.Now().UTC()
	if err := handler.save(entry); err != nil {
		response.InternalError(ctx, "审核未保存，请重试")
		return
	}
	// 退回不能再审核通过：只保留记录和原因，删除素材，释放存储。
	if entry.Status == "rejected" {
		handler.removeMedia(entry.ID)
	}
	response.Success(ctx, entry)
}

// Withdraw 由作者撤回：删除记录和素材，不再计入任何额度，无法恢复。
func (handler *StudioHandler) Withdraw(ctx *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(ctx)
	if !ok {
		response.Unauthorized(ctx, "请先登录")
		return
	}
	handler.mu.Lock()
	defer handler.mu.Unlock()
	entry, err := handler.read(ctx.Param("id"))
	if err != nil || entry.OwnerID != subject.UserID {
		response.NotFound(ctx, "投稿不存在")
		return
	}
	if err := handler.remove(entry.ID); err != nil {
		response.InternalError(ctx, "撤回未完成，请重试")
		return
	}
	entry.Status, entry.Reason = "withdrawn", "作者撤回"
	response.Success(ctx, entry)
}

// AdminDelete 由管理员删除任意投稿的记录和素材，用于清理存储；操作经管理员审计中间件记录。
func (handler *StudioHandler) AdminDelete(ctx *gin.Context) {
	role, ok := middleware.GetUserRoleFromContext(ctx)
	if !ok || role != service.RoleAdmin {
		response.Forbidden(ctx, "只有管理员可以删除投稿")
		return
	}
	handler.mu.Lock()
	defer handler.mu.Unlock()
	entry, err := handler.read(ctx.Param("id"))
	if err != nil {
		response.NotFound(ctx, "投稿不存在")
		return
	}
	if err := handler.remove(entry.ID); err != nil {
		response.InternalError(ctx, "删除未完成，请重试")
		return
	}
	response.Success(ctx, gin.H{"id": entry.ID, "deleted": true})
}

// remove 先删记录再删素材：记录删除失败时什么都不改；素材残留只记日志，不影响额度统计。
func (handler *StudioHandler) remove(id string) error {
	if !studioIDPattern.MatchString(id) {
		return os.ErrNotExist
	}
	if err := os.Remove(filepath.Join(handler.root, id+".json")); err != nil {
		return err
	}
	handler.removeMedia(id)
	return nil
}
