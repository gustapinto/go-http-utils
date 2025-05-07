package bucket

import (
	"math"
	"sync"
	"time"
)

type Map struct {
	capacity   int64
	refillRate int64
	tokens     map[string]int64
	lastRefill map[string]time.Time
	mu         sync.Mutex
}

func NewKeyedBucket(capacity, refillRate int64) *Map {
	return &Map{
		mu:         sync.Mutex{},
		tokens:     map[string]int64{},
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: map[string]time.Time{},
	}
}

func (ktb *Map) refill(key string) {
	tokens, exists := ktb.tokens[key]
	if !exists {
		ktb.tokens[key] = ktb.capacity
		ktb.lastRefill[key] = time.Now()
		return
	}

	now := time.Now()
	lastRefill := ktb.lastRefill[key]

	timeSinceLastRefill := now.Sub(lastRefill).Seconds()
	tokensToRefill := ktb.refillRate * int64(math.Floor(timeSinceLastRefill))

	tokens += tokensToRefill
	if tokens > ktb.capacity {
		tokens = ktb.capacity
	}

	ktb.tokens[key] = tokens
	ktb.lastRefill[key] = now
}

func (ktb *Map) Allow(key string) bool {
	return ktb.AllowN(key, 1)
}

func (ktb *Map) AllowN(key string, requestedTokens int64) bool {
	ktb.mu.Lock()
	defer ktb.mu.Unlock()

	ktb.refill(key)
	if ktb.tokens[key] <= 0 {
		return false
	}

	ktb.tokens[key] -= requestedTokens
	return true
}
