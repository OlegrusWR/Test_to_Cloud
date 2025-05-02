package ratelimiter

import (
	"sync"
	"time"
)

type TokenBucket struct{
	mu sync.Mutex
	tokens int
	capacity int
	rate time.Duration // lowercase
	lastrefil time.Time
}

func NewBucket(capacity int, rate time.Duration) *TokenBucket{
	if rate <= 0 {
        rate = time.Second 
    }
	return &TokenBucket{
		tokens: capacity,
		capacity: capacity,
		lastrefil: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()

    effectiveRate := tb.rate
    if effectiveRate <= 0 {
        effectiveRate = time.Second
        tb.rate = effectiveRate 
    }

    now := time.Now()
    elapsed := now.Sub(tb.lastrefil)
    
    if effectiveRate > 0 {
        tokensToAdd := int(elapsed / effectiveRate)
        if tokensToAdd > 0 {
            tb.tokens = min(tb.capacity, tb.tokens + tokensToAdd)
            tb.lastrefil = now
        }
    }

    if tb.tokens > 0 {
        tb.tokens--
        return true
    }
    return false
}