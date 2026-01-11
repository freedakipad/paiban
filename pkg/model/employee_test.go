package model

import (
	"testing"
)

func TestEmployee_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"active员工", "active", true},
		{"inactive员工", "inactive", false},
		{"leave员工", "leave", false},
		{"空状态", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Employee{Status: tt.status}
			if result := e.IsActive(); result != tt.expected {
				t.Errorf("IsActive() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestEmployee_HasSkill(t *testing.T) {
	e := &Employee{
		Skills: []string{"cooking", "service", "cleaning"},
	}

	tests := []struct {
		skill    string
		expected bool
	}{
		{"cooking", true},
		{"service", true},
		{"cleaning", true},
		{"management", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.skill, func(t *testing.T) {
			if result := e.HasSkill(tt.skill); result != tt.expected {
				t.Errorf("HasSkill(%s) = %v, expected %v", tt.skill, result, tt.expected)
			}
		})
	}
}

func TestEmployee_HasCertification(t *testing.T) {
	e := &Employee{
		Certifications: []string{"health_cert", "no_criminal_record"},
	}

	tests := []struct {
		cert     string
		expected bool
	}{
		{"health_cert", true},
		{"no_criminal_record", true},
		{"nurse_cert", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.cert, func(t *testing.T) {
			if result := e.HasCertification(tt.cert); result != tt.expected {
				t.Errorf("HasCertification(%s) = %v, expected %v", tt.cert, result, tt.expected)
			}
		})
	}
}

func TestEmployee_CanServeLocation(t *testing.T) {
	// 无限制的员工
	e1 := &Employee{}
	loc := Location{Latitude: 39.9, Longitude: 116.4}
	if !e1.CanServeLocation(loc) {
		t.Error("无限制员工应能服务任何位置")
	}

	// 有距离限制的员工
	e2 := &Employee{
		HomeLocation: &Location{Latitude: 39.9, Longitude: 116.4},
		ServiceArea:  &ServiceArea{MaxRadius: 5},
	}

	// 近距离位置
	nearLoc := Location{Latitude: 39.91, Longitude: 116.41}
	if !e2.CanServeLocation(nearLoc) {
		t.Error("近距离位置应该可服务")
	}

	// 远距离位置
	farLoc := Location{Latitude: 40.5, Longitude: 117.0}
	if e2.CanServeLocation(farLoc) {
		t.Error("远距离位置不应该可服务")
	}
}

