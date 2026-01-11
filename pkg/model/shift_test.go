package model

import (
	"testing"
	"time"
)

func TestAssignment_WorkingHours(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected float64
	}{
		{
			name:     "8小时工作",
			start:    time.Date(2026, 1, 11, 9, 0, 0, 0, time.Local),
			end:      time.Date(2026, 1, 11, 17, 0, 0, 0, time.Local),
			expected: 8.0,
		},
		{
			name:     "4小时半工作",
			start:    time.Date(2026, 1, 11, 9, 0, 0, 0, time.Local),
			end:      time.Date(2026, 1, 11, 13, 30, 0, 0, time.Local),
			expected: 4.5,
		},
		{
			name:     "跨天夜班",
			start:    time.Date(2026, 1, 11, 22, 0, 0, 0, time.Local),
			end:      time.Date(2026, 1, 12, 6, 0, 0, 0, time.Local),
			expected: 8.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Assignment{StartTime: tt.start, EndTime: tt.end}
			if result := a.WorkingHours(); result != tt.expected {
				t.Errorf("WorkingHours() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestAssignment_IsOnDate(t *testing.T) {
	a := &Assignment{Date: "2026-01-11"}

	if !a.IsOnDate("2026-01-11") {
		t.Error("应该返回true")
	}
	if a.IsOnDate("2026-01-12") {
		t.Error("应该返回false")
	}
}

func TestShift_DurationHours(t *testing.T) {
	tests := []struct {
		name      string
		duration  int
		breakTime int
		expected  float64
	}{
		{"8小时班含1小时休息", 480, 60, 7.0},
		{"4小时班无休息", 240, 0, 4.0},
		{"6小时班含30分钟休息", 360, 30, 5.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Shift{Duration: tt.duration, BreakTime: tt.breakTime}
			if result := s.DurationHours(); result != tt.expected {
				t.Errorf("DurationHours() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestShift_IsNightShift(t *testing.T) {
	nightShift := &Shift{ShiftType: "night"}
	dayShift := &Shift{ShiftType: "day"}

	if !nightShift.IsNightShift() {
		t.Error("夜班应返回true")
	}
	if dayShift.IsNightShift() {
		t.Error("日班应返回false")
	}
}

func TestShift_IsSplitShift(t *testing.T) {
	splitShift := &Shift{ShiftType: "split"}
	normalShift := &Shift{ShiftType: "morning"}

	if !splitShift.IsSplitShift() {
		t.Error("两头班应返回true")
	}
	if normalShift.IsSplitShift() {
		t.Error("普通班应返回false")
	}
}

