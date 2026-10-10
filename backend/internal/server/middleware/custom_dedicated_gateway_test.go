//go:build unit

package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type customDedicatedTrackedBody struct {
	io.ReadCloser
	reads int
}

func (body *customDedicatedTrackedBody) Read(buffer []byte) (int, error) {
	body.reads++
	return body.ReadCloser.Read(buffer)
}

func TestCustomDedicatedGatewayOwnsRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, sample := range []struct {
		name, path, body string
		limit            int64
		status           int
	}{
		{"large_body", "/v1/responses", `{"input":"` + strings.Repeat("a", (2<<20)+1) + `"}`, 4 << 20, http.StatusOK},
		{"gateway_limit", "/v1/responses", strings.Repeat("a", 1025), 1024, http.StatusRequestEntityTooLarge},
		{"images", "/v1/responses", `{"tools":[{"type":"image_generation"}]}`, 4096, http.StatusOK},
		{"async_images", "/v1/images/generations/async", `{}`, 4096, http.StatusOK},
		{"multipart", "/v1/images/edits", "--multipart-body", 4096, http.StatusOK},
		{"new_endpoint", "/v1/new-endpoint", "not-json", 4096, http.StatusOK},
		{"deep_json", "/v1/responses", strings.Repeat("[", 70) + "0" + strings.Repeat("]", 70), 4096, http.StatusOK},
	} {
		t.Run(sample.name, func(t *testing.T) {
			database, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { _ = database.Close() })
			keys := &service.APIKeyService{}
			service.NewCustomDedicatedService(database, nil, keys, &service.GatewayService{}, &service.OpenAIGatewayService{})
			now := time.Now()
			bindingRows := func() *sqlmock.Rows {
				return sqlmock.NewRows([]string{"id", "user_id", "account_id", "group_id", "label", "expires_at", "revoked_at", "updated_at", "user_ids"}).AddRow(1, 11, 22, 33, "包号", now.Add(time.Hour), nil, now, []byte(`[11]`))
			}
			mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts WHERE deleted_at IS NULL").WithArgs(int64(33)).WillReturnRows(bindingRows())
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"valid"}).AddRow(true))
			group := int64(33)
			key, err := keys.PrepareCustomDedicatedBilling(context.Background(), &service.APIKey{ID: 7, UserID: 11, GroupID: &group})
			require.NoError(t, err)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .* FROM custom_dedicated_accounts WHERE id=.*FOR UPDATE").WithArgs(int64(1)).WillReturnRows(bindingRows())
			mock.ExpectQuery("SELECT clock_timestamp").WillReturnRows(sqlmock.NewRows([]string{"accepted_at"}).AddRow(now))
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"valid"}).AddRow(true))
			mock.ExpectExec("INSERT INTO custom_dedicated_request_leases.*finished_at").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			body := &customDedicatedTrackedBody{ReadCloser: io.NopCloser(strings.NewReader(sample.body))}
			router := gin.New()
			router.Use(RequestBodyLimit(sample.limit))
			router.Use(func(ctx *gin.Context) { customDedicatedBillingAdmission(ctx, keys, key) })
			router.POST(sample.path, func(ctx *gin.Context) {
				require.Zero(t, body.reads)
				payload, readErr := io.ReadAll(ctx.Request.Body)
				if readErr != nil {
					ctx.Status(http.StatusRequestEntityTooLarge)
					return
				}
				ctx.Data(http.StatusOK, "application/octet-stream", payload)
			})
			request := httptest.NewRequest(http.MethodPost, sample.path, nil)
			request.Body = body
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			require.Equal(t, sample.status, recorder.Code)
			if sample.status == http.StatusOK {
				require.Equal(t, sample.body, recorder.Body.String())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
