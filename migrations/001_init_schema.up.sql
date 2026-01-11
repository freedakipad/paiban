-- PaiBan 排班引擎 - 初始化数据库Schema
-- Migration: 001_init_schema
-- ====================================

-- 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ====================================
-- 基础表
-- ====================================

-- 组织/机构表
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'general' CHECK (type IN ('restaurant', 'factory', 'housekeeping', 'nursing', 'general')),
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 员工表
CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'leave')),
    hire_date DATE NOT NULL,
    position VARCHAR(100),
    skills JSONB DEFAULT '[]',
    certifications JSONB DEFAULT '[]',
    hourly_rate DECIMAL(10, 2) DEFAULT 0,
    preferences JSONB,
    service_area JSONB,
    home_location JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (org_id, code)
);

-- 员工可用性表
CREATE TABLE IF NOT EXISTS employee_availability (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('available', 'unavailable', 'preferred')),
    time_ranges JSONB,
    reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (employee_id, date)
);

-- 员工合同表
CREATE TABLE IF NOT EXISTS employee_contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    contract_type VARCHAR(20) NOT NULL CHECK (contract_type IN ('full_time', 'part_time', 'temp')),
    min_hours_per_week INT DEFAULT 0,
    max_hours_per_week INT DEFAULT 44,
    max_hours_per_day INT DEFAULT 10,
    max_overtime_per_week INT DEFAULT 12,
    rest_days_per_week INT DEFAULT 1,
    valid_from DATE NOT NULL,
    valid_to DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 班次定义表
CREATE TABLE IF NOT EXISTS shifts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    duration INT NOT NULL, -- 分钟
    break_time INT DEFAULT 0, -- 休息时间（分钟）
    shift_type VARCHAR(20) NOT NULL DEFAULT 'regular' CHECK (shift_type IN ('morning', 'afternoon', 'evening', 'night', 'split', 'regular')),
    color VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (org_id, code)
);

-- 班次需求表
CREATE TABLE IF NOT EXISTS shift_requirements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    shift_id UUID NOT NULL REFERENCES shifts(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    position VARCHAR(100),
    min_employees INT NOT NULL DEFAULT 1,
    max_employees INT NOT NULL DEFAULT 10,
    opt_employees INT, -- 最优人数
    skills JSONB DEFAULT '[]',
    priority INT DEFAULT 5 CHECK (priority >= 1 AND priority <= 10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 排班计划表
CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    version INT DEFAULT 1,
    created_by UUID,
    published_at TIMESTAMP WITH TIME ZONE,
    constraint_config JSONB,
    statistics JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 排班分配表
CREATE TABLE IF NOT EXISTS assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    shift_id UUID NOT NULL REFERENCES shifts(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    position VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'confirmed', 'completed', 'cancelled')),
    is_overtime BOOLEAN DEFAULT false,
    is_swapped BOOLEAN DEFAULT false,
    original_employee_id UUID REFERENCES employees(id),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 约束配置表
CREATE TABLE IF NOT EXISTS constraints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(100) NOT NULL,
    category VARCHAR(10) NOT NULL CHECK (category IN ('hard', 'soft')),
    weight INT DEFAULT 50 CHECK (weight >= 1 AND weight <= 100),
    config JSONB NOT NULL DEFAULT '{}',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 场景模板表
CREATE TABLE IF NOT EXISTS scenario_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    scenario VARCHAR(50) NOT NULL CHECK (scenario IN ('restaurant', 'factory', 'housekeeping', 'nursing')),
    description TEXT,
    constraints JSONB NOT NULL,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 换班请求表
CREATE TABLE IF NOT EXISTS swap_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    requestor_id UUID NOT NULL REFERENCES employees(id),
    target_id UUID REFERENCES employees(id),
    source_assignment UUID NOT NULL REFERENCES assignments(id),
    target_assignment UUID REFERENCES assignments(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    reason TEXT,
    reviewed_by UUID REFERENCES employees(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ====================================
-- 派出服务相关表
-- ====================================

-- 客户表
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    address TEXT NOT NULL,
    location JSONB,
    type VARCHAR(20) NOT NULL DEFAULT 'individual' CHECK (type IN ('individual', 'household', 'enterprise')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    level VARCHAR(20),
    notes TEXT,
    preferences JSONB,
    preferred_emp_ids JSONB DEFAULT '[]',
    blocked_emp_ids JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (org_id, code)
);

-- 服务订单表
CREATE TABLE IF NOT EXISTS service_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    order_no VARCHAR(50) NOT NULL,
    service_type VARCHAR(50) NOT NULL,
    service_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    duration INT NOT NULL, -- 分钟
    address TEXT NOT NULL,
    location JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'assigned', 'in_progress', 'completed', 'cancelled')),
    employee_id UUID REFERENCES employees(id),
    skills JSONB DEFAULT '[]',
    priority INT DEFAULT 5 CHECK (priority >= 1 AND priority <= 10),
    notes TEXT,
    amount DECIMAL(10, 2) DEFAULT 0,
    assigned_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (org_id, order_no)
);

-- 服务记录表
CREATE TABLE IF NOT EXISTS service_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    check_in_time TIMESTAMP WITH TIME ZONE,
    check_out_time TIMESTAMP WITH TIME ZONE,
    check_in_location JSONB,
    check_out_location JSONB,
    actual_minutes INT,
    service_items JSONB DEFAULT '[]',
    rating INT CHECK (rating >= 1 AND rating <= 5),
    feedback TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'checked_in', 'checked_out', 'verified')),
    photos JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 护理计划表（长护险）
CREATE TABLE IF NOT EXISTS care_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    plan_no VARCHAR(50) NOT NULL UNIQUE,
    level INT NOT NULL CHECK (level >= 1 AND level <= 6),
    start_date DATE NOT NULL,
    end_date DATE,
    weekly_hours INT NOT NULL,
    service_items JSONB NOT NULL DEFAULT '[]',
    frequency VARCHAR(50) NOT NULL,
    primary_carer_id UUID REFERENCES employees(id),
    backup_carer_ids JSONB DEFAULT '[]',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'expired')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 客户-员工服务历史表
CREATE TABLE IF NOT EXISTS customer_employee_history (
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    service_count INT DEFAULT 0,
    total_minutes INT DEFAULT 0,
    avg_rating DECIMAL(3, 2),
    last_service_at TIMESTAMP WITH TIME ZONE,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (customer_id, employee_id)
);

-- ====================================
-- 索引
-- ====================================

CREATE INDEX IF NOT EXISTS idx_employees_org ON employees(org_id);
CREATE INDEX IF NOT EXISTS idx_employees_status ON employees(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employee_availability_date ON employee_availability(date);
CREATE INDEX IF NOT EXISTS idx_shifts_org ON shifts(org_id);
CREATE INDEX IF NOT EXISTS idx_shift_requirements_date ON shift_requirements(date);
CREATE INDEX IF NOT EXISTS idx_schedules_org_dates ON schedules(org_id, start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_assignments_schedule ON assignments(schedule_id);
CREATE INDEX IF NOT EXISTS idx_assignments_employee_date ON assignments(employee_id, date);
CREATE INDEX IF NOT EXISTS idx_assignments_date ON assignments(date);
CREATE INDEX IF NOT EXISTS idx_constraints_org ON constraints(org_id);
CREATE INDEX IF NOT EXISTS idx_templates_scenario ON scenario_templates(scenario);
CREATE INDEX IF NOT EXISTS idx_customers_org ON customers(org_id);
CREATE INDEX IF NOT EXISTS idx_service_orders_customer ON service_orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_service_orders_date ON service_orders(service_date);
CREATE INDEX IF NOT EXISTS idx_service_orders_status ON service_orders(status);
CREATE INDEX IF NOT EXISTS idx_service_records_order ON service_records(order_id);
CREATE INDEX IF NOT EXISTS idx_care_plans_customer ON care_plans(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_employee_history ON customer_employee_history(customer_id, employee_id);

-- ====================================
-- 触发器：自动更新 updated_at
-- ====================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON employees FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_shifts_updated_at BEFORE UPDATE ON shifts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_schedules_updated_at BEFORE UPDATE ON schedules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_assignments_updated_at BEFORE UPDATE ON assignments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_constraints_updated_at BEFORE UPDATE ON constraints FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_service_orders_updated_at BEFORE UPDATE ON service_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_care_plans_updated_at BEFORE UPDATE ON care_plans FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

