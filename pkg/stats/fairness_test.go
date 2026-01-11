package stats

import (
	"testing"
	"time"
)

func TestFairnessAnalyzer_Analyze(t *testing.T) {
	analyzer := NewFairnessAnalyzer()

	employees := []*EmployeeInfo{
		{ID: "emp1", Name: "员工1"},
		{ID: "emp2", Name: "员工2"},
	}

	now := time.Now()
	assignments := []*AssignmentInfo{
		{
			EmployeeID:   "emp1",
			EmployeeName: "员工1",
			Date:         "2026-01-11",
			StartTime:    now,
			EndTime:      now.Add(8 * time.Hour),
		},
		{
			EmployeeID:   "emp1",
			EmployeeName: "员工1",
			Date:         "2026-01-12",
			StartTime:    now,
			EndTime:      now.Add(8 * time.Hour),
		},
		{
			EmployeeID:   "emp2",
			EmployeeName: "员工2",
			Date:         "2026-01-11",
			StartTime:    now,
			EndTime:      now.Add(8 * time.Hour),
		},
	}

	metrics := analyzer.Analyze(assignments, employees)

	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// 员工1有16小时，员工2有8小时，应有一定差异
	if metrics.WorkloadGini < 0 || metrics.WorkloadGini > 1 {
		t.Errorf("Gini coefficient should be between 0 and 1, got %f", metrics.WorkloadGini)
	}

	if len(metrics.EmployeeStats) != 2 {
		t.Errorf("Expected 2 employee stats, got %d", len(metrics.EmployeeStats))
	}
}

func TestFairnessAnalyzer_EmptyInput(t *testing.T) {
	analyzer := NewFairnessAnalyzer()

	metrics := analyzer.Analyze(nil, nil)

	if metrics == nil {
		t.Fatal("Should return empty metrics for nil input")
	}
}

func TestFairnessAnalyzer_PerfectFairness(t *testing.T) {
	analyzer := NewFairnessAnalyzer()

	employees := []*EmployeeInfo{
		{ID: "emp1", Name: "员工1"},
		{ID: "emp2", Name: "员工2"},
	}

	now := time.Now()
	// 完全相同的工时分配
	assignments := []*AssignmentInfo{
		{EmployeeID: "emp1", Date: "2026-01-11", StartTime: now, EndTime: now.Add(8 * time.Hour)},
		{EmployeeID: "emp2", Date: "2026-01-11", StartTime: now, EndTime: now.Add(8 * time.Hour)},
	}

	metrics := analyzer.Analyze(assignments, employees)

	// 完全相同应该Gini=0
	if metrics.WorkloadGini > 0.01 {
		t.Errorf("Perfect fairness should have Gini near 0, got %f", metrics.WorkloadGini)
	}
}

func TestFairnessAnalyzer_OverallScore(t *testing.T) {
	analyzer := NewFairnessAnalyzer()

	employees := []*EmployeeInfo{
		{ID: "emp1", Name: "员工1"},
	}

	now := time.Now()
	assignments := []*AssignmentInfo{
		{EmployeeID: "emp1", Date: "2026-01-11", StartTime: now, EndTime: now.Add(8 * time.Hour)},
	}

	metrics := analyzer.Analyze(assignments, employees)

	// 分数应该在0-100之间
	if metrics.OverallFairnessScore < 0 || metrics.OverallFairnessScore > 100 {
		t.Errorf("Score should be 0-100, got %f", metrics.OverallFairnessScore)
	}
}
