package handler

import (
	"encoding/binary"
	"hash/crc32"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStudioSaveRenameFailureCleansTemporaryRecord(t *testing.T) {
	root := t.TempDir()
	id := strings.Repeat("a", 32)
	target := filepath.Join(root, id+".json")
	require.NoError(t, os.Mkdir(target, 0700))
	handler := &StudioHandler{root: root}
	err := handler.save(StudioEntry{StudioPublicEntry: StudioPublicEntry{ID: id, Title: "保留原始保存错误"}})
	require.Error(t, err)
	files, readErr := os.ReadDir(root)
	require.NoError(t, readErr)
	require.Len(t, files, 1, "重命名失败后不得残留临时记录")
	require.Equal(t, id+".json", files[0].Name())
	require.True(t, files[0].IsDir(), "失败不得替换原目标")
}

func TestStudioImageDecoderBoundaries(t *testing.T) {
	webp, err := os.ReadFile(filepath.Join("testdata", "studio-valid.webp"))
	require.NoError(t, err)
	truncatedWebP := []byte{'R', 'I', 'F', 'F', 18, 0, 0, 0, 'W', 'E', 'B', 'P', 'V', 'P', '8', 'L', 5, 0, 0, 0, 47, 0}
	require.Equal(t, "image/webp", http.DetectContentType(truncatedWebP))
	// 保留合法PNG的IHDR结构与CRC，只改尺寸，避免为超限案例分配巨幅位图。
	withDimensions := func(width, height uint32) []byte {
		content := studioPNG(t)
		binary.BigEndian.PutUint32(content[16:20], width)
		binary.BigEndian.PutUint32(content[20:24], height)
		binary.BigEndian.PutUint32(content[29:33], crc32.ChecksumIEEE(content[12:29]))
		return content
	}
	for _, sample := range []struct {
		name    string
		content []byte
		status  int
	}{
		{"valid webp", webp, http.StatusCreated},
		{"truncated webp", truncatedWebP, http.StatusBadRequest},
		{"below minimum", withDimensions(31, 32), http.StatusBadRequest},
		{"over pixel limit", withDimensions(4801, 5000), http.StatusBadRequest},
	} {
		t.Run(sample.name, func(t *testing.T) {
			handler := &StudioHandler{root: t.TempDir()}
			response := studioUpload(t, studioTestRouter(handler), 1, sample.content, nil)
			require.Equal(t, sample.status, response.Code, response.Body.String())
			if sample.status == http.StatusBadRequest {
				entries, listErr := handler.entries()
				require.NoError(t, listErr)
				require.Empty(t, entries, "拒绝的图片不得留下投稿记录")
				files, readErr := os.ReadDir(handler.root)
				require.NoError(t, readErr)
				require.Empty(t, files, "拒绝的图片不得留下媒体或临时文件")
			}
		})
	}
}
