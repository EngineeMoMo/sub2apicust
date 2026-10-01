//go:build embed

package web

import (
	"bytes"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var familyImagePath = regexp.MustCompile(`^assets/[a-z0-9-]+\.(webp|mp4)$`)

const familyRecipeCSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data: https: http://localhost:* http://127.0.0.1:* http://[::1]:*; connect-src 'self' https: http://localhost:* http://127.0.0.1:* http://[::1]:*; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'self'"
// [CUSTOM] 工坊仅可读取同源审核通过的公开作品目录，不开放外部模型连接。
const familyStudioCSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; media-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'self'"

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
	allowed := filename == "index.html" || filename == "theme.css" || (product == "studio" && (filename == "host-client.js" || filename == "new-gallery.mjs"))
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
	http.ServeContent(ctx.Writer, ctx.Request, filename, time.Time{}, bytes.NewReader(content))
	ctx.Abort()
	return true
}
