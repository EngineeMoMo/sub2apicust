package service

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCustomPaymentQueryGateConcurrentAndCooldown(t *testing.T) {
	gate := &customPaymentQueryGate{}
	now := time.Now()
	var allowed atomic.Int64
	var workers sync.WaitGroup
	for workerIndex := 0; workerIndex < 32; workerIndex++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if gate.begin(1, now) {
				allowed.Add(1)
			}
		}()
	}
	workers.Wait()
	require.EqualValues(t, 1, allowed.Load())
	require.False(t, gate.begin(1, now.Add(time.Hour)))
	gate.finish(1, now)
	require.False(t, gate.begin(1, now.Add(4*time.Second)))
	require.True(t, gate.begin(1, now.Add(5*time.Second)))
	require.True(t, gate.begin(2, now))
}

func TestCustomPaymentQueryGateMemoryBoundAndReclaim(t *testing.T) {
	gate := &customPaymentQueryGate{}
	now := time.Now()
	for orderID := int64(1); orderID <= customPaymentPublicQueryMaxOrders; orderID++ {
		require.True(t, gate.begin(orderID, now))
	}
	require.False(t, gate.begin(customPaymentPublicQueryMaxOrders+1, now))
	gate.finish(1, now)
	require.False(t, gate.begin(customPaymentPublicQueryMaxOrders+1, now.Add(time.Second)))
	require.True(t, gate.begin(customPaymentPublicQueryMaxOrders+1, now.Add(5*time.Second)))
	require.Len(t, gate.orders, customPaymentPublicQueryMaxOrders)
}
