package service

import (
	"sync"
	"time"
)

// translationCacheTTL controls how long a message's translation outcome is
// retained, so repeat triggers (multiple reactors, multiple trigger emojis,
// re-adding a reaction) reuse the cached result instead of calling the
// translation API again.
const translationCacheTTL = 1 * time.Hour

// cachedTranslation holds the completed translation decision for a single
// (message, target language) pair.
type cachedTranslation struct {
	skip             bool // translation was skipped (same language, empty, or identical result)
	translatedText   string
	detectedSource   string
	channelDelivered bool // a channel reply has already been posted for this key
}

// translationCache is a thread-safe in-memory TTL cache keyed by
// "<messageID>|<normalized target lang>".
type translationCache struct {
	mu    sync.Mutex
	items map[string]*translationCacheItem
}

type translationCacheItem struct {
	value     cachedTranslation
	expiresAt time.Time
}

func newTranslationCache() *translationCache {
	c := &translationCache{items: make(map[string]*translationCacheItem)}
	go c.cleanup()
	return c
}

func (c *translationCache) get(key string) (cachedTranslation, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.items[key]
	if !ok || time.Now().After(item.expiresAt) {
		return cachedTranslation{}, false
	}
	return item.value, true
}

func (c *translationCache) set(key string, value cachedTranslation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = &translationCacheItem{value: value, expiresAt: time.Now().Add(translationCacheTTL)}
}

func (c *translationCache) cleanup() {
	ticker := time.NewTicker(translationCacheTTL)
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
