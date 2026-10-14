//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthIPNetworkTargetNormalization(t *testing.T) {
	for raw, expected := range map[string]string{
		" 210.230.206.* ":           "210.230.206.0/24",
		"210.230.*.*":               "210.230.0.0/16",
		"210.*.*.*":                 "210.0.0.0/8",
		"210.230.206.27/24":         "210.230.206.0/24",
		"210.230.206.27/32":         "210.230.206.27",
		"210.230.206.27":            "210.230.206.27",
		"::ffff:210.230.206.27/120": "210.230.206.0/24",
		"::ffff:210.230.206.27/128": "210.230.206.27",
		"2001:0db8:0001::123/48":    "2001:db8:1::/48",
		"2001:0db8::123/128":        "2001:db8::123",
	} {
		actual, err := NormalizeAuthIPTarget(raw)
		require.NoError(t, err, raw)
		require.Equal(t, expected, actual, raw)
	}
	for _, raw := range []string{"*", "*.*.*.*", "210.*.206.*", "210.230.206.2*", "210.230.*", "210.230.999.*", "210.230.206.0/33", "0.0.0.0/0", "::/0", "::ffff:210.230.206.0/96", "::ffff:210.230.206.0/80", "fe80::1%eth0", "2001:db8::*"} {
		_, err := NormalizeAuthIPTarget(raw)
		require.Error(t, err, raw)
	}
	_, err := NormalizeAuthIP("210.230.206.*")
	require.Error(t, err, "真实请求来源仍只接受单IP")
}

func TestAuthIPAggregationTargets(t *testing.T) {
	config := AuthIPBlacklistConfig{AggregationMode: AuthIPAggregationNetwork}
	for raw, expected := range map[string]string{
		"210.230.206.27":        "210.230.206.0/24",
		"210.230.206.23":        "210.230.206.0/24",
		"210.230.207.23":        "210.230.207.0/24",
		"::ffff:210.230.206.27": "210.230.206.0/24",
		"2001:db8:1:2::abcd":    "2001:db8:1:2::/64",
	} {
		actual, err := AuthIPAggregationTarget(raw, config)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	}
	actual, err := AuthIPAggregationTarget("210.230.206.27", AuthIPBlacklistConfig{})
	require.NoError(t, err)
	require.Equal(t, "210.230.206.27", actual)
	config.NetworkIPv4Prefix = 33
	_, err = AuthIPAggregationTarget("210.230.206.27", config)
	require.Error(t, err)
}
