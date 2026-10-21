//go:build embed

package web

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestFamilyProductStaticBoundaries(t *testing.T) {
	assets := fstest.MapFS{
		"recipes/index.html":                        {Data: []byte("recipe-product")},
		"recipes/bundle.js":                         {Data: []byte("recipe-script")},
		"studio/index.html":                         {Data: []byte("studio-product")},
		"studio/app.mjs":                            {Data: []byte("studio-script")},
		"studio/host-client.js":                     {Data: []byte("host-adapter")},
		"studio/new-gallery.mjs":                    {Data: []byte("new-gallery")},
		"studio/generate.mjs":                       {Data: []byte("generate-module")},
		"studio/videos.mjs":                         {Data: []byte("clip-manifest")},
		"studio/video-cache.mjs":                    {Data: []byte("clip-cache")},
		"studio/video-plan.mjs":                     {Data: []byte("not-public")},
		"studio/assets/film.mp4":                    {Data: []byte("0123456789")},
		"studio/assets/clip-ink-koi-0123456789.mp4": {Data: []byte("abcdefghij")},
		"studio/assets/example.webp":                {Data: []byte("image-content")},
		"studio/assets/private.png":                 {Data: []byte("not-public")},
		"studio/README.md":                          {Data: []byte("not-public")},
	}
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		if !serveFamilyProduct(ctx, assets) {
			ctx.String(http.StatusOK, "main-spa")
		}
	})
	tests := []struct {
		name, method, target, body string
		status                     int
	}{
		{"recipe", "GET", "/recipes/", "recipe-product", 200},
		{"studio", "GET", "/studio/", "studio-product", 200},
		{"module", "GET", "/studio/app.mjs", "studio-script", 200},
		{"adapter", "GET", "/studio/host-client.js", "host-adapter", 200},
		{"new-gallery", "GET", "/studio/new-gallery.mjs", "new-gallery", 200},
		{"generate", "GET", "/studio/generate.mjs", "generate-module", 200},
		{"recipe-generate", "GET", "/recipes/generate.mjs", "", 404},
		{"clip-manifest", "GET", "/studio/videos.mjs", "clip-manifest", 200},
		{"clip-cache", "GET", "/studio/video-cache.mjs", "clip-cache", 200},
		{"video-plan", "GET", "/studio/video-plan.mjs", "", 404},
		{"recipe-clip-manifest", "GET", "/recipes/videos.mjs", "", 404},
		{"image", "GET", "/studio/assets/example.webp", "image-content", 200},
		{"head", "HEAD", "/recipes/", "", 200},
		{"missing", "GET", "/recipes/missing.js", "", 404},
		{"source", "GET", "/studio/README.md", "", 404},
		{"raw-image", "GET", "/studio/assets/private.png", "", 404},
		{"listing", "GET", "/studio/assets/", "", 404},
		{"traversal", "GET", "/studio/../index.html", "", 404},
		{"encoded-traversal", "GET", "/recipes/%2e%2e/index.html", "", 404},
		{"post", "POST", "/recipes/", "", 405},
		{"delete", "DELETE", "/studio/", "", 405},
		{"main-route", "GET", "/dashboard", "main-spa", 200},
	}
	for _, scenario := range tests {
		t.Run(scenario.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(scenario.method, scenario.target, nil))
			require.Equal(t, scenario.status, response.Code)
			require.Equal(t, scenario.body, response.Body.String())
			if scenario.name != "main-route" {
				// 成功的文本每次复查、图片缓存一天，错误与非 GET 仍不缓存。
				wantCache := "no-store"
				if scenario.status == 200 {
					wantCache = "no-cache"
					if scenario.name == "image" {
						wantCache = "public, max-age=86400"
					}
					require.NotEmpty(t, response.Header().Get("ETag"))
				}
				require.Equal(t, wantCache, response.Header().Get("Cache-Control"))
				require.NotContains(t, response.Header().Get("Content-Security-Policy"), "unsafe-inline")
				require.Equal(t, "SAMEORIGIN", response.Header().Get("X-Frame-Options"))
				require.Contains(t, response.Header().Get("Content-Security-Policy"), "frame-ancestors 'self'")
			}
			if scenario.name == "module" {
				require.Equal(t, "text/javascript; charset=utf-8", response.Header().Get("Content-Type"))
			}
			if scenario.status == 405 {
				require.Equal(t, "GET, HEAD", response.Header().Get("Allow"))
			}
		})
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/recipes?example=meeting", nil))
	require.Equal(t, 308, response.Code)
	require.Equal(t, "/recipes/?example=meeting", response.Header().Get("Location"))
	request := httptest.NewRequest("GET", "/studio/assets/film.mp4", nil)
	request.Header.Set("Range", "bytes=2-5")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, 206, response.Code)
	require.Equal(t, "2345", response.Body.String())
	require.Equal(t, "video/mp4", response.Header().Get("Content-Type"))
	require.Equal(t, "bytes 2-5/10", response.Header().Get("Content-Range"))
	require.Contains(t, response.Header().Get("Content-Security-Policy"), "media-src 'self'")
	require.Contains(t, response.Header().Get("Content-Security-Policy"), "connect-src 'self'")
	require.NotContains(t, response.Header().Get("Content-Security-Policy"), "connect-src https:")
	require.NotContains(t, response.Header().Get("Content-Security-Policy"), "https:")
	require.Equal(t, "public, max-age=86400", response.Header().Get("Cache-Control"))
	require.NotEmpty(t, response.Header().Get("ETag"))

	// 带内容摘要的图生短视频可长期缓存：重复观看命中浏览器缓存或 304，不再下载全文；仍支持 Range。
	clip := "/studio/assets/clip-ink-koi-0123456789.mp4"
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", clip, nil))
	require.Equal(t, 200, response.Code)
	require.Equal(t, "abcdefghij", response.Body.String())
	require.Equal(t, "video/mp4", response.Header().Get("Content-Type"))
	require.Equal(t, "public, max-age=31536000, immutable", response.Header().Get("Cache-Control"))
	require.Equal(t, `"0123456789"`, response.Header().Get("ETag"))
	require.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
	request = httptest.NewRequest("GET", clip, nil)
	request.Header.Set("If-None-Match", `"0123456789"`)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, 304, response.Code)
	require.Empty(t, response.Body.String())
	request = httptest.NewRequest("GET", clip, nil)
	request.Header.Set("Range", "bytes=0-3")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, 206, response.Code)
	require.Equal(t, "abcd", response.Body.String())
	for _, target := range []string{"/studio/assets/clip-ink-koi-012345678.mp4", "/studio/assets/clip-ink-koi-ABCDEF0123.mp4", "/studio/assets/clips/ink-koi.mp4"} {
		response = httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest("GET", target, nil))
		require.Equal(t, 404, response.Code, target)
		require.Equal(t, "no-store", response.Header().Get("Cache-Control"), target)
	}

	// 控制台内嵌的工坊在用户授权后可以直连 Images API 并显示返回的图片；公开页仍保持只读同源。
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/studio/?embedded=1", nil))
	require.Equal(t, 200, response.Code)
	embedded := response.Header().Get("Content-Security-Policy")
	require.Contains(t, embedded, "connect-src 'self' https:")
	require.Contains(t, embedded, "img-src 'self' data: https:")
	require.Contains(t, embedded, "script-src 'self';")
	require.Contains(t, embedded, "frame-ancestors 'self'")
	require.NotContains(t, embedded, "unsafe-inline")
	require.NotContains(t, embedded, "unsafe-eval")
	// 浏览器会忽略 CSP 里的 IPv6 字面地址并在控制台报错，三份策略都不应包含。
	for _, policy := range []string{familyRecipeCSP, familyStudioCSP, familyStudioEmbeddedCSP} {
		require.NotContains(t, policy, "[::1]")
	}
	require.Contains(t, embedded, "media-src 'self' blob:;")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest("GET", "/studio/", nil))
	public := response.Header().Get("Content-Security-Policy")
	require.NotContains(t, public, "https:")
	// 公开页与内嵌页都只多出 blob:，用于播放从同源下载并缓存在浏览器里的短视频。
	require.Contains(t, public, "media-src 'self' blob:;")
	require.Contains(t, public, "connect-src 'self';")
}

// 省流量：文本按内容摘要 ETag、支持时返回 gzip，重复访问 304；图片缓存一天并可复查；小文件不压缩。
func TestFamilyProductCachingAndCompression(t *testing.T) {
	script := []byte(strings.Repeat("export const gallery = 'magic studio';\n", 200))
	assets := fstest.MapFS{
		"studio/app.mjs":            {Data: script},
		"studio/host-client.js":     {Data: []byte("tiny")},
		"studio/assets/sample.webp": {Data: []byte("webp-image-bytes")},
	}
	router := gin.New()
	router.Use(func(ctx *gin.Context) { serveFamilyProduct(ctx, assets) })
	get := func(target string, headers map[string]string) *httptest.ResponseRecorder {
		request := httptest.NewRequest("GET", target, nil)
		for key, value := range headers {
			request.Header.Set(key, value)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}

	plain := get("/studio/app.mjs", nil)
	require.Equal(t, 200, plain.Code)
	require.Equal(t, string(script), plain.Body.String())
	require.Empty(t, plain.Header().Get("Content-Encoding"))
	require.Equal(t, "no-cache", plain.Header().Get("Cache-Control"))
	require.Equal(t, "Accept-Encoding", plain.Header().Get("Vary"))
	etag := plain.Header().Get("ETag")
	require.Regexp(t, `^"[a-f0-9]{16}"$`, etag)

	zipped := get("/studio/app.mjs", map[string]string{"Accept-Encoding": "br, gzip;q=0.8"})
	require.Equal(t, 200, zipped.Code)
	require.Equal(t, "gzip", zipped.Header().Get("Content-Encoding"))
	require.Equal(t, strings.TrimSuffix(etag, `"`)+`-gz"`, zipped.Header().Get("ETag"))
	require.Equal(t, strconv.Itoa(zipped.Body.Len()), zipped.Header().Get("Content-Length"))
	require.Less(t, zipped.Body.Len(), len(script)/5)
	reader, err := gzip.NewReader(bytes.NewReader(zipped.Body.Bytes()))
	require.NoError(t, err)
	inflated, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, script, inflated)
	require.Contains(t, zipped.Header().Get("Content-Security-Policy"), "script-src 'self'")

	again := get("/studio/app.mjs", map[string]string{"Accept-Encoding": "gzip", "If-None-Match": zipped.Header().Get("ETag")})
	require.Equal(t, 304, again.Code)
	require.Empty(t, again.Body.String())
	require.Equal(t, 304, get("/studio/app.mjs", map[string]string{"If-None-Match": etag}).Code)
	// 版本不匹配（换了压缩方式或内容变了）时返回完整内容。
	require.Equal(t, 200, get("/studio/app.mjs", map[string]string{"If-None-Match": zipped.Header().Get("ETag")}).Code)
	require.Equal(t, 200, get("/studio/app.mjs", map[string]string{"Accept-Encoding": "gzip", "If-None-Match": `"0000000000000000-gz"`}).Code)

	refused := get("/studio/app.mjs", map[string]string{"Accept-Encoding": "gzip;q=0, identity"})
	require.Empty(t, refused.Header().Get("Content-Encoding"))
	require.Equal(t, string(script), refused.Body.String())

	small := get("/studio/host-client.js", map[string]string{"Accept-Encoding": "gzip"})
	require.Equal(t, "tiny", small.Body.String())
	require.Empty(t, small.Header().Get("Content-Encoding"))
	require.Equal(t, "no-cache", small.Header().Get("Cache-Control"))

	image := get("/studio/assets/sample.webp", map[string]string{"Accept-Encoding": "gzip"})
	require.Equal(t, "webp-image-bytes", image.Body.String())
	require.Empty(t, image.Header().Get("Content-Encoding"))
	require.Equal(t, "public, max-age=86400", image.Header().Get("Cache-Control"))
	require.Equal(t, 304, get("/studio/assets/sample.webp", map[string]string{"If-None-Match": image.Header().Get("ETag")}).Code)

	for header, want := range map[string]bool{"gzip": true, "GZIP": true, "deflate, gzip;q=1.0": true, "*": true, "gzip;q=0": false, "gzip; q=0.0": false, "br": false, "": false, "identity": false} {
		require.Equal(t, want, acceptsGzip(header), header)
	}
}

func TestFamilyProductsInBothEmbeddedServers(t *testing.T) {
	server, err := NewFrontendServer(nil)
	require.NoError(t, err)
	for _, handler := range []gin.HandlerFunc{server.Middleware(), ServeEmbeddedFrontend()} {
		router := gin.New()
		router.Use(handler)
		for _, target := range []string{"/recipes/", "/studio/"} {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest("GET", target, nil))
			require.Equal(t, 200, response.Code)
			require.Contains(t, response.Body.String(), `name="mofa-api-site" content="same-origin"`)
			require.NotContains(t, response.Body.String(), `<script>`)
		}
		// 真实打包产物：gzip 解压后与嵌入文件逐字节一致。
		original, err := fs.ReadFile(frontendFS, "dist/studio/app.mjs")
		require.NoError(t, err)
		request := httptest.NewRequest("GET", "/studio/app.mjs", nil)
		request.Header.Set("Accept-Encoding", "gzip")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, 200, response.Code)
		require.Equal(t, "gzip", response.Header().Get("Content-Encoding"))
		reader, err := gzip.NewReader(bytes.NewReader(response.Body.Bytes()))
		require.NoError(t, err)
		inflated, err := io.ReadAll(reader)
		require.NoError(t, err)
		require.Equal(t, original, inflated)
	}
}
