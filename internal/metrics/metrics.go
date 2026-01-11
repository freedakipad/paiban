// Package metrics 提供Prometheus监控指标
package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// MetricsRegistry 指标注册表
type MetricsRegistry struct {
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	mu         sync.RWMutex
}

// Counter 计数器
type Counter struct {
	Name   string
	Help   string
	Labels []string
	values map[string]float64
	mu     sync.RWMutex
}

// Gauge 仪表盘
type Gauge struct {
	Name   string
	Help   string
	Labels []string
	values map[string]float64
	mu     sync.RWMutex
}

// Histogram 直方图
type Histogram struct {
	Name    string
	Help    string
	Labels  []string
	Buckets []float64
	counts  map[string][]int
	sums    map[string]float64
	mu      sync.RWMutex
}

var (
	registry *MetricsRegistry
	once     sync.Once
)

// GetRegistry 获取全局注册表
func GetRegistry() *MetricsRegistry {
	once.Do(func() {
		registry = &MetricsRegistry{
			counters:   make(map[string]*Counter),
			gauges:     make(map[string]*Gauge),
			histograms: make(map[string]*Histogram),
		}
		initDefaultMetrics()
	})
	return registry
}

// initDefaultMetrics 初始化默认指标
func initDefaultMetrics() {
	// 请求计数器
	registry.NewCounter("paiban_http_requests_total", "HTTP请求总数", []string{"method", "path", "status"})
	
	// 请求延迟直方图
	registry.NewHistogram("paiban_http_request_duration_seconds", "HTTP请求延迟", 
		[]string{"method", "path"},
		[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0})
	
	// 排班生成计数器
	registry.NewCounter("paiban_schedule_generation_total", "排班生成次数", []string{"scenario", "status"})
	
	// 排班生成延迟
	registry.NewHistogram("paiban_schedule_generation_duration_seconds", "排班生成延迟",
		[]string{"scenario"},
		[]float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0})
	
	// 约束评估计数器
	registry.NewCounter("paiban_constraint_evaluations_total", "约束评估次数", []string{"constraint_type", "result"})
	
	// 活动任务数
	registry.NewGauge("paiban_active_tasks", "当前活动任务数", []string{})
	
	// 数据库连接池
	registry.NewGauge("paiban_db_connections", "数据库连接数", []string{"state"})
	
	// 优化迭代次数
	registry.NewCounter("paiban_optimizer_iterations_total", "优化器迭代次数", []string{"optimizer_type"})
	
	// 解决方案质量分数
	registry.NewGauge("paiban_solution_score", "解决方案质量分数", []string{"org_id"})
	
	// 公平性指数
	registry.NewGauge("paiban_fairness_gini", "公平性基尼系数", []string{"org_id", "metric_type"})
	
	// 覆盖率
	registry.NewGauge("paiban_coverage_rate", "班次覆盖率", []string{"org_id"})
}

// NewCounter 创建计数器
func (r *MetricsRegistry) NewCounter(name, help string, labels []string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()

	counter := &Counter{
		Name:   name,
		Help:   help,
		Labels: labels,
		values: make(map[string]float64),
	}
	r.counters[name] = counter
	return counter
}

// NewGauge 创建仪表盘
func (r *MetricsRegistry) NewGauge(name, help string, labels []string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()

	gauge := &Gauge{
		Name:   name,
		Help:   help,
		Labels: labels,
		values: make(map[string]float64),
	}
	r.gauges[name] = gauge
	return gauge
}

// NewHistogram 创建直方图
func (r *MetricsRegistry) NewHistogram(name, help string, labels []string, buckets []float64) *Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()

	histogram := &Histogram{
		Name:    name,
		Help:    help,
		Labels:  labels,
		Buckets: buckets,
		counts:  make(map[string][]int),
		sums:    make(map[string]float64),
	}
	r.histograms[name] = histogram
	return histogram
}

// GetCounter 获取计数器
func (r *MetricsRegistry) GetCounter(name string) *Counter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.counters[name]
}

// GetGauge 获取仪表盘
func (r *MetricsRegistry) GetGauge(name string) *Gauge {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.gauges[name]
}

// GetHistogram 获取直方图
func (r *MetricsRegistry) GetHistogram(name string) *Histogram {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.histograms[name]
}

// Counter methods

// Inc 增加计数
func (c *Counter) Inc(labelValues ...string) {
	c.Add(1, labelValues...)
}

// Add 增加指定值
func (c *Counter) Add(value float64, labelValues ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := labelKey(labelValues)
	c.values[key] += value
}

// Gauge methods

// Set 设置值
func (g *Gauge) Set(value float64, labelValues ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	key := labelKey(labelValues)
	g.values[key] = value
}

// Inc 增加
func (g *Gauge) Inc(labelValues ...string) {
	g.Add(1, labelValues...)
}

// Dec 减少
func (g *Gauge) Dec(labelValues ...string) {
	g.Add(-1, labelValues...)
}

// Add 增加指定值
func (g *Gauge) Add(value float64, labelValues ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	key := labelKey(labelValues)
	g.values[key] += value
}

// Histogram methods

// Observe 记录观测值
func (h *Histogram) Observe(value float64, labelValues ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	key := labelKey(labelValues)
	
	if _, exists := h.counts[key]; !exists {
		h.counts[key] = make([]int, len(h.Buckets)+1)
	}
	
	// 找到对应的bucket
	for i, bucket := range h.Buckets {
		if value <= bucket {
			h.counts[key][i]++
		}
	}
	h.counts[key][len(h.Buckets)]++ // +Inf bucket
	
	h.sums[key] += value
}

// labelKey 生成标签键
func labelKey(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	key := ""
	for i, l := range labels {
		if i > 0 {
			key += ","
		}
		key += l
	}
	return key
}

// Handler 返回Prometheus格式的指标HTTP处理器
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		
		registry := GetRegistry()
		registry.mu.RLock()
		defer registry.mu.RUnlock()

		// 输出计数器
		for _, counter := range registry.counters {
			fmt.Fprintf(w, "# HELP %s %s\n", counter.Name, counter.Help)
			fmt.Fprintf(w, "# TYPE %s counter\n", counter.Name)
			
			counter.mu.RLock()
			for key, value := range counter.values {
				if key == "" {
					fmt.Fprintf(w, "%s %f\n", counter.Name, value)
				} else {
					fmt.Fprintf(w, "%s{%s} %f\n", counter.Name, formatLabels(counter.Labels, key), value)
				}
			}
			counter.mu.RUnlock()
		}

		// 输出仪表盘
		for _, gauge := range registry.gauges {
			fmt.Fprintf(w, "# HELP %s %s\n", gauge.Name, gauge.Help)
			fmt.Fprintf(w, "# TYPE %s gauge\n", gauge.Name)
			
			gauge.mu.RLock()
			for key, value := range gauge.values {
				if key == "" {
					fmt.Fprintf(w, "%s %f\n", gauge.Name, value)
				} else {
					fmt.Fprintf(w, "%s{%s} %f\n", gauge.Name, formatLabels(gauge.Labels, key), value)
				}
			}
			gauge.mu.RUnlock()
		}

		// 输出直方图
		for _, histogram := range registry.histograms {
			fmt.Fprintf(w, "# HELP %s %s\n", histogram.Name, histogram.Help)
			fmt.Fprintf(w, "# TYPE %s histogram\n", histogram.Name)
			
			histogram.mu.RLock()
			for key, counts := range histogram.counts {
				cumulative := 0
				for i, bucket := range histogram.Buckets {
					cumulative += counts[i]
					if key == "" {
						fmt.Fprintf(w, "%s_bucket{le=\"%f\"} %d\n", histogram.Name, bucket, cumulative)
					} else {
						fmt.Fprintf(w, "%s_bucket{%s,le=\"%f\"} %d\n", histogram.Name, formatLabels(histogram.Labels, key), bucket, cumulative)
					}
				}
				cumulative += counts[len(histogram.Buckets)]
				if key == "" {
					fmt.Fprintf(w, "%s_bucket{le=\"+Inf\"} %d\n", histogram.Name, cumulative)
					fmt.Fprintf(w, "%s_sum %f\n", histogram.Name, histogram.sums[key])
					fmt.Fprintf(w, "%s_count %d\n", histogram.Name, cumulative)
				} else {
					fmt.Fprintf(w, "%s_bucket{%s,le=\"+Inf\"} %d\n", histogram.Name, formatLabels(histogram.Labels, key), cumulative)
					fmt.Fprintf(w, "%s_sum{%s} %f\n", histogram.Name, formatLabels(histogram.Labels, key), histogram.sums[key])
					fmt.Fprintf(w, "%s_count{%s} %d\n", histogram.Name, formatLabels(histogram.Labels, key), cumulative)
				}
			}
			histogram.mu.RUnlock()
		}
	})
}

// formatLabels 格式化标签
func formatLabels(names []string, values string) string {
	vals := splitLabelKey(values)
	result := ""
	for i, name := range names {
		if i > 0 {
			result += ","
		}
		val := ""
		if i < len(vals) {
			val = vals[i]
		}
		result += fmt.Sprintf("%s=\"%s\"", name, val)
	}
	return result
}

// splitLabelKey 分割标签键
func splitLabelKey(key string) []string {
	if key == "" {
		return nil
	}
	var result []string
	current := ""
	for _, c := range key {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}

// RecordRequestMetrics 记录请求指标
func RecordRequestMetrics(method, path string, status int, duration time.Duration) {
	registry := GetRegistry()
	
	// 请求计数
	counter := registry.GetCounter("paiban_http_requests_total")
	if counter != nil {
		counter.Inc(method, path, fmt.Sprintf("%d", status))
	}
	
	// 请求延迟
	histogram := registry.GetHistogram("paiban_http_request_duration_seconds")
	if histogram != nil {
		histogram.Observe(duration.Seconds(), method, path)
	}
}

// RecordScheduleGeneration 记录排班生成指标
func RecordScheduleGeneration(scenario string, success bool, duration time.Duration) {
	registry := GetRegistry()
	
	status := "success"
	if !success {
		status = "failure"
	}
	
	counter := registry.GetCounter("paiban_schedule_generation_total")
	if counter != nil {
		counter.Inc(scenario, status)
	}
	
	histogram := registry.GetHistogram("paiban_schedule_generation_duration_seconds")
	if histogram != nil {
		histogram.Observe(duration.Seconds(), scenario)
	}
}

// RecordConstraintEvaluation 记录约束评估指标
func RecordConstraintEvaluation(constraintType string, satisfied bool) {
	registry := GetRegistry()
	
	result := "satisfied"
	if !satisfied {
		result = "violated"
	}
	
	counter := registry.GetCounter("paiban_constraint_evaluations_total")
	if counter != nil {
		counter.Inc(constraintType, result)
	}
}

// SetSolutionScore 设置解决方案质量分数
func SetSolutionScore(orgID string, score float64) {
	registry := GetRegistry()
	gauge := registry.GetGauge("paiban_solution_score")
	if gauge != nil {
		gauge.Set(score, orgID)
	}
}

// SetFairnessGini 设置公平性基尼系数
func SetFairnessGini(orgID, metricType string, gini float64) {
	registry := GetRegistry()
	gauge := registry.GetGauge("paiban_fairness_gini")
	if gauge != nil {
		gauge.Set(gini, orgID, metricType)
	}
}

// SetCoverageRate 设置覆盖率
func SetCoverageRate(orgID string, rate float64) {
	registry := GetRegistry()
	gauge := registry.GetGauge("paiban_coverage_rate")
	if gauge != nil {
		gauge.Set(rate, orgID)
	}
}

