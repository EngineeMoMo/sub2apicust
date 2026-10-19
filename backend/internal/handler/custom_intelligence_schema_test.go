//go:build unit

package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func intelligenceV6Summary(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("testdata/intelligence_summary_v6.json")
	require.NoError(t, err)
	var summary map[string]any
	require.NoError(t, json.Unmarshal(raw, &summary))
	return summary
}

func TestIntelligenceRealV6SummaryMatchesSourceDisplay(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low"))
	source.override("low", intelligenceV6Summary(t))
	status, body := intelligenceTestGet(t, source.cache(newIntelligenceTestClock()))
	require.Equal(t, http.StatusOK, status)
	require.False(t, body.Data.RefreshFailed)
	require.Len(t, body.Data.Points, 1)
	point := body.Data.Points[0]
	require.Equal(t, "ok", point.State)
	require.Equal(t, "complete", point.ScoreStatus)
	require.Equal(t, 77.45784471832991, *point.Score)
	require.Equal(t, 64, *point.Coverage)
	require.Equal(t, 106, *point.GradedRuns)
	require.Equal(t, 127, *point.StatisticsCount)
	require.Equal(t, 115, *point.PriceSamples)
	require.Equal(t, 121, *point.DurationSamples)
	require.InDelta(t, 34.268215/115, *point.AveragePriceUSD, 1e-12)
	require.InDelta(t, 50372863.51603189/121/60000, *point.AverageMinutes, 1e-12)
}

func TestIntelligenceV6KeepsContractValidation(t *testing.T) {
	binding := intelligenceTestBinding("low")
	selection := binding.ModelEfforts[0]
	for name, overrides := range map[string]map[string]any{
		"未知汇总版本": {"schema": "radar-bench-public-summary-v7-150-details"},
		"计分规则变化": {"scoring_policy_version": "radar-bench-v3-best-of-5"},
		"计分版本变化": {"score_version": "radar-bench-v3-100"},
		"题库变化":   {"benchmark": "different-benchmark"},
		"满分变化":   {"score_max": 100},
		"判分题次越界": {"graded_runs_count": 193},
		"覆盖题数越界": {"coverage": 65},
	} {
		t.Run(name, func(t *testing.T) {
			summary := intelligenceV6Summary(t)
			for key, value := range overrides {
				summary[key] = value
			}
			_, err := intelligenceSummaryPoint(intelligenceDecodeSummary(t, summary), binding, selection)
			require.ErrorIs(t, err, errIntelligenceInvalid)
		})
	}
}
