/**
 * PaiBan API 控制台 - 场景数据
 * 各业务场景的示例数据定义
 */

// ========== 日期工具函数 ==========
function getTodayDate() {
  return new Date().toISOString().split('T')[0];
}

function getNextWeekDate() {
  const d = new Date();
  d.setDate(d.getDate() + 7);
  return d.toISOString().split('T')[0];
}

function getDateOffset(days) {
  const d = new Date();
  d.setDate(d.getDate() + days);
  return d.toISOString().split('T')[0];
}

// ========== 需求生成函数 ==========
function generateRestaurantRequirements() {
  const morningShift = "550e8400-e29b-41d4-a716-446655440010";
  const eveningShift = "550e8400-e29b-41d4-a716-446655440012";
  
  const reqs = [];
  for (let i = 0; i < 7; i++) {
    const d = new Date(); d.setDate(d.getDate() + i);
    const dayOfWeek = d.getDay();
    const isWeekend = (dayOfWeek === 0 || dayOfWeek === 6);
    
    reqs.push({ shift_id: morningShift, date: getDateOffset(i), position: "服务员", min_employees: 2, priority: 8, note: "早班服务" });
    reqs.push({ shift_id: eveningShift, date: getDateOffset(i), position: "服务员", min_employees: isWeekend ? 3 : 2, priority: 8, note: isWeekend ? "周末晚班" : "晚班服务" });
    reqs.push({ shift_id: morningShift, date: getDateOffset(i), position: "厨师", min_employees: 1, priority: 9, note: "早班备餐" });
    reqs.push({ shift_id: eveningShift, date: getDateOffset(i), position: "厨师", min_employees: 1, priority: 9, note: "晚班备餐" });
  }
  return reqs;
}

function generateFactoryRequirements() {
  const shifts = ["650e8400-e29b-41d4-a716-446655440010", "650e8400-e29b-41d4-a716-446655440011", "650e8400-e29b-41d4-a716-446655440012"];
  const names = ["白班", "中班", "夜班"];
  return Array.from({length: 7}, (_, i) => shifts.map((s, idx) => ({ shift_id: s, date: getDateOffset(i), min_employees: 2, priority: 9, note: names[idx] + "产线" }))).flat();
}

function generateHousekeepingRequirements() {
  const reqs = [];
  for (let i = 0; i < 7; i++) {
    reqs.push({ shift_id: "750e8400-e29b-41d4-a716-446655440010", date: getDateOffset(i), position: "保洁员", min_employees: 2, priority: 7, note: "日常保洁订单" });
    if (i % 2 === 0) {
      reqs.push({ shift_id: "750e8400-e29b-41d4-a716-446655440011", date: getDateOffset(i), position: "保洁员", min_employees: 1, priority: 6, note: "下午保洁" });
    }
  }
  return reqs;
}

function generateNursingRequirements() {
  const reqs = [];
  for (let i = 0; i < 7; i++) {
    reqs.push({ shift_id: "850e8400-e29b-41d4-a716-446655440010", date: getDateOffset(i), position: "护理员", min_employees: 2, priority: 10, note: "上午护理" });
    reqs.push({ shift_id: "850e8400-e29b-41d4-a716-446655440011", date: getDateOffset(i), position: "护理员", min_employees: 2, priority: 10, note: "下午护理" });
  }
  return reqs;
}

// ========== 场景数据 ==========
const scenarioData = {
  // 餐饮门店场景
  restaurant: {
    method: 'POST',
    endpoint: '/api/v1/schedule/generate',
    body: {
      org_id: "550e8400-e29b-41d4-a716-446655440000",
      start_date: getTodayDate(),
      end_date: getNextWeekDate(),
      scenario: "restaurant",
      employees: [
        { 
          id: "550e8400-e29b-41d4-a716-446655440001", 
          name: "张三", 
          position: "服务员", 
          skills: ["收银", "点餐", "传菜"], 
          status: "active",
          preferences: { preferred_shifts: ["M"], avoid_shifts: ["E"], max_hours_per_week: 40 } 
        },
        { 
          id: "550e8400-e29b-41d4-a716-446655440002", 
          name: "李四", 
          position: "服务员", 
          skills: ["点餐", "传菜"], 
          status: "active",
          preferences: { preferred_shifts: ["M"], avoid_days: [0, 6] } 
        },
        { 
          id: "550e8400-e29b-41d4-a716-446655440005", 
          name: "钱七", 
          position: "服务员", 
          skills: ["收银"], 
          status: "active" 
        },
        { 
          id: "550e8400-e29b-41d4-a716-446655440006", 
          name: "孙八", 
          position: "服务员", 
          skills: ["点餐", "传菜"], 
          status: "active",
          preferences: { preferred_shifts: ["E"] } 
        },
        { 
          id: "550e8400-e29b-41d4-a716-446655440007", 
          name: "周九", 
          position: "服务员", 
          skills: ["收银", "点餐"], 
          status: "active" 
        },
        { 
          id: "550e8400-e29b-41d4-a716-446655440008", 
          name: "吴十", 
          position: "服务员", 
          skills: ["传菜"], 
          status: "active" 
        },
        { 
          id: "550e8400-e29b-41d4-a716-446655440003", 
          name: "王五", 
          position: "厨师", 
          skills: ["炒菜", "凉菜"], 
          status: "active",
          preferences: { max_hours_per_week: 35 } 
        },
        { 
          id: "550e8400-e29b-41d4-a716-446655440004", 
          name: "赵六", 
          position: "厨师", 
          skills: ["炒菜", "面点"], 
          status: "active" 
        },
        { 
          id: "550e8400-e29b-41d4-a716-446655440009", 
          name: "郑十一", 
          position: "厨师", 
          skills: ["炒菜", "凉菜", "面点"], 
          status: "active" 
        }
      ],
      shifts: [
        { id: "550e8400-e29b-41d4-a716-446655440010", name: "早班", code: "M", start_time: "09:00", end_time: "14:00", duration: 300, type: "morning" },
        { id: "550e8400-e29b-41d4-a716-446655440011", name: "午班", code: "A", start_time: "11:00", end_time: "14:00", duration: 180, type: "afternoon" },
        { id: "550e8400-e29b-41d4-a716-446655440012", name: "晚班", code: "E", start_time: "17:00", end_time: "22:00", duration: 300, type: "evening" },
        { id: "550e8400-e29b-41d4-a716-446655440013", name: "两头班", code: "S", start_time: "11:00", end_time: "21:00", duration: 480, type: "split" }
      ],
      requirements: generateRestaurantRequirements(),
      constraints: { max_hours_per_week: 44, min_rest_hours: 11, max_consecutive_days: 6 },
      options: { timeout_seconds: 30, optimization_level: 2, respect_preferences: true }
    }
  },

  // 工厂产线场景
  factory: {
    method: 'POST',
    endpoint: '/api/v1/schedule/generate',
    body: {
      org_id: "650e8400-e29b-41d4-a716-446655440000",
      start_date: getTodayDate(),
      end_date: getNextWeekDate(),
      scenario: "factory",
      employees: [
        { id: "650e8400-e29b-41d4-a716-446655440001", name: "工人A", position: "操作工", skills: ["数控车床", "焊接"], status: "active" },
        { id: "650e8400-e29b-41d4-a716-446655440002", name: "工人B", position: "操作工", skills: ["数控车床"], status: "active" },
        { id: "650e8400-e29b-41d4-a716-446655440003", name: "工人C", position: "操作工", skills: ["焊接", "装配"], status: "active" },
        { id: "650e8400-e29b-41d4-a716-446655440004", name: "工人D", position: "质检员", skills: ["质量检验"], status: "active" },
        { id: "650e8400-e29b-41d4-a716-446655440005", name: "工人E", position: "操作工", skills: ["装配"], status: "active" },
        { id: "650e8400-e29b-41d4-a716-446655440006", name: "工人F", position: "操作工", skills: ["数控车床", "装配"], status: "active" }
      ],
      shifts: [
        { id: "650e8400-e29b-41d4-a716-446655440010", name: "白班", code: "D", start_time: "08:00", end_time: "16:00", duration: 480, type: "morning" },
        { id: "650e8400-e29b-41d4-a716-446655440011", name: "中班", code: "S", start_time: "16:00", end_time: "00:00", duration: 480, type: "afternoon" },
        { id: "650e8400-e29b-41d4-a716-446655440012", name: "夜班", code: "N", start_time: "00:00", end_time: "08:00", duration: 480, type: "night" }
      ],
      requirements: generateFactoryRequirements(),
      constraints: { max_hours_per_week: 48, min_rest_hours: 8, max_consecutive_nights: 3 },
      options: { timeout_seconds: 30, optimization_level: 2 }
    }
  },

  // 家政服务场景
  housekeeping: {
    method: 'POST',
    endpoint: '/api/v1/schedule/generate',
    body: {
      org_id: "750e8400-e29b-41d4-a716-446655440000",
      start_date: getTodayDate(),
      end_date: getNextWeekDate(),
      scenario: "housekeeping",
      employees: [
        { id: "750e8400-e29b-41d4-a716-446655440001", name: "阿姨A", position: "保洁员", skills: ["日常保洁", "开荒保洁"], status: "active" },
        { id: "750e8400-e29b-41d4-a716-446655440002", name: "阿姨B", position: "保洁员", skills: ["日常保洁", "擦玻璃"], status: "active" },
        { id: "750e8400-e29b-41d4-a716-446655440003", name: "阿姨C", position: "月嫂", skills: ["月嫂服务", "育儿嫂"], status: "active" },
        { id: "750e8400-e29b-41d4-a716-446655440004", name: "阿姨D", position: "保洁员", skills: ["日常保洁"], status: "active" }
      ],
      shifts: [
        { id: "750e8400-e29b-41d4-a716-446655440010", name: "上午时段", code: "AM", start_time: "08:00", end_time: "12:00", duration: 240, type: "morning" },
        { id: "750e8400-e29b-41d4-a716-446655440011", name: "下午时段", code: "PM", start_time: "14:00", end_time: "18:00", duration: 240, type: "afternoon" },
        { id: "750e8400-e29b-41d4-a716-446655440012", name: "全天服务", code: "FD", start_time: "08:00", end_time: "18:00", duration: 480, type: "morning" }
      ],
      requirements: generateHousekeepingRequirements(),
      constraints: { max_orders_per_day: 3, skill_match_required: true },
      options: { timeout_seconds: 30, respect_preferences: true }
    }
  },

  // 长护险/护理场景
  nursing: {
    method: 'POST',
    endpoint: '/api/v1/schedule/generate',
    body: {
      org_id: "850e8400-e29b-41d4-a716-446655440000",
      start_date: getTodayDate(),
      end_date: getNextWeekDate(),
      scenario: "nursing",
      employees: [
        { id: "850e8400-e29b-41d4-a716-446655440001", name: "护理员A", position: "护理员", skills: ["基础护理", "康复护理"], status: "active" },
        { id: "850e8400-e29b-41d4-a716-446655440002", name: "护理员B", position: "护理员", skills: ["基础护理", "生活照料"], status: "active" },
        { id: "850e8400-e29b-41d4-a716-446655440003", name: "护理员C", position: "高级护理员", skills: ["康复护理", "基础护理"], status: "active" },
        { id: "850e8400-e29b-41d4-a716-446655440004", name: "护理员D", position: "护理员", skills: ["生活照料", "基础护理"], status: "active" }
      ],
      shifts: [
        { id: "850e8400-e29b-41d4-a716-446655440010", name: "上午护理", code: "AM", start_time: "08:00", end_time: "12:00", duration: 240, type: "morning" },
        { id: "850e8400-e29b-41d4-a716-446655440011", name: "下午护理", code: "PM", start_time: "14:00", end_time: "18:00", duration: 240, type: "afternoon" },
        { id: "850e8400-e29b-41d4-a716-446655440012", name: "全日护理", code: "FD", start_time: "08:00", end_time: "17:00", duration: 480, type: "morning" }
      ],
      requirements: generateNursingRequirements(),
      constraints: { continuity_required: true, max_patients_per_day: 4 },
      options: { timeout_seconds: 30, optimization_level: 3 }
    }
  }
};

// ========== 场景元数据（用于显示） ==========
const scenarioMeta = {
  restaurant: {
    name: '餐饮门店',
    icon: '🍜',
    color: '#f85149',
    description: '适用于餐厅、咖啡店等服务业的员工排班',
    features: ['早/晚班', '技能匹配', '周末加班']
  },
  factory: {
    name: '工厂产线',
    icon: '🏭',
    color: '#a371f7',
    description: '适用于制造业三班倒、连续生产场景',
    features: ['三班制', '倒班规则', '夜班限制']
  },
  housekeeping: {
    name: '家政服务',
    icon: '🏠',
    color: '#3fb950',
    description: '适用于家政公司的订单派工场景',
    features: ['技能匹配', '区域优化', '订单优先']
  },
  nursing: {
    name: '长护险/护理',
    icon: '💊',
    color: '#58a6ff',
    description: '适用于护理机构的护理员排班',
    features: ['护理计划', '连续性', '资质要求']
  }
};
