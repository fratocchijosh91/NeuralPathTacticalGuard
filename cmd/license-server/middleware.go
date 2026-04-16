package main

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitorEntry
	limit    int
	window   time.Duration
}

type visitorEntry struct {
	count    int
	windowAt time.Time
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	rl := &ipRateLimiter{
		visitors: make(map[string]*visitorEntry),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists || now.After(v.windowAt) {
		rl.visitors[ip] = &visitorEntry{count: 1, windowAt: now.Add(rl.window)}
		return true
	}

	v.count++
	return v.count <= rl.limit
}

func (rl *ipRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.After(v.windowAt) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func rateLimitMiddleware(rl *ipRateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		if !rl.allow(ip) {
			writeJSON(w, http.StatusTooManyRequests, activationResponse{
				Message: "troppe richieste, riprova tra poco",
			})
			return
		}
		next(w, r)
	}
}

func requireAPIKeyMiddleware(expectedAPIKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(expectedAPIKey) == "" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"message": "endpoint admin non configurato",
			})
			return
		}

		provided := strings.TrimSpace(r.Header.Get("X-API-Key"))
		if provided == "" {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				provided = strings.TrimSpace(authHeader[7:])
			}
		}
		if provided != expectedAPIKey {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"message": "api key non valida",
			})
			return
		}
		next(w, r)
	}
}

func auditLogMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next(rec, r)

		log.Printf("AUDIT method=%s path=%s ip=%s status=%d duration=%s ua=%q",
			r.Method,
			r.URL.Path,
			extractIP(r),
			rec.statusCode,
			time.Since(start).Round(time.Millisecond),
			r.UserAgent(),
		)
	}
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := splitFirst(xff, ",")
		return trimSpace(parts)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return trimSpace(xri)
	}
	host := r.RemoteAddr
	if idx := lastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

func splitFirst(s, sep string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			return s[:i]
		}
	}
	return s
}

func lastIndex(s, sep string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == sep[0] {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}
