-- PaiBan 排班引擎 - 数据库初始化脚本
-- ====================================

-- 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ====================================
-- 基础表
-- ====================================

-- 组织表
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'general', -- restaurant/factory/housekeeping/nursing
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 约束配置表
CREATE TABLE IF NOT EXISTS constraints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES organizations(id),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    category VARCHAR(10) NOT NULL CHECK (category IN ('hard', 'soft')),
    weight INT DEFAULT 50 CHECK (weight >= 1 AND weight <= 100),
    config JSONB NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 场景模板表
CREATE TABLE IF NOT EXISTS scenario_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    scenario VARCHAR(50) NOT NULL, -- restaurant/factory/housekeeping/nursing
    description TEXT,
    constraints JSONB NOT NULL,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ====================================
-- 索引
-- ====================================

CREATE INDEX IF NOT EXISTS idx_constraints_org ON constraints(org_id);
CREATE INDEX IF NOT EXISTS idx_constraints_type ON constraints(type);
CREATE INDEX IF NOT EXISTS idx_templates_scenario ON scenario_templates(scenario);

-- ====================================
-- 初始数据：场景模板
-- ====================================

INSERT INTO scenario_templates (name, scenario, description, constraints, is_default) VALUES
(
    '餐饮门店标准模板',
    'restaurant',
    '适用于餐饮门店的标准约束配置',
    '{
        "hard": [
            {"type": "max_hours_per_day", "config": {"max_hours": 10}},
            {"type": "max_hours_per_week", "config": {"max_hours": 44}},
            {"type": "min_rest_between_shifts", "config": {"min_hours": 10}},
            {"type": "max_consecutive_days", "config": {"max_days": 6}}
        ],
        "soft": [
            {"type": "employee_preference", "weight": 60},
            {"type": "workload_balance", "weight": 80},
            {"type": "minimize_overtime", "weight": 70}
        ]
    }',
    true
),
(
    '工厂三班倒模板',
    'factory',
    '适用于工厂三班倒的约束配置',
    '{
        "hard": [
            {"type": "max_hours_per_day", "config": {"max_hours": 8}},
            {"type": "max_hours_per_week", "config": {"max_hours": 44}},
            {"type": "shift_rotation_pattern", "config": {"pattern": "三班倒"}},
            {"type": "max_consecutive_night_shifts", "config": {"max_nights": 4}}
        ],
        "soft": [
            {"type": "team_together", "weight": 70},
            {"type": "workload_balance", "weight": 60}
        ]
    }',
    true
),
(
    '家政服务模板',
    'housekeeping',
    '适用于家政服务的约束配置',
    '{
        "hard": [
            {"type": "max_hours_per_day", "config": {"max_hours": 10}},
            {"type": "service_area_match", "config": {"max_distance_km": 15}},
            {"type": "travel_time_buffer", "config": {"min_buffer_minutes": 30}},
            {"type": "skill_required", "config": {}}
        ],
        "soft": [
            {"type": "customer_preference", "weight": 90},
            {"type": "minimize_travel_distance", "weight": 70},
            {"type": "workload_balance", "weight": 60}
        ]
    }',
    true
),
(
    '长护险标准模板',
    'nursing',
    '适用于长期护理保险服务的约束配置',
    '{
        "hard": [
            {"type": "certification_level", "config": {}},
            {"type": "care_plan_compliance", "config": {"enforce_frequency": true}},
            {"type": "service_area_match", "config": {"max_distance_km": 10}},
            {"type": "max_patients_per_day", "config": {"max_patients": 6}}
        ],
        "soft": [
            {"type": "caregiver_continuity", "weight": 95},
            {"type": "service_time_regularity", "weight": 85},
            {"type": "workload_balance", "weight": 60}
        ]
    }',
    true
)
ON CONFLICT DO NOTHING;

-- ====================================
-- 完成
-- ====================================

SELECT 'PaiBan 数据库初始化完成' AS message;

