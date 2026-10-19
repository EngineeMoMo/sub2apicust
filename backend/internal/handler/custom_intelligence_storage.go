package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const intelligenceSnapshotLimit = 1 << 20

type intelligenceStoredSnapshot struct {
	Version       int                  `json:"version"`
	ScoreVersion  string               `json:"score_version"`
	PolicyVersion string               `json:"scoring_policy_version"`
	CostBasis     string               `json:"cost_basis"`
	TaskSet       string               `json:"task_set_sha256"`
	Snapshot      intelligenceSnapshot `json:"snapshot"`
}

func (cache *intelligenceCache) configureStorage(dataDir string) {
	if strings.TrimSpace(dataDir) == "" {
		return
	}
	cache.storagePath = filepath.Join(dataDir, "intelligence-cache", "last-success.json")
	if err := cache.loadSnapshot(); err != nil && !os.IsNotExist(err) {
		slog.Warn("intelligence snapshot load failed", "error", err)
	}
}

func (cache *intelligenceCache) loadSnapshot() error {
	file, err := os.Open(cache.storagePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	raw, err := io.ReadAll(io.LimitReader(file, intelligenceSnapshotLimit+1))
	if err != nil {
		return err
	}
	if len(raw) > intelligenceSnapshotLimit {
		return intelligenceInvalid("stored snapshot exceeds size limit")
	}
	var stored intelligenceStoredSnapshot
	if err := json.Unmarshal(raw, &stored); err != nil {
		return err
	}
	if err := cache.validateStoredSnapshot(stored); err != nil {
		return err
	}
	snapshot := stored.Snapshot
	snapshot.taskSet = stored.TaskSet
	for index := range snapshot.Points {
		if snapshot.Points[index].State == "ok" {
			snapshot.Points[index].State = "retained"
		}
	}
	cache.snapshot, cache.failed = &snapshot, true
	return nil
}

func (cache *intelligenceCache) saveSnapshot(snapshot *intelligenceSnapshot) error {
	if cache.storagePath == "" {
		return nil
	}
	stored := intelligenceStoredSnapshot{Version: 1, ScoreVersion: intelligenceScoreVersion, PolicyVersion: intelligencePolicyVersion,
		CostBasis: intelligenceCostBasis, TaskSet: snapshot.taskSet, Snapshot: *snapshot}
	if err := cache.validateStoredSnapshot(stored); err != nil {
		return err
	}
	raw, err := json.Marshal(stored)
	if err != nil {
		return err
	}
	if len(raw) > intelligenceSnapshotLimit {
		return intelligenceInvalid("stored snapshot exceeds size limit")
	}
	dir := filepath.Dir(cache.storagePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".last-success-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()
	if _, err := file.Write(raw); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), cache.storagePath)
}

func (cache *intelligenceCache) validateStoredSnapshot(stored intelligenceStoredSnapshot) error {
	snapshot := stored.Snapshot
	checkedAt, err := time.Parse(time.RFC3339, snapshot.CheckedAt)
	if stored.Version != 1 || stored.ScoreVersion != intelligenceScoreVersion || stored.PolicyVersion != intelligencePolicyVersion || stored.CostBasis != intelligenceCostBasis || err != nil || checkedAt.After(cache.now().Add(5*time.Minute)) {
		return intelligenceInvalid("stored snapshot version or date")
	}
	binding := intelligenceBinding{Enabled: true, ProductionReadVerified: true, Benchmark: snapshot.Benchmark,
		CatalogVersion: snapshot.CatalogVersion, TaskSetSHA256: stored.TaskSet, TotalTasks: snapshot.RequiredTasks,
		ScoreVersion: intelligenceScoreVersion, ScoreMax: snapshot.ScoreMax}
	for _, point := range snapshot.Points {
		binding.ModelEfforts = append(binding.ModelEfforts, intelligenceSelection{Model: point.Model, Effort: point.Effort})
	}
	if err := validateIntelligenceBinding(binding); err != nil {
		return err
	}
	available := false
	for index, point := range snapshot.Points {
		if point.State == "unavailable" {
			if point.Score != nil || point.Coverage != nil || point.GradedRuns != nil || point.AverageMinutes != nil || point.DurationSamples != nil || point.AveragePriceUSD != nil || point.PriceSamples != nil || point.StatisticsCount != nil || point.ReadAt != "" || point.ScoreStatus != "" || point.StatisticsState != "" {
				return intelligenceInvalid("stored unavailable point has metrics")
			}
			continue
		}
		if point.State != "ok" && point.State != "retained" {
			return intelligenceInvalid("stored point state")
		}
		readAt, err := time.Parse(time.RFC3339, point.ReadAt)
		if err != nil || readAt.After(checkedAt) {
			return intelligenceInvalid("stored point date")
		}
		summary := intelligenceSummary{Schema: intelligenceSummarySchema, Benchmark: binding.Benchmark,
			CatalogVersion: binding.CatalogVersion, TaskSetSHA256: binding.TaskSetSHA256,
			ScoreVersion: intelligenceScoreVersion, ScoreMax: intelligenceScoreMax, PolicyVersion: intelligencePolicyVersion,
			Model: point.Model, Effort: point.Effort, Score: point.Score, ScoreStatus: point.ScoreStatus,
			Coverage: intelligenceStoredCount(point.Coverage), RequiredTasks: intelligenceStoredCount(&snapshot.RequiredTasks),
			GradedRunsCount: intelligenceStoredCount(point.GradedRuns)}
		if _, err := intelligenceSummaryPoint(summary, binding, binding.ModelEfforts[index]); err != nil {
			return err
		}
		if point.StatisticsCount != nil && (*point.StatisticsCount < 0 || *point.StatisticsCount > 1e9) {
			return intelligenceInvalid("stored statistics count")
		}
		for _, metric := range []struct {
			average *float64
			samples *int
		}{{point.AverageMinutes, point.DurationSamples}, {point.AveragePriceUSD, point.PriceSamples}} {
			if metric.samples != nil && (*metric.samples < 0 || point.StatisticsCount == nil || *metric.samples > *point.StatisticsCount) {
				return intelligenceInvalid("stored metric samples")
			}
			if metric.average != nil && (math.IsNaN(*metric.average) || math.IsInf(*metric.average, 0) || *metric.average < 0 || metric.samples == nil || *metric.samples == 0) {
				return intelligenceInvalid("stored metric average")
			}
			if metric.samples != nil && (*metric.samples > 0) != (metric.average != nil) {
				return intelligenceInvalid("stored metric missing average")
			}
		}
		switch point.StatisticsState {
		case "available":
			if point.StatisticsCount == nil || point.PriceSamples == nil || point.DurationSamples == nil {
				return intelligenceInvalid("stored statistics missing counts")
			}
		case "not_provided", "invalid_source_statistics":
			if point.StatisticsCount != nil || point.PriceSamples != nil || point.DurationSamples != nil || point.AveragePriceUSD != nil || point.AverageMinutes != nil {
				return intelligenceInvalid("stored unavailable statistics have metrics")
			}
		default:
			return intelligenceInvalid("stored statistics state")
		}
		available = true
	}
	if !available {
		return intelligenceInvalid("stored snapshot has no usable points")
	}
	return nil
}

func intelligenceStoredCount(value *int) *float64 {
	if value == nil {
		return nil
	}
	number := float64(*value)
	return &number
}
