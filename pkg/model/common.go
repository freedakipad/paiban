// Package model 定义排班引擎的核心数据模型
package model

import (
	"math"
	"time"

	"github.com/google/uuid"
)

// ScenarioType 场景类型
type ScenarioType string

const (
	ScenarioRestaurant   ScenarioType = "restaurant"   // 餐饮
	ScenarioFactory      ScenarioType = "factory"      // 工厂
	ScenarioHousekeeping ScenarioType = "housekeeping" // 家政
	ScenarioNursing      ScenarioType = "nursing"      // 长护险
)

// ConstraintCategory 约束类别
type ConstraintCategory string

const (
	ConstraintHard ConstraintCategory = "hard" // 硬约束（必须满足）
	ConstraintSoft ConstraintCategory = "soft" // 软约束（尽量满足）
)

// BaseModel 基础模型（包含通用字段）
type BaseModel struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

// NewBaseModel 创建新的基础模型
func NewBaseModel() BaseModel {
	now := time.Now()
	return BaseModel{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Organization 组织/机构
type Organization struct {
	BaseModel
	Name     string       `json:"name" db:"name"`
	Code     string       `json:"code" db:"code"`
	Type     ScenarioType `json:"type" db:"type"`
	Settings JSONMap      `json:"settings" db:"settings"`
}

// JSONMap 用于存储 JSONB 数据
type JSONMap map[string]interface{}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Duration 返回时间范围的持续时间
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// Overlaps 检查两个时间范围是否重叠
func (tr TimeRange) Overlaps(other TimeRange) bool {
	return tr.Start.Before(other.End) && other.Start.Before(tr.End)
}

// Contains 检查时间范围是否包含某个时间点
func (tr TimeRange) Contains(t time.Time) bool {
	return !t.Before(tr.Start) && t.Before(tr.End)
}

// DateRange 日期范围
type DateRange struct {
	StartDate string `json:"start_date"` // YYYY-MM-DD
	EndDate   string `json:"end_date"`   // YYYY-MM-DD
}

// Location 地理位置
type Location struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city,omitempty"`
	District  string  `json:"district,omitempty"`
}

// Distance 计算两个位置之间的距离（公里）
// 使用 Haversine 公式
func (l Location) Distance(other Location) float64 {
	const earthRadius = 6371.0 // 地球半径（公里）

	lat1Rad := l.Latitude * math.Pi / 180
	lat2Rad := other.Latitude * math.Pi / 180
	deltaLat := (other.Latitude - l.Latitude) * math.Pi / 180
	deltaLon := (other.Longitude - l.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
