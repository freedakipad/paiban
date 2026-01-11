// Package logger 提供统一的日志框架
package logger

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	once   sync.Once
	logger zerolog.Logger
)

// Level 日志级别
type Level = zerolog.Level

const (
	DebugLevel = zerolog.DebugLevel
	InfoLevel  = zerolog.InfoLevel
	WarnLevel  = zerolog.WarnLevel
	ErrorLevel = zerolog.ErrorLevel
	FatalLevel = zerolog.FatalLevel
)

// Config 日志配置
type Config struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"` // json/console
	Output     string `yaml:"output" json:"output"` // stdout/stderr/file
	FilePath   string `yaml:"file_path,omitempty" json:"file_path,omitempty"`
	TimeFormat string `yaml:"time_format,omitempty" json:"time_format,omitempty"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Format:     "console",
		Output:     "stdout",
		TimeFormat: time.RFC3339,
	}
}

// Init 初始化日志器
func Init(cfg Config) {
	once.Do(func() {
		level := parseLevel(cfg.Level)
		zerolog.SetGlobalLevel(level)

		var output io.Writer
		switch cfg.Output {
		case "stderr":
			output = os.Stderr
		case "file":
			if cfg.FilePath != "" {
				f, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err == nil {
					output = f
				} else {
					output = os.Stdout
				}
			} else {
				output = os.Stdout
			}
		default:
			output = os.Stdout
		}

		if cfg.Format == "console" {
			output = zerolog.ConsoleWriter{
				Out:        output,
				TimeFormat: cfg.TimeFormat,
			}
		}

		logger = zerolog.New(output).With().Timestamp().Logger()
	})
}

// parseLevel 解析日志级别
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Get 获取日志器
func Get() *zerolog.Logger {
	if logger.GetLevel() == zerolog.Disabled {
		Init(DefaultConfig())
	}
	return &logger
}

// WithContext 从上下文创建日志器
func WithContext(ctx context.Context) *zerolog.Logger {
	l := Get().With().Logger()
	
	// 添加请求ID
	if reqID, ok := ctx.Value("request_id").(string); ok {
		l = l.With().Str("request_id", reqID).Logger()
	}
	
	// 添加组织ID
	if orgID, ok := ctx.Value("org_id").(string); ok {
		l = l.With().Str("org_id", orgID).Logger()
	}
	
	return &l
}

// Debug 记录调试日志
func Debug() *zerolog.Event {
	return Get().Debug()
}

// Info 记录信息日志
func Info() *zerolog.Event {
	return Get().Info()
}

// Warn 记录警告日志
func Warn() *zerolog.Event {
	return Get().Warn()
}

// Error 记录错误日志
func Error() *zerolog.Event {
	return Get().Error()
}

// Fatal 记录致命错误日志
func Fatal() *zerolog.Event {
	return Get().Fatal()
}

// WithError 添加错误信息
func WithError(err error) *zerolog.Event {
	return Get().Error().Err(err)
}

// WithField 添加字段
func WithField(key string, value interface{}) *zerolog.Logger {
	l := Get().With().Interface(key, value).Logger()
	return &l
}

// WithFields 添加多个字段
func WithFields(fields map[string]interface{}) *zerolog.Logger {
	ctx := Get().With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	l := ctx.Logger()
	return &l
}

// SchedulerLogger 排班引擎专用日志器
type SchedulerLogger struct {
	base *zerolog.Logger
}

// NewSchedulerLogger 创建排班引擎日志器
func NewSchedulerLogger() *SchedulerLogger {
	l := Get().With().Str("component", "scheduler").Logger()
	return &SchedulerLogger{base: &l}
}

// StartSchedule 记录排班开始
func (l *SchedulerLogger) StartSchedule(scheduleID string, employees, days int) {
	l.base.Info().
		Str("schedule_id", scheduleID).
		Int("employees", employees).
		Int("days", days).
		Msg("开始生成排班")
}

// ConstraintViolation 记录约束违反
func (l *SchedulerLogger) ConstraintViolation(constraint, details string) {
	l.base.Warn().
		Str("constraint", constraint).
		Str("details", details).
		Msg("约束违反")
}

// ScheduleComplete 记录排班完成
func (l *SchedulerLogger) ScheduleComplete(scheduleID string, duration time.Duration, score float64) {
	l.base.Info().
		Str("schedule_id", scheduleID).
		Dur("duration", duration).
		Float64("score", score).
		Msg("排班生成完成")
}

