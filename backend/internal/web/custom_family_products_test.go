//go:build embed

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestFamilyProductStaticBoundaries(t *testing.T) {
	assets := fstest.MapFS{
		"recipes/index.html":         {Data: []byte("recipe-product")},
		"recipes/bundle.js":          {Data: []byte("recipe-script")},
		"studio/index.html":          {Data: []byte("studio-product")},
		"studio/app.mjs":             {Data: []byte("studio-script")},
		"studio/host-client.js":      {Data: []byte("host-adapter")},
		"studio/new-gallery.mjs":     {Data: []byte("new-gallery")},
		"studio/assets/film.mp4":     {Data: []byte("0123456789")},
		"studio/assets/example.webp": {Data: []byte("image-content")},
		"studio/assets/private.png":  {Data: []byte("not-public")},
		"studio/README.md":           {Data: []byte("not-public")},
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
				require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
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
	}
}
