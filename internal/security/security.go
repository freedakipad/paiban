// Package security 提供安全功能
package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidAPIKey     = errors.New("无效的API密钥")
	ErrExpiredAPIKey     = errors.New("API密钥已过期")
	ErrRateLimitExceeded = errors.New("请求频率超限")
	ErrInvalidSignature  = errors.New("无效的签名")
)

// APIKey API密钥
type APIKey struct {
	Key       string     `json:"key"`
	Secret    string     `json:"-"` // 不序列化
	TenantID  string     `json:"tenant_id"`
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"` // 权限范围
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Enabled   bool       `json:"enabled"`
}

// IsValid 检查密钥是否有效
func (k *APIKey) IsValid() bool {
	if !k.Enabled {
		return false
	}
	if k.ExpiresAt != nil && k.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// HasScope 检查密钥是否有某权限
func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}

// APIKeyManager API密钥管理器
type APIKeyManager struct {
	keys map[string]*APIKey // key -> APIKey
	mu   sync.RWMutex
}

// NewAPIKeyManager 创建密钥管理器
func NewAPIKeyManager() *APIKeyManager {
	return &APIKeyManager{
		keys: make(map[string]*APIKey),
	}
}

// GenerateKey 生成新密钥
func (m *APIKeyManager) GenerateKey(tenantID, name string, scopes []string, expiresIn *time.Duration) (*APIKey, error) {
	key, err := generateRandomString(32)
	if err != nil {
		return nil, err
	}

	secret, err := generateRandomString(64)
	if err != nil {
		return nil, err
	}

	apiKey := &APIKey{
		Key:       "pk_" + key,
		Secret:    secret,
		TenantID:  tenantID,
		Name:      name,
		Scopes:    scopes,
		CreatedAt: time.Now(),
		Enabled:   true,
	}

	if expiresIn != nil {
		expiresAt := time.Now().Add(*expiresIn)
		apiKey.ExpiresAt = &expiresAt
	}

	m.mu.Lock()
	m.keys[apiKey.Key] = apiKey
	m.mu.Unlock()

	return apiKey, nil
}

// Validate 验证密钥
func (m *APIKeyManager) Validate(key string) (*APIKey, error) {
	m.mu.RLock()
	apiKey, exists := m.keys[key]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrInvalidAPIKey
	}

	if !apiKey.IsValid() {
		return nil, ErrExpiredAPIKey
	}

	return apiKey, nil
}

// Revoke 撤销密钥
func (m *APIKeyManager) Revoke(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if apiKey, exists := m.keys[key]; exists {
		apiKey.Enabled = false
	}
}

// Delete 删除密钥
func (m *APIKeyManager) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.keys, key)
}

// RateLimiter 请求频率限制器
type RateLimiter struct {
	requests map[string][]time.Time // key -> request timestamps
	limit    int                    // 时间窗口内最大请求数
	window   time.Duration          // 时间窗口
	mu       sync.Mutex
}

// NewRateLimiter 创建频率限制器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// 启动清理协程
	go rl.cleanup()

	return rl
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// 获取时间窗口内的请求
	reqs := rl.requests[key]
	var validReqs []time.Time
	for _, t := range reqs {
		if t.After(windowStart) {
			validReqs = append(validReqs, t)
		}
	}

	// 检查是否超限
	if len(validReqs) >= rl.limit {
		return false
	}

	// 记录新请求
	validReqs = append(validReqs, now)
	rl.requests[key] = validReqs

	return true
}

// cleanup 定期清理过期数据
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for key, reqs := range rl.requests {
			var validReqs []time.Time
			for _, t := range reqs {
				if t.After(windowStart) {
					validReqs = append(validReqs, t)
				}
			}
			if len(validReqs) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validReqs
			}
		}
		rl.mu.Unlock()
	}
}

// SignatureVerifier 签名验证器
type SignatureVerifier struct {
	secretKey string
}

// NewSignatureVerifier 创建签名验证器
func NewSignatureVerifier(secretKey string) *SignatureVerifier {
	return &SignatureVerifier{secretKey: secretKey}
}

// GenerateSignature 生成签名
func (v *SignatureVerifier) GenerateSignature(payload string, timestamp int64) string {
	message := payload + ":" + string(rune(timestamp))
	h := hmac.New(sha256.New, []byte(v.secretKey))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// Verify 验证签名
func (v *SignatureVerifier) Verify(payload, signature string, timestamp int64, maxAge time.Duration) bool {
	// 检查时间戳是否过期
	requestTime := time.Unix(timestamp, 0)
	if time.Since(requestTime) > maxAge {
		return false
	}

	// 验证签名
	expectedSig := v.GenerateSignature(payload, timestamp)
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// ExtractAPIKey 从请求中提取API密钥
func ExtractAPIKey(r *http.Request) string {
	// 1. 从 Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// 2. 从 X-API-Key header
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}

	// 3. 从 query parameter
	if key := r.URL.Query().Get("api_key"); key != "" {
		return key
	}

	return ""
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// HashPassword 哈希密码
func HashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyPassword 验证密码
func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}

// SanitizeInput 清理输入（防止注入）
func SanitizeInput(input string) string {
	// 基本清理
	input = strings.TrimSpace(input)
	// 移除可能的SQL注入字符
	dangerous := []string{"--", ";", "/*", "*/", "xp_", "@@"}
	for _, d := range dangerous {
		input = strings.ReplaceAll(input, d, "")
	}
	return input
}
