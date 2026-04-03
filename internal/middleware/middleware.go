package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/williamprado/foto-magica-profissional/internal/auth"
)

type ctxKey string

const (
	userKey   ctxKey = "user"
	tenantKey ctxKey = "tenant"
)

type UserClaims struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
	Role     string
	Email    string
}

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		logger.Info("http_request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(started)),
		)
	}
}

func Auth(jwt auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "missing bearer token"}})
			c.Abort()
			return
		}

		claims, err := jwt.Verify(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "invalid token"}})
			c.Abort()
			return
		}

		userClaims := UserClaims{
			UserID:   claims.UserID,
			TenantID: claims.TenantID,
			Role:     claims.Role,
			Email:    claims.Email,
		}
		c.Set(string(userKey), userClaims)
		c.Set(string(tenantKey), claims.TenantID)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), userKey, userClaims))
		c.Next()
	}
}

func MustUser(c *gin.Context) UserClaims {
	claims, _ := c.MustGet(string(userKey)).(UserClaims)
	return claims
}

func TenantScoped() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestedTenant := c.GetHeader("X-Tenant-ID")
		if requestedTenant == "" {
			c.Next()
			return
		}

		claims := MustUser(c)
		if claims.TenantID.String() != requestedTenant {
			c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "tenant_scope_violation", "message": "tenant mismatch"}})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireRole(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := MustUser(c)
		for _, role := range allowed {
			if claims.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": "insufficient role"}})
		c.Abort()
	}
}

func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	type counter struct {
		count int
		reset time.Time
	}

	var (
		mu       sync.Mutex
		counters = map[string]counter{}
	)

	return func(c *gin.Context) {
		key := c.ClientIP()
		mu.Lock()
		current := counters[key]
		if time.Now().After(current.reset) {
			current = counter{reset: time.Now().Add(time.Minute)}
		}
		current.count++
		counters[key] = current
		mu.Unlock()

		if current.count > requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "rate_limited", "message": "generation limit exceeded"}})
			c.Abort()
			return
		}

		c.Next()
	}
}
