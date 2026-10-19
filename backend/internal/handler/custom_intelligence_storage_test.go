//go:build unit

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIntelligenceSnapshotSurvivesRestartAndSourceFailure(t *testing.T) {
	for _, failure := range []string{"未知汇总版本", "源站繁忙"} {
		t.Run(failure, func(t *testing.T) {
			source := newIntelligenceTestSource(t, intelligenceTestBinding("low"))
			clock := newIntelligenceTestClock()
			dataDir := t.TempDir()
			cache := source.cache(clock)
			cache.configureStorage(dataDir)
			status, first := intelligenceTestGet(t, cache)
			require.Equal(t, http.StatusOK, status)
			original, err := os.ReadFile(cache.storagePath)
			require.NoError(t, err)
			if failure == "未知汇总版本" {
				source.override("low", map[string]any{"schema": "radar-bench-public-summary-v99"})
			} else {
				source.setStatus("binding", http.StatusServiceUnavailable, "120")
			}
			restarted := source.cache(clock)
			restarted.configureStorage(dataDir)
			status, fallback := intelligenceTestGet(t, restarted)
			require.Equal(t, http.StatusOK, status)
			require.True(t, fallback.Data.RefreshFailed)
			require.Equal(t, first.Data.CheckedAt, fallback.Data.CheckedAt)
			require.Equal(t, "retained", fallback.Data.Points[0].State)
			require.Equal(t, first.Data.Points[0].ReadAt, fallback.Data.Points[0].ReadAt)
			require.Equal(t, *first.Data.Points[0].Score, *fallback.Data.Points[0].Score)
			onFailure, err := os.ReadFile(restarted.storagePath)
			require.NoError(t, err)
			require.Equal(t, original, onFailure)

			source.setStatus("binding", 0, "")
			source.override("low", map[string]any{"schema": intelligenceDetailsSchema, "score": 99.0})
			clock.Advance(intelligenceFreshFor)
			status, recovered := intelligenceTestGet(t, restarted)
			require.Equal(t, http.StatusOK, status)
			require.False(t, recovered.Data.RefreshFailed)
			require.Equal(t, "ok", recovered.Data.Points[0].State)
			require.Equal(t, 99.0, *recovered.Data.Points[0].Score)
			files, err := os.ReadDir(filepath.Dir(restarted.storagePath))
			require.NoError(t, err)
			require.Len(t, files, 1)
			reloaded := source.cache(clock)
			reloaded.configureStorage(dataDir)
			require.Equal(t, 99.0, *reloaded.snapshot.Points[0].Score)
		})
	}
}

func TestIntelligenceStoredPartialSnapshotKeepsMissingDistinct(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low", "medium"))
	source.setStatus("medium", http.StatusInternalServerError, "")
	clock := newIntelligenceTestClock()
	cache := source.cache(clock)
	dataDir := t.TempDir()
	cache.configureStorage(dataDir)
	status, _ := intelligenceTestGet(t, cache)
	require.Equal(t, http.StatusOK, status)
	restored := source.cache(clock)
	restored.configureStorage(dataDir)
	require.NotNil(t, restored.snapshot)
	require.Equal(t, "retained", restored.snapshot.Points[0].State)
	require.Equal(t, "unavailable", restored.snapshot.Points[1].State)
	require.Nil(t, restored.snapshot.Points[1].Score)
	require.Nil(t, restored.snapshot.Points[1].AveragePriceUSD)
}

func TestIntelligenceInvalidStoredSnapshotDoesNotBlockFreshRead(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low"))
	clock := newIntelligenceTestClock()
	seed := source.cache(clock)
	snapshot, _ := seed.current(t.Context())
	require.NotNil(t, snapshot)
	valid, err := json.Marshal(intelligenceStoredSnapshot{Version: 1, ScoreVersion: intelligenceScoreVersion,
		PolicyVersion: intelligencePolicyVersion, CostBasis: intelligenceCostBasis, TaskSet: snapshot.taskSet, Snapshot: *snapshot})
	require.NoError(t, err)
	for name, change := range map[string]func(*intelligenceStoredSnapshot){
		"计分版本变化": func(stored *intelligenceStoredSnapshot) { stored.ScoreVersion = "unsupported" },
		"计分规则变化": func(stored *intelligenceStoredSnapshot) { stored.PolicyVersion = "unsupported" },
		"费用口径变化": func(stored *intelligenceStoredSnapshot) { stored.CostBasis = "unsupported" },
		"均值缺失":   func(stored *intelligenceStoredSnapshot) { stored.Snapshot.Points[0].AveragePriceUSD = nil },
		"文件版本变化": func(stored *intelligenceStoredSnapshot) { stored.Version = 2 },
		"题库摘要无效": func(stored *intelligenceStoredSnapshot) { stored.TaskSet = "invalid" },
		"重复档位": func(stored *intelligenceStoredSnapshot) {
			stored.Snapshot.Points = append(stored.Snapshot.Points, stored.Snapshot.Points[0])
		},
		"越界分数": func(stored *intelligenceStoredSnapshot) { *stored.Snapshot.Points[0].Score = 151 },
		"未知状态": func(stored *intelligenceStoredSnapshot) { stored.Snapshot.Points[0].State = "unexpected" },
		"未来时间": func(stored *intelligenceStoredSnapshot) {
			stored.Snapshot.CheckedAt = clock.Now().Add(time.Hour).Format(time.RFC3339)
		},
		"非法读取时间": func(stored *intelligenceStoredSnapshot) { stored.Snapshot.Points[0].ReadAt = "invalid" },
		"样本越界":   func(stored *intelligenceStoredSnapshot) { *stored.Snapshot.Points[0].PriceSamples = 999 },
		"负费用":    func(stored *intelligenceStoredSnapshot) { *stored.Snapshot.Points[0].AveragePriceUSD = -1 },
		"无可用数据":  func(stored *intelligenceStoredSnapshot) { stored.Snapshot.Points = nil },
		"JSON损坏": nil,
		"文件超限":   nil,
	} {
		t.Run(name, func(t *testing.T) {
			var stored intelligenceStoredSnapshot
			require.NoError(t, json.Unmarshal(valid, &stored))
			raw := []byte("invalid-json")
			if change != nil {
				change(&stored)
				raw, err = json.Marshal(stored)
				require.NoError(t, err)
			} else if name == "文件超限" {
				raw = []byte(strings.Repeat(" ", intelligenceSnapshotLimit+1))
			}
			cache := source.cache(clock)
			cache.storagePath = filepath.Join(t.TempDir(), "last-success.json")
			require.NoError(t, os.WriteFile(cache.storagePath, raw, 0600))
			require.Error(t, cache.loadSnapshot())
			require.Nil(t, cache.snapshot)
			status, body := intelligenceTestGet(t, cache)
			require.Equal(t, http.StatusOK, status)
			require.False(t, body.Data.RefreshFailed)
		})
	}
}

func TestIntelligenceSnapshotWriteFailureKeepsEndpointAvailable(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low"))
	cache := source.cache(newIntelligenceTestClock())
	blocker := filepath.Join(t.TempDir(), "not-a-directory")
	require.NoError(t, os.WriteFile(blocker, []byte("preserve"), 0600))
	cache.storagePath = filepath.Join(blocker, "last-success.json")
	status, body := intelligenceTestGet(t, cache)
	require.Equal(t, http.StatusOK, status)
	require.False(t, body.Data.RefreshFailed)
	raw, err := os.ReadFile(blocker)
	require.NoError(t, err)
	require.Equal(t, "preserve", string(raw))
}

func TestIntelligenceSelectionCapacityMatchesFrontend(t *testing.T) {
	binding := intelligenceTestBinding()
	for index := range intelligenceMaxSelections {
		binding.ModelEfforts = append(binding.ModelEfforts, intelligenceSelection{Model: "model", Effort: fmt.Sprintf("effort-%d", index)})
	}
	require.NoError(t, validateIntelligenceBinding(binding))
	binding.ModelEfforts = append(binding.ModelEfforts, intelligenceSelection{Model: "model", Effort: "over-limit"})
	require.ErrorIs(t, validateIntelligenceBinding(binding), errIntelligenceInvalid)
}

func TestIntelligenceStoredSnapshotPreservesZeroAndMissingStatistics(t *testing.T) {
	for name, overrides := range map[string]map[string]any{
		"零分":     {"score": 0},
		"无判分":    {"score": nil, "coverage": 0, "score_status": "missing_current_result", "graded_runs_count": 0, "statistics_state": "not_provided", "statistics": nil},
		"未提供统计":  {"statistics_state": "not_provided", "statistics": nil},
		"统计状态变化": {"statistics_state": "unknown_statistics"},
	} {
		t.Run(name, func(t *testing.T) {
			source := newIntelligenceTestSource(t, intelligenceTestBinding("low"))
			source.override("low", overrides)
			clock := newIntelligenceTestClock()
			cache := source.cache(clock)
			cache.configureStorage(t.TempDir())
			status, body := intelligenceTestGet(t, cache)
			require.Equal(t, http.StatusOK, status)
			restored := source.cache(clock)
			restored.storagePath = cache.storagePath
			require.NoError(t, restored.loadSnapshot())
			expected := body.Data.Points[0]
			expected.State = "retained"
			require.Equal(t, expected, restored.snapshot.Points[0])
		})
	}
}
