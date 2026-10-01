package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func studioTestRouter(handler *StudioHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		owner, _ := strconv.ParseInt(ctx.GetHeader("Test-Owner"), 10, 64)
		if owner > 0 {
			ctx.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: owner})
			ctx.Set(string(middleware.ContextKeyUserRole), ctx.GetHeader("Test-Role"))
		}
		ctx.Next()
	})
	router.POST("/submit", handler.Submit)
	router.GET("/mine", handler.Mine)
	router.GET("/review", handler.ReviewList)
	router.POST("/review/:id", handler.Review)
	router.POST("/withdraw/:id", handler.Withdraw)
	router.GET("/private/:id", handler.PrivateMedia)
	router.GET("/public/:id", handler.PublicMedia)
	router.GET("/gallery", handler.PublicList)
	return router
}
func studioCall(router *gin.Engine, method, path, body string, owner int64, role string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Test-Owner", strconv.FormatInt(owner, 10))
	request.Header.Set("Test-Role", role)
	output := httptest.NewRecorder()
	router.ServeHTTP(output, request)
	return output
}
func studioPNG(t *testing.T) []byte {
	t.Helper()
	var content bytes.Buffer
	require.NoError(t, png.Encode(&content, image.NewRGBA(image.Rect(0, 0, 32, 32))))
	return content.Bytes()
}
func studioUpload(t *testing.T, router *gin.Engine, owner int64, content []byte, overrides map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	fields := map[string]string{"title": "测试作品", "author": "作者昵称", "category": "奇幻风景", "style": "水彩", "media": "image", "prompt_kind": "actual", "prompt": "一片有层次的暖色幻想山谷与湖泊", "model": "投稿人自己的模型", "rights_confirmed": "true", "publish_consent": "true"}
	for key, value := range overrides {
		fields[key] = value
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		require.NoError(t, writer.WriteField(key, value))
	}
	file, err := writer.CreateFormFile("file", "uploaded.png")
	require.NoError(t, err)
	_, err = file.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	request := httptest.NewRequest(http.MethodPost, "/submit", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Test-Owner", strconv.FormatInt(owner, 10))
	request.Header.Set("Test-Role", "user")
	output := httptest.NewRecorder()
	router.ServeHTTP(output, request)
	return output
}
func studioEntry(t *testing.T, output *httptest.ResponseRecorder) StudioEntry {
	t.Helper()
	var envelope struct {
		Data StudioEntry `json:"data"`
	}
	require.NoError(t, json.Unmarshal(output.Body.Bytes(), &envelope))
	require.NotEmpty(t, envelope.Data.ID)
	return envelope.Data
}
func TestStudioPrivatePublishUnpublishAndWithdraw(t *testing.T) {
	handler := &StudioHandler{root: filepath.Join(t.TempDir(), "submissions")}
	router := studioTestRouter(handler)
	uploaded := studioUpload(t, router, 1, studioPNG(t), nil)
	require.Equal(t, http.StatusCreated, uploaded.Code, uploaded.Body.String())
	entry := studioEntry(t, uploaded)
	require.Equal(t, "pending", entry.Status)
	require.Equal(t, http.StatusNotFound, studioCall(router, "GET", "/public/"+entry.ID, "", 0, "").Code)
	require.Equal(t, http.StatusNotFound, studioCall(router, "GET", "/private/"+entry.ID, "", 2, "user").Code)
	require.Equal(t, http.StatusOK, studioCall(router, "GET", "/private/"+entry.ID, "", 1, "user").Code)
	require.Contains(t, studioCall(router, "GET", "/gallery", "", 0, "").Body.String(), `"data":[]`)
	approve := `{"action":"approve","expected_status":"pending","review_confirmed":true}`
	require.Equal(t, http.StatusForbidden, studioCall(router, "POST", "/review/"+entry.ID, approve, 1, "user").Code)
	require.Equal(t, http.StatusBadRequest, studioCall(router, "POST", "/review/"+entry.ID, `{"action":"approve","expected_status":"pending"}`, 9, "admin").Code)
	require.Equal(t, http.StatusOK, studioCall(router, "POST", "/review/"+entry.ID, approve, 9, "admin").Code)
	require.Equal(t, http.StatusConflict, studioCall(router, "POST", "/review/"+entry.ID, approve, 9, "admin").Code)
	gallery := studioCall(router, "GET", "/gallery", "", 0, "")
	require.Contains(t, gallery.Body.String(), entry.ID)
	for _, field := range []string{"owner_id", "reviewer_id", "reason", "mime"} {
		require.NotContains(t, gallery.Body.String(), field)
	}
	request := httptest.NewRequest("GET", "/public/"+entry.ID, nil)
	request.Header.Set("Range", "bytes=0-9")
	output := httptest.NewRecorder()
	router.ServeHTTP(output, request)
	require.Equal(t, http.StatusPartialContent, output.Code)
	require.Len(t, output.Body.Bytes(), 10)
	require.Equal(t, "no-store", output.Header().Get("Cache-Control"))
	require.Equal(t, "nosniff", output.Header().Get("X-Content-Type-Options"))
	require.Equal(t, http.StatusBadRequest, studioCall(router, "POST", "/review/"+entry.ID, `{"action":"unpublish","expected_status":"published","reason":"短"}`, 9, "admin").Code)
	require.Equal(t, http.StatusOK, studioCall(router, "POST", "/review/"+entry.ID, `{"action":"unpublish","expected_status":"published","reason":"授权需要补充"}`, 9, "admin").Code)
	require.Equal(t, http.StatusNotFound, studioCall(router, "GET", "/public/"+entry.ID, "", 0, "").Code)
	require.Equal(t, http.StatusNotFound, studioCall(router, "POST", "/withdraw/"+entry.ID, "", 2, "user").Code)
	require.Equal(t, http.StatusOK, studioCall(router, "POST", "/withdraw/"+entry.ID, "", 1, "user").Code)
	require.Equal(t, http.StatusBadRequest, studioCall(router, "POST", "/review/"+entry.ID, `{"action":"approve","expected_status":"withdrawn","review_confirmed":true}`, 9, "admin").Code)
	require.Equal(t, http.StatusUnauthorized, studioCall(router, "GET", "/mine", "", 0, "").Code)
	require.Equal(t, http.StatusForbidden, studioCall(router, "GET", "/review", "", 1, "user").Code)
	require.Contains(t, studioCall(router, "GET", "/mine", "", 1, "user").Body.String(), "withdrawn")
	require.Contains(t, studioCall(router, "GET", "/mine", "", 2, "user").Body.String(), `"data":[]`)
}
func TestStudioUploadValidationAndQuota(t *testing.T) {
	handler := &StudioHandler{root: t.TempDir()}
	router := studioTestRouter(handler)
	require.Equal(t, http.StatusUnauthorized, studioUpload(t, router, 0, studioPNG(t), nil).Code)
	for name, fields := range map[string]map[string]string{
		"rights": {"rights_confirmed": "false"}, "consent": {"publish_consent": "false"},
		"prompt": {"prompt": "短"}, "category": {"category": "未知"}, "style": {"style": "未知"}, "source": {"prompt_kind": "pretend"},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, http.StatusBadRequest, studioUpload(t, router, 1, studioPNG(t), fields).Code)
		})
	}
	for _, content := range [][]byte{[]byte("<svg xmlns='http://www.w3.org/2000/svg'><script>alert(1)</script></svg>"), []byte("<html>not image</html>"), {1, 2, 3}} {
		require.Equal(t, http.StatusBadRequest, studioUpload(t, router, 1, content, nil).Code)
	}
	require.Equal(t, http.StatusRequestEntityTooLarge, studioUpload(t, router, 1, make([]byte, studioFileLimit+1), nil).Code)
	for index := 0; index < 5; index++ {
		require.Equal(t, http.StatusCreated, studioUpload(t, router, 1, studioPNG(t), nil).Code)
	}
	require.Equal(t, http.StatusTooManyRequests, studioUpload(t, router, 1, studioPNG(t), nil).Code)
	require.Equal(t, http.StatusCreated, studioUpload(t, router, 2, studioPNG(t), nil).Code)
	require.Equal(t, http.StatusNotFound, studioCall(router, "GET", "/public/..", "", 0, "").Code)
}
func TestStudioRestartAndCorruptRecordFailClosed(t *testing.T) {
	root := t.TempDir()
	handler := &StudioHandler{root: root}
	router := studioTestRouter(handler)
	entry := studioEntry(t, studioUpload(t, router, 1, studioPNG(t), nil))
	restarted := studioTestRouter(&StudioHandler{root: root})
	require.Contains(t, studioCall(restarted, "GET", "/mine", "", 1, "user").Body.String(), entry.ID)
	require.Equal(t, http.StatusNotFound, studioCall(restarted, "GET", "/public/"+entry.ID, "", 0, "").Code)
	require.Equal(t, http.StatusOK, studioCall(restarted, "POST", "/review/"+entry.ID, `{"action":"reject","expected_status":"pending","reason":"请补充素材授权"}`, 9, "admin").Code)
	require.Equal(t, http.StatusNotFound, studioCall(restarted, "GET", "/public/"+entry.ID, "", 0, "").Code)
	require.NoError(t, os.WriteFile(filepath.Join(root, entry.ID+".json"), []byte("{broken"), 0600))
	require.Equal(t, http.StatusInternalServerError, studioCall(restarted, "GET", "/gallery", "", 0, "").Code)
	require.Equal(t, http.StatusNotFound, studioCall(restarted, "GET", "/public/"+entry.ID, "", 0, "").Code)
}
func TestStudioMP4SignatureRangeAndPublishedWithdrawal(t *testing.T) {
	handler := &StudioHandler{root: t.TempDir()}
	router := studioTestRouter(handler)
	signature := []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'm', 'p', '4', '2', 0, 0, 0, 0, 'm', 'p', '4', '2', 'i', 's', 'o', 'm'}
	upload := studioUpload(t, router, 1, signature, map[string]string{"media": "video", "prompt_kind": "reference"})
	require.Equal(t, http.StatusCreated, upload.Code, upload.Body.String())
	entry := studioEntry(t, upload)
	require.Equal(t, "video/mp4", entry.MIME)
	require.Equal(t, http.StatusNotFound, studioCall(router, "GET", "/public/"+entry.ID, "", 0, "").Code)
	require.Equal(t, http.StatusOK, studioCall(router, "GET", "/private/"+entry.ID, "", 9, "admin").Code)
	require.Equal(t, http.StatusOK, studioCall(router, "POST", "/review/"+entry.ID, `{"action":"approve","expected_status":"pending","review_confirmed":true}`, 9, "admin").Code)
	request := httptest.NewRequest("GET", "/public/"+entry.ID, nil)
	request.Header.Set("Range", "bytes=4-11")
	output := httptest.NewRecorder()
	router.ServeHTTP(output, request)
	require.Equal(t, http.StatusPartialContent, output.Code)
	require.Equal(t, "video/mp4", output.Header().Get("Content-Type"))
	require.Equal(t, "ftypmp42", output.Body.String())
	require.Equal(t, http.StatusOK, studioCall(router, "POST", "/withdraw/"+entry.ID, "", 1, "user").Code)
	require.Equal(t, http.StatusNotFound, studioCall(router, "GET", "/public/"+entry.ID, "", 0, "").Code)
	require.Contains(t, studioCall(router, "GET", "/gallery", "", 0, "").Body.String(), `"data":[]`)
}
