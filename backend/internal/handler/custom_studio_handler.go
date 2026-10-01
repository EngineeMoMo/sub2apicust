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
	root string
	mu   sync.RWMutex
}

func RegisterStudioRoutes(v1 *gin.RouterGroup, dataDir string, jwtAuth, adminAuth, auditLog gin.HandlerFunc, settingService *service.SettingService, limiter *middleware.PanelRateLimiter) {
	handler := &StudioHandler{root: filepath.Join(dataDir, "studio-submissions")}
	public := v1.Group("/studio/gallery")
	public.Use(limiter.Global())
	public.GET("", handler.PublicList)
	public.GET("/:id/media", handler.PublicMedia)
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
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		response.BadRequest(ctx, "请选择图片或视频文件")
		return
	}
	defer studioCleanup("upload file", file.Close)
	content, err := io.ReadAll(io.LimitReader(file, studioFileLimit+1))
	if err != nil || len(content) > studioFileLimit {
		response.Error(ctx, http.StatusRequestEntityTooLarge, "视频文件不能超过30MB")
		return
	}
	entry.MIME = http.DetectContentType(content)
	if entry.Media == "image" {
		if len(content) > 12<<20 || !studioContains([]string{"image/png", "image/jpeg", "image/webp"}, entry.MIME) {
			response.BadRequest(ctx, "图片仅支持PNG、JPEG、WebP，最多12MB")
			return
		}
		config, _, err := image.DecodeConfig(bytes.NewReader(content))
		if err != nil || config.Width < 32 || config.Height < 32 || int64(config.Width)*int64(config.Height) > 24000000 {
			response.BadRequest(ctx, "图片无法解析或尺寸不符：至少32像素，最多2400万像素")
			return
		}
	} else if entry.MIME != "video/mp4" {
		response.BadRequest(ctx, "视频仅支持MP4，请先确认可完整播放")
		return
	}
	entry.Size = int64(len(content))
	handler.mu.Lock()
	defer handler.mu.Unlock()
	entries, err := handler.entries()
	if err != nil {
		response.InternalError(ctx, "投稿存储不可用，请稍后重试")
		return
	}
	var size int64
	owned, pending := 0, 0
	for _, existing := range entries {
		size += existing.Size
		if existing.OwnerID == subject.UserID {
			owned++
			if existing.Status == "pending" {
				pending++
			}
		}
	}
	if len(entries) >= 1000 || size+entry.Size > 512<<20 || owned >= 20 || pending >= 5 {
		response.Error(ctx, http.StatusTooManyRequests, "投稿额度已满：每人最多5份待审核、20份投稿；请联系管理员清理存储")
		return
	}
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		response.InternalError(ctx, "无法创建投稿编号")
		return
	}
	entry.ID = hex.EncodeToString(token)
	entry.MediaURL = "/api/v1/studio/gallery/" + entry.ID + "/media"
	if err := os.MkdirAll(handler.root, 0700); err != nil {
		response.InternalError(ctx, "投稿目录不可写")
		return
	}
	mediaPath := filepath.Join(handler.root, entry.ID+".media")
	if err := os.WriteFile(mediaPath, content, 0600); err != nil {
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
	response.Success(ctx, entry)
}
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
	entry.Status, entry.Reason = "withdrawn", "作者撤回"
	if err := handler.save(entry); err != nil {
		response.InternalError(ctx, "撤回未保存，请重试")
		return
	}
	response.Success(ctx, entry)
}
