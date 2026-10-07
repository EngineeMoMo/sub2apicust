//go:build unit

package repository

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCustomSubscriptionOpeningMigrationPreservesUsageAndExpiry(t *testing.T) {
	dsn := os.Getenv("DEDICATED_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires an isolated PostgreSQL database")
	}
	u, err := url.Parse(dsn)
	require.NoError(t, err)
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := fmt.Sprintf("subscription_opening_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, err)
		_ = admin.Close()
	})
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := sql.Open("postgres", u.String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`CREATE TABLE user_subscriptions (
		id BIGINT PRIMARY KEY, starts_at TIMESTAMPTZ NOT NULL, expires_at TIMESTAMPTZ NOT NULL,
		deleted_at TIMESTAMPTZ, daily_window_start TIMESTAMPTZ, weekly_window_start TIMESTAMPTZ,
		monthly_window_start TIMESTAMPTZ, weekly_usage_usd NUMERIC DEFAULT 3, monthly_usage_usd NUMERIC DEFAULT 4)`)
	require.NoError(t, err)
	start := time.Now().UTC().Truncate(time.Second).Add(-10 * 24 * time.Hour)
	expiry := start.Add(60 * 24 * time.Hour)
	_, err = db.Exec(`INSERT INTO user_subscriptions (id, starts_at, expires_at, weekly_window_start, monthly_window_start)
		VALUES (1,$1,$2,NULL,NULL),(2,$1,$2,$3,$3),(3,$1,$2,$1,$1),(4,$1,$4,NULL,NULL)`,
		start, expiry, start.Add(2*24*time.Hour), start.Add(24*time.Hour))
	require.NoError(t, err)
	migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "244_custom_subscription_opening_windows.sql"))
	require.NoError(t, err)
	for attempt := 0; attempt < 2; attempt++ {
		_, err = db.Exec(string(migration))
		require.NoError(t, err)
		for _, id := range []int{1, 2, 3} {
			var weekly, monthly, expires time.Time
			var usedWeekly, usedMonthly float64
			err = db.QueryRow(`SELECT weekly_window_start, monthly_window_start, expires_at, weekly_usage_usd, monthly_usage_usd FROM user_subscriptions WHERE id=$1`, id).
				Scan(&weekly, &monthly, &expires, &usedWeekly, &usedMonthly)
			require.NoError(t, err)
			expectedWeekly := start.Add(7 * 24 * time.Hour)
			if id == 3 {
				expectedWeekly = start
			}
			require.True(t, expectedWeekly.Equal(weekly))
			require.True(t, start.Equal(monthly))
			require.True(t, expiry.Equal(expires))
			require.Equal(t, 3.0, usedWeekly)
			require.Equal(t, 4.0, usedMonthly)
		}
		var untouched bool
		require.NoError(t, db.QueryRow(`SELECT weekly_window_start IS NULL AND monthly_window_start IS NULL FROM user_subscriptions WHERE id=4`).Scan(&untouched))
		require.True(t, untouched)
	}
}
