// Package middleware 提供HTTP中间件
package middleware

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/paiban/paiban/internal/security"
	"github.com/paiban/paiban/internal/tenant"
)

// AuthConfig 认证配置
type AuthConfig struct {
	APIKeyManager   *security.APIKeyManager
	TenantManager   *tenant.TenantManager
	RateLimiter     *security.RateLimiter
	SkipPaths       []string // 跳过认证的路径
	EnableRateLimit bool
}

// AuthMiddleware 认证中间件
func AuthMiddleware(config *AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否跳过认证
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 提取API密钥
			apiKey := security.ExtractAPIKey(r)
			if apiKey == "" {
				http.Error(w, `{"error":"missing_api_key","message":"API密钥未提供"}`, http.StatusUnauthorized)
				return
			}

			// 验证API密钥
			key, err := config.APIKeyManager.Validate(apiKey)
			if err != nil {
				log.Printf("API密钥验证失败: %s, err=%v", apiKey[:10]+"...", err)
				http.Error(w, `{"error":"invalid_api_key","message":"无效的API密钥"}`, http.StatusUnauthorized)
				return
			}

			// 获取租户
			t, err := config.TenantManager.Get(key.TenantID)
			if err != nil {
				http.Error(w, `{"error":"tenant_error","message":"租户不可用"}`, http.StatusForbidden)
				return
			}

			// 检查频率限制
			if config.EnableRateLimit && config.RateLimiter != nil {
				if !config.RateLimiter.Allow(key.TenantID) {
					http.Error(w, `{"error":"rate_limit","message":"请求频率超限"}`, http.StatusTooManyRequests)
					return
				}
			}

			// 将租户信息添加到上下文
			ctx := tenant.WithTenant(r.Context(), t)
			r = r.WithContext(ctx)

			// 添加租户信息到响应头
			w.Header().Set("X-Tenant-ID", t.ID.String())

			next.ServeHTTP(w, r)
		})
	}
}

// RequireScope 权限范围检查中间件
func RequireScope(scope string, keyManager *security.APIKeyManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := security.ExtractAPIKey(r)
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			key, err := keyManager.Validate(apiKey)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			if !key.HasScope(scope) {
				http.Error(w, `{"error":"forbidden","message":"权限不足"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取租户信息
		tenantInfo := "anonymous"
		if t, ok := tenant.FromContext(r.Context()); ok {
			tenantInfo = t.Code
		}

		log.Printf("[%s] %s %s - tenant=%s", r.Method, r.URL.Path, r.RemoteAddr, tenantInfo)
		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 安全相关响应头
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware 恢复中间件（捕获panic）
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, `{"error":"internal_error","message":"服务器内部错误"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

func generateRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("req_%x", b[:8])
}

