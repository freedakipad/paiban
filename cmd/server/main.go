// PaiBan 排班引擎服务
// 主程序入口

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/internal/handler"
	"github.com/paiban/paiban/internal/metrics"
	"github.com/paiban/paiban/pkg/logger"
)

// 构建信息（通过 ldflags 注入）
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 初始化日志
	logger.Init(logger.Config{
		Level:  os.Getenv("APP_LOG_LEVEL"),
		Format: "console",
	})

	// 打印版本信息
	fmt.Printf("PaiBan 排班引擎 v%s\n", Version)
	fmt.Printf("Build: %s (%s)\n", BuildTime, GitCommit)
	fmt.Println()

	// 获取端口配置
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "7012"
	}

	// 创建处理器
	scheduleHandler := handler.NewScheduleHandler()

	// 创建 HTTP 服务器
	mux := http.NewServeMux()

	// ========================================
	// 系统端点
	// ========================================

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"paiban"}`))
	})

	// 版本信息端点
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"version":"%s","build_time":"%s","git_commit":"%s"}`, Version, BuildTime, GitCommit)
	})

	// ========================================
	// API v1 端点
	// ========================================

	// API 根路由
	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"message": "PaiBan 排班引擎 API v1",
			"endpoints": {
				"schedule": {
					"generate": "POST /api/v1/schedule/generate",
					"validate": "POST /api/v1/schedule/validate"
				},
				"constraints": {
					"templates": "GET /api/v1/constraints/templates"
				},
				"stats": {
					"fairness": "POST /api/v1/stats/fairness",
					"coverage": "POST /api/v1/stats/coverage",
					"workload": "POST /api/v1/stats/workload"
				},
				"dispatch": {
					"single": "POST /api/v1/dispatch/single",
					"batch": "POST /api/v1/dispatch/batch",
					"route": "POST /api/v1/dispatch/route"
				}
			}
		}`))
	})

	// 排班生成 API
	mux.HandleFunc("/api/v1/schedule/generate", scheduleHandler.Generate)

	// 排班验证 API
	mux.HandleFunc("/api/v1/schedule/validate", scheduleHandler.Validate)

	// 约束模板 API
	mux.HandleFunc("/api/v1/constraints/templates", handleConstraintTemplates)

	// 约束库 API - 返回后端支持的所有约束及参数定义
	mux.HandleFunc("/api/v1/constraints/library", handleConstraintLibrary)

	// ========================================
	// 统计分析 API
	// ========================================

	// 公平性分析 API
	mux.HandleFunc("/api/v1/stats/fairness", handler.GetFairnessHandler)

	// 覆盖率分析 API
	mux.HandleFunc("/api/v1/stats/coverage", handler.GetCoverageHandler)

	// 工作量统计 API
	mux.HandleFunc("/api/v1/stats/workload", handler.GetWorkloadHandler)

	// ========================================
	// 派出服务 API
	// ========================================

	// 智能派单 API
	mux.HandleFunc("/api/v1/dispatch/single", handler.DispatchHandler)

	// 批量派单 API
	mux.HandleFunc("/api/v1/dispatch/batch", handler.BatchDispatchHandler)

	// 最优路线 API
	mux.HandleFunc("/api/v1/dispatch/route", handler.OptimalRouteHandler)

	// ========================================
	// 监控端点
	// ========================================

	// Prometheus 指标端点
	mux.Handle("/metrics", metrics.Handler())

	// ========================================
	// 中间件
	// ========================================

	// 创建带中间件的处理器
	// 中间件执行顺序：requestID -> rateLimit -> cors -> logging -> handler
	handler := requestIDMiddleware(rateLimitMiddleware(corsMiddleware(loggingMiddleware(mux))))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 启动服务器（非阻塞）
	go func() {
		logger.Info().
			Str("port", port).
			Str("version", Version).
			Str("url", fmt.Sprintf("http://localhost:%s", port)).
			Str("api_docs", fmt.Sprintf("http://localhost:%s/api/v1/", port)).
			Msg("服务器启动")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("服务器启动失败")
			os.Exit(1)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("服务器关闭失败")
		os.Exit(1)
	}

	logger.Info().Msg("服务器已关闭")
}

// requestIDMiddleware 请求ID追踪中间件
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 尝试从请求头获取 Request ID，没有则生成新的
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置响应头
		w.Header().Set("X-Request-ID", requestID)

		// 将 Request ID 存储到 context 中
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loggingMiddleware 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 获取 Request ID
		requestID, _ := r.Context().Value("request_id").(string)
		
		// 包装ResponseWriter以捕获状态码
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		
		duration := time.Since(start)
		
		logger.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.statusCode).
			Dur("duration", duration).
			Msg("请求处理")
		
		// 记录Prometheus指标
		metrics.RecordRequestMetrics(r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

// responseWriter 包装ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RateLimiter 简单的令牌桶限流器
type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // 每秒添加的令牌数
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter 创建限流器
func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
	return &RateLimiter{
		tokens:     requestsPerSecond,
		maxTokens:  requestsPerSecond * 2, // 允许突发流量
		refillRate: requestsPerSecond,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

var globalRateLimiter = NewRateLimiter(100) // 默认 100 QPS

// rateLimitMiddleware 限流中间件
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !globalRateLimiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   true,
				"code":    "RATE_LIMITED",
				"message": "请求过于频繁，请稍后重试",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware CORS中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ConstraintRule 约束规则
type ConstraintRule struct {
	Name        string `json:"name"`
	Type        string `json:"type"`        // hard/soft
	Category    string `json:"category"`    // 约束类别
	Description string `json:"description"` // 约束描述
	Default     string `json:"default"`     // 默认值
}

// ConstraintTemplate 约束模板
type ConstraintTemplate struct {
	Scenario    string           `json:"scenario"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Constraints []ConstraintRule `json:"constraints"` // 约束规则列表
}

// ConstraintTemplatesResponse 约束模板响应
type ConstraintTemplatesResponse struct {
	Templates []ConstraintTemplate `json:"templates"`
}

// ConstraintParam 约束参数定义
type ConstraintParam struct {
	Name        string `json:"name"`        // 参数名称
	Type        string `json:"type"`        // 参数类型: int, float, string, bool, array
	Description string `json:"description"` // 参数描述
	Default     string `json:"default"`     // 默认值
	Min         string `json:"min,omitempty"`  // 最小值(可选)
	Max         string `json:"max,omitempty"`  // 最大值(可选)
}

// ConstraintDefinition 约束定义（约束库中的完整定义）
type ConstraintDefinition struct {
	Name        string            `json:"name"`        // 约束唯一标识
	DisplayName string            `json:"display_name"` // 显示名称
	Type        string            `json:"type"`        // hard/soft
	Category    string            `json:"category"`    // 分类
	Description string            `json:"description"` // 详细描述
	Scenarios   []string          `json:"scenarios"`   // 适用场景
	Params      []ConstraintParam `json:"params"`      // 可配置参数
}

// ConstraintLibraryResponse 约束库响应
type ConstraintLibraryResponse struct {
	Library []ConstraintDefinition `json:"library"`
}

// handleConstraintTemplates 处理约束模板请求
func handleConstraintTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 通用硬约束
	commonHardConstraints := []ConstraintRule{
		{Name: "max_hours_per_day", Type: "hard", Category: "工时限制", Description: "每日最大工时", Default: "10小时"},
		{Name: "max_hours_per_week", Type: "hard", Category: "工时限制", Description: "每周最大工时", Default: "44小时"},
		{Name: "min_rest_between_shifts", Type: "hard", Category: "休息保障", Description: "班次间最小休息时间", Default: "11小时"},
		{Name: "max_consecutive_days", Type: "hard", Category: "休息保障", Description: "最大连续工作天数", Default: "6天"},
		{Name: "skill_required", Type: "hard", Category: "资质要求", Description: "技能与岗位匹配", Default: "必须满足"},
	}

	// 通用软约束
	commonSoftConstraints := []ConstraintRule{
		{Name: "workload_balance", Type: "soft", Category: "公平性", Description: "工作量均衡", Default: "权重60"},
		{Name: "employee_preference", Type: "soft", Category: "偏好", Description: "员工偏好考虑", Default: "权重50"},
		{Name: "minimize_overtime", Type: "soft", Category: "成本优化", Description: "减少加班", Default: "权重70"},
	}

	templates := []ConstraintTemplate{
		{
			Scenario:    "restaurant",
			Name:        "餐饮门店标准模板",
			Description: "适用于餐饮门店的标准约束配置，包含高峰期人员配置、工时限制等",
			Constraints: append(append(commonHardConstraints,
				ConstraintRule{Name: "industry_certification", Type: "hard", Category: "资质要求", Description: "健康证等行业资质", Default: "必须持有"},
				ConstraintRule{Name: "peak_hours_coverage", Type: "soft", Category: "服务保障", Description: "高峰期人员覆盖", Default: "11:00-13:00, 17:00-20:00 最少3人"},
				ConstraintRule{Name: "split_shift", Type: "soft", Category: "排班模式", Description: "两头班支持", Default: "每周最多2次"},
			), commonSoftConstraints...),
		},
		{
			Scenario:    "factory",
			Name:        "工厂三班倒模板",
			Description: "适用于工厂三班倒的约束配置，包含倒班规则、产线覆盖等",
			Constraints: append(append(commonHardConstraints,
				ConstraintRule{Name: "shift_rotation", Type: "hard", Category: "排班模式", Description: "倒班轮换规则", Default: "早-中-晚轮换"},
				ConstraintRule{Name: "production_line_coverage", Type: "hard", Category: "服务保障", Description: "产线24小时覆盖", Default: "必须满足"},
				ConstraintRule{Name: "handover_overlap", Type: "soft", Category: "交接", Description: "交接班重叠时间", Default: "15分钟"},
			), commonSoftConstraints...),
		},
		{
			Scenario:    "housekeeping",
			Name:        "家政服务模板",
			Description: "适用于家政服务的约束配置，包含服务区域、路程时间等",
			Constraints: append(append(commonHardConstraints,
				ConstraintRule{Name: "service_area", Type: "hard", Category: "区域限制", Description: "服务区域匹配", Default: "必须在服务范围内"},
				ConstraintRule{Name: "travel_time", Type: "soft", Category: "效率优化", Description: "路程时间考虑", Default: "尽量减少"},
				ConstraintRule{Name: "time_window", Type: "hard", Category: "服务保障", Description: "服务时间窗口", Default: "必须在客户指定时段"},
			), commonSoftConstraints...),
		},
		{
			Scenario:    "nursing",
			Name:        "长护险服务模板",
			Description: "适用于长期护理保险服务的约束配置，包含护理计划、资质等级等",
			Constraints: append(append(commonHardConstraints,
				ConstraintRule{Name: "nursing_qualification", Type: "hard", Category: "资质要求", Description: "护理资质等级", Default: "必须持有护理证"},
				ConstraintRule{Name: "service_continuity", Type: "soft", Category: "服务质量", Description: "服务连续性", Default: "优先安排熟悉的护理员"},
				ConstraintRule{Name: "max_patients_per_day", Type: "hard", Category: "服务质量", Description: "每日最大服务患者数", Default: "4人"},
			), commonSoftConstraints...),
		},
	}

	response := ConstraintTemplatesResponse{Templates: templates}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleConstraintLibrary 处理约束库请求 - 返回后端支持的所有约束定义
func handleConstraintLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 完整的约束库 - 后端实际支持的所有约束
	library := []ConstraintDefinition{
		// ========================================
		// 通用硬约束
		// ========================================
		{
			Name:        "max_hours_per_day",
			DisplayName: "每日最大工时",
			Type:        "hard",
			Category:    "工时限制",
			Description: "限制员工每天的最大工作时长，超过则排班无效。适用于所有行业的基础劳动法规要求。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "max_hours", Type: "int", Description: "最大工时(小时)", Default: "10", Min: "6", Max: "14"},
			},
		},
		{
			Name:        "max_hours_per_week",
			DisplayName: "每周最大工时",
			Type:        "hard",
			Category:    "工时限制",
			Description: "限制员工每周的累计工作时长，确保符合劳动法规定。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "max_hours", Type: "int", Description: "最大工时(小时)", Default: "44", Min: "36", Max: "60"},
			},
		},
		{
			Name:        "min_rest_between_shifts",
			DisplayName: "班次间最小休息时间",
			Type:        "hard",
			Category:    "休息保障",
			Description: "确保员工在两个班次之间有足够的休息时间，防止过度疲劳。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "min_hours", Type: "int", Description: "最小休息时间(小时)", Default: "11", Min: "8", Max: "14"},
			},
		},
		{
			Name:        "max_consecutive_days",
			DisplayName: "最大连续工作天数",
			Type:        "hard",
			Category:    "休息保障",
			Description: "限制员工连续工作的最大天数，确保员工有足够的休息日。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "max_days", Type: "int", Description: "最大连续天数", Default: "6", Min: "4", Max: "7"},
			},
		},
		{
			Name:        "skill_required",
			DisplayName: "技能与岗位匹配",
			Type:        "hard",
			Category:    "资质要求",
			Description: "确保分配的员工具备该岗位所需的技能和资质。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params:      []ConstraintParam{},
		},
		{
			Name:        "industry_certification",
			DisplayName: "行业资质认证",
			Type:        "hard",
			Category:    "资质要求",
			Description: "检查员工是否持有行业必需的资质证书（如健康证、护理证等）。",
			Scenarios:   []string{"restaurant", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "required_certs", Type: "array", Description: "必需证书列表", Default: "健康证"},
			},
		},
		// ========================================
		// 通用软约束
		// ========================================
		{
			Name:        "workload_balance",
			DisplayName: "工作量均衡",
			Type:        "soft",
			Category:    "公平性",
			Description: "尽量使各员工的工作量分布均匀，提高公平性和员工满意度。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "60", Min: "0", Max: "100"},
				{Name: "tolerance", Type: "float", Description: "容忍偏差百分比", Default: "20", Min: "5", Max: "50"},
			},
		},
		{
			Name:        "employee_preference",
			DisplayName: "员工偏好考虑",
			Type:        "soft",
			Category:    "偏好",
			Description: "尽量满足员工对班次、休息日等的个人偏好。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "50", Min: "0", Max: "100"},
			},
		},
		{
			Name:        "minimize_overtime",
			DisplayName: "减少加班",
			Type:        "soft",
			Category:    "成本优化",
			Description: "优化排班以减少加班时间，降低人力成本。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "70", Min: "0", Max: "100"},
				{Name: "standard_hours", Type: "int", Description: "标准工时(周)", Default: "40"},
			},
		},
		// ========================================
		// 餐饮行业特有约束
		// ========================================
		{
			Name:        "peak_hours_coverage",
			DisplayName: "高峰期人员覆盖",
			Type:        "soft",
			Category:    "服务保障",
			Description: "确保在用餐高峰期有足够的员工在岗，提供优质服务。",
			Scenarios:   []string{"restaurant"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "90", Min: "0", Max: "100"},
				{Name: "peak_hours", Type: "array", Description: "高峰时段", Default: "11:00-13:00,17:00-20:00"},
				{Name: "min_staff", Type: "int", Description: "最少员工数", Default: "3", Min: "1", Max: "10"},
			},
		},
		{
			Name:        "split_shift",
			DisplayName: "两头班支持",
			Type:        "soft",
			Category:    "排班模式",
			Description: "允许员工上午和晚间分别上班（两头班），中间休息。适合餐饮业高峰期排班。",
			Scenarios:   []string{"restaurant"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "60", Min: "0", Max: "100"},
				{Name: "max_per_week", Type: "int", Description: "每周最多两头班次数", Default: "2", Min: "0", Max: "5"},
				{Name: "allow", Type: "bool", Description: "是否允许两头班", Default: "true"},
			},
		},
		{
			Name:        "position_coverage",
			DisplayName: "岗位覆盖",
			Type:        "soft",
			Category:    "服务保障",
			Description: "确保各时段都有足够的不同岗位员工（如收银、服务员、厨师等）。",
			Scenarios:   []string{"restaurant"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "80", Min: "0", Max: "100"},
			},
		},
		// ========================================
		// 工厂产线特有约束
		// ========================================
		{
			Name:        "shift_rotation",
			DisplayName: "倒班轮换规则",
			Type:        "hard",
			Category:    "排班模式",
			Description: "设定倒班模式（如三班倒），确保班次按规律轮换。",
			Scenarios:   []string{"factory"},
			Params: []ConstraintParam{
				{Name: "pattern", Type: "string", Description: "轮换模式", Default: "三班倒"},
				{Name: "rotation_days", Type: "int", Description: "轮换周期(天)", Default: "7", Min: "3", Max: "14"},
			},
		},
		{
			Name:        "max_consecutive_nights",
			DisplayName: "最大连续夜班",
			Type:        "hard",
			Category:    "休息保障",
			Description: "限制员工连续上夜班的天数，保护员工健康。",
			Scenarios:   []string{"factory"},
			Params: []ConstraintParam{
				{Name: "max_nights", Type: "int", Description: "最大连续夜班天数", Default: "4", Min: "2", Max: "7"},
			},
		},
		{
			Name:        "production_line_coverage",
			DisplayName: "产线24小时覆盖",
			Type:        "hard",
			Category:    "生产保障",
			Description: "确保生产线在指定时间段内有足够的人员覆盖。",
			Scenarios:   []string{"factory"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "100", Min: "0", Max: "100"},
			},
		},
		{
			Name:        "team_together",
			DisplayName: "团队协作",
			Type:        "soft",
			Category:    "协作",
			Description: "尽量安排同一团队的成员在相同班次工作，提高团队协作效率。",
			Scenarios:   []string{"factory"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "70", Min: "0", Max: "100"},
			},
		},
		// ========================================
		// 家政服务特有约束
		// ========================================
		{
			Name:        "service_area",
			DisplayName: "服务区域匹配",
			Type:        "hard",
			Category:    "区域限制",
			Description: "确保员工只被分配到其覆盖的服务区域内的订单。",
			Scenarios:   []string{"housekeeping", "nursing"},
			Params:      []ConstraintParam{},
		},
		{
			Name:        "time_window",
			DisplayName: "服务时间窗口",
			Type:        "hard",
			Category:    "服务保障",
			Description: "确保服务安排在客户指定的时间窗口内进行。",
			Scenarios:   []string{"housekeeping", "nursing"},
			Params:      []ConstraintParam{},
		},
		{
			Name:        "travel_time",
			DisplayName: "路程时间优化",
			Type:        "soft",
			Category:    "效率优化",
			Description: "优化派单顺序，减少员工在不同客户之间的通勤时间。",
			Scenarios:   []string{"housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "80", Min: "0", Max: "100"},
				{Name: "max_travel_minutes", Type: "int", Description: "最大通勤时间(分钟)", Default: "60"},
			},
		},
		// ========================================
		// 长护险/护理特有约束
		// ========================================
		{
			Name:        "nursing_qualification",
			DisplayName: "护理资质等级",
			Type:        "hard",
			Category:    "资质要求",
			Description: "确保护理员具备服务所需的护理资质等级。",
			Scenarios:   []string{"nursing"},
			Params: []ConstraintParam{
				{Name: "required_level", Type: "string", Description: "要求等级", Default: "初级护理员"},
			},
		},
		{
			Name:        "service_continuity",
			DisplayName: "服务连续性",
			Type:        "soft",
			Category:    "服务质量",
			Description: "优先安排熟悉患者情况的护理员，提高护理连续性和患者满意度。",
			Scenarios:   []string{"nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "85", Min: "0", Max: "100"},
			},
		},
		{
			Name:        "max_patients_per_day",
			DisplayName: "每日最大服务患者数",
			Type:        "hard",
			Category:    "服务质量",
			Description: "限制护理员每天服务的最大患者数量，确保服务质量。",
			Scenarios:   []string{"nursing"},
			Params: []ConstraintParam{
				{Name: "max_patients", Type: "int", Description: "最大患者数", Default: "4", Min: "1", Max: "8"},
			},
		},
	}

	response := ConstraintLibraryResponse{Library: library}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
