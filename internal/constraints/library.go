// Package constraints 约束系统
package constraints

// ConstraintParam 约束参数定义
type ConstraintParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // int, float, string, bool, array
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
	Min         string `json:"min,omitempty"`
	Max         string `json:"max,omitempty"`
}

// ConstraintDefinition 约束定义
type ConstraintDefinition struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	Type        string            `json:"type"`     // hard 硬约束, soft 软约束
	Category    string            `json:"category"` // 分类
	Description string            `json:"description"`
	Scenarios   []string          `json:"scenarios"` // 适用场景
	Params      []ConstraintParam `json:"params"`
}

// LibraryResponse 约束库响应
type LibraryResponse struct {
	Library []ConstraintDefinition `json:"library"`
}

// GetLibrary 获取完整的约束库
func GetLibrary() []ConstraintDefinition {
	return []ConstraintDefinition{
		// =====================================================
		// 通用硬约束
		// =====================================================
		{
			Name:        "max_hours_per_day",
			DisplayName: "每日最大工时",
			Type:        "hard",
			Category:    "工时限制",
			Description: "限制员工每天的最大工作时长，超过则排班无效。适用于所有行业的基础劳动法规要求。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "max_hours", Type: "int", Description: "最大工时(小时)", Default: "10", Min: "6", Max: "14"},
			},
		},
		{
			Name:        "max_hours_per_week",
			DisplayName: "每周最大工时",
			Type:        "hard",
			Category:    "工时限制",
			Description: "限制员工每周的累计工作时长，确保符合劳动法规定。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "max_hours", Type: "int", Description: "最大工时(小时)", Default: "44", Min: "36", Max: "60"},
			},
		},
		{
			Name:        "min_hours_per_week",
			DisplayName: "每周最小工时",
			Type:        "soft",
			Category:    "工时限制",
			Description: "确保员工每周有足够的排班，保障基本工作量。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "min_hours", Type: "int", Description: "最小工时(小时)", Default: "20", Min: "8", Max: "40"},
				{Name: "weight", Type: "int", Description: "优化权重", Default: "50", Min: "0", Max: "100"},
			},
		},
		{
			Name:        "min_rest_between_shifts",
			DisplayName: "班次间最小休息时间",
			Type:        "hard",
			Category:    "休息保障",
			Description: "确保员工在两个班次之间有足够的休息时间，防止过度疲劳。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "min_hours", Type: "int", Description: "最小休息时间(小时)", Default: "11", Min: "8", Max: "14"},
			},
		},
		{
			Name:        "max_consecutive_days",
			DisplayName: "最大连续工作天数",
			Type:        "hard",
			Category:    "休息保障",
			Description: "限制员工连续工作的最大天数，确保员工有足够的休息日。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "max_days", Type: "int", Description: "最大连续天数", Default: "6", Min: "4", Max: "7"},
			},
		},
		{
			Name:        "skill_required",
			DisplayName: "技能与岗位匹配",
			Type:        "hard",
			Category:    "资质要求",
			Description: "确保分配的员工具备该岗位所需的技能和资质。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params:      []ConstraintParam{},
		},
		{
			Name:        "industry_certification",
			DisplayName: "行业资质认证",
			Type:        "hard",
			Category:    "资质要求",
			Description: "检查员工是否持有行业必需的资质证书（如健康证、护理证等）。",
			Scenarios:   []string{"restaurant", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "required_certs", Type: "array", Description: "必需证书列表", Default: "健康证"},
			},
		},
		{
			Name:        "employee_unavailable",
			DisplayName: "员工不可用时间",
			Type:        "hard",
			Category:    "时间限制",
			Description: "在员工标记为不可用的时间段内不进行排班（如请假、个人事务）。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params:      []ConstraintParam{},
		},
		{
			Name:        "fixed_shift",
			DisplayName: "固定班次约束",
			Type:        "hard",
			Category:    "排班模式",
			Description: "部分员工有固定的班次安排（如只上早班或只上夜班）。",
			Scenarios:   []string{"restaurant", "factory"},
			Params:      []ConstraintParam{},
		},

		// =====================================================
		// 通用软约束
		// =====================================================
		{
			Name:        "workload_balance",
			DisplayName: "工作量均衡",
			Type:        "soft",
			Category:    "公平性",
			Description: "尽量使各员工的工作量分布均匀，提高公平性和员工满意度。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "60", Min: "0", Max: "100"},
				{Name: "tolerance", Type: "float", Description: "容忍偏差百分比", Default: "20", Min: "5", Max: "50"},
			},
		},
		{
			Name:        "employee_preference",
			DisplayName: "员工偏好考虑",
			Type:        "soft",
			Category:    "偏好",
			Description: "尽量满足员工对班次、休息日等的个人偏好。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "50", Min: "0", Max: "100"},
			},
		},
		{
			Name:        "minimize_overtime",
			DisplayName: "减少加班",
			Type:        "soft",
			Category:    "成本优化",
			Description: "优化排班以减少加班时间，降低人力成本。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "70", Min: "0", Max: "100"},
				{Name: "standard_hours", Type: "int", Description: "标准工时(周)", Default: "40"},
			},
		},
		{
			Name:        "senior_junior_pair",
			DisplayName: "新老搭配",
			Type:        "soft",
			Category:    "协作",
			Description: "尽量安排老员工与新员工搭配工作，促进经验传承和培训。",
			Scenarios:   []string{"restaurant", "factory", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "40", Min: "0", Max: "100"},
				{Name: "senior_months", Type: "int", Description: "老员工工龄门槛(月)", Default: "12"},
			},
		},
		{
			Name:        "holiday_handling",
			DisplayName: "法定假日处理",
			Type:        "soft",
			Category:    "休息保障",
			Description: "在法定假日优先安排愿意加班的员工，并提供加班补贴。",
			Scenarios:   []string{"restaurant", "factory", "housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "80", Min: "0", Max: "100"},
				{Name: "bonus_rate", Type: "float", Description: "假日加班倍率", Default: "2.0", Min: "1.0", Max: "3.0"},
			},
		},

		// =====================================================
		// 餐饮行业特有约束
		// =====================================================
		{
			Name:        "peak_hours_coverage",
			DisplayName: "高峰期人员覆盖",
			Type:        "soft",
			Category:    "服务保障",
			Description: "确保在用餐高峰期有足够的员工在岗，提供优质服务。",
			Scenarios:   []string{"restaurant"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "90", Min: "0", Max: "100"},
				{Name: "peak_hours", Type: "array", Description: "高峰时段", Default: "11:00-13:00,17:00-20:00"},
				{Name: "min_staff", Type: "int", Description: "最少员工数", Default: "3", Min: "1", Max: "10"},
			},
		},
		{
			Name:        "split_shift",
			DisplayName: "两头班支持",
			Type:        "soft",
			Category:    "排班模式",
			Description: "允许员工上午和晚间分别上班（两头班），中间休息。适合餐饮业高峰期排班。",
			Scenarios:   []string{"restaurant"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "60", Min: "0", Max: "100"},
				{Name: "max_per_week", Type: "int", Description: "每周最多两头班次数", Default: "2", Min: "0", Max: "5"},
				{Name: "allow", Type: "bool", Description: "是否允许两头班", Default: "true"},
			},
		},
		{
			Name:        "position_coverage",
			DisplayName: "岗位覆盖",
			Type:        "soft",
			Category:    "服务保障",
			Description: "确保各时段都有足够的不同岗位员工（如收银、服务员、厨师等）。",
			Scenarios:   []string{"restaurant"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "80", Min: "0", Max: "100"},
			},
		},

		// =====================================================
		// 工厂产线特有约束
		// =====================================================
		{
			Name:        "shift_rotation",
			DisplayName: "倒班轮换规则",
			Type:        "hard",
			Category:    "排班模式",
			Description: "设定倒班模式（如三班倒），确保班次按规律轮换。",
			Scenarios:   []string{"factory"},
			Params: []ConstraintParam{
				{Name: "pattern", Type: "string", Description: "轮换模式", Default: "三班倒"},
				{Name: "rotation_days", Type: "int", Description: "轮换周期(天)", Default: "7", Min: "3", Max: "14"},
			},
		},
		{
			Name:        "max_consecutive_nights",
			DisplayName: "最大连续夜班",
			Type:        "hard",
			Category:    "休息保障",
			Description: "限制员工连续上夜班的天数，保护员工健康。",
			Scenarios:   []string{"factory"},
			Params: []ConstraintParam{
				{Name: "max_nights", Type: "int", Description: "最大连续夜班天数", Default: "4", Min: "2", Max: "7"},
			},
		},
		{
			Name:        "production_line_coverage",
			DisplayName: "产线24小时覆盖",
			Type:        "hard",
			Category:    "生产保障",
			Description: "确保生产线在指定时间段内有足够的人员覆盖。",
			Scenarios:   []string{"factory"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "100", Min: "0", Max: "100"},
			},
		},
		{
			Name:        "team_together",
			DisplayName: "团队协作",
			Type:        "soft",
			Category:    "协作",
			Description: "尽量安排同一团队的成员在相同班次工作，提高团队协作效率。",
			Scenarios:   []string{"factory"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "70", Min: "0", Max: "100"},
			},
		},

		// =====================================================
		// 家政服务特有约束
		// =====================================================
		{
			Name:        "service_area",
			DisplayName: "服务区域匹配",
			Type:        "hard",
			Category:    "区域限制",
			Description: "确保员工只被分配到其覆盖的服务区域内的订单。",
			Scenarios:   []string{"housekeeping", "nursing"},
			Params:      []ConstraintParam{},
		},
		{
			Name:        "time_window",
			DisplayName: "服务时间窗口",
			Type:        "hard",
			Category:    "服务保障",
			Description: "确保服务安排在客户指定的时间窗口内进行。",
			Scenarios:   []string{"housekeeping", "nursing"},
			Params:      []ConstraintParam{},
		},
		{
			Name:        "travel_time",
			DisplayName: "路程时间优化",
			Type:        "soft",
			Category:    "效率优化",
			Description: "优化派单顺序，减少员工在不同客户之间的通勤时间。",
			Scenarios:   []string{"housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "80", Min: "0", Max: "100"},
				{Name: "max_travel_minutes", Type: "int", Description: "最大通勤时间(分钟)", Default: "60"},
			},
		},
		{
			Name:        "customer_preference",
			DisplayName: "客户偏好",
			Type:        "soft",
			Category:    "服务质量",
			Description: "尽量安排客户指定或偏好的服务人员。",
			Scenarios:   []string{"housekeeping", "nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "60", Min: "0", Max: "100"},
			},
		},

		// =====================================================
		// 长护险/护理特有约束
		// =====================================================
		{
			Name:        "nursing_qualification",
			DisplayName: "护理资质等级",
			Type:        "hard",
			Category:    "资质要求",
			Description: "确保护理员具备服务所需的护理资质等级。",
			Scenarios:   []string{"nursing"},
			Params: []ConstraintParam{
				{Name: "required_level", Type: "string", Description: "要求等级", Default: "初级护理员"},
			},
		},
		{
			Name:        "service_continuity",
			DisplayName: "服务连续性",
			Type:        "soft",
			Category:    "服务质量",
			Description: "优先安排熟悉患者情况的护理员，提高护理连续性和患者满意度。",
			Scenarios:   []string{"nursing"},
			Params: []ConstraintParam{
				{Name: "weight", Type: "int", Description: "优化权重", Default: "85", Min: "0", Max: "100"},
			},
		},
		{
			Name:        "max_patients_per_day",
			DisplayName: "每日最大服务患者数",
			Type:        "hard",
			Category:    "服务质量",
			Description: "限制护理员每天服务的最大患者数量，确保服务质量。",
			Scenarios:   []string{"nursing"},
			Params: []ConstraintParam{
				{Name: "max_patients", Type: "int", Description: "最大患者数", Default: "4", Min: "1", Max: "8"},
			},
		},
		{
			Name:        "care_plan_compliance",
			DisplayName: "护理计划合规",
			Type:        "hard",
			Category:    "服务质量",
			Description: "确保排班安排符合患者的护理计划要求（服务频次、时间等）。",
			Scenarios:   []string{"nursing"},
			Params:      []ConstraintParam{},
		},
	}
}
