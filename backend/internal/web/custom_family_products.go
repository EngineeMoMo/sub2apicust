//go:build embed

package web

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// [CUSTOM] 配方与工坊静态文件省流量：文本按内容摘要做 ETag 并要求每次复查（发布后立即生效，重复访问只回 304），
// 图片缓存一天；大于 1KB 的文本在客户端支持时返回 gzip，压缩结果按内容摘要缓存在内存，同一内容只压缩一次。
var familyGzipCache sync.Map

const familyGzipMinBytes = 1024

func familyTextFile(filename string) bool {
	switch path.Ext(filename) {
	case ".html", ".css", ".js", ".mjs":
		return true
	}
	return false
}

// acceptsGzip 按 Accept-Encoding 的逐项取值判断，q=0 视为拒绝。
func acceptsGzip(header string) bool {
	for _, part := range strings.Split(header, ",") {
		fields := strings.Split(part, ";")
		name := strings.ToLower(strings.TrimSpace(fields[0]))
		if name != "gzip" && name != "*" {
			continue
		}
		accepted := true
		for _, field := range fields[1:] {
			key, value, ok := strings.Cut(strings.TrimSpace(field), "=")
			if ok && strings.EqualFold(key, "q") {
				if q, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil && q <= 0 {
					accepted = false
				}
			}
		}
		if accepted {
			return true
		}
	}
	return false
}

func familyGzip(digest string, content []byte) []byte {
	if cached, ok := familyGzipCache.Load(digest); ok {
		if compressed, valid := cached.([]byte); valid {
			return compressed
		}
	}
	var buffer bytes.Buffer
	writer, _ := gzip.NewWriterLevel(&buffer, gzip.BestCompression)
	_, _ = writer.Write(content)
	_ = writer.Close()
	compressed := buffer.Bytes()
	familyGzipCache.Store(digest, compressed)
	return compressed
}

var familyImagePath = regexp.MustCompile(`^assets/[a-z0-9-]+\.(webp|mp4)$`)

// [CUSTOM] 工坊图生短视频文件名带内容摘要（clip-<条目>-<10位sha256>.mp4），内容变化即换地址，
// 可让浏览器长期缓存，重复观看不再占用服务器流量；其余产品文件仍不缓存。
var familyClipPath = regexp.MustCompile(`^assets/clip-[a-z0-9-]+-([a-f0-9]{10})\.mp4$`)

const familyRecipeCSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data: https: http://localhost:* http://127.0.0.1:*; connect-src 'self' https: http://localhost:* http://127.0.0.1:*; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'self'"

// [CUSTOM] 公开浏览的工坊仅可读取同源审核通过的公开作品目录，不开放外部模型连接。
// media-src 的 blob: 只用于播放本页从同源下载并缓存在浏览器里的短视频，不放开任何外部来源。
const familyStudioCSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; media-src 'self' blob:; connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'self'"

// [CUSTOM] 控制台内嵌（?embedded=1）的工坊可以在用户明确授权后直连所选 Images API，
// 并显示接口返回的 Base64 或 HTTPS 图片；脚本仍只允许同源，不放开 unsafe-inline。
const familyStudioEmbeddedCSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data: https: http://localhost:* http://127.0.0.1:*; media-src 'self' blob:; connect-src 'self' https: http://localhost:* http://127.0.0.1:*; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'self'"

func serveFamilyProduct(ctx *gin.Context, assets fs.FS) bool {
	requestPath := ctx.Request.URL.Path
	product := ""
	for _, candidate := range []string{"recipes", "studio"} {
		prefix := "/" + candidate
		if requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/") {
			product = candidate
			break
		}
	}
	if product == "" {
		return false
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Header("X-Frame-Options", "SAMEORIGIN")
	ctx.Header("Referrer-Policy", "no-referrer")
	policy := familyStudioCSP
	if product == "recipes" {
		policy = familyRecipeCSP
	} else if ctx.Request.URL.Query().Get("embedded") == "1" {
		policy = familyStudioEmbeddedCSP
	}
	ctx.Header("Content-Security-Policy", policy)
	if ctx.Request.Method != http.MethodGet && ctx.Request.Method != http.MethodHead {
		ctx.Header("Allow", "GET, HEAD")
		ctx.AbortWithStatus(http.StatusMethodNotAllowed)
		return true
	}
	if requestPath == "/"+product {
		location := requestPath + "/"
		if ctx.Request.URL.RawQuery != "" {
			location += "?" + ctx.Request.URL.RawQuery
		}
		ctx.Redirect(http.StatusPermanentRedirect, location)
		ctx.Abort()
		return true
	}
	filename := strings.TrimPrefix(requestPath, "/"+product+"/")
	if filename == "" {
		filename = "index.html"
	}
	allowed := filename == "index.html" || filename == "theme.css" || (product == "studio" && (filename == "host-client.js" || filename == "new-gallery.mjs" || filename == "generate.mjs" || filename == "videos.mjs" || filename == "video-cache.mjs"))
	if product == "recipes" {
		allowed = allowed || filename == "bundle.js"
	} else {
		allowed = allowed || filename == "app.mjs" || filename == "core.mjs" || filename == "icons.mjs" || filename == "catalog.mjs" || filename == "assets/mofa-mark-flat.png" || familyImagePath.MatchString(filename)
	}
	if !allowed {
		ctx.AbortWithStatus(http.StatusNotFound)
		return true
	}
	content, err := fs.ReadFile(assets, product+"/"+filename)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return true
	}
	contentTypes := map[string]string{
		".html": "text/html; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".js":   "text/javascript; charset=utf-8",
		".mjs":  "text/javascript; charset=utf-8",
		".webp": "image/webp",
		".png":  "image/png",
		".mp4":  "video/mp4",
	}
	ctx.Header("Content-Type", contentTypes[path.Ext(filename)])
	if match := familyClipPath.FindStringSubmatch(filename); product == "studio" && match != nil {
		ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
		ctx.Header("ETag", `"`+match[1]+`"`)
		http.ServeContent(ctx.Writer, ctx.Request, filename, time.Time{}, bytes.NewReader(content))
		ctx.Abort()
		return true
	}
	sum := sha256.Sum256(content)
	digest := hex.EncodeToString(sum[:8])
	if familyTextFile(filename) {
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("Vary", "Accept-Encoding")
		if len(content) >= familyGzipMinBytes && acceptsGzip(ctx.GetHeader("Accept-Encoding")) {
			compressed := familyGzip(digest, content)
			ctx.Header("Content-Encoding", "gzip")
			ctx.Header("Content-Length", strconv.Itoa(len(compressed)))
			ctx.Header("ETag", `"`+digest+`-gz"`)
			http.ServeContent(ctx.Writer, ctx.Request, filename, time.Time{}, bytes.NewReader(compressed))
			ctx.Abort()
			return true
		}
	} else {
		ctx.Header("Cache-Control", "public, max-age=86400")
	}
	ctx.Header("ETag", `"`+digest+`"`)
	http.ServeContent(ctx.Writer, ctx.Request, filename, time.Time{}, bytes.NewReader(content))
	ctx.Abort()
	return true
}
