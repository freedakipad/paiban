// Package optimizer 提供排班优化算法
package optimizer

import (
	"context"
	"hash/fnv"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/paiban/paiban/pkg/model"
)

// OptimizationConfig 优化配置
type OptimizationConfig struct {
	MaxIterations    int           `json:"max_iterations"`     // 最大迭代次数
	MaxTime          time.Duration `json:"max_time"`           // 最大运行时间
	InitialTemp      float64       `json:"initial_temp"`       // 模拟退火初始温度
	CoolingRate      float64       `json:"cooling_rate"`       // 冷却速率
	TabuSize         int           `json:"tabu_size"`          // 禁忌表大小
	NeighborhoodSize int           `json:"neighborhood_size"`  // 邻域大小
	ParallelWorkers  int           `json:"parallel_workers"`   // 并行工作数
	StopOnPlateau    bool          `json:"stop_on_plateau"`    // 平台期停止
	PlateauThreshold int           `json:"plateau_threshold"`  // 平台期阈值（无改进迭代次数）
}

// DefaultOptConfig 默认优化配置
func DefaultOptConfig() *OptimizationConfig {
	return &OptimizationConfig{
		MaxIterations:    1000,
		MaxTime:          30 * time.Second,
		InitialTemp:      100.0,
		CoolingRate:      0.99,
		TabuSize:         50,
		NeighborhoodSize: 20,
		ParallelWorkers:  4,
		StopOnPlateau:    true,
		PlateauThreshold: 100,
	}
}

// Solution 表示一个排班方案
type Solution struct {
	Assignments []*model.Assignment
	Score       float64
	Violations  []string
	Feasible    bool
}

// Clone 深拷贝解决方案
func (s *Solution) Clone() *Solution {
	clone := &Solution{
		Assignments: make([]*model.Assignment, len(s.Assignments)),
		Score:       s.Score,
		Violations:  make([]string, len(s.Violations)),
		Feasible:    s.Feasible,
	}
	for i, a := range s.Assignments {
		cloneA := *a
		clone.Assignments[i] = &cloneA
	}
	copy(clone.Violations, s.Violations)
	return clone
}

// ConstraintEvaluator 约束评估器接口
type ConstraintEvaluator interface {
	Evaluate(assignments []*model.Assignment, employees []*model.Employee, shifts []*model.Shift) (float64, []string)
}

// LocalSearchOptimizer 局部搜索优化器
type LocalSearchOptimizer struct {
	config     *OptimizationConfig
	evaluator  ConstraintEvaluator
	neighbors  *NeighborhoodGenerator
	tabuList   *TabuList
	rng        *rand.Rand
	mu         sync.Mutex
}

// NewLocalSearchOptimizer 创建局部搜索优化器
func NewLocalSearchOptimizer(config *OptimizationConfig, evaluator ConstraintEvaluator) *LocalSearchOptimizer {
	if config == nil {
		config = DefaultOptConfig()
	}
	return &LocalSearchOptimizer{
		config:    config,
		evaluator: evaluator,
		neighbors: NewNeighborhoodGenerator(),
		tabuList:  NewTabuList(config.TabuSize),
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// OptimizeContext 优化上下文
type OptimizeContext struct {
	Employees []*model.Employee
	Shifts    []*model.Shift
}

// Optimize 优化排班方案
func (o *LocalSearchOptimizer) Optimize(ctx context.Context, initial *Solution, employees []*model.Employee, shifts []*model.Shift) (*Solution, error) {
	start := time.Now()
	
	current := initial.Clone()
	best := current.Clone()
	
	temperature := o.config.InitialTemp
	noImprovementCount := 0
	
	optCtx := &OptimizeContext{
		Employees: employees,
		Shifts:    shifts,
	}
	
	log.Printf("开始局部搜索优化: max_iterations=%d, max_time=%s, initial_score=%.2f",
		o.config.MaxIterations, o.config.MaxTime, current.Score)

	for i := 0; i < o.config.MaxIterations; i++ {
		// 检查超时和取消
		select {
		case <-ctx.Done():
			log.Println("优化被取消")
			return best, ctx.Err()
		default:
		}
		
		if time.Since(start) > o.config.MaxTime {
			log.Println("达到最大运行时间")
			break
		}

		// 生成邻域解
		neighbors := o.generateNeighbors(current, employees, shifts)
		if len(neighbors) == 0 {
			continue
		}

		// 评估邻域解
		bestNeighbor := o.evaluateBestNeighbor(neighbors, optCtx)
		if bestNeighbor == nil {
			continue
		}

		// 检查是否在禁忌表中
		moveKey := o.getMoveKey(current, bestNeighbor)
		inTabu := o.tabuList.Contains(moveKey)

		// 模拟退火接受准则
		accept := false
		if bestNeighbor.Score < current.Score {
			// 更优解，接受
			accept = true
		} else if !inTabu {
			// 较差解，以概率接受（模拟退火）
			delta := bestNeighbor.Score - current.Score
			prob := boltzmannProbability(delta, temperature)
			if o.rng.Float64() < prob {
				accept = true
			}
		}

		if accept {
			current = bestNeighbor
			o.tabuList.Add(moveKey)

			// 更新最优解
			if current.Score < best.Score {
				best = current.Clone()
				noImprovementCount = 0
				log.Printf("发现更优解: iteration=%d, score=%.2f", i, best.Score)
			} else {
				noImprovementCount++
			}
		} else {
			noImprovementCount++
		}

		// 检查平台期
		if o.config.StopOnPlateau && noImprovementCount >= o.config.PlateauThreshold {
			log.Printf("达到平台期阈值，停止优化: iterations=%d, no_improvement=%d", i, noImprovementCount)
			break
		}

		// 降温
		temperature *= o.config.CoolingRate
	}

	elapsed := time.Since(start)
	log.Printf("局部搜索优化完成: initial=%.2f, final=%.2f, improvement=%.2f, elapsed=%s",
		initial.Score, best.Score, initial.Score-best.Score, elapsed)

	return best, nil
}

// generateNeighbors 生成邻域解
func (o *LocalSearchOptimizer) generateNeighbors(current *Solution, employees []*model.Employee, shifts []*model.Shift) []*Solution {
	neighbors := make([]*Solution, 0, o.config.NeighborhoodSize)
	
	for i := 0; i < o.config.NeighborhoodSize; i++ {
		neighbor := o.neighbors.GenerateNeighbor(current, employees, shifts)
		if neighbor != nil {
			neighbors = append(neighbors, neighbor)
		}
	}
	
	return neighbors
}

// evaluateBestNeighbor 评估并返回最优邻域解
func (o *LocalSearchOptimizer) evaluateBestNeighbor(neighbors []*Solution, optCtx *OptimizeContext) *Solution {
	if len(neighbors) == 0 {
		return nil
	}

	var best *Solution
	var bestScore float64 = -1

	for _, neighbor := range neighbors {
		// 评估约束
		score, violations := o.evaluateSolution(neighbor.Assignments, optCtx)
		neighbor.Score = score
		neighbor.Violations = violations
		neighbor.Feasible = len(violations) == 0

		if best == nil || score < bestScore {
			best = neighbor
			bestScore = score
		}
	}

	return best
}

// evaluateSolution 评估解决方案
func (o *LocalSearchOptimizer) evaluateSolution(assignments []*model.Assignment, optCtx *OptimizeContext) (float64, []string) {
	if o.evaluator == nil {
		return 0, nil
	}
	return o.evaluator.Evaluate(assignments, optCtx.Employees, optCtx.Shifts)
}

// getMoveKey 获取移动的唯一键
func (o *LocalSearchOptimizer) getMoveKey(_, to *Solution) uint64 {
	return hashAssignments(to.Assignments)
}

// hashAssignments 计算分配的哈希 (使用FNV-1a算法)
func hashAssignments(assignments []*model.Assignment) uint64 {
	if len(assignments) == 0 {
		return 0
	}
	h := fnv.New64a()
	for _, a := range assignments {
		h.Write(a.EmployeeID[:])
		h.Write(a.ShiftID[:])
		h.Write([]byte(a.Date))
	}
	return h.Sum64()
}

// boltzmannProbability 计算模拟退火的接受概率
// delta: 能量差 (new - old)
// temperature: 当前温度
func boltzmannProbability(delta, temperature float64) float64 {
	if delta <= 0 {
		return 1.0 // 更优解总是接受
	}
	if temperature <= 0 {
		return 0.0 // 温度为0时不接受更差的解
	}
	return math.Exp(-delta / temperature)
}

// TabuList 禁忌表（使用uint64哈希作为键提高性能）
type TabuList struct {
	items    map[uint64]struct{}
	order    []uint64
	maxSize  int
	mu       sync.RWMutex
}

// NewTabuList 创建禁忌表
func NewTabuList(size int) *TabuList {
	return &TabuList{
		items:   make(map[uint64]struct{}),
		order:   make([]uint64, 0, size),
		maxSize: size,
	}
}

// Add 添加到禁忌表
func (t *TabuList) Add(key uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.items[key]; exists {
		return
	}

	// 超出容量时移除最旧的
	if len(t.order) >= t.maxSize {
		oldest := t.order[0]
		t.order = t.order[1:]
		delete(t.items, oldest)
	}

	t.items[key] = struct{}{}
	t.order = append(t.order, key)
}

// Contains 检查是否在禁忌表中
func (t *TabuList) Contains(key uint64) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, exists := t.items[key]
	return exists
}

// Clear 清空禁忌表
func (t *TabuList) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.items = make(map[uint64]struct{})
	t.order = t.order[:0]
}

