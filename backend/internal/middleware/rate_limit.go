package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"commercial-diving-decompression-control/backend/internal/util"
	"github.com/gin-gonic/gin"
)

type visitorWindow struct {
	started time.Time
	count   int
}

type LocalRateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	visitors map[string]visitorWindow
}

func NewLocalRateLimiter(limit int) *LocalRateLimiter {
	return &LocalRateLimiter{limit: limit, window: time.Minute, visitors: map[string]visitorWindow{}}
}

func (l *LocalRateLimiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP()
		l.mu.Lock()
		state, exists := l.visitors[key]
		if !exists || now.Sub(state.started) >= l.window {
			state = visitorWindow{started: now}
		}
		state.count++
		l.visitors[key] = state
		remaining := l.limit - state.count
		reset := state.started.Add(l.window)
		l.mu.Unlock()

		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(l.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
		if state.count > l.limit {
			c.Header("Retry-After", strconv.Itoa(int(time.Until(reset).Seconds())+1))
			c.Abort()
			c.JSON(http.StatusTooManyRequests, util.Envelope{
				Error:     &util.Error{Code: "RATE_LIMITED", Message: "local request limit exceeded"},
				RequestID: util.RequestID(c),
			})
			return
		}
		c.Next()
	}
}
