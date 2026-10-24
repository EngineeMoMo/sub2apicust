package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllPlatformsIncludesEveryConcretePlatform(t *testing.T) {
	require.ElementsMatch(t, []string{
		"anthropic",
		"openai",
		"gemini",
		"antigravity",
		"grok",
		"kimi",
		"zhipu",
		"deepseek",
		"minimax",
		"opencode_go",
		"typesafe",
		"command_code",
		"cline",
		"kiro", // [CUSTOM] 错误规则须覆盖已登记的 Kiro 原生平台。
	}, AllPlatforms())
}
