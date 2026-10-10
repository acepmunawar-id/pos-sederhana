// Package middleware contains Gin middleware: JWT auth, role guard, logging, recovery and request ID.
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/posumkm/backend/pkg/hash"
	"github.com/posumkm/backend/pkg/jwt"
	"github.com/posumkm/backend/pkg/response"
)

const (
	ctxUserID = "auth.user_id"
	ctxRole   = "auth.role"
	ctxName   = "auth.name"
)

// Auth validates the Bearer access token and stores the user identity in the context.
func Auth(tokens *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		raw, ok := strings.CutPrefix(h, "Bearer ")
		if !ok || raw == "" {
			response.Error(c, apperror.Unauthorized("Silakan login terlebih dahulu"))
			return
		}
		claims, err := tokens.Parse(raw)
		if err != nil {
			response.Error(c, apperror.Unauthorized("Sesi tidak valid atau sudah berakhir"))
			return
		}
		id, err := claims.UserID()
		if err != nil {
			response.Error(c, apperror.Unauthorized("Sesi tidak valid"))
			return
		}
		c.Set(ctxUserID, id)
		c.Set(ctxRole, claims.Role)
		c.Set(ctxName, claims.Name)
		c.Next()
	}
}

// RequireRoles allows the request only when the authenticated role is one of roles.
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !slices.Contains(roles, Role(c)) {
			response.Error(c, apperror.Forbidden("Anda tidak memiliki akses ke fitur ini"))
			return
		}
		c.Next()
	}
}

func UserID(c *gin.Context) int64 { return c.GetInt64(ctxUserID) }
func Role(c *gin.Context) string  { return c.GetString(ctxRole) }

// RequestID propagates X-Request-ID or generates one.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" || len(id) > 64 {
			id, _ = hash.RandomToken(9)
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// Logger writes one structured log line per request.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}
		log.LogAttrs(c.Request.Context(), level, "http request",
			slog.String("request_id", c.GetString("request_id")),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.ClientIP()),
			slog.Int64("user_id", c.GetInt64(ctxUserID)),
		)
	}
}

// Recovery converts panics into a 500 response in the standard envelope.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered", "error", fmt.Sprint(r), "stack", string(debug.Stack()),
					"request_id", c.GetString("request_id"))
				response.Error(c, apperror.New(http.StatusInternalServerError, "Terjadi kesalahan pada server"))
			}
		}()
		c.Next()
	}
}
