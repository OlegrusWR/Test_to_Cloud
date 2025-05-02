package ratelimiter

import (
	"sync"
	"time"
)

type TokenBucket struct{
	mu sync.Mutex
	tokens int
	capacity int
	rate time.Duration
	lastrefil time.Time
}

func NewBucket(capacity int, rate time.Duration) *TokenBucket{
	return &TokenBucket{
		tokens: capacity,
		capacity: capacity,
		lastrefil: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	past := now.Sub(tb.lastrefil)
	tokkensToAdd := int(past / tb.rate)

	if tokkensToAdd > 0{
		tb.tokens = min(tb.tokens+tokkensToAdd, tb.capacity)
	}
	if tb.tokens > 0{
		tb.tokens--
		return true
	}
	return false
}