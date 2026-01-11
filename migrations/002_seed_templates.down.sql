-- PaiBan 排班引擎 - 删除场景模板数据
-- Migration: 002_seed_templates (DOWN)
-- ====================================

DELETE FROM scenario_templates WHERE is_default = true;

