package model

import (
	"testing"
)

func TestLocation_Distance(t *testing.T) {
	tests := []struct {
		name     string
		loc1     Location
		loc2     Location
		expected float64
		delta    float64
	}{
		{
			name: "同一位置",
			loc1: Location{Latitude: 39.9042, Longitude: 116.4074},
			loc2: Location{Latitude: 39.9042, Longitude: 116.4074},
			expected: 0,
			delta:    0.001,
		},
		{
			name: "北京到上海",
			loc1: Location{Latitude: 39.9042, Longitude: 116.4074}, // 北京
			loc2: Location{Latitude: 31.2304, Longitude: 121.4737}, // 上海
			expected: 1066, // 约1066公里
			delta:    10,
		},
		{
			name: "短距离",
			loc1: Location{Latitude: 39.9042, Longitude: 116.4074},
			loc2: Location{Latitude: 39.9142, Longitude: 116.4174}, // 约1.4公里
			expected: 1.4,
			delta:    0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.loc1.Distance(tt.loc2)
			if result < tt.expected-tt.delta || result > tt.expected+tt.delta {
				t.Errorf("Distance() = %v, expected %v ± %v", result, tt.expected, tt.delta)
			}
		})
	}
}

func TestNewBaseModel(t *testing.T) {
	base := NewBaseModel()
	
	if base.ID.String() == "" {
		t.Error("ID should not be empty")
	}
	if base.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if base.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

