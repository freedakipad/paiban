-- PaiBan 排班引擎 - 回滚初始化Schema
-- Migration: 001_init_schema (DOWN)
-- ====================================

-- 删除触发器
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;
DROP TRIGGER IF EXISTS update_shifts_updated_at ON shifts;
DROP TRIGGER IF EXISTS update_schedules_updated_at ON schedules;
DROP TRIGGER IF EXISTS update_assignments_updated_at ON assignments;
DROP TRIGGER IF EXISTS update_constraints_updated_at ON constraints;
DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;
DROP TRIGGER IF EXISTS update_service_orders_updated_at ON service_orders;
DROP TRIGGER IF EXISTS update_care_plans_updated_at ON care_plans;

-- 删除函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除派出服务相关表
DROP TABLE IF EXISTS customer_employee_history;
DROP TABLE IF EXISTS care_plans;
DROP TABLE IF EXISTS service_records;
DROP TABLE IF EXISTS service_orders;
DROP TABLE IF EXISTS customers;

-- 删除排班相关表
DROP TABLE IF EXISTS swap_requests;
DROP TABLE IF EXISTS scenario_templates;
DROP TABLE IF EXISTS constraints;
DROP TABLE IF EXISTS assignments;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS shift_requirements;
DROP TABLE IF EXISTS shifts;
DROP TABLE IF EXISTS employee_contracts;
DROP TABLE IF EXISTS employee_availability;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS organizations;

-- 删除扩展
DROP EXTENSION IF EXISTS "uuid-ossp";

