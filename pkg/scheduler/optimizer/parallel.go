// Package optimizer 提供排班优化算法
package optimizer

import (
	"context"
	"log"
	"sync"

	"github.com/paiban/paiban/pkg/model"
)

// ParallelEvaluator 并行评估器
type ParallelEvaluator struct {
	workers   int
	evaluator ConstraintEvaluator
}

// NewParallelEvaluator 创建并行评估器
func NewParallelEvaluator(workers int, evaluator ConstraintEvaluator) *ParallelEvaluator {
	if workers <= 0 {
		workers = 4
	}
	return &ParallelEvaluator{
		workers:   workers,
		evaluator: evaluator,
	}
}

// EvaluationResult 评估结果
type EvaluationResult struct {
	Index      int
	Solution   *Solution
	Score      float64
	Violations []string
	Feasible   bool
}

// EvaluateBatch 并行评估一批解决方案
func (p *ParallelEvaluator) EvaluateBatch(ctx context.Context, solutions []*Solution, optCtx *OptimizeContext) []EvaluationResult {
	if len(solutions) == 0 {
		return nil
	}

	resultChan := make(chan EvaluationResult, len(solutions))
	jobChan := make(chan struct {
		index    int
		solution *Solution
	}, len(solutions))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				select {
				case <-ctx.Done():
					return
				default:
					result := p.evaluateSingle(job.solution, optCtx)
					result.Index = job.index
					resultChan <- result
				}
			}
		}()
	}

	// 发送任务
	go func() {
		for i, sol := range solutions {
			jobChan <- struct {
				index    int
				solution *Solution
			}{i, sol}
		}
		close(jobChan)
	}()

	// 等待完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	results := make([]EvaluationResult, len(solutions))
	for result := range resultChan {
		results[result.Index] = result
	}

	return results
}

// evaluateSingle 评估单个解决方案
func (p *ParallelEvaluator) evaluateSingle(solution *Solution, optCtx *OptimizeContext) EvaluationResult {
	if p.evaluator == nil {
		return EvaluationResult{
			Solution: solution,
			Score:    0,
			Feasible: true,
		}
	}

	score, violations := p.evaluator.Evaluate(solution.Assignments, optCtx.Employees, optCtx.Shifts)

	return EvaluationResult{
		Solution:   solution,
		Score:      score,
		Violations: violations,
		Feasible:   len(violations) == 0,
	}
}

// FindBest 从结果中找出最优解
func (p *ParallelEvaluator) FindBest(results []EvaluationResult) *EvaluationResult {
	if len(results) == 0 {
		return nil
	}

	best := &results[0]
	for i := 1; i < len(results); i++ {
		if results[i].Score < best.Score {
			best = &results[i]
		}
	}
	return best
}

// ParallelOptimizer 并行优化器
type ParallelOptimizer struct {
	config      *OptimizationConfig
	evaluator   *ParallelEvaluator
	neighbors   *NeighborhoodGenerator
}

// NewParallelOptimizer 创建并行优化器
func NewParallelOptimizer(config *OptimizationConfig, constraintEvaluator ConstraintEvaluator) *ParallelOptimizer {
	if config == nil {
		config = DefaultOptConfig()
	}
	return &ParallelOptimizer{
		config:    config,
		evaluator: NewParallelEvaluator(config.ParallelWorkers, constraintEvaluator),
		neighbors: NewNeighborhoodGenerator(),
	}
}

// OptimizeParallel 并行优化
func (p *ParallelOptimizer) OptimizeParallel(ctx context.Context, initial *Solution, employees []*model.Employee, shifts []*model.Shift) (*Solution, error) {
	current := initial.Clone()
	best := current.Clone()

	optCtx := &OptimizeContext{
		Employees: employees,
		Shifts:    shifts,
	}

	log.Printf("开始并行优化: workers=%d, neighborhood_size=%d",
		p.config.ParallelWorkers, p.config.NeighborhoodSize)

	noImprovementCount := 0

	for iter := 0; iter < p.config.MaxIterations; iter++ {
		select {
		case <-ctx.Done():
			return best, ctx.Err()
		default:
		}

		// 并行生成邻域解
		neighbors := p.generateNeighborsParallel(ctx, current, employees, shifts, p.config.NeighborhoodSize)
		if len(neighbors) == 0 {
			continue
		}

		// 并行评估
		results := p.evaluator.EvaluateBatch(ctx, neighbors, optCtx)
		
		// 找出最优邻域解
		bestResult := p.evaluator.FindBest(results)
		if bestResult == nil {
			continue
		}

		// 更新当前解
		if bestResult.Score < current.Score {
			current = bestResult.Solution.Clone()
			current.Score = bestResult.Score
			current.Violations = bestResult.Violations
			current.Feasible = bestResult.Feasible

			// 更新最优解
			if current.Score < best.Score {
				best = current.Clone()
				noImprovementCount = 0
				log.Printf("并行优化发现更优解: iteration=%d, score=%.2f, violations=%d",
					iter, best.Score, len(best.Violations))
			}
		} else {
			noImprovementCount++
		}

		// 检查平台期
		if p.config.StopOnPlateau && noImprovementCount >= p.config.PlateauThreshold {
			log.Printf("并行优化达到平台期: iterations=%d", iter)
			break
		}
	}

	log.Printf("并行优化完成: initial=%.2f, final=%.2f", initial.Score, best.Score)

	return best, nil
}

// generateNeighborsParallel 并行生成邻域解
func (p *ParallelOptimizer) generateNeighborsParallel(ctx context.Context, current *Solution, employees []*model.Employee, shifts []*model.Shift, count int) []*Solution {
	resultChan := make(chan *Solution, count)
	
	var wg sync.WaitGroup
	batchSize := count / p.config.ParallelWorkers
	if batchSize < 1 {
		batchSize = 1
	}

	for i := 0; i < p.config.ParallelWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			localGen := NewNeighborhoodGenerator()
			
			for j := 0; j < batchSize; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					neighbor := localGen.GenerateNeighbor(current, employees, shifts)
					if neighbor != nil {
						resultChan <- neighbor
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make([]*Solution, 0, count)
	for neighbor := range resultChan {
		results = append(results, neighbor)
	}

	return results
}

// IslandOptimizer 岛屿模型并行优化器
// 多个独立种群并行进化，定期交换最优解
type IslandOptimizer struct {
	config        *OptimizationConfig
	evaluator     ConstraintEvaluator
	islandCount   int
	migrationRate float64
}

// NewIslandOptimizer 创建岛屿模型优化器
func NewIslandOptimizer(config *OptimizationConfig, evaluator ConstraintEvaluator, islandCount int) *IslandOptimizer {
	if islandCount < 2 {
		islandCount = 2
	}
	return &IslandOptimizer{
		config:        config,
		evaluator:     evaluator,
		islandCount:   islandCount,
		migrationRate: 0.1,
	}
}

// Island 岛屿（独立种群）
type Island struct {
	ID        int
	Best      *Solution
	Current   *Solution
	Optimizer *LocalSearchOptimizer
}

// OptimizeIslands 岛屿模型并行优化
func (io *IslandOptimizer) OptimizeIslands(ctx context.Context, initial *Solution, employees []*model.Employee, shifts []*model.Shift) (*Solution, error) {
	// 创建岛屿
	islands := make([]*Island, io.islandCount)
	for i := 0; i < io.islandCount; i++ {
		islands[i] = &Island{
			ID:        i,
			Best:      initial.Clone(),
			Current:   initial.Clone(),
			Optimizer: NewLocalSearchOptimizer(io.config, io.evaluator),
		}
	}

	// 并行运行岛屿
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	for i := 0; i < io.islandCount; i++ {
		wg.Add(1)
		go func(island *Island) {
			defer wg.Done()
			
			result, err := island.Optimizer.Optimize(ctx, island.Current, employees, shifts)
			if err == nil {
				mu.Lock()
				island.Best = result
				mu.Unlock()
			}
		}(islands[i])
	}

	wg.Wait()

	// 找出全局最优解
	globalBest := islands[0].Best
	for _, island := range islands[1:] {
		if island.Best.Score < globalBest.Score {
			globalBest = island.Best
		}
	}

	log.Printf("岛屿模型优化完成: islands=%d, best_score=%.2f", io.islandCount, globalBest.Score)

	return globalBest, nil
}

