package stats

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCoverageAnalyzer_Analyze(t *testing.T) {
	analyzer := NewCoverageAnalyzer()

	now := time.Now()
	shift1ID := uuid.New().String()
	shift2ID := uuid.New().String()

	shifts := []*ShiftInfo{
		{
			ID:        shift1ID,
			Date:      "2026-01-11",
			Type:      "morning",
			StartTime: now,
			EndTime:   now.Add(8 * time.Hour),
		},
		{
			ID:        shift2ID,
			Date:      "2026-01-11",
			Type:      "evening",
			StartTime: now.Add(8 * time.Hour),
			EndTime:   now.Add(16 * time.Hour),
		},
	}

	assignments := []*AssignmentInfo{
		{
			ShiftID:      shift1ID,
			EmployeeID:   uuid.New().String(),
			EmployeeName: "员工1",
			Date:         "2026-01-11",
			StartTime:    now,
			EndTime:      now.Add(8 * time.Hour),
		},
	}

	metrics := analyzer.Analyze(shifts, assignments)

	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// 2个班次，1个被分配，覆盖率应为50%
	if metrics.OverallCoverage != 50 {
		t.Errorf("Expected 50%% coverage, got %.1f%%", metrics.OverallCoverage)
	}

	if len(metrics.UncoveredShifts) != 1 {
		t.Errorf("Expected 1 uncovered shift, got %d", len(metrics.UncoveredShifts))
	}
}

func TestCoverageAnalyzer_FullCoverage(t *testing.T) {
	analyzer := NewCoverageAnalyzer()

	now := time.Now()
	shiftID := uuid.New().String()

	shifts := []*ShiftInfo{
		{ID: shiftID, Date: "2026-01-11", Type: "morning", StartTime: now, EndTime: now.Add(8 * time.Hour)},
	}

	assignments := []*AssignmentInfo{
		{ShiftID: shiftID, EmployeeID: uuid.New().String(), Date: "2026-01-11", StartTime: now, EndTime: now.Add(8 * time.Hour)},
	}

	metrics := analyzer.Analyze(shifts, assignments)

	if metrics.OverallCoverage != 100 {
		t.Errorf("Expected 100%% coverage, got %.1f%%", metrics.OverallCoverage)
	}

	if len(metrics.UncoveredShifts) != 0 {
		t.Errorf("Expected 0 uncovered shifts, got %d", len(metrics.UncoveredShifts))
	}
}

func TestCoverageAnalyzer_EmptyInput(t *testing.T) {
	analyzer := NewCoverageAnalyzer()

	metrics := analyzer.Analyze(nil, nil)

	if metrics == nil {
		t.Fatal("Should return metrics for nil input")
	}

	if metrics.OverallCoverage != 100 {
		t.Errorf("Empty shifts should have 100%% coverage, got %.1f%%", metrics.OverallCoverage)
	}
}

func TestCoverageAnalyzer_SetMinStaffRequirements(t *testing.T) {
	analyzer := NewCoverageAnalyzer()

	requirements := map[int]int{
		9:  3, // 9点需要3人
		10: 4, // 10点需要4人
	}

	analyzer.SetMinStaffRequirements(requirements)

	// 验证设置成功
	if len(analyzer.minStaffPerHour) != 2 {
		t.Errorf("Expected 2 requirements, got %d", len(analyzer.minStaffPerHour))
	}
}

func TestCoverageAnalyzer_DailyCoverage(t *testing.T) {
	analyzer := NewCoverageAnalyzer()

	now := time.Now()
	shift1 := uuid.New().String()
	shift2 := uuid.New().String()

	shifts := []*ShiftInfo{
		{ID: shift1, Date: "2026-01-11", Type: "morning", StartTime: now, EndTime: now.Add(8 * time.Hour)},
		{ID: shift2, Date: "2026-01-12", Type: "morning", StartTime: now, EndTime: now.Add(8 * time.Hour)},
	}

	assignments := []*AssignmentInfo{
		{ShiftID: shift1, EmployeeID: "emp1", Date: "2026-01-11", StartTime: now, EndTime: now.Add(8 * time.Hour)},
	}

	metrics := analyzer.Analyze(shifts, assignments)

	// 检查每日覆盖情况
	if len(metrics.DailyCoverage) != 2 {
		t.Errorf("Expected 2 daily coverage entries, got %d", len(metrics.DailyCoverage))
	}
}
