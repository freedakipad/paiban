// Package optimizer 提供排班优化算法
package optimizer

import (
	"math/rand"
	"time"

	"github.com/paiban/paiban/pkg/model"
)

// MoveType 邻域移动类型
type MoveType int

const (
	MoveSwap      MoveType = iota // 交换两个员工的班次
	MoveRelocate                  // 重新分配员工到不同班次
	MoveInsert                    // 插入新分配
	MoveRemove                    // 移除分配
	Move2Opt                      // 2-opt改进
	MoveChain                     // 链式移动
)

// Move 邻域移动操作
type Move struct {
	Type        MoveType
	Assignment1 *model.Assignment
	Assignment2 *model.Assignment
	Employee    *model.Employee
	Shift       *model.Shift
	Index1      int
	Index2      int
}

// NeighborhoodGenerator 邻域生成器
type NeighborhoodGenerator struct {
	rng        *rand.Rand
	moveWeights map[MoveType]float64
}

// NewNeighborhoodGenerator 创建邻域生成器
func NewNeighborhoodGenerator() *NeighborhoodGenerator {
	return &NeighborhoodGenerator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		moveWeights: map[MoveType]float64{
			MoveSwap:     0.35, // 35% 交换
			MoveRelocate: 0.30, // 30% 重新分配
			MoveInsert:   0.15, // 15% 插入
			MoveRemove:   0.10, // 10% 移除
			Move2Opt:     0.05, // 5% 2-opt
			MoveChain:    0.05, // 5% 链式移动
		},
	}
}

// GenerateNeighbor 生成邻域解
func (n *NeighborhoodGenerator) GenerateNeighbor(current *Solution, employees []*model.Employee, shifts []*model.Shift) *Solution {
	if current == nil || len(current.Assignments) == 0 {
		return nil
	}

	moveType := n.selectMoveType()
	
	switch moveType {
	case MoveSwap:
		return n.generateSwapMove(current, employees)
	case MoveRelocate:
		return n.generateRelocateMove(current, employees, shifts)
	case MoveInsert:
		return n.generateInsertMove(current, employees, shifts)
	case MoveRemove:
		return n.generateRemoveMove(current)
	case Move2Opt:
		return n.generate2OptMove(current)
	case MoveChain:
		return n.generateChainMove(current, employees)
	default:
		return n.generateSwapMove(current, employees)
	}
}

// selectMoveType 按权重选择移动类型
func (n *NeighborhoodGenerator) selectMoveType() MoveType {
	r := n.rng.Float64()
	cumulative := 0.0
	
	for moveType, weight := range n.moveWeights {
		cumulative += weight
		if r < cumulative {
			return moveType
		}
	}
	
	return MoveSwap
}

// generateSwapMove 生成交换移动
// 交换两个员工在不同班次上的分配
func (n *NeighborhoodGenerator) generateSwapMove(current *Solution, _ []*model.Employee) *Solution {
	if len(current.Assignments) < 2 {
		return nil
	}

	neighbor := current.Clone()
	
	// 随机选择两个分配
	i := n.rng.Intn(len(neighbor.Assignments))
	j := n.rng.Intn(len(neighbor.Assignments))
	for j == i {
		j = n.rng.Intn(len(neighbor.Assignments))
	}

	// 交换员工ID
	neighbor.Assignments[i].EmployeeID, neighbor.Assignments[j].EmployeeID = 
		neighbor.Assignments[j].EmployeeID, neighbor.Assignments[i].EmployeeID

	return neighbor
}

// generateRelocateMove 生成重新分配移动
// 将某个员工分配到不同的班次
func (n *NeighborhoodGenerator) generateRelocateMove(current *Solution, _ []*model.Employee, shifts []*model.Shift) *Solution {
	if len(current.Assignments) == 0 || len(shifts) == 0 {
		return nil
	}

	neighbor := current.Clone()
	
	// 随机选择一个分配
	idx := n.rng.Intn(len(neighbor.Assignments))
	assignment := neighbor.Assignments[idx]
	
	// 随机选择一个新班次
	newShift := shifts[n.rng.Intn(len(shifts))]
	
	// 检查是否是不同班次
	if assignment.ShiftID == newShift.ID {
		return nil
	}
	
	// 更新分配
	assignment.ShiftID = newShift.ID

	return neighbor
}

// generateInsertMove 生成插入移动
// 为未分配的班次添加员工
func (n *NeighborhoodGenerator) generateInsertMove(current *Solution, employees []*model.Employee, shifts []*model.Shift) *Solution {
	if len(employees) == 0 || len(shifts) == 0 {
		return nil
	}

	neighbor := current.Clone()
	
	// 查找未分配的班次
	assignedShifts := make(map[string]bool)
	for _, a := range neighbor.Assignments {
		assignedShifts[a.ShiftID.String()] = true
	}
	
	var unassignedShifts []*model.Shift
	for _, s := range shifts {
		if !assignedShifts[s.ID.String()] {
			unassignedShifts = append(unassignedShifts, s)
		}
	}
	
	if len(unassignedShifts) == 0 {
		return nil
	}
	
	// 随机选择未分配的班次和员工
	shift := unassignedShifts[n.rng.Intn(len(unassignedShifts))]
	employee := employees[n.rng.Intn(len(employees))]
	
	// 创建新分配
	newAssignment := &model.Assignment{
		ShiftID:    shift.ID,
		EmployeeID: employee.ID,
		Status:     "scheduled",
	}
	
	neighbor.Assignments = append(neighbor.Assignments, newAssignment)
	return neighbor
}

// generateRemoveMove 生成移除移动
// 移除某个分配
func (n *NeighborhoodGenerator) generateRemoveMove(current *Solution) *Solution {
	if len(current.Assignments) <= 1 {
		return nil
	}

	neighbor := current.Clone()
	
	// 随机选择一个分配移除
	idx := n.rng.Intn(len(neighbor.Assignments))
	neighbor.Assignments = append(neighbor.Assignments[:idx], neighbor.Assignments[idx+1:]...)
	
	return neighbor
}

// generate2OptMove 生成2-opt移动
// 反转分配序列的一段
func (n *NeighborhoodGenerator) generate2OptMove(current *Solution) *Solution {
	if len(current.Assignments) < 4 {
		return nil
	}

	neighbor := current.Clone()
	
	// 随机选择两个位置
	i := n.rng.Intn(len(neighbor.Assignments) - 1)
	j := i + 2 + n.rng.Intn(len(neighbor.Assignments)-i-2)
	if j >= len(neighbor.Assignments) {
		j = len(neighbor.Assignments) - 1
	}
	
	// 反转i到j之间的序列
	for left, right := i, j; left < right; left, right = left+1, right-1 {
		neighbor.Assignments[left], neighbor.Assignments[right] = 
			neighbor.Assignments[right], neighbor.Assignments[left]
	}
	
	return neighbor
}

// generateChainMove 生成链式移动
// 将多个分配进行链式重新分配
func (n *NeighborhoodGenerator) generateChainMove(current *Solution, _ []*model.Employee) *Solution {
	if len(current.Assignments) < 3 {
		return nil
	}

	neighbor := current.Clone()
	
	// 随机选择链长度 (2-4)
	chainLen := 2 + n.rng.Intn(3)
	if chainLen > len(neighbor.Assignments) {
		chainLen = len(neighbor.Assignments)
	}
	
	// 随机选择链的起始位置
	indices := make([]int, chainLen)
	for i := 0; i < chainLen; i++ {
		indices[i] = n.rng.Intn(len(neighbor.Assignments))
	}
	
	// 链式移动员工ID
	firstEmployee := neighbor.Assignments[indices[0]].EmployeeID
	
	for i := 0; i < chainLen-1; i++ {
		neighbor.Assignments[indices[i]].EmployeeID = neighbor.Assignments[indices[i+1]].EmployeeID
	}
	
	neighbor.Assignments[indices[chainLen-1]].EmployeeID = firstEmployee
	
	return neighbor
}

// GenerateBatch 批量生成邻域解
func (n *NeighborhoodGenerator) GenerateBatch(current *Solution, employees []*model.Employee, shifts []*model.Shift, count int) []*Solution {
	results := make([]*Solution, 0, count)
	
	for i := 0; i < count; i++ {
		neighbor := n.GenerateNeighbor(current, employees, shifts)
		if neighbor != nil {
			results = append(results, neighbor)
		}
	}
	
	return results
}

// SetMoveWeights 设置移动类型权重
func (n *NeighborhoodGenerator) SetMoveWeights(weights map[MoveType]float64) {
	n.moveWeights = weights
}

