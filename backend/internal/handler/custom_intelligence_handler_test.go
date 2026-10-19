//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// 2026-10-12 从 codex-reset-radar.pages.dev 实取的 gpt-6.1-sol/low 公开汇总原文（含源站返回的 history）。
const intelligenceRealSummary = `{"schema":"radar-bench-public-summary-v5-150-daily-graded-count","benchmark":"radar-bench-final64-20261004","score_version":"radar-bench-v2-150","score_max":150,"catalog_version":"dradar-four-library-deployment007-server024-20261004-config64","task_set_sha256":"f75ff09b6117cd83d3ab6ea744fe379ac117676c042538d5df5e38489af27bc4","model":"gpt-6.1-sol","effort":"low","score":81.75471971832992,"coverage":64,"required_tasks":64,"score_status":"complete","graded_runs_count":83,"source_counts":{"deepswe":15,"pompeii":16,"tb4":14,"science":19},"scoring_policy_version":"radar-bench-v2-last3-per-task-equal-weight-150","statistics":{"sample_count":83,"cost_basis":"official_standard_api_equivalent; no estimates or claimed cash charges","metrics":{"elapsed_ms":{"total":null,"mean":null,"known_total":40180886,"known_samples":82,"missing_samples":1,"missing_reason":"source_evidence_incomplete"},"api_equivalent_usd":{"total":null,"mean":null,"known_total":23.478868,"known_samples":71,"missing_samples":12,"missing_reason":"source_evidence_incomplete"}}},"statistics_state":"available","history":{"schema":"radar-bench-public-history-v1-150-daily","score_version":"radar-bench-v2-150","score_max":150,"model":"gpt-6.1-sol","effort":"low","time_zone":"UTC","bucket":"day","basis":"last canonical score per UTC grading day; days without grades omitted","updated_bucket":"2026-10-11T00:00:00.000Z","points":[{"at":"2026-10-06T00:00:00.000Z","score":112.5,"coverage":12,"required_tasks":64,"score_status":"provisional"},{"at":"2026-10-07T00:00:00.000Z","score":110,"coverage":15,"required_tasks":64,"score_status":"provisional"},{"at":"2026-10-08T00:00:00.000Z","score":84.24288987258913,"coverage":63,"required_tasks":64,"score_status":"provisional"},{"at":"2026-10-09T00:00:00.000Z","score":82.92659471832992,"coverage":64,"required_tasks":64,"score_status":"complete"},{"at":"2026-10-10T00:00:00.000Z","score":81.75471971832992,"coverage":64,"required_tasks":64,"score_status":"complete"},{"at":"2026-10-11T00:00:00.000Z","score":81.75471971832992,"coverage":64,"required_tasks":64,"score_status":"complete"}]},"history_state":"available"}`

type intelligenceTestClock struct {
	mu sync.Mutex
	at time.Time
}

func newIntelligenceTestClock() *intelligenceTestClock {
	return &intelligenceTestClock{at: time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC)}
}

func (clock *intelligenceTestClock) Now() time.Time {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	return clock.at
}

func (clock *intelligenceTestClock) Advance(duration time.Duration) {
	clock.mu.Lock()
	clock.at = clock.at.Add(duration)
	clock.mu.Unlock()
}

// 与实取汇总一致的绑定；按需只保留部分档位。
func intelligenceTestBinding(efforts ...string) intelligenceBinding {
	binding := intelligenceBinding{Enabled: true, ProductionReadVerified: true, Benchmark: "radar-bench-final64-20261004",
		CatalogVersion: "dradar-four-library-deployment007-server024-20261004-config64",
		TaskSetSHA256:  "f75ff09b6117cd83d3ab6ea744fe379ac117676c042538d5df5e38489af27bc4",
		TotalTasks:     64, ScoreVersion: intelligenceScoreVersion, ScoreMax: intelligenceScoreMax}
	for _, effort := range efforts {
		binding.ModelEfforts = append(binding.ModelEfforts, intelligenceSelection{Model: "gpt-6.1-sol", Effort: effort})
	}
	return binding
}

// 以实取原文为底，按档位改写 model／effort，并跟随绑定的题库标识。
func intelligenceTestSummary(t *testing.T, binding intelligenceBinding, selection intelligenceSelection, overrides map[string]any) map[string]any {
	t.Helper()
	summary := map[string]any{}
	require.NoError(t, json.Unmarshal([]byte(intelligenceRealSummary), &summary))
	summary["model"], summary["effort"] = selection.Model, selection.Effort
	summary["benchmark"], summary["catalog_version"], summary["task_set_sha256"] = binding.Benchmark, binding.CatalogVersion, binding.TaskSetSHA256
	for key, value := range overrides {
		summary[key] = value
	}
	return summary
}

type intelligenceTestSource struct {
	t      *testing.T
	server *httptest.Server

	mu         sync.Mutex
	binding    intelligenceBinding
	summaries  map[intelligenceSelection]map[string]any
	status     map[string]int // key：binding 或 effort
	retryAfter string
	gate       map[string]chan struct{}
	queries    []string

	bindingHits atomic.Int32
	summaryHits atomic.Int32
}

func newIntelligenceTestSource(t *testing.T, binding intelligenceBinding) *intelligenceTestSource {
	source := &intelligenceTestSource{t: t, status: map[string]int{}, gate: map[string]chan struct{}{}}
	source.setBinding(binding)
	source.server = httptest.NewServer(http.HandlerFunc(source.serve))
	t.Cleanup(source.server.Close)
	return source
}

func (source *intelligenceTestSource) setBinding(binding intelligenceBinding) {
	source.mu.Lock()
	defer source.mu.Unlock()
	source.binding = binding
	source.summaries = map[intelligenceSelection]map[string]any{}
	for _, selection := range binding.ModelEfforts {
		source.summaries[selection] = intelligenceTestSummary(source.t, binding, selection, nil)
	}
}

func (source *intelligenceTestSource) override(effort string, overrides map[string]any) {
	source.mu.Lock()
	defer source.mu.Unlock()
	selection := intelligenceSelection{Model: "gpt-6.1-sol", Effort: effort}
	source.summaries[selection] = intelligenceTestSummary(source.t, source.binding, selection, overrides)
}

func (source *intelligenceTestSource) setStatus(key string, status int, retryAfter string) {
	source.mu.Lock()
	defer source.mu.Unlock()
	source.status[key], source.retryAfter = status, retryAfter
}

// block 让该档位请求挂起，返回放行函数；测试失败提前结束时也会在关闭服务前放行，避免卡住。
func (source *intelligenceTestSource) block(effort string) func() {
	source.mu.Lock()
	defer source.mu.Unlock()
	gate := make(chan struct{})
	source.gate[effort] = gate
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	source.t.Cleanup(release)
	return release
}

func (source *intelligenceTestSource) serve(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Cookie") != "" || request.Header.Get("Authorization") != "" || request.Header.Get("User-Agent") != intelligenceUserAgent {
		source.t.Errorf("unexpected request headers: %v", request.Header)
	}
	source.mu.Lock()
	var payload any
	key := "binding"
	switch {
	case request.URL.Path == intelligenceBindingPath:
		source.bindingHits.Add(1)
		if request.Header.Get("X-Radar-Bench-Selection") != "" {
			source.t.Errorf("binding request carries selection mark")
		}
		payload = source.binding
	case request.URL.Path == intelligenceSummaryPath:
		source.summaryHits.Add(1)
		source.queries = append(source.queries, request.URL.RawQuery)
		key = request.URL.Query().Get("effort")
		payload = source.summaries[intelligenceSelection{Model: request.URL.Query().Get("model"), Effort: key}]
		if request.Header.Get("X-Radar-Bench-Selection") != intelligenceSelectionMark {
			source.status[key] = http.StatusPreconditionRequired
		}
	default:
		key = "unknown"
		source.status[key] = http.StatusNotFound
	}
	body, err := json.Marshal(payload)
	status, retryAfter, gate := source.status[key], source.retryAfter, source.gate[key]
	source.mu.Unlock()
	if err != nil {
		source.t.Errorf("marshal fake payload: %v", err)
		return
	}
	if gate != nil {
		<-gate
	}
	if status != 0 {
		if retryAfter != "" {
			writer.Header().Set("Retry-After", retryAfter)
		}
		writer.WriteHeader(status)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, _ = writer.Write(body)
}

func (source *intelligenceTestSource) cache(clock *intelligenceTestClock) *intelligenceCache {
	cache := newIntelligenceCache("", false)
	cache.origin, cache.now = source.server.URL, clock.Now
	return cache
}

type intelligenceTestBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		intelligenceSnapshot
		RefreshFailed bool `json:"refresh_failed"`
	} `json:"data"`
}

func intelligenceTestGet(t *testing.T, cache *intelligenceCache) (int, intelligenceTestBody) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/intelligence", cache.Get)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/intelligence", nil))
	var body intelligenceTestBody
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body), recorder.Body.String())
	return recorder.Code, body
}

// 实取汇总经代读后与源站页面展示一致：81.8 分、83 题次、$0.33、8.2 分钟（截图同格）。
func TestIntelligenceRealSummaryMatchesSourceDisplay(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low", "medium"))
	clock := newIntelligenceTestClock()
	status, body := intelligenceTestGet(t, source.cache(clock))
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, body.Code)
	snapshot := body.Data
	require.False(t, snapshot.RefreshFailed)
	require.Equal(t, "radar-bench-final64-20261004", snapshot.Benchmark)
	require.Equal(t, 64, snapshot.RequiredTasks)
	require.Equal(t, float64(150), snapshot.ScoreMax)
	require.Equal(t, "2026-10-12T10:00:00Z", snapshot.CheckedAt)
	require.Len(t, snapshot.Points, 2)
	require.Equal(t, []string{"low", "medium"}, []string{snapshot.Points[0].Effort, snapshot.Points[1].Effort}, "按绑定顺序输出")
	point := snapshot.Points[0]
	require.Equal(t, "ok", point.State)
	require.Equal(t, "complete", point.ScoreStatus)
	require.Equal(t, 81.8, math.Round(*point.Score*10)/10)
	require.Equal(t, 64, *point.Coverage)
	require.Equal(t, 83, *point.GradedRuns)
	require.Equal(t, 0.33, math.Round(*point.AveragePriceUSD*100)/100)
	require.InDelta(t, 23.478868/71, *point.AveragePriceUSD, 1e-12, "均值 = 已知合计 ÷ 已知样本，不按总样本摊薄")
	require.Equal(t, 8.2, math.Round(*point.AverageMinutes*10)/10)
	require.Equal(t, []int{71, 82, 83}, []int{*point.PriceSamples, *point.DurationSamples, *point.StatisticsCount})
	require.Equal(t, "available", point.StatisticsState)
	require.Equal(t, "2026-10-12T10:00:00Z", point.ReadAt)

	require.EqualValues(t, 1, source.bindingHits.Load())
	require.EqualValues(t, 2, source.summaryHits.Load())
	require.ElementsMatch(t, []string{"model=gpt-6.1-sol&effort=low&view=summary", "model=gpt-6.1-sol&effort=medium&view=summary"}, source.queries)
}

// 0 分是有效成绩；无覆盖、未提供统计时为 null，不当作 0 分或免费。
func TestIntelligenceKeepsZeroDistinctFromMissing(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low", "medium"))
	source.override("low", map[string]any{"score": 0, "coverage": 1, "score_status": "provisional", "graded_runs_count": 1, "statistics": nil, "statistics_state": "not_provided"})
	source.override("medium", map[string]any{"score": nil, "coverage": 0, "score_status": "missing_current_result", "graded_runs_count": 0})
	cache := source.cache(newIntelligenceTestClock())
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/intelligence", cache.Get)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/intelligence", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
	raw := recorder.Body.String()
	require.Contains(t, raw, `"effort":"low","state":"ok","score":0,"coverage":1`)
	require.Contains(t, raw, `"effort":"medium","state":"ok","score":null,"coverage":0,"score_status":"missing_current_result"`)
	require.Contains(t, raw, `"average_price_usd":null,"price_samples":null,"average_minutes":null,"duration_samples":null,"statistics_samples":null,"statistics_state":"not_provided"`)
	require.NotContains(t, raw, "task_set", "内部题库摘要不对外输出")
}

func intelligenceDecodeSummary(t *testing.T, summary map[string]any) intelligenceSummary {
	t.Helper()
	raw, err := json.Marshal(summary)
	require.NoError(t, err)
	var decoded intelligenceSummary
	require.NoError(t, json.Unmarshal(raw, &decoded))
	return decoded
}

func TestIntelligenceRejectsInconsistentSummary(t *testing.T) {
	binding := intelligenceTestBinding("low")
	selection := binding.ModelEfforts[0]
	_, err := intelligenceSummaryPoint(intelligenceDecodeSummary(t, intelligenceTestSummary(t, binding, selection, nil)), binding, selection)
	require.NoError(t, err, "实取原文必须通过")
	for name, overrides := range map[string]map[string]any{
		"汇总版本变化":    {"schema": "radar-bench-public-summary-v6"},
		"计分版本变化":    {"score_version": "radar-bench-v3-100"},
		"满分变化":      {"score_max": 100},
		"计分规则变化":    {"scoring_policy_version": "radar-bench-v3-best-of-5"},
		"题库摘要不一致":   {"task_set_sha256": strings.Repeat("0", 64)},
		"档位与请求不一致":  {"effort": "high"},
		"分数超过满分":    {"score": 150.5},
		"负分":        {"score": -1},
		"有覆盖却无分数":   {"score": nil},
		"无覆盖却有分数":   {"coverage": 0, "score_status": "missing_current_result", "graded_runs_count": 0},
		"缺少覆盖题数":    {"coverage": nil},
		"覆盖题数为小数":   {"coverage": 10.5},
		"覆盖超过总题数":   {"coverage": 65},
		"总题数与绑定不一致": {"required_tasks": 65},
		"状态与覆盖不一致":  {"score_status": "provisional"},
		"判分题次少于覆盖":  {"graded_runs_count": 63},
		"判分题次超过三倍":  {"graded_runs_count": 193},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := intelligenceSummaryPoint(intelligenceDecodeSummary(t, intelligenceTestSummary(t, binding, selection, overrides)), binding, selection)
			require.ErrorIs(t, err, errIntelligenceInvalid)
		})
	}
}

// 统计异常只让费用／耗时缺失，分数照常展示，与源站页面一致。
func TestIntelligenceInvalidStatisticsKeepScore(t *testing.T) {
	binding := intelligenceTestBinding("low")
	selection := binding.ModelEfforts[0]
	metric := func(summary map[string]any, name string) map[string]any {
		return summary["statistics"].(map[string]any)["metrics"].(map[string]any)[name].(map[string]any)
	}
	for name, mutate := range map[string]func(map[string]any){
		"费用口径变化":     func(summary map[string]any) { summary["statistics"].(map[string]any)["cost_basis"] = "estimated" },
		"样本数不一致":     func(summary map[string]any) { metric(summary, "api_equivalent_usd")["missing_samples"] = 13 },
		"有已知样本却无合计":  func(summary map[string]any) { metric(summary, "elapsed_ms")["known_total"] = nil },
		"负费用":        func(summary map[string]any) { metric(summary, "api_equivalent_usd")["known_total"] = -1 },
		"状态可用却缺少统计":  func(summary map[string]any) { summary["statistics"] = nil },
		"未知统计状态":     func(summary map[string]any) { summary["statistics_state"] = "future_state" },
		"源站标记统计来源异常": func(summary map[string]any) { summary["statistics_state"] = "invalid_source_statistics" },
	} {
		t.Run(name, func(t *testing.T) {
			summary := intelligenceTestSummary(t, binding, selection, nil)
			mutate(summary)
			point, err := intelligenceSummaryPoint(intelligenceDecodeSummary(t, summary), binding, selection)
			require.NoError(t, err)
			require.Equal(t, "invalid_source_statistics", point.StatisticsState)
			require.NotNil(t, point.Score)
			require.Nil(t, point.AveragePriceUSD)
			require.Nil(t, point.AverageMinutes)
		})
	}

	// 无任何已知费用（截图 GPT-5.6 Luna xhigh 显示 –）：费用为 null、已知样本 0，耗时照常。
	summary := intelligenceTestSummary(t, binding, selection, nil)
	price := metric(summary, "api_equivalent_usd")
	price["known_total"], price["known_samples"], price["missing_samples"] = nil, 0, 83
	point, err := intelligenceSummaryPoint(intelligenceDecodeSummary(t, summary), binding, selection)
	require.NoError(t, err)
	require.Equal(t, "available", point.StatisticsState)
	require.Nil(t, point.AveragePriceUSD)
	require.Equal(t, 0, *point.PriceSamples)
	require.NotNil(t, point.AverageMinutes)
}

// 单个档位失败：保留同一题库上次成功值并标记 retained；题库变化后旧值不再沿用。
func TestIntelligencePartialFailureRetainsSameCatalogOnly(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low", "medium"))
	clock := newIntelligenceTestClock()
	cache := source.cache(clock)
	first, _ := cache.current(context.Background())
	require.NotNil(t, first)

	clock.Advance(intelligenceFreshFor + time.Minute)
	source.override("low", map[string]any{"score": 90.0})
	source.setStatus("medium", http.StatusInternalServerError, "")
	second, failed := cache.current(context.Background())
	require.False(t, failed, "部分失败仍发布新快照")
	require.Equal(t, "ok", second.Points[0].State)
	require.Equal(t, 90.0, *second.Points[0].Score)
	require.Equal(t, "retained", second.Points[1].State)
	require.Equal(t, *first.Points[1].Score, *second.Points[1].Score)
	require.Equal(t, first.Points[1].ReadAt, second.Points[1].ReadAt, "保留原读取时间")
	require.True(t, cache.nextAt.Equal(clock.Now().Add(intelligenceRetryMin)), "5xx 一分钟后再读")

	clock.Advance(2 * time.Minute)
	changed := intelligenceTestBinding("low", "medium")
	changed.CatalogVersion = "dradar-next-catalog"
	source.setBinding(changed)
	third, _ := cache.current(context.Background())
	require.Equal(t, "dradar-next-catalog", third.CatalogVersion)
	require.Equal(t, "ok", third.Points[0].State)
	require.Equal(t, "unavailable", third.Points[1].State)
	require.Nil(t, third.Points[1].Score, "不把旧题库的分数挂到新题库")
}

// 整轮失败：继续返回上次快照并标记 refresh_failed，遵守 Retry-After。
func TestIntelligenceTotalFailureKeepsSnapshotAndBacksOff(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low"))
	clock := newIntelligenceTestClock()
	cache := source.cache(clock)
	first, _ := cache.current(context.Background())
	clock.Advance(intelligenceFreshFor)
	source.setStatus("binding", http.StatusServiceUnavailable, "120")
	status, body := intelligenceTestGet(t, cache)
	require.Equal(t, http.StatusOK, status)
	require.True(t, body.Data.RefreshFailed)
	require.Equal(t, first.CheckedAt, body.Data.CheckedAt)
	require.True(t, cache.nextAt.Equal(clock.Now().Add(2*time.Minute)))
	clock.Advance(time.Minute)
	_, failed := cache.current(context.Background())
	require.True(t, failed)
	require.EqualValues(t, 2, source.bindingHits.Load(), "退避期内不再请求源站")
}

// 首次读取就失败时返回 503，一分钟内不重复请求源站。
func TestIntelligenceColdFailureReturnsUnavailable(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low"))
	source.setStatus("binding", http.StatusBadGateway, "")
	clock := newIntelligenceTestClock()
	cache := source.cache(clock)
	status, body := intelligenceTestGet(t, cache)
	require.Equal(t, http.StatusServiceUnavailable, status)
	require.Contains(t, body.Message, "评测数据源暂时不可用")
	status, _ = intelligenceTestGet(t, cache)
	require.Equal(t, http.StatusServiceUnavailable, status)
	require.EqualValues(t, 1, source.bindingHits.Load())
}

// 同时打开页面只触发一轮读取；新鲜期内刷新只读缓存。
func TestIntelligenceSingleFlightAndFreshWindow(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low", "medium"))
	release := source.block("medium")
	clock := newIntelligenceTestClock()
	cache := source.cache(clock)
	snapshots := make([]*intelligenceSnapshot, 8)
	var wg sync.WaitGroup
	for index := range snapshots {
		wg.Add(1)
		go func() {
			defer wg.Done()
			snapshots[index], _ = cache.current(context.Background())
		}()
	}
	require.Eventually(t, func() bool { return source.summaryHits.Load() == 2 }, 5*time.Second, 5*time.Millisecond)
	release()
	wg.Wait()
	for _, snapshot := range snapshots {
		require.NotNil(t, snapshot)
		require.Same(t, snapshots[0], snapshot)
	}
	require.EqualValues(t, 1, source.bindingHits.Load())

	clock.Advance(intelligenceFreshFor - time.Second)
	_, _ = cache.current(context.Background())
	require.EqualValues(t, 1, source.bindingHits.Load())
	clock.Advance(time.Second)
	_, _ = cache.current(context.Background())
	require.EqualValues(t, 2, source.bindingHits.Load())
}

// 已有旧数据时源站变慢：最多等 staleWait 后先返回旧数据，读取完成后再更新。
func TestIntelligenceSlowRefreshServesPreviousSnapshot(t *testing.T) {
	source := newIntelligenceTestSource(t, intelligenceTestBinding("low"))
	clock := newIntelligenceTestClock()
	cache := source.cache(clock)
	cache.staleWait = 50 * time.Millisecond
	first, _ := cache.current(context.Background())
	clock.Advance(intelligenceFreshFor)
	source.override("low", map[string]any{"score": 99.0})
	release := source.block("low")
	started := time.Now()
	stale, failed := cache.current(context.Background())
	require.Less(t, time.Since(started), 5*time.Second)
	require.Same(t, first, stale)
	require.False(t, failed)
	release()
	require.Eventually(t, func() bool {
		snapshot, _ := cache.current(context.Background())
		return snapshot != first && *snapshot.Points[0].Score == 99.0
	}, 5*time.Second, 5*time.Millisecond)
}

// 源站繁忙后本轮不再发起剩余请求。
func TestIntelligenceBusySourceStopsRemainingReads(t *testing.T) {
	binding := intelligenceTestBinding("low")
	source := newIntelligenceTestSource(t, binding)
	source.setStatus("low", http.StatusTooManyRequests, "120")
	cache := source.cache(newIntelligenceTestClock())
	var busy atomic.Bool
	result := cache.readSummary(context.Background(), binding, binding.ModelEfforts[0], &busy)
	require.Error(t, result.err)
	require.True(t, busy.Load())
	require.Equal(t, 2*time.Minute, intelligenceRetry(result.err))
	result = cache.readSummary(context.Background(), binding, binding.ModelEfforts[0], &busy)
	require.ErrorIs(t, result.err, errIntelligenceSkipped)
	require.EqualValues(t, 1, source.summaryHits.Load())
}

// 不跟随跳转、不接受超大响应、代理初始化失败不悄悄直连。
func TestIntelligenceRejectsRedirectOversizeAndProxyFallback(t *testing.T) {
	var elsewhere atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { elsewhere.Add(1) }))
	t.Cleanup(target.Close)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == intelligenceBindingPath {
			http.Redirect(writer, request, target.URL+intelligenceBindingPath, http.StatusFound)
			return
		}
		_, _ = writer.Write([]byte(`{"pad":"` + strings.Repeat("x", intelligenceSummaryLimit) + `"}`))
	}))
	t.Cleanup(server.Close)
	cache := newIntelligenceCache("", false)
	cache.origin = server.URL
	var binding intelligenceBinding
	err := cache.get(context.Background(), intelligenceBindingPath, false, intelligenceBindingLimit, &binding)
	var sourceErr *intelligenceSourceError
	require.ErrorAs(t, err, &sourceErr)
	require.Equal(t, http.StatusFound, sourceErr.status)
	require.Zero(t, elsewhere.Load())
	var summary intelligenceSummary
	require.ErrorIs(t, cache.get(context.Background(), intelligenceSummaryPath, true, intelligenceSummaryLimit, &summary), errIntelligenceInvalid)

	blocked := newIntelligenceCache("ftp://proxy.invalid:21", false)
	require.Nil(t, blocked.client)
	require.Error(t, blocked.get(context.Background(), intelligenceBindingPath, false, intelligenceBindingLimit, &binding))
	direct := newIntelligenceCache("ftp://proxy.invalid:21", true)
	require.NotNil(t, direct.client, "仅管理员显式允许时回退直连")
}

func TestIntelligenceRetryAfterAndRetryPolicy(t *testing.T) {
	now := time.Date(2026, 10, 12, 10, 0, 0, 0, time.UTC)
	for value, expected := range map[string]time.Duration{
		"":      0,
		"120":   2 * time.Minute,
		"5":     intelligenceRetryMin,
		"99999": intelligenceRetryMax,
		"Inf":   intelligenceRetryMax,
		"NaN":   0,
		"-3":    0,
		"later": 0,
		now.Add(5 * time.Minute).Format(http.TimeFormat): 5 * time.Minute,
		now.Add(-time.Minute).Format(http.TimeFormat):    0,
	} {
		require.Equal(t, expected, intelligenceRetryAfter(value, now), value)
	}
	require.Equal(t, intelligenceRetryMin, intelligenceRetry(fmt.Errorf("x: %w", context.DeadlineExceeded)))
	require.Equal(t, intelligenceRetryMin, intelligenceRetry(&intelligenceSourceError{status: http.StatusTooManyRequests}))
	require.Equal(t, intelligenceFreshFor, intelligenceRetry(&intelligenceSourceError{status: http.StatusPreconditionRequired}), "缺选择标记等 4xx 不快速重试")
	require.Equal(t, intelligenceFreshFor, intelligenceRetry(intelligenceInvalid("schema")))

	binding := intelligenceTestBinding("low", "medium")
	cache := &intelligenceCache{now: newIntelligenceTestClock().Now}
	okPoint := intelligenceResult{point: intelligencePoint{Model: "gpt-6.1-sol", Effort: "low"}}
	_, retry, err := cache.merge(binding, []intelligenceResult{okPoint, {err: errors.New("timeout")}}, nil)
	require.NoError(t, err)
	require.Equal(t, intelligenceRetryMin, retry)
	_, retry, err = cache.merge(binding, []intelligenceResult{okPoint, {err: &intelligenceSourceError{status: http.StatusTooManyRequests, retry: 5 * time.Minute}}}, nil)
	require.NoError(t, err)
	require.Equal(t, 5*time.Minute, retry)
	_, retry, err = cache.merge(binding, []intelligenceResult{okPoint, okPoint}, nil)
	require.NoError(t, err)
	require.Equal(t, intelligenceFreshFor, retry)
	_, _, err = cache.merge(binding, []intelligenceResult{{err: errIntelligenceSkipped}, {err: errIntelligenceSkipped}}, nil)
	require.Error(t, err, "全部失败不替换缓存")
}

// 接口与 /intelligence 页面一样要求登录；未登录在访问源站前就被拒绝。
func TestIntelligenceRouteRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	reject := func(ctx *gin.Context) { ctx.AbortWithStatus(http.StatusUnauthorized) }
	RegisterIntelligenceRoutes(router.Group("/api/v1"), t.TempDir(), "", false, reject, nil, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/intelligence", nil))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}
