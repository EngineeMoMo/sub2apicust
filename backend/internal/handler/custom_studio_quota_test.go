package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// 回归 2026-10-11 审查：单个用户反复上传再撤回曾占满全站 512MB，其他用户无法投稿。
func TestStudioWithdrawFreesStorage(t *testing.T) {
	handler := &StudioHandler{root: t.TempDir()}
	router := studioTestRouter(handler)
	large := make([]byte, studioImageLimit) // 合法 PNG 文件头，补足到图片上限
	copy(large, studioPNG(t))
	for range 3 {
		entry := studioEntry(t, studioUpload(t, router, 1, large, nil))
		require.Equal(t, http.StatusOK, studioCall(router, "POST", "/withdraw/"+entry.ID, "", 1, "user").Code)
	}
	files, err := os.ReadDir(handler.root)
	require.NoError(t, err)
	require.Empty(t, files, "撤回后不残留记录、素材或临时文件")
	require.Equal(t, http.StatusCreated, studioUpload(t, router, 2, studioPNG(t), nil).Code)
}

func TestStudioQuotaCountsOnlyKeptMedia(t *testing.T) {
	entries := make([]StudioEntry, 0, 50)
	for range 43 {
		entries = append(entries, StudioEntry{OwnerID: 1, Status: "rejected", Size: studioImageLimit})
	}
	require.Empty(t, studioQuota(entries, 2, studioImageLimit), "退回已删除素材，不占存储")
	require.Contains(t, studioQuota(entries, 1, 1), "退回记录", "退回记录仍计入本人20份上限")
	published := make([]StudioEntry, 0, 43)
	for range 43 {
		published = append(published, StudioEntry{OwnerID: 3, Status: "published", Size: studioImageLimit})
	}
	require.Contains(t, studioQuota(published, 2, studioImageLimit), "全站投稿存储已满")
	pending := []StudioEntry{{OwnerID: 1, Status: "pending"}, {OwnerID: 1, Status: "pending"}, {OwnerID: 1, Status: "pending"}, {OwnerID: 1, Status: "pending"}, {OwnerID: 1, Status: "pending"}}
	require.Contains(t, studioQuota(pending, 1, 1), "5份投稿在等待审核")
	require.Empty(t, studioQuota(pending, 2, 1))
}

func TestStudioRejectRemovesMediaAndAdminDelete(t *testing.T) {
	handler := &StudioHandler{root: t.TempDir()}
	router := studioTestRouter(handler)
	rejected := studioEntry(t, studioUpload(t, router, 1, studioPNG(t), nil))
	require.Equal(t, http.StatusOK, studioCall(router, "POST", "/review/"+rejected.ID, `{"action":"reject","expected_status":"pending","reason":"请补充素材授权"}`, 9, "admin").Code)
	_, err := os.Stat(filepath.Join(handler.root, rejected.ID+".media"))
	require.ErrorIs(t, err, os.ErrNotExist, "退回删除素材")
	mine := studioCall(router, "GET", "/mine", "", 1, "user").Body.String()
	require.Contains(t, mine, "请补充素材授权", "退回原因仍保留给作者")
	require.Equal(t, http.StatusNotFound, studioCall(router, "GET", "/private/"+rejected.ID, "", 1, "user").Code)

	kept := studioEntry(t, studioUpload(t, router, 1, studioPNG(t), nil))
	require.Equal(t, http.StatusForbidden, studioCall(router, "DELETE", "/admin/"+kept.ID, "", 1, "user").Code)
	require.Equal(t, http.StatusNotFound, studioCall(router, "DELETE", "/admin/"+"0123456789abcdef0123456789abcdef", "", 9, "admin").Code)
	require.Equal(t, http.StatusOK, studioCall(router, "DELETE", "/admin/"+kept.ID, "", 9, "admin").Code)
	require.Equal(t, http.StatusOK, studioCall(router, "DELETE", "/admin/"+rejected.ID, "", 9, "admin").Code)
	files, err := os.ReadDir(handler.root)
	require.NoError(t, err)
	require.Empty(t, files)
}

func TestStudioUploadConcurrencySlots(t *testing.T) {
	handler := &StudioHandler{root: t.TempDir()}
	release, ok := handler.acquireUpload(1)
	require.True(t, ok)
	_, ok = handler.acquireUpload(1)
	require.False(t, ok, "同一用户同时只处理一份上传")
	releases := []func(){release}
	for user := int64(2); user <= studioUploadConcurrent; user++ {
		next, ok := handler.acquireUpload(user)
		require.True(t, ok)
		releases = append(releases, next)
	}
	_, ok = handler.acquireUpload(99)
	require.False(t, ok, "全站并发上传有上限")
	for _, next := range releases {
		next()
	}
	again, ok := handler.acquireUpload(1)
	require.True(t, ok)
	again()
}
