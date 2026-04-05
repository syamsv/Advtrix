package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	count    int
	resetAt  time.Time
}

// Limiter is an in-memory per-key rate limiter.
type Limiter struct {
	mu      sync.Mutex
	entries map[string]*entry
	limit   int
	window  time.Duration
}

// New creates a limiter allowing `limit` requests per `window` per key.
func New(limit int, window time.Duration) *Limiter {
	l := &Limiter{
		entries: make(map[string]*entry),
		limit:   limit,
		window:  window,
	}
	go l.cleanup()
	return l
}

// Allow returns true if the key has not exceeded the rate limit.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, exists := l.entries[key]
	if !exists || now.After(e.resetAt) {
		l.entries[key] = &entry{count: 1, resetAt: now.Add(l.window)}
		return true
	}

	if e.count >= l.limit {
		return false
	}

	e.count++
	return true
}

func (l *Limiter) cleanup() {
	ticker := time.NewTicker(l.window)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for k, e := range l.entries {
			if now.After(e.resetAt) {
				delete(l.entries, k)
			}
		}
		l.mu.Unlock()
	}
}
