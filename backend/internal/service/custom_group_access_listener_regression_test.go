//go:build unit

package service

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCustomGroupAccessPostgresListenerBacklogRecoveryAndStop(t *testing.T) {
	dsn := os.Getenv("CUSTOM_GROUP_ACCESS_TEST_DSN")
	if dsn == "" {
		t.Skip("CUSTOM_GROUP_ACCESS_TEST_DSN 未配置，未验证真实通知积压与退出")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	database, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var databaseName string
	require.NoError(t, database.QueryRowContext(ctx, "SELECT current_database()").Scan(&databaseName))
	require.Equal(t, "group_access_test", databaseName)
	application := fmt.Sprintf("group_access_backlog_%d", time.Now().UnixNano())
	access := &CustomGroupAccessService{cache: newCustomGroupAccessCache()}
	stop := access.startChangeListener(dsn + " application_name=" + application)
	t.Cleanup(stop)
	healthy := func() bool {
		access.cache.mu.Lock()
		defer access.cache.mu.Unlock()
		return time.Now().Before(access.cache.healthyUntil)
	}
	require.Eventually(t, healthy, 5*time.Second, 10*time.Millisecond)
	for attempt := 0; attempt < 3; attempt++ {
		access.cache.mu.Lock()
		_, notifyErr := database.ExecContext(ctx, `SELECT pg_notify('custom_group_access_changed', json_build_object('kind','group','id',series)::text) FROM generate_series(1,1000) AS series`)
		time.Sleep(1100 * time.Millisecond)
		access.cache.mu.Unlock()
		require.NoError(t, notifyErr)
		time.Sleep(4 * time.Second)
		require.True(t, healthy(), "通知积压结束后心跳必须继续，不能与pq转发队列互等")
	}
	_, err = database.ExecContext(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=current_database() AND application_name=$1", application)
	require.NoError(t, err)
	require.Eventually(t, healthy, 6*time.Second, 20*time.Millisecond)
	require.Eventually(t, func() bool {
		var count int
		err := database.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND application_name=$1", application).Scan(&count)
		return err == nil && count == 1
	}, 6*time.Second, 20*time.Millisecond)
	stopped := make(chan struct{})
	go func() { stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		t.Fatal("通知监听退出超时")
	}
	require.False(t, healthy())
	stop()
}

func TestCustomGroupAccessListenerConnectionDeadlineAndClose(t *testing.T) {
	for _, mode := range []string{"timeout", "stop"} {
		t.Run(mode, func(t *testing.T) {
			connection, peer := net.Pipe()
			t.Cleanup(func() { _ = connection.Close(); _ = peer.Close() })
			dialer := &customGroupAccessDialer{conn: connection}
			done := make(chan error, 1)
			go func() { _, err := connection.Read(make([]byte, 1)); done <- err }()
			if mode == "timeout" {
				dialer.deadline(time.Now().Add(30 * time.Millisecond))
			} else {
				dialer.close()
			}
			select {
			case err := <-done:
				require.Error(t, err)
			case <-time.After(time.Second):
				t.Fatal("失联连接必须可超时和退出")
			}
		})
	}
}
