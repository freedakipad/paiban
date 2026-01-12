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
  // 工作地点列表（模拟真实订单地址）
  const workLocations = [
    { address: "浦东新区金桥路888号", latitude: 31.2341, longitude: 121.6045, district: "浦东新区" },
    { address: "徐汇区衡山路100号", latitude: 31.2090, longitude: 121.4450, district: "徐汇区" },
    { address: "长宁区延安西路1088号", latitude: 31.2180, longitude: 121.4200, district: "长宁区" },
    { address: "静安区南京西路1266号", latitude: 31.2290, longitude: 121.4550, district: "静安区" },
    { address: "闵行区莘松路380号", latitude: 31.1150, longitude: 121.3900, district: "闵行区" },
    { address: "浦东新区陆家嘴环路1000号", latitude: 31.2400, longitude: 121.5000, district: "浦东新区" },
    { address: "徐汇区漕溪北路88号", latitude: 31.1900, longitude: 121.4370, district: "徐汇区" }
  ];
  const reqs = [];
  for (let i = 0; i < 7; i++) {
    reqs.push({ 
      shift_id: "750e8400-e29b-41d4-a716-446655440010", 
      date: getDateOffset(i), 
      position: "保洁员", 
      min_employees: 2, 
      priority: 7, 
      note: "日常保洁订单",
      work_location: workLocations[i % workLocations.length]
    });
    if (i % 2 === 0) {
      reqs.push({ 
        shift_id: "750e8400-e29b-41d4-a716-446655440011", 
        date: getDateOffset(i), 
        position: "保洁员", 
        min_employees: 1, 
        priority: 6, 
        note: "下午保洁",
        work_location: workLocations[(i + 3) % workLocations.length]
      });
    }
  }
  return reqs;
}

function generateNursingRequirements() {
  // 患者住址列表（模拟上门护理地址）
  const patientLocations = [
    { address: "浦东新区世纪大道1号", latitude: 31.2335, longitude: 121.5250, district: "浦东新区", patient: "张老" },
    { address: "黄浦区人民广场附近", latitude: 31.2320, longitude: 121.4750, district: "黄浦区", patient: "李奶奶" },
    { address: "徐汇区田林路200号", latitude: 31.1780, longitude: 121.4180, district: "徐汇区", patient: "王老" },
    { address: "长宁区虹桥路1000号", latitude: 31.2050, longitude: 121.4100, district: "长宁区", patient: "刘奶奶" },
    { address: "静安区江宁路500号", latitude: 31.2400, longitude: 121.4600, district: "静安区", patient: "陈老" },
    { address: "闵行区七宝镇", latitude: 31.1500, longitude: 121.3600, district: "闵行区", patient: "赵老" },
    { address: "普陀区曹杨路800号", latitude: 31.2450, longitude: 121.4100, district: "普陀区", patient: "孙奶奶" }
  ];
  const reqs = [];
  for (let i = 0; i < 7; i++) {
    reqs.push({ 
      shift_id: "850e8400-e29b-41d4-a716-446655440010", 
      date: getDateOffset(i), 
      position: "护理员", 
      min_employees: 2, 
      priority: 10, 
      note: "上午护理",
      work_location: patientLocations[i % patientLocations.length]
    });
    reqs.push({ 
      shift_id: "850e8400-e29b-41d4-a716-446655440011", 
      date: getDateOffset(i), 
      position: "护理员", 
      min_employees: 2, 
      priority: 10, 
      note: "下午护理",
      work_location: patientLocations[(i + 3) % patientLocations.length]
    });
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
        { 
          id: "750e8400-e29b-41d4-a716-446655440001", 
          name: "阿姨A", 
          position: "保洁员", 
          skills: ["日常保洁", "开荒保洁"], 
          status: "active",
          home_location: { address: "浦东新区张江镇", latitude: 31.2041, longitude: 121.5901, district: "浦东新区" },
          service_area: { districts: ["浦东新区"], max_radius: 10 }
        },
        { 
          id: "750e8400-e29b-41d4-a716-446655440002", 
          name: "阿姨B", 
          position: "保洁员", 
          skills: ["日常保洁", "擦玻璃"], 
          status: "active",
          home_location: { address: "徐汇区徐家汇", latitude: 31.1956, longitude: 121.4375, district: "徐汇区" },
          service_area: { districts: ["徐汇区", "长宁区"], max_radius: 8 }
        },
        { 
          id: "750e8400-e29b-41d4-a716-446655440003", 
          name: "阿姨C", 
          position: "月嫂", 
          skills: ["月嫂服务", "育儿嫂"], 
          status: "active",
          home_location: { address: "静安区南京西路", latitude: 31.2304, longitude: 121.4737, district: "静安区" },
          service_area: { max_radius: 15 }
        },
        { 
          id: "750e8400-e29b-41d4-a716-446655440004", 
          name: "阿姨D", 
          position: "保洁员", 
          skills: ["日常保洁"], 
          status: "active",
          home_location: { address: "闵行区莘庄镇", latitude: 31.1131, longitude: 121.3849, district: "闵行区" },
          service_area: { districts: ["闵行区", "徐汇区"], max_radius: 12 }
        }
      ],
      shifts: [
        { id: "750e8400-e29b-41d4-a716-446655440010", name: "上午时段", code: "AM", start_time: "08:00", end_time: "12:00", duration: 240, type: "morning" },
        { id: "750e8400-e29b-41d4-a716-446655440011", name: "下午时段", code: "PM", start_time: "14:00", end_time: "18:00", duration: 240, type: "afternoon" },
        { id: "750e8400-e29b-41d4-a716-446655440012", name: "全天服务", code: "FD", start_time: "08:00", end_time: "18:00", duration: 480, type: "morning" }
      ],
      requirements: generateHousekeepingRequirements(),
      constraints: { max_orders_per_day: 3, skill_match_required: true, max_travel_time: 60 },
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
        { 
          id: "850e8400-e29b-41d4-a716-446655440001", 
          name: "护理员A", 
          position: "护理员", 
          skills: ["基础护理", "康复护理"], 
          status: "active",
          home_location: { address: "浦东新区杨高路", latitude: 31.2200, longitude: 121.5300, district: "浦东新区" },
          service_area: { districts: ["浦东新区", "黄浦区"], max_radius: 15 }
        },
        { 
          id: "850e8400-e29b-41d4-a716-446655440002", 
          name: "护理员B", 
          position: "护理员", 
          skills: ["基础护理", "生活照料"], 
          status: "active",
          home_location: { address: "徐汇区龙华路", latitude: 31.1850, longitude: 121.4400, district: "徐汇区" },
          service_area: { districts: ["徐汇区", "长宁区", "闵行区"], max_radius: 12 }
        },
        { 
          id: "850e8400-e29b-41d4-a716-446655440003", 
          name: "护理员C", 
          position: "高级护理员", 
          skills: ["康复护理", "基础护理"], 
          status: "active",
          home_location: { address: "静安区北京西路", latitude: 31.2350, longitude: 121.4500, district: "静安区" },
          service_area: { districts: ["静安区", "普陀区", "黄浦区"], max_radius: 10 }
        },
        { 
          id: "850e8400-e29b-41d4-a716-446655440004", 
          name: "护理员D", 
          position: "护理员", 
          skills: ["生活照料", "基础护理"], 
          status: "active",
          home_location: { address: "长宁区古北路", latitude: 31.2100, longitude: 121.4000, district: "长宁区" },
          service_area: { districts: ["长宁区", "闵行区"], max_radius: 15 }
        }
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
