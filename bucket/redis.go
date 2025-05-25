package bucket

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	_tokensKeyFormat      = "BUCKET/%s/TOKENS"
	_lastRefillKeyFormat  = "BUCKET/%s/LAST_REFILL"
	_lastRefillTimeFormat = time.RFC3339
)

type Redis struct {
	capacity   int64
	refillRate int64
	mu         sync.Mutex
	client     *redis.Client
}

func NewRedis(capacity, refillRate int64, client *redis.Client) *Redis {
	return &Redis{
		mu:         sync.Mutex{},
		capacity:   capacity,
		refillRate: refillRate,
		client:     client,
	}
}

func (rtb *Redis) setTokens(key string, tokens int64) error {
	tokensKey := fmt.Sprintf(_tokensKeyFormat, key)

	return rtb.client.Set(context.Background(), tokensKey, tokens, 0).Err()
}

func (rtb *Redis) getOrSetTokens(key string) (int64, error) {
	tokensKey := fmt.Sprintf(_tokensKeyFormat, key)
	tokensRes := rtb.client.Get(context.Background(), tokensKey)
	if tokensRes.Err() != nil {
		if errors.Is(tokensRes.Err(), redis.Nil) {
			return rtb.capacity, rtb.setTokens(key, rtb.capacity)
		}
	}

	var tokens int64
	if err := tokensRes.Scan(&tokens); err != nil {
		return 0, err
	}

	return tokens, nil
}

func (rtb *Redis) getLastRefill(key string) (time.Time, error) {
	refillKey := fmt.Sprintf(_lastRefillKeyFormat, key)
	refillRes := rtb.client.Get(context.Background(), refillKey)
	if refillRes.Err() != nil {
		if errors.Is(refillRes.Err(), redis.Nil) {
			return time.Now(), nil
		}

		return time.Time{}, refillRes.Err()
	}

	lastRefill, err := time.Parse(_lastRefillTimeFormat, refillRes.Val())
	if err != nil {
		lastRefill = time.Now()
	}

	return lastRefill, nil
}

func (rtb *Redis) setLastRefill(key string, lastRefill time.Time) error {
	refillKey := fmt.Sprintf(_lastRefillKeyFormat, key)
	lastRefillVal := lastRefill.Format(_lastRefillTimeFormat)

	return rtb.client.Set(context.Background(), refillKey, lastRefillVal, 0).Err()
}

func (rtb *Redis) refill(key string) error {
	now := time.Now()

	tokens, err := rtb.getOrSetTokens(key)
	if err != nil {
		return err
	}

	lastRefill, err := rtb.getLastRefill(key)
	if err != nil {
		return err
	}

	timeSinceLastRefill := now.Sub(lastRefill).Seconds()
	tokensToRefill := rtb.refillRate * int64(math.Floor(timeSinceLastRefill))

	tokens += tokensToRefill
	if tokens > rtb.capacity {
		tokens = rtb.capacity
	}

	if err := rtb.setLastRefill(key, now); err != nil {
		return err
	}

	return rtb.setTokens(key, tokens)
}

func (rtb *Redis) Allow(key string) (bool, error) {
	return rtb.AllowN(key, 1)
}

func (rtb *Redis) AllowN(key string, requestedTokens int64) (bool, error) {
	rtb.mu.Lock()
	defer rtb.mu.Unlock()

	if err := rtb.refill(key); err != nil {
		return false, err
	}

	tokens, err := rtb.getOrSetTokens(key)
	if err != nil {
		return false, err
	}

	if tokens <= 0 {
		return false, nil
	}

	newTokens := tokens - requestedTokens
	if err := rtb.setTokens(key, newTokens); err != nil {
		return false, err
	}

	return true, nil
}
