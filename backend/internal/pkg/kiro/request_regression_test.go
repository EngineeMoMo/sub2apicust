package kiro

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKiroPreservesUnknownDatedModelIDs(t *testing.T) {
	for _, model := range []string{"gpt-5.3-20260101", "custom-sonnet-20261001", "vendor.model-20261023"} {
		resolved, err := ResolveModel(model)
		require.NoError(t, err)
		require.Equal(t, model, resolved)
	}
	for _, model := range []string{"gpt-5[1m]", "custom[1m]", "claude-opus-4-6[1m]"} {
		_, err := ResolveModel(model)
		require.Error(t, err)
	}
	resolved, err := ResolveModel("claude-sonnet-4-6-20260101[1m]")
	require.NoError(t, err)
	require.Equal(t, "claude-sonnet-4.6-1m", resolved)
}

func TestKiroRejectsNonTextSystem(t *testing.T) {
	_, err := BuildPayload([]byte(`{"system":[{"type":"image","source":{"type":"base64","data":"aGk="}}],"messages":[{"role":"user","content":"hi"}]}`), "claude-sonnet-4-6", "profile", "conversation")
	require.ErrorContains(t, err, "system 仅支持文本")
}

func TestKiroMergesAdjacentAssistantToolBlocks(t *testing.T) {
	payload, err := BuildPayload([]byte(`{"messages":[{"role":"user","content":"hi"},{"role":"assistant","content":[{"type":"tool_use","id":"one","name":"first","input":{}}]},{"role":"assistant","content":[{"type":"tool_use","id":"two","name":"second","input":{}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"one","content":"1"},{"type":"tool_result","tool_use_id":"two","content":"2"}]}]}`), "claude-sonnet-4-6", "profile", "conversation")
	require.NoError(t, err)
	require.Len(t, payload.ConversationState.History, 2)
	require.Len(t, payload.ConversationState.History[1].Assistant.Tools, 2)
	require.Len(t, payload.ConversationState.Current["userInputMessage"].Context.Results, 2)
}
