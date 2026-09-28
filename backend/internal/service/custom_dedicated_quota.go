package service

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func customDedicatedNumber(raw any) (float64, bool) {
	if raw == nil {
		return 0, false
	}
	number, err := strconv.ParseFloat(fmt.Sprint(raw), 64)
	return number, err == nil && !math.IsNaN(number) && !math.IsInf(number, 0) && number >= 0
}

func customDedicatedAnthropicWindow(extra map[string]any, utilizationKey, resetKey string) *UsageProgress {
	utilization, valid := customDedicatedNumber(extra[utilizationKey])
	if !valid {
		return nil
	}
	window := &UsageProgress{Utilization: utilization * 100}
	if reset, valid := customDedicatedNumber(extra[resetKey]); valid && reset > 0 && reset < 253402300800 {
		stamp := time.Unix(int64(reset), 0)
		window.ResetsAt = &stamp
	}
	return window
}

func customDedicatedPassiveUsage(account *Account, now time.Time) *UsageInfo {
	usage := &UsageInfo{Source: "passive"}
	if account == nil {
		return usage
	}
	timestampKey := "passive_usage_sampled_at"
	if account.Platform == PlatformOpenAI {
		timestampKey = "codex_usage_updated_at"
		if _, valid := customDedicatedNumber(account.Extra["codex_5h_used_percent"]); valid {
			usage.FiveHour = buildCodexUsageProgressFromExtra(account.Extra, "5h", now)
		}
		if _, valid := customDedicatedNumber(account.Extra["codex_7d_used_percent"]); valid {
			usage.SevenDay = buildCodexUsageProgressFromExtra(account.Extra, "7d", now)
		}
	} else {
		if utilization, valid := customDedicatedNumber(account.Extra["session_window_utilization"]); valid {
			usage.FiveHour = &UsageProgress{Utilization: utilization * 100, ResetsAt: account.SessionWindowEnd}
		}
		usage.SevenDay = customDedicatedAnthropicWindow(account.Extra, "passive_usage_7d_utilization", "passive_usage_7d_reset")
		usage.SevenDayFable = customDedicatedAnthropicWindow(account.Extra, "passive_usage_7d_oi_utilization", "passive_usage_7d_oi_reset")
	}
	if raw, exists := account.Extra[timestampKey]; exists && raw != nil {
		if stamp, err := time.Parse(time.RFC3339, fmt.Sprint(raw)); err == nil {
			usage.UpdatedAt = &stamp
		}
	}
	return usage
}
