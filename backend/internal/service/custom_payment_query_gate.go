package service

import (
	"sync"
	"time"
)

const customPaymentPublicQueryInterval = 5 * time.Second
const customPaymentPublicQueryMaxOrders = 4096

type customPaymentQuerySlot struct {
	inFlight  bool
	nextQuery time.Time
}

type customPaymentQueryGate struct {
	mutex  sync.Mutex
	orders map[int64]customPaymentQuerySlot
}

func (gate *customPaymentQueryGate) begin(orderID int64, now time.Time) bool {
	gate.mutex.Lock()
	defer gate.mutex.Unlock()
	if gate.orders == nil {
		gate.orders = make(map[int64]customPaymentQuerySlot)
	}
	current, exists := gate.orders[orderID]
	if exists && (current.inFlight || now.Before(current.nextQuery)) {
		return false
	}
	if !exists && len(gate.orders) >= customPaymentPublicQueryMaxOrders {
		for existingID, slot := range gate.orders {
			if !slot.inFlight && !now.Before(slot.nextQuery) {
				delete(gate.orders, existingID)
			}
		}
		if len(gate.orders) >= customPaymentPublicQueryMaxOrders {
			return false
		}
	}
	gate.orders[orderID] = customPaymentQuerySlot{inFlight: true}
	return true
}

func (gate *customPaymentQueryGate) finish(orderID int64, now time.Time) {
	gate.mutex.Lock()
	defer gate.mutex.Unlock()
	gate.orders[orderID] = customPaymentQuerySlot{nextQuery: now.Add(customPaymentPublicQueryInterval)}
}
