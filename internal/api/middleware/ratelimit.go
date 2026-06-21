package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateLimitEntry struct {
	lastSeen time.Time
	requests int
}

var (
	clients = make(map[string]*rateLimitEntry)
	mu      sync.Mutex
)

func init() {
	// Background cleaner goroutine to evict stale client tracking entries
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			mu.Lock()
			now := time.Now()
			for ip, entry := range clients {
				if now.Sub(entry.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// RateLimitMiddleware limits each IP to a maximum of 100 requests per minute.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		// Get real IP from reverse proxies if available
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			ip = strings.TrimSpace(parts[0])
		} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
			ip = xri
		} else {
			// Strip port from RemoteAddr
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				ip = ip[:idx]
			}
		}

		mu.Lock()
		entry, exists := clients[ip]
		now := time.Now()

		if !exists {
			entry = &rateLimitEntry{lastSeen: now, requests: 0}
			clients[ip] = entry
		}

		// Reset count if a minute has passed since last check
		if now.Sub(entry.lastSeen) > 1*time.Minute {
			entry.requests = 0
			entry.lastSeen = now
		}

		if entry.requests >= 100 {
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error": {"code": "TOO_MANY_REQUESTS", "message": "Rate limit exceeded. Please wait a minute."}}`))
			return
		}

		entry.requests++
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
