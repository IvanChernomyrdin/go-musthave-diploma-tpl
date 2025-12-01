package middleware

import (
	"net/http"
	"sync"
	"time"
)

type Limiter struct {
	mu        sync.Mutex
	requests  map[string]int
	lastReset time.Time
}

func LimitRequestsMiddleware(maxRequests int, timeout time.Duration) func(http.Handler) http.Handler {
	limiter := &Limiter{requests: make(map[string]int), lastReset: time.Now()}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limiter.mu.Lock()
			defer limiter.mu.Unlock()

			ip := r.RemoteAddr

			// Проверка и обновление состояния лимита запросов
			if time.Since(limiter.lastReset) > timeout {
				limiter.requests = make(map[string]int)
				limiter.lastReset = time.Now()
			}

			if limiter.requests[ip] >= maxRequests {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			limiter.requests[ip]++
			h.ServeHTTP(w, r)
		})

	}
}
