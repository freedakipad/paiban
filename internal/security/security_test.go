package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIKey_IsValid(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name     string
		key      *APIKey
		expected bool
	}{
		{
			name:     "有效密钥",
			key:      &APIKey{Enabled: true},
			expected: true,
		},
		{
			name:     "禁用密钥",
			key:      &APIKey{Enabled: false},
			expected: false,
		},
		{
			name:     "未过期密钥",
			key:      &APIKey{Enabled: true, ExpiresAt: &future},
			expected: true,
		},
		{
			name:     "已过期密钥",
			key:      &APIKey{Enabled: true, ExpiresAt: &past},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.key.IsValid(); result != tt.expected {
				t.Errorf("IsValid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestAPIKey_HasScope(t *testing.T) {
	key := &APIKey{
		Scopes: []string{"schedule", "dispatch"},
	}

	if !key.HasScope("schedule") {
		t.Error("应有schedule权限")
	}
	if !key.HasScope("dispatch") {
		t.Error("应有dispatch权限")
	}
	if key.HasScope("admin") {
		t.Error("不应有admin权限")
	}

	// 测试通配符
	key2 := &APIKey{Scopes: []string{"*"}}
	if !key2.HasScope("anything") {
		t.Error("通配符应匹配任何权限")
	}
}

func TestAPIKeyManager_GenerateKey(t *testing.T) {
	manager := NewAPIKeyManager()

	key, err := manager.GenerateKey("tenant1", "测试密钥", []string{"schedule"}, nil)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if key.Key == "" || key.Secret == "" {
		t.Error("Key and Secret should not be empty")
	}
	if key.TenantID != "tenant1" {
		t.Errorf("Expected TenantID='tenant1', got %s", key.TenantID)
	}
	if !key.Enabled {
		t.Error("New key should be enabled")
	}
}

func TestAPIKeyManager_Validate(t *testing.T) {
	manager := NewAPIKeyManager()

	key, _ := manager.GenerateKey("tenant1", "测试", []string{"schedule"}, nil)

	// 验证有效密钥
	validKey, err := manager.Validate(key.Key)
	if err != nil {
		t.Errorf("Validate failed: %v", err)
	}
	if validKey.Key != key.Key {
		t.Error("Got wrong key")
	}

	// 验证无效密钥
	_, err = manager.Validate("invalid_key")
	if err != ErrInvalidAPIKey {
		t.Errorf("Expected ErrInvalidAPIKey, got: %v", err)
	}
}

func TestAPIKeyManager_Revoke(t *testing.T) {
	manager := NewAPIKeyManager()

	key, _ := manager.GenerateKey("tenant1", "测试", []string{"schedule"}, nil)
	manager.Revoke(key.Key)

	_, err := manager.Validate(key.Key)
	if err != ErrExpiredAPIKey {
		t.Errorf("Expected ErrExpiredAPIKey after revoke, got: %v", err)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(5, time.Second)

	// 前5次应该允许
	for i := 0; i < 5; i++ {
		if !limiter.Allow("client1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 第6次应该拒绝
	if limiter.Allow("client1") {
		t.Error("Request 6 should be denied")
	}

	// 不同客户端应该允许
	if !limiter.Allow("client2") {
		t.Error("Different client should be allowed")
	}
}

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(r *http.Request)
		expected string
	}{
		{
			name: "从Bearer提取",
			setup: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer test_key")
			},
			expected: "test_key",
		},
		{
			name: "从X-API-Key提取",
			setup: func(r *http.Request) {
				r.Header.Set("X-API-Key", "api_key_123")
			},
			expected: "api_key_123",
		},
		{
			name: "从query参数提取",
			setup: func(r *http.Request) {
				q := r.URL.Query()
				q.Set("api_key", "query_key")
				r.URL.RawQuery = q.Encode()
			},
			expected: "query_key",
		},
		{
			name:     "无密钥",
			setup:    func(r *http.Request) {},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			tt.setup(req)

			result := ExtractAPIKey(req)
			if result != tt.expected {
				t.Errorf("ExtractAPIKey() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "secret123"
	hash := HashPassword(password)

	if hash == "" {
		t.Error("Hash should not be empty")
	}
	if hash == password {
		t.Error("Hash should not equal password")
	}

	// 相同密码应产生相同哈希
	hash2 := HashPassword(password)
	if hash != hash2 {
		t.Error("Same password should produce same hash")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "secret123"
	hash := HashPassword(password)

	if !VerifyPassword(password, hash) {
		t.Error("Correct password should verify")
	}
	if VerifyPassword("wrong", hash) {
		t.Error("Wrong password should not verify")
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"test--drop", "testdrop"},
		{"select;delete", "selectdelete"},
		{"normal text", "normal text"},
	}

	for _, tt := range tests {
		result := SanitizeInput(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeInput(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
