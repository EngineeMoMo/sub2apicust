package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/httpclient"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// 智力效率页面的数据源：Codex Radar「雷达分 · 雷达Bench v1」公开汇总（用户 2026-10-12 指定）。
// 源站汇总接口不返回 CORS 头，且非默认档位必须带显式选择标记（否则 428），浏览器无法直接读取，
// 所以由服务端按源站页面同样的方式逐个档位读取，校验后缓存。只读源站公开数据，不发送本站用户信息。
const (
	intelligenceOrigin        = "https://codex-reset-radar.pages.dev"
	intelligenceBindingPath   = "/data/radar-bench-binding.json"
	intelligenceSummaryPath   = "/api/radar-bench-score"
	intelligenceSelectionMark = "user-selected-v1"
	intelligenceUserAgent     = "Sub2API-Intelligence"

	// 只接受已核对过字段含义的契约版本；源站升级版本后显示「格式变化」，不猜测新字段。
	intelligenceScoreVersion  = "radar-bench-v2-150"
	intelligenceSummarySchema = "radar-bench-public-summary-v5-150-daily-graded-count"
	intelligencePolicyVersion = "radar-bench-v2-last3-per-task-equal-weight-150"
	intelligenceCostBasis     = "official_standard_api_equivalent; no estimates or claimed cash charges"
	intelligenceScoreMax      = 150

	intelligenceMaxSelections = 64
	intelligenceConcurrency   = 3 // 与源站页面自身的读取并发一致
	intelligenceBindingLimit  = 1 << 20
	intelligenceSummaryLimit  = 256 << 10

	intelligenceFreshFor       = 10 * time.Minute // 源站汇总自身缓存 60 秒，这里放宽，避免本站访问放大到源站
	intelligenceRetryMin       = time.Minute
	intelligenceRetryMax       = 30 * time.Minute
	intelligenceRequestTimeout = 15 * time.Second
	intelligenceRoundTimeout   = 40 * time.Second
	intelligenceColdWait       = 45 * time.Second // 首次没有任何数据时等待一轮读取
	intelligenceStaleWait      = 20 * time.Second // 已有旧数据时最多等这么久，超时先返回旧数据
)

var (
	intelligenceNamePattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
	intelligenceDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	errIntelligenceInvalid    = errors.New("intelligence source data invalid")
	errIntelligenceSkipped    = errors.New("intelligence read skipped after source busy")
)

type intelligenceSelection struct {
	Model  string `json:"model"`
	Effort string `json:"effort"`
}

// 源站绑定文件：当前题库摘要及全部模型／档位。
type intelligenceBinding struct {
	Enabled                bool                    `json:"enabled"`
	ProductionReadVerified bool                    `json:"production_read_verified"`
	Benchmark              string                  `json:"benchmark"`
	CatalogVersion         string                  `json:"catalog_version"`
	TaskSetSHA256          string                  `json:"task_set_sha256"`
	TotalTasks             int                     `json:"total_tasks"`
	ScoreVersion           string                  `json:"score_version"`
	ScoreMax               float64                 `json:"score_max"`
	ModelEfforts           []intelligenceSelection `json:"model_efforts"`
}

type intelligenceMetric struct {
	KnownTotal     *float64 `json:"known_total"`
	KnownSamples   *float64 `json:"known_samples"`
	MissingSamples *float64 `json:"missing_samples"`
}

// 源站单个模型／档位的公开汇总；计数用 float64 解码后再校验为非负整数。
type intelligenceSummary struct {
	Schema          string   `json:"schema"`
	Benchmark       string   `json:"benchmark"`
	ScoreVersion    string   `json:"score_version"`
	ScoreMax        float64  `json:"score_max"`
	CatalogVersion  string   `json:"catalog_version"`
	TaskSetSHA256   string   `json:"task_set_sha256"`
	Model           string   `json:"model"`
	Effort          string   `json:"effort"`
	Score           *float64 `json:"score"`
	Coverage        *float64 `json:"coverage"`
	RequiredTasks   *float64 `json:"required_tasks"`
	ScoreStatus     string   `json:"score_status"`
	GradedRunsCount *float64 `json:"graded_runs_count"`
	PolicyVersion   string   `json:"scoring_policy_version"`
	StatisticsState string   `json:"statistics_state"`
	Statistics      *struct {
		SampleCount *float64 `json:"sample_count"`
		CostBasis   string   `json:"cost_basis"`
		Metrics     struct {
			ElapsedMS *intelligenceMetric `json:"elapsed_ms"`
			USD       *intelligenceMetric `json:"api_equivalent_usd"`
		} `json:"metrics"`
	} `json:"statistics"`
}

// 页面展示的一个模型／档位。state：ok 本轮读取成功；retained 本轮失败、保留上次成功值；unavailable 尚无成功读取。
// 缺失一律为 null，不当作 0 分或免费。
type intelligencePoint struct {
	Model           string   `json:"model"`
	Effort          string   `json:"effort"`
	State           string   `json:"state"`
	Score           *float64 `json:"score"`
	Coverage        *int     `json:"coverage"`
	ScoreStatus     string   `json:"score_status"`
	GradedRuns      *int     `json:"graded_runs"`
	AveragePriceUSD *float64 `json:"average_price_usd"`
	PriceSamples    *int     `json:"price_samples"`
	AverageMinutes  *float64 `json:"average_minutes"`
	DurationSamples *int     `json:"duration_samples"`
	StatisticsCount *int     `json:"statistics_samples"`
	StatisticsState string   `json:"statistics_state"`
	ReadAt          string   `json:"read_at"`
}

type intelligenceSnapshot struct {
	Benchmark      string              `json:"benchmark"`
	CatalogVersion string              `json:"catalog_version"`
	ScoreMax       float64             `json:"score_max"`
	RequiredTasks  int                 `json:"required_tasks"`
	CheckedAt      string              `json:"checked_at"`
	Points         []intelligencePoint `json:"points"`
	taskSet        string
}

type intelligenceResponse struct {
	intelligenceSnapshot
	RefreshFailed bool `json:"refresh_failed"`
}

// 进程内缓存：只在有人打开页面时读取源站，同一时间只跑一轮，不是常驻定时采集。
type intelligenceCache struct {
	client    *http.Client
	clientErr error
	origin    string
	now       func() time.Time
	coldWait  time.Duration
	staleWait time.Duration

	mu       sync.Mutex
	snapshot *intelligenceSnapshot
	failed   bool
	nextAt   time.Time
	running  chan struct{}
}

func newIntelligenceCache(proxyURL string, allowDirectOnProxyError bool) *intelligenceCache {
	cache := &intelligenceCache{origin: intelligenceOrigin, now: time.Now, coldWait: intelligenceColdWait, staleWait: intelligenceStaleWait}
	// 与定价、更新检查相同：复用 update.proxy_url；代理初始化失败默认不回退直连，避免暴露服务器真实 IP。
	client, err := httpclient.GetClient(httpclient.Options{Timeout: intelligenceRequestTimeout, ProxyURL: proxyURL})
	if err != nil {
		if strings.TrimSpace(proxyURL) != "" && !allowDirectOnProxyError {
			slog.Warn("proxy client init failed, all requests will fail", "service", "intelligence", "error", err)
			cache.clientErr = fmt.Errorf("proxy client init failed and direct fallback is disabled: %w", err)
			return cache
		}
		client = &http.Client{Timeout: intelligenceRequestTimeout}
	}
	cloned := *client
	// 只读固定源站，不跟随跳转到其他地址。
	cloned.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	cache.client = &cloned
	return cache
}

// RegisterIntelligenceRoutes 注册智力效率数据接口：与 /intelligence 页面一样只对登录用户开放，沿用面板全局限流。
// 刷新只读本站缓存，不会把每次访问放大成源站请求。
func RegisterIntelligenceRoutes(v1 *gin.RouterGroup, proxyURL string, allowDirectOnProxyError bool, jwtAuth gin.HandlerFunc, settingService *service.SettingService, limiter *middleware.PanelRateLimiter) {
	cache := newIntelligenceCache(proxyURL, allowDirectOnProxyError)
	group := v1.Group("/intelligence")
	group.Use(jwtAuth, middleware.BackendModeUserGuard(settingService), limiter.Global())
	group.GET("", cache.Get)
}

func (cache *intelligenceCache) Get(ctx *gin.Context) {
	snapshot, failed := cache.current(ctx.Request.Context())
	if snapshot == nil {
		response.Error(ctx, http.StatusServiceUnavailable, "评测数据源暂时不可用，请稍后重试")
		return
	}
	response.Success(ctx, intelligenceResponse{intelligenceSnapshot: *snapshot, RefreshFailed: failed})
}

// current 返回缓存；过期时启动一轮读取。已有旧数据最多等 staleWait，超时先返回旧数据，读取在后台完成。
func (cache *intelligenceCache) current(ctx context.Context) (*intelligenceSnapshot, bool) {
	cache.mu.Lock()
	done := cache.running
	if done == nil && !cache.now().Before(cache.nextAt) {
		done = make(chan struct{})
		cache.running = done
		go cache.refresh(done)
	}
	snapshot, failed := cache.snapshot, cache.failed
	cache.mu.Unlock()
	if done == nil {
		return snapshot, failed
	}
	wait := cache.staleWait
	if snapshot == nil {
		wait = cache.coldWait
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
	case <-ctx.Done():
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	return cache.snapshot, cache.failed
}

func (cache *intelligenceCache) refresh(done chan struct{}) {
	ctx, cancel := context.WithTimeout(context.Background(), intelligenceRoundTimeout)
	defer cancel()
	cache.mu.Lock()
	previous := cache.snapshot
	cache.mu.Unlock()
	snapshot, retry, err := cache.safeRead(ctx, previous)
	cache.mu.Lock()
	defer cache.mu.Unlock()
	defer close(done)
	cache.running = nil
	cache.nextAt = cache.now().Add(retry)
	if err != nil {
		cache.failed = true
		slog.Warn("intelligence source refresh failed", "error", err, "retry_after", retry.String())
		return
	}
	cache.snapshot, cache.failed = snapshot, false
}

func (cache *intelligenceCache) safeRead(ctx context.Context, previous *intelligenceSnapshot) (snapshot *intelligenceSnapshot, retry time.Duration, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("intelligence_refresh_panic", "recover", recovered)
			snapshot, retry, err = nil, intelligenceFreshFor, fmt.Errorf("%w: refresh panic", errIntelligenceInvalid)
		}
	}()
	return cache.read(ctx, previous)
}

type intelligenceResult struct {
	point intelligencePoint
	err   error
}

// read 读取一轮：绑定文件 → 全部模型／档位汇总（并发 3）。源站返回 429／502／503 后不再发起剩余请求。
func (cache *intelligenceCache) read(ctx context.Context, previous *intelligenceSnapshot) (*intelligenceSnapshot, time.Duration, error) {
	var binding intelligenceBinding
	if err := cache.get(ctx, intelligenceBindingPath, false, intelligenceBindingLimit, &binding); err != nil {
		return nil, intelligenceRetry(err), fmt.Errorf("read binding: %w", err)
	}
	if err := validateIntelligenceBinding(binding); err != nil {
		return nil, intelligenceRetry(err), err
	}
	results := make([]intelligenceResult, len(binding.ModelEfforts))
	jobs := make(chan int)
	var busy atomic.Bool
	var wg sync.WaitGroup
	for range intelligenceConcurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				results[index] = cache.readSummary(ctx, binding, binding.ModelEfforts[index], &busy)
			}
		}()
	}
	for index := range binding.ModelEfforts {
		jobs <- index
	}
	close(jobs)
	wg.Wait()
	return cache.merge(binding, results, previous)
}

func (cache *intelligenceCache) readSummary(ctx context.Context, binding intelligenceBinding, selection intelligenceSelection, busy *atomic.Bool) (result intelligenceResult) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("intelligence_summary_panic", "recover", recovered)
			result = intelligenceResult{err: fmt.Errorf("%w: summary panic", errIntelligenceInvalid)}
		}
	}()
	if busy.Load() {
		return intelligenceResult{err: errIntelligenceSkipped}
	}
	// 参数顺序与源站页面一致。
	path := intelligenceSummaryPath + "?model=" + url.QueryEscape(selection.Model) + "&effort=" + url.QueryEscape(selection.Effort) + "&view=summary"
	var summary intelligenceSummary
	if err := cache.get(ctx, path, true, intelligenceSummaryLimit, &summary); err != nil {
		var sourceErr *intelligenceSourceError
		if errors.As(err, &sourceErr) && intelligenceBusy(sourceErr.status) {
			busy.Store(true)
		}
		return intelligenceResult{err: fmt.Errorf("%s/%s: %w", selection.Model, selection.Effort, err)}
	}
	point, err := intelligenceSummaryPoint(summary, binding, selection)
	if err != nil {
		return intelligenceResult{err: fmt.Errorf("%s/%s: %w", selection.Model, selection.Effort, err)}
	}
	point.ReadAt = cache.now().UTC().Format(time.RFC3339)
	return intelligenceResult{point: point}
}

// merge 合并本轮结果：失败的档位沿用同一题库上次成功的值（标记 retained），从未成功的标记 unavailable。
// 全部失败时不替换缓存。
func (cache *intelligenceCache) merge(binding intelligenceBinding, results []intelligenceResult, previous *intelligenceSnapshot) (*intelligenceSnapshot, time.Duration, error) {
	retained := map[intelligenceSelection]intelligencePoint{}
	if previous != nil && previous.Benchmark == binding.Benchmark && previous.CatalogVersion == binding.CatalogVersion && previous.taskSet == binding.TaskSetSHA256 {
		for _, point := range previous.Points {
			if point.State != "unavailable" {
				retained[intelligenceSelection{Model: point.Model, Effort: point.Effort}] = point
			}
		}
	}
	snapshot := &intelligenceSnapshot{Benchmark: binding.Benchmark, CatalogVersion: binding.CatalogVersion, ScoreMax: binding.ScoreMax,
		RequiredTasks: binding.TotalTasks, CheckedAt: cache.now().UTC().Format(time.RFC3339), Points: make([]intelligencePoint, 0, len(results)), taskSet: binding.TaskSetSHA256}
	retry, failures := intelligenceFreshFor, 0
	var firstErr error
	for index, result := range results {
		selection := binding.ModelEfforts[index]
		if result.err == nil {
			result.point.State = "ok"
			snapshot.Points = append(snapshot.Points, result.point)
			continue
		}
		if failures == 0 {
			retry, firstErr = 0, result.err
		}
		failures++
		retry = max(retry, intelligenceRetry(result.err))
		if point, ok := retained[selection]; ok {
			point.State = "retained"
			snapshot.Points = append(snapshot.Points, point)
		} else {
			snapshot.Points = append(snapshot.Points, intelligencePoint{Model: selection.Model, Effort: selection.Effort, State: "unavailable"})
		}
	}
	if failures == len(results) {
		return nil, retry, fmt.Errorf("all %d selections failed, first: %w", failures, firstErr)
	}
	if failures > 0 {
		slog.Warn("intelligence source partially failed", "failed", failures, "total", len(results), "first_error", firstErr, "retry_after", retry.String())
	}
	return snapshot, retry, nil
}

type intelligenceSourceError struct {
	status int
	retry  time.Duration
}

func (err *intelligenceSourceError) Error() string {
	return fmt.Sprintf("intelligence source HTTP %d", err.status)
}

// get 发起固定路径的匿名 GET：不带本站 Cookie、Token 或用户信息。summary 请求带源站页面同样的显式选择标记。
func (cache *intelligenceCache) get(ctx context.Context, path string, selected bool, limit int64, target any) error {
	if cache.clientErr != nil {
		return cache.clientErr
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, cache.origin+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", intelligenceUserAgent)
	if selected {
		request.Header.Set("X-Radar-Bench-Selection", intelligenceSelectionMark)
	}
	resp, err := cache.client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return &intelligenceSourceError{status: resp.StatusCode, retry: intelligenceRetryAfter(resp.Header.Get("Retry-After"), cache.now())}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return err
	}
	if int64(len(body)) > limit {
		return fmt.Errorf("%w: response exceeds %d bytes", errIntelligenceInvalid, limit)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("%w: %v", errIntelligenceInvalid, err)
	}
	return nil
}

// intelligenceRetry 决定失败后的重试间隔：源站给出 Retry-After 时遵守；繁忙／5xx／网络错误 1 分钟；
// 数据格式不符和其他 4xx 不会很快恢复，按正常周期 10 分钟。
func intelligenceRetry(err error) time.Duration {
	var sourceErr *intelligenceSourceError
	if errors.As(err, &sourceErr) {
		if sourceErr.retry > 0 {
			return sourceErr.retry
		}
		if sourceErr.status >= 400 && sourceErr.status < 500 && sourceErr.status != http.StatusTooManyRequests {
			return intelligenceFreshFor
		}
		return intelligenceRetryMin
	}
	if errors.Is(err, errIntelligenceInvalid) {
		return intelligenceFreshFor
	}
	return intelligenceRetryMin
}

// intelligenceRetryAfter 解析秒数或 HTTP 日期，限制在 1–30 分钟；无效值返回 0。
func intelligenceRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	var wait time.Duration
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		if math.IsNaN(seconds) || seconds <= 0 {
			return 0
		}
		wait = time.Duration(math.Min(seconds, intelligenceRetryMax.Seconds()) * float64(time.Second))
	} else if at, err := http.ParseTime(value); err == nil {
		wait = at.Sub(now)
		if wait <= 0 {
			return 0
		}
	} else {
		return 0
	}
	return min(max(wait, intelligenceRetryMin), intelligenceRetryMax)
}

func intelligenceBusy(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusBadGateway || status == http.StatusServiceUnavailable
}

func intelligenceInvalid(reason string) error {
	return fmt.Errorf("%w: %s", errIntelligenceInvalid, reason)
}

// validateIntelligenceBinding 校验绑定文件：已启用、计分版本一致、题库摘要及模型／档位列表合法。
func validateIntelligenceBinding(binding intelligenceBinding) error {
	switch {
	case !binding.Enabled || !binding.ProductionReadVerified:
		return intelligenceInvalid("binding not enabled")
	case binding.ScoreVersion != intelligenceScoreVersion || binding.ScoreMax != intelligenceScoreMax:
		return intelligenceInvalid("unsupported score version")
	case !intelligenceDigestPattern.MatchString(binding.TaskSetSHA256) || binding.Benchmark == "" || len(binding.Benchmark) > 200 || binding.CatalogVersion == "" || len(binding.CatalogVersion) > 200:
		return intelligenceInvalid("benchmark identity")
	case binding.TotalTasks <= 0 || binding.TotalTasks > 1000:
		return intelligenceInvalid("total tasks")
	case len(binding.ModelEfforts) == 0 || len(binding.ModelEfforts) > intelligenceMaxSelections:
		return intelligenceInvalid("selection count")
	}
	seen := make(map[intelligenceSelection]bool, len(binding.ModelEfforts))
	for _, selection := range binding.ModelEfforts {
		if !intelligenceNamePattern.MatchString(selection.Model) || !intelligenceNamePattern.MatchString(selection.Effort) || seen[selection] {
			return intelligenceInvalid("selection")
		}
		seen[selection] = true
	}
	return nil
}

// intelligenceSummaryPoint 按源站页面同样的约束校验公开汇总，只取展示字段；不重算分数、不补零。
func intelligenceSummaryPoint(summary intelligenceSummary, binding intelligenceBinding, selection intelligenceSelection) (intelligencePoint, error) {
	if summary.Schema != intelligenceSummarySchema || summary.ScoreVersion != intelligenceScoreVersion || summary.ScoreMax != intelligenceScoreMax || summary.PolicyVersion != intelligencePolicyVersion {
		return intelligencePoint{}, intelligenceInvalid("summary version")
	}
	if summary.Benchmark != binding.Benchmark || summary.CatalogVersion != binding.CatalogVersion || summary.TaskSetSHA256 != binding.TaskSetSHA256 {
		return intelligencePoint{}, intelligenceInvalid("summary benchmark differs from binding")
	}
	if summary.Model != selection.Model || summary.Effort != selection.Effort {
		return intelligencePoint{}, intelligenceInvalid("summary selection differs from request")
	}
	required, ok := intelligenceCount(summary.RequiredTasks)
	if !ok || required != binding.TotalTasks {
		return intelligencePoint{}, intelligenceInvalid("required tasks")
	}
	coverage, ok := intelligenceCount(summary.Coverage)
	if !ok || coverage > required {
		return intelligencePoint{}, intelligenceInvalid("coverage")
	}
	status := "provisional"
	switch coverage {
	case 0:
		status = "missing_current_result"
	case required:
		status = "complete"
	}
	if summary.ScoreStatus != status {
		return intelligencePoint{}, intelligenceInvalid("score status differs from coverage")
	}
	// 无覆盖时分数必须为 null；有覆盖时 0 分也是有效成绩。
	if (coverage == 0) != (summary.Score == nil) || (summary.Score != nil && (*summary.Score < 0 || *summary.Score > intelligenceScoreMax)) {
		return intelligencePoint{}, intelligenceInvalid("score")
	}
	point := intelligencePoint{Model: selection.Model, Effort: selection.Effort, Score: summary.Score, Coverage: &coverage, ScoreStatus: status}
	if summary.GradedRunsCount != nil {
		// 每题最多计最近 3 次有效判分。
		runs, ok := intelligenceCount(summary.GradedRunsCount)
		if !ok || runs < coverage || runs > coverage*3 {
			return intelligencePoint{}, intelligenceInvalid("graded runs")
		}
		point.GradedRuns = &runs
	}
	intelligenceStatistics(&point, summary)
	return point, nil
}

// intelligenceStatistics 只用源站公开聚合：均值 = 已知样本合计 ÷ 已知样本数，与源站页面算法一致；
// 部分样本缺失时仍给已知样本均值并带样本数。统计异常只让费用／耗时缺失，不影响分数。
func intelligenceStatistics(point *intelligencePoint, summary intelligenceSummary) {
	point.StatisticsState = summary.StatisticsState
	switch summary.StatisticsState {
	case "available":
	case "not_provided":
		return
	default:
		point.StatisticsState = "invalid_source_statistics"
		return
	}
	stats := summary.Statistics
	if stats == nil || stats.CostBasis != intelligenceCostBasis {
		point.StatisticsState = "invalid_source_statistics"
		return
	}
	samples, ok := intelligenceCount(stats.SampleCount)
	price, priceKnown, priceOK := intelligenceMean(stats.Metrics.USD, samples)
	elapsed, elapsedKnown, elapsedOK := intelligenceMean(stats.Metrics.ElapsedMS, samples)
	if !ok || !priceOK || !elapsedOK {
		point.StatisticsState = "invalid_source_statistics"
		return
	}
	if elapsed != nil {
		minutes := *elapsed / 60000
		elapsed = &minutes
	}
	point.StatisticsCount, point.AveragePriceUSD, point.PriceSamples, point.AverageMinutes, point.DurationSamples = &samples, price, priceKnown, elapsed, elapsedKnown
}

func intelligenceMean(metric *intelligenceMetric, samples int) (*float64, *int, bool) {
	if metric == nil || (metric.KnownTotal != nil && *metric.KnownTotal < 0) {
		return nil, nil, false
	}
	known, knownOK := intelligenceCount(metric.KnownSamples)
	missing, missingOK := intelligenceCount(metric.MissingSamples)
	if !knownOK || !missingOK || known+missing != samples {
		return nil, nil, false
	}
	if known == 0 {
		return nil, &known, metric.KnownTotal == nil
	}
	if metric.KnownTotal == nil {
		return nil, nil, false
	}
	mean := *metric.KnownTotal / float64(known)
	return &mean, &known, true
}

// intelligenceCount 把 JSON 数字转成非负整数计数；缺失、小数或负数视为无效。
func intelligenceCount(value *float64) (int, bool) {
	if value == nil || *value < 0 || *value > 1e9 || *value != math.Trunc(*value) {
		return 0, false
	}
	return int(*value), true
}
