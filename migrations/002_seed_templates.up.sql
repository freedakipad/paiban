-- PaiBan 排班引擎 - 初始化场景模板数据
-- Migration: 002_seed_templates
-- ====================================

-- 餐饮门店标准模板
INSERT INTO scenario_templates (name, scenario, description, constraints, is_default) VALUES
(
    '餐饮门店标准模板',
    'restaurant',
    '适用于餐饮门店的标准约束配置，支持高峰期排班、两头班等特殊需求',
    '{
        "hard": [
            {"type": "max_hours_per_day", "config": {"max_hours": 10}},
            {"type": "max_hours_per_week", "config": {"max_hours": 44}},
            {"type": "min_rest_between_shifts", "config": {"min_hours": 10}},
            {"type": "max_consecutive_days", "config": {"max_days": 6}},
            {"type": "skill_required", "config": {}}
        ],
        "soft": [
            {"type": "employee_preference", "weight": 60, "config": {}},
            {"type": "workload_balance", "weight": 80, "config": {}},
            {"type": "minimize_overtime", "weight": 70, "config": {}},
            {"type": "peak_hours_coverage", "weight": 90, "config": {"peak_hours": ["11:00-13:00", "17:00-20:00"]}}
        ]
    }',
    true
),
-- 工厂三班倒模板
(
    '工厂三班倒模板',
    'factory',
    '适用于工厂三班倒的约束配置，确保产线连续运转和员工轮班公平',
    '{
        "hard": [
            {"type": "max_hours_per_day", "config": {"max_hours": 8}},
            {"type": "max_hours_per_week", "config": {"max_hours": 44}},
            {"type": "shift_rotation_pattern", "config": {"pattern": "三班倒", "rotation_days": 7}},
            {"type": "max_consecutive_night_shifts", "config": {"max_nights": 4}},
            {"type": "min_rest_between_shifts", "config": {"min_hours": 12}},
            {"type": "production_line_coverage", "config": {}}
        ],
        "soft": [
            {"type": "team_together", "weight": 70, "config": {}},
            {"type": "workload_balance", "weight": 60, "config": {}},
            {"type": "skill_match", "weight": 75, "config": {}},
            {"type": "minimize_shift_changes", "weight": 50, "config": {}}
        ]
    }',
    true
),
-- 家政服务模板
(
    '家政服务标准模板',
    'housekeeping',
    '适用于家政服务的派单约束配置，优化服务质量和路线效率',
    '{
        "hard": [
            {"type": "max_hours_per_day", "config": {"max_hours": 10}},
            {"type": "service_area_match", "config": {"max_distance_km": 15}},
            {"type": "travel_time_buffer", "config": {"min_buffer_minutes": 30}},
            {"type": "skill_required", "config": {}},
            {"type": "max_orders_per_day", "config": {"max_orders": 6}}
        ],
        "soft": [
            {"type": "customer_preference", "weight": 90, "config": {}},
            {"type": "minimize_travel_distance", "weight": 70, "config": {}},
            {"type": "workload_balance", "weight": 60, "config": {}},
            {"type": "service_continuity", "weight": 85, "config": {"prefer_same_worker": true}},
            {"type": "time_slot_preference", "weight": 65, "config": {}}
        ]
    }',
    true
),
-- 长护险标准模板
(
    '长护险服务模板',
    'nursing',
    '适用于长期护理保险服务的约束配置，强调服务连续性和护理质量',
    '{
        "hard": [
            {"type": "certification_level", "config": {"require_match": true}},
            {"type": "care_plan_compliance", "config": {"enforce_frequency": true, "enforce_duration": true}},
            {"type": "service_area_match", "config": {"max_distance_km": 10}},
            {"type": "max_patients_per_day", "config": {"max_patients": 6}},
            {"type": "min_service_interval", "config": {"min_hours": 2}}
        ],
        "soft": [
            {"type": "caregiver_continuity", "weight": 95, "config": {"primary_carer_bonus": 50}},
            {"type": "service_time_regularity", "weight": 85, "config": {"time_variance_penalty": 10}},
            {"type": "workload_balance", "weight": 60, "config": {}},
            {"type": "patient_preference", "weight": 80, "config": {}},
            {"type": "minimize_travel", "weight": 55, "config": {}}
        ]
    }',
    true
)
ON CONFLICT DO NOTHING;

