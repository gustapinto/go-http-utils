package bucket

import (
	"math"
	"sync"
	"time"
)

type Simple struct {
	tokens     int64
	capacity   int64
	refillRate int64
	mu         sync.Mutex
	lastRefill time.Time
}

func NewTokenBucket(capacity, refillRate int64) *Simple {
	return &Simple{
		mu:         sync.Mutex{},
		tokens:     capacity, // Starts full
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *Simple) refill() {
	now := time.Now()
	timeSinceLastRefill := now.Sub(tb.lastRefill).Seconds()
	tokensToRefill := tb.refillRate * int64(math.Floor(timeSinceLastRefill))

	if tokensToRefill > tb.capacity {
		tokensToRefill = tb.capacity
	}

	tb.tokens += tokensToRefill
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now
}

func (tb *Simple) Allow() bool {
	return tb.AllowN(1)
}

func (tb *Simple) AllowN(requestedTokens int64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	if tb.tokens <= 0 {
		return false
	}

	tb.tokens -= requestedTokens
	return true
}
