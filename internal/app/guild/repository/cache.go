package repository

import (
	"sync"
	"time"
)

// GuildCache is a thread-safe in-memory TTL cache for guild configuration data.
// It is used to avoid hitting the database on every Discord event.
type GuildCache[T any] struct {
	mu    sync.RWMutex
	items map[int64]*entry[T]
	ttl   time.Duration
}

type entry[T any] struct {
	value     T
	expiresAt time.Time
}

// NewGuildCache creates a new GuildCache with the given TTL duration.
func NewGuildCache[T any](ttl time.Duration) *GuildCache[T] {
	c := &GuildCache[T]{
		items: make(map[int64]*entry[T]),
		ttl:   ttl,
	}
	// Background cleanup goroutine
	go c.cleanup()
	return c
}

// Get returns the cached value for the given guild ID, along with a hit indicator.
func (c *GuildCache[T]) Get(guildID int64) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.items[guildID]
	if !ok || time.Now().After(e.expiresAt) {
		var zero T
		return zero, false
	}
	return e.value, true
}

// Set stores a value in the cache with the configured TTL.
func (c *GuildCache[T]) Set(guildID int64, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[guildID] = &entry[T]{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes the cached value for the given guild ID immediately.
func (c *GuildCache[T]) Invalidate(guildID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, guildID)
}

// cleanup periodically evicts expired entries to prevent memory leaks.
func (c *GuildCache[T]) cleanup() {
	ticker := time.NewTicker(c.ttl * 2)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, e := range c.items {
			if now.After(e.expiresAt) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}
