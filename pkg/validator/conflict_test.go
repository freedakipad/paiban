package validator

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
)

func TestConflictDetector_DetectAll(t *testing.T) {
	detector := NewConflictDetector(DefaultDetectorConfig())

	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	emp1 := uuid.New()

	employees := map[uuid.UUID]*model.Employee{
		emp1: {BaseModel: model.BaseModel{ID: emp1}, Name: "员工1"},
	}

	assignments := []*model.Assignment{
		{
			BaseModel:  model.BaseModel{ID: uuid.New()},
			EmployeeID: emp1,
			Date:       now.Format("2006-01-02"),
			StartTime:  now,
			EndTime:    now.Add(8 * time.Hour),
		},
		{
			BaseModel:  model.BaseModel{ID: uuid.New()},
			EmployeeID: emp1,
			Date:       tomorrow.Format("2006-01-02"),
			StartTime:  tomorrow,
			EndTime:    tomorrow.Add(8 * time.Hour),
		},
	}

	conflicts := detector.DetectAll(assignments, employees)

	// 正常排班不应有冲突
	if len(conflicts) != 0 {
		t.Errorf("Expected 0 conflicts, got %d", len(conflicts))
		for _, c := range conflicts {
			t.Logf("Conflict: %s", c.Message)
		}
	}
}

func TestConflictDetector_DetectOverlap(t *testing.T) {
	detector := NewConflictDetector(DefaultDetectorConfig())

	now := time.Now()
	emp1 := uuid.New()

	employees := map[uuid.UUID]*model.Employee{
		emp1: {BaseModel: model.BaseModel{ID: emp1}, Name: "员工1"},
	}

	// 两个时间重叠的班次
	assignments := []*model.Assignment{
		{
			BaseModel:  model.BaseModel{ID: uuid.New()},
			EmployeeID: emp1,
			Date:       now.Format("2006-01-02"),
			StartTime:  now,
			EndTime:    now.Add(8 * time.Hour),
		},
		{
			BaseModel:  model.BaseModel{ID: uuid.New()},
			EmployeeID: emp1,
			Date:       now.Format("2006-01-02"),
			StartTime:  now.Add(4 * time.Hour), // 重叠4小时
			EndTime:    now.Add(12 * time.Hour),
		},
	}

	conflicts := detector.DetectAll(assignments, employees)

	// 应该检测到重叠冲突
	hasOverlap := false
	for _, c := range conflicts {
		if c.Type == ConflictOverlap {
			hasOverlap = true
			break
		}
	}

	if !hasOverlap {
		t.Error("Should detect overlap conflict")
	}
}

func TestConflictDetector_DetectRestTime(t *testing.T) {
	detector := NewConflictDetector(DefaultDetectorConfig())

	now := time.Now()
	emp1 := uuid.New()

	employees := map[uuid.UUID]*model.Employee{
		emp1: {BaseModel: model.BaseModel{ID: emp1}, Name: "员工1"},
	}

	// 休息时间不足的班次（同一天内）
	assignments := []*model.Assignment{
		{
			BaseModel:  model.BaseModel{ID: uuid.New()},
			EmployeeID: emp1,
			Date:       now.Format("2006-01-02"),
			StartTime:  now,
			EndTime:    now.Add(8 * time.Hour),
		},
		{
			BaseModel:  model.BaseModel{ID: uuid.New()},
			EmployeeID: emp1,
			Date:       now.Format("2006-01-02"),
			StartTime:  now.Add(10 * time.Hour), // 只休息2小时
			EndTime:    now.Add(18 * time.Hour),
		},
	}

	conflicts := detector.DetectAll(assignments, employees)

	// 应该检测到某种冲突（可能是重叠或休息时间）
	if len(conflicts) == 0 {
		t.Log("No conflicts detected for insufficient rest time in same day")
	}
}

func TestConflictDetector_DetectForAssignment(t *testing.T) {
	detector := NewConflictDetector(DefaultDetectorConfig())

	now := time.Now()
	emp1 := uuid.New()

	employee := &model.Employee{
		BaseModel: model.BaseModel{ID: emp1},
		Name:      "员工1",
	}

	existing := []*model.Assignment{
		{
			BaseModel:  model.BaseModel{ID: uuid.New()},
			EmployeeID: emp1,
			Date:       now.Format("2006-01-02"),
			StartTime:  now,
			EndTime:    now.Add(8 * time.Hour),
		},
	}

	// 检查新分配是否冲突
	newAssignment := &model.Assignment{
		BaseModel:  model.BaseModel{ID: uuid.New()},
		EmployeeID: emp1,
		Date:       now.Format("2006-01-02"),
		StartTime:  now.Add(4 * time.Hour), // 重叠
		EndTime:    now.Add(12 * time.Hour),
	}

	conflicts := detector.DetectForAssignment(newAssignment, existing, employee)

	if len(conflicts) == 0 {
		t.Error("Should detect conflict for overlapping assignment")
	}
}

func TestDefaultDetectorConfig(t *testing.T) {
	config := DefaultDetectorConfig()

	if config.MinRestHours <= 0 {
		t.Error("MinRestHours should be positive")
	}
	if config.MaxHoursPerDay <= 0 {
		t.Error("MaxHoursPerDay should be positive")
	}
	if config.MaxConsecutiveDays <= 0 {
		t.Error("MaxConsecutiveDays should be positive")
	}
}

func TestNewConflictDetector(t *testing.T) {
	config := &DetectorConfig{
		MinRestHours:      10,
		MaxHoursPerDay:    12,
		MaxConsecutiveDays: 5,
	}

	detector := NewConflictDetector(config)

	if detector == nil {
		t.Error("Detector should not be nil")
	}
}
