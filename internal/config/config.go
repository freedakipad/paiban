// Package config 提供配置管理
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 应用配置
type Config struct {
	App        AppConfig        `yaml:"app"`
	Database   DatabaseConfig   `yaml:"database"`
	Redis      RedisConfig      `yaml:"redis"`
	API        APIConfig        `yaml:"api"`
	Scheduler  SchedulerConfig  `yaml:"scheduler"`
	Dispatcher DispatcherConfig `yaml:"dispatcher"`
	Metrics    MetricsConfig    `yaml:"metrics"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name     string `yaml:"name"`
	Env      string `yaml:"env"`
	Port     int    `yaml:"port"`
	LogLevel string `yaml:"log_level"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Name            string        `yaml:"name"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// DSN 返回数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// Addr 返回Redis地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// APIConfig API配置
type APIConfig struct {
	RateLimit int           `yaml:"rate_limit"`
	Timeout   time.Duration `yaml:"timeout"`
	CORS      CORSConfig    `yaml:"cors"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	Enabled bool     `yaml:"enabled"`
	Origins []string `yaml:"origins"`
}

// SchedulerConfig 排班引擎配置
type SchedulerConfig struct {
	DefaultTimeout    time.Duration `yaml:"default_timeout"`
	MaxIterations     int           `yaml:"max_iterations"`
	OptimizationLevel int           `yaml:"optimization_level"` // 1=快速, 2=平衡, 3=最优
}

// DispatcherConfig 派单引擎配置
type DispatcherConfig struct {
	DefaultTimeout time.Duration `yaml:"default_timeout"`
	OptimizeRoute  bool          `yaml:"optimize_route"`
	MaxDistanceKm  float64       `yaml:"max_distance_km"`
}

// MetricsConfig 监控配置
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// Load 从环境变量加载配置
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:     getEnv("APP_NAME", "paiban"),
			Env:      getEnv("APP_ENV", "development"),
			Port:     getEnvInt("APP_PORT", 7012),
			LogLevel: getEnv("APP_LOG_LEVEL", "info"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			Name:            getEnv("DB_NAME", "paiban"),
			User:            getEnv("DB_USER", "paiban"),
			Password:        getEnv("DB_PASSWORD", "paiban123"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
		},
		API: APIConfig{
			RateLimit: getEnvInt("API_RATE_LIMIT", 100),
			Timeout:   getEnvDuration("API_TIMEOUT", 30*time.Second),
			CORS: CORSConfig{
				Enabled: getEnvBool("API_CORS_ENABLED", true),
				Origins: []string{"*"},
			},
		},
		Scheduler: SchedulerConfig{
			DefaultTimeout:    getEnvDuration("SCHEDULER_TIMEOUT", 30*time.Second),
			MaxIterations:     getEnvInt("SCHEDULER_MAX_ITERATIONS", 1000),
			OptimizationLevel: getEnvInt("SCHEDULER_OPTIMIZATION_LEVEL", 2),
		},
		Dispatcher: DispatcherConfig{
			DefaultTimeout: getEnvDuration("DISPATCHER_TIMEOUT", 5*time.Second),
			OptimizeRoute:  getEnvBool("DISPATCHER_OPTIMIZE_ROUTE", true),
			MaxDistanceKm:  getEnvFloat("DISPATCHER_MAX_DISTANCE", 15.0),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvBool("METRICS_ENABLED", true),
			Path:    getEnv("METRICS_PATH", "/metrics"),
		},
	}

	return cfg, nil
}

// IsDevelopment 检查是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction 检查是否为生产环境
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// IsTest 检查是否为测试环境
func (c *Config) IsTest() bool {
	return c.App.Env == "test"
}

// 辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
