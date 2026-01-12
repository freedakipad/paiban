/**
 * 餐饮门店智能排班系统 - 数据模型
 * 支持连锁门店多店管理
 */

// ===== 连锁门店配置 =====
// 默认门店数据
const defaultStores = [
    {
        id: 'store-001',
        name: '总店',
        code: 'HQ',
        address: '北京市朝阳区三里屯路123号',
        phone: '010-12345678',
        manager: '张经理',
        location: { lat: 39.9334, lng: 116.4521 },
        openTime: '09:00',
        closeTime: '22:00',
        status: 'active',
        type: 'flagship',  // flagship-旗舰店, standard-标准店, express-快餐店
        capacity: 80,      // 座位数
        createdAt: '2023-01-01'
    },
    {
        id: 'store-002',
        name: '望京分店',
        code: 'WJ',
        address: '北京市朝阳区望京西路168号',
        phone: '010-23456789',
        manager: '李经理',
        location: { lat: 40.0024, lng: 116.4697 },
        openTime: '09:00',
        closeTime: '22:00',
        status: 'active',
        type: 'standard',
        capacity: 50,
        createdAt: '2023-06-15'
    },
    {
        id: 'store-003',
        name: '国贸分店',
        code: 'GM',
        address: '北京市朝阳区国贸大厦B座1层',
        phone: '010-34567890',
        manager: '王经理',
        location: { lat: 39.9087, lng: 116.4597 },
        openTime: '10:00',
        closeTime: '21:00',
        status: 'active',
        type: 'express',
        capacity: 30,
        createdAt: '2024-01-10'
    }
];

// 门店类型配置
const STORE_TYPES = {
    flagship: { label: '旗舰店', icon: '🏪', color: '#f59e0b', minStaff: 8 },
    standard: { label: '标准店', icon: '🏬', color: '#3b82f6', minStaff: 5 },
    express: { label: '快餐店', icon: '🍱', color: '#10b981', minStaff: 3 }
};

// 默认班次配置
const defaultShifts = [
    {
        id: 'shift-morning',
        name: '早班',
        code: 'M',
        startTime: '09:00',
        endTime: '14:00',
        color: '#f59e0b',
        hours: 5
    },
    {
        id: 'shift-afternoon',
        name: '午班',
        code: 'A',
        startTime: '11:00',
        endTime: '14:00',
        color: '#10b981',
        hours: 3
    },
    {
        id: 'shift-evening',
        name: '晚班',
        code: 'E',
        startTime: '17:00',
        endTime: '22:00',
        color: '#8b5cf6',
        hours: 5
    },
    {
        id: 'shift-split',
        name: '两头班',
        code: 'S',
        startTime: '11:00',
        endTime: '21:00',
        color: '#ec4899',
        hours: 8,
        note: '中间休息2小时'
    }
];

// 默认员工数据 - 支持多门店归属
// 注意：大部分员工不能跨店调配，只有少数资深/机动员工可以
const defaultEmployees = [
    {
        id: 'emp-001',
        name: '张三',
        position: '服务员',
        skills: ['收银', '点餐', '传菜'],
        phone: '13812341234',
        hireDate: '2024-03-15',
        status: 'active',
        storeId: 'store-001',      // 所属门店
        canTransfer: true,         // 普通员工不跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: ['shift-evening'],
            avoidDays: [0, 6], // 周日、周六
            maxHoursPerWeek: 40
        }
    },
    {
        id: 'emp-002',
        name: '李四',
        position: '服务员',
        skills: ['点餐', '传菜'],
        phone: '13956785678',
        hireDate: '2024-06-01',
        status: 'active',
        storeId: 'store-001',
        canTransfer: true,          // 可跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: ['shift-morning'],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    {
        id: 'emp-003',
        name: '王五',
        position: '厨师',
        skills: ['炒菜', '凉菜'],
        phone: '13790129012',
        hireDate: '2023-11-20',
        status: 'active',
        storeId: 'store-001',
        canTransfer: false,   // 厨师不跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [0], // 周日
            maxHoursPerWeek: 44
        }
    },
    {
        id: 'emp-004',
        name: '赵六',
        position: '厨师',
        skills: ['炒菜', '面点'],
        phone: '13634563456',
        hireDate: '2024-01-10',
        status: 'active',
        storeId: 'store-001',
        canTransfer: false,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: ['shift-evening'],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    {
        id: 'emp-005',
        name: '钱七',
        position: '服务员',
        skills: ['收银'],
        phone: '13578907890',
        hireDate: '2024-08-01',
        status: 'active',
        storeId: 'store-001',
        canTransfer: true,          // ★ 机动人员，可跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 40
        }
    },
    {
        id: 'emp-006',
        name: '孙八',
        position: '服务员',
        skills: ['点餐', '传菜'],
        phone: '13423452345',
        hireDate: '2024-05-15',
        status: 'active',
        storeId: 'store-002',      // 望京分店
        canTransfer: true,          // 可跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: ['shift-afternoon'],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    {
        id: 'emp-007',
        name: '周九',
        position: '服务员',
        skills: ['收银', '点餐'],
        phone: '13367896789',
        hireDate: '2024-02-20',
        status: 'active',
        storeId: 'store-002',      // 望京分店
        canTransfer: true,          // 可跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: ['shift-split'],
            avoidDays: [6], // 周六
            maxHoursPerWeek: 40
        }
    },
    {
        id: 'emp-008',
        name: '吴十',
        position: '服务员',
        skills: ['传菜'],
        phone: '13201230123',
        hireDate: '2024-09-01',
        status: 'inactive',
        storeId: 'store-002',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    {
        id: 'emp-009',
        name: '郑十一',
        position: '厨师',
        skills: ['炒菜', '凉菜', '面点'],
        phone: '13145674567',
        hireDate: '2023-08-10',
        status: 'active',
        storeId: 'store-001',
        canTransfer: true,        // 资深厨师可以跨店支援
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 望京分店员工
    {
        id: 'emp-010',
        name: '陈十二',
        position: '厨师',
        skills: ['炒菜', '凉菜'],
        phone: '13511112222',
        hireDate: '2023-10-01',
        status: 'active',
        storeId: 'store-002',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    {
        id: 'emp-011',
        name: '林十三',
        position: '服务员',
        skills: ['收银', '点餐', '传菜'],
        phone: '13522223333',
        hireDate: '2024-02-15',
        status: 'active',
        storeId: 'store-002',
        canTransfer: true,          // ★ 机动人员，可跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: ['shift-morning'],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 国贸分店员工
    {
        id: 'emp-012',
        name: '黄十四',
        position: '厨师',
        skills: ['炒菜', '面点'],
        phone: '13533334444',
        hireDate: '2024-01-20',
        status: 'active',
        storeId: 'store-003',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    {
        id: 'emp-013',
        name: '杨十五',
        position: '服务员',
        skills: ['收银', '点餐'],
        phone: '13544445555',
        hireDate: '2024-03-01',
        status: 'active',
        storeId: 'store-003',
        canTransfer: true,          // 可跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [0],
            maxHoursPerWeek: 40
        }
    },
    {
        id: 'emp-014',
        name: '吕十六',
        position: '服务员',
        skills: ['点餐', '传菜'],
        phone: '13555556666',
        hireDate: '2024-04-01',
        status: 'active',
        storeId: 'store-003',
        canTransfer: true,          // 可跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: ['shift-evening'],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // ===== 按排班报告建议新增员工 =====
    // 总店 +3服务员
    {
        id: 'emp-015',
        name: '冯十七',
        position: '服务员',
        skills: ['收银', '点餐'],
        phone: '13666667777',
        hireDate: '2026-01-12',
        status: 'active',
        storeId: 'store-001',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: ['shift-morning'],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    {
        id: 'emp-016',
        name: '卫十八',
        position: '服务员',
        skills: ['点餐', '传菜'],
        phone: '13777778888',
        hireDate: '2026-01-12',
        status: 'active',
        storeId: 'store-001',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: ['shift-evening'],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    {
        id: 'emp-017',
        name: '蒋十九',
        position: '服务员',
        skills: ['传菜', '清洁'],
        phone: '13888889999',
        hireDate: '2026-01-12',
        status: 'active',
        storeId: 'store-001',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [0],
            maxHoursPerWeek: 44
        }
    },
    // 总店 +1厨师
    {
        id: 'emp-018',
        name: '沈二十',
        position: '厨师',
        skills: ['炒菜', '凉菜'],
        phone: '13999990000',
        hireDate: '2026-01-12',
        status: 'active',
        storeId: 'store-001',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 望京分店 +1服务员
    {
        id: 'emp-019',
        name: '韩廿一',
        position: '服务员',
        skills: ['点餐', '传菜'],
        phone: '13100001111',
        hireDate: '2026-01-12',
        status: 'active',
        storeId: 'store-002',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 望京分店 +1厨师
    {
        id: 'emp-020',
        name: '杨廿二',
        position: '厨师',
        skills: ['炒菜', '面点'],
        phone: '13100002222',
        hireDate: '2026-01-12',
        status: 'active',
        storeId: 'store-002',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: ['shift-morning'],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 国贸分店 +1厨师
    {
        id: 'emp-021',
        name: '朱廿三',
        position: '厨师',
        skills: ['炒菜', '凉菜'],
        phone: '13100003333',
        hireDate: '2026-01-12',
        status: 'active',
        storeId: 'store-003',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [6],
            maxHoursPerWeek: 44
        }
    },
    // 国贸分店 +1服务员
    {
        id: 'emp-022',
        name: '何廿四',
        position: '服务员',
        skills: ['点餐', '收银'],
        phone: '13100004444',
        hireDate: '2026-01-15',
        status: 'active',
        storeId: 'store-003',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 望京分店 +1厨师
    {
        id: 'emp-023',
        name: '吴廿五',
        position: '厨师',
        skills: ['炒菜', '面点'],
        phone: '13100005555',
        hireDate: '2026-01-15',
        status: 'active',
        storeId: 'store-002',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 国贸分店 +1厨师
    {
        id: 'emp-024',
        name: '郑廿六',
        position: '厨师',
        skills: ['炒菜', '凉菜'],
        phone: '13100006666',
        hireDate: '2026-01-15',
        status: 'active',
        storeId: 'store-003',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 望京分店 +1服务员
    {
        id: 'emp-025',
        name: '王廿七',
        position: '服务员',
        skills: ['点餐', '收银'],
        phone: '13100007777',
        hireDate: '2026-01-15',
        status: 'active',
        storeId: 'store-002',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 总店 +1厨师
    {
        id: 'emp-026',
        name: '李廿八',
        position: '厨师',
        skills: ['炒菜', '面点'],
        phone: '13100008888',
        hireDate: '2026-01-15',
        status: 'active',
        storeId: 'store-001',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 总店 +1服务员
    {
        id: 'emp-027',
        name: '陈廿九',
        position: '服务员',
        skills: ['点餐', '收银'],
        phone: '13100009999',
        hireDate: '2026-01-15',
        status: 'active',
        storeId: 'store-001',
        canTransfer: true,
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 可跨店服务员（机动人员）
    {
        id: 'emp-028',
        name: '刘三十',
        position: '服务员',
        skills: ['点餐', '收银', '迎宾'],
        phone: '13100010000',
        hireDate: '2026-01-15',
        status: 'active',
        storeId: 'store-001',
        canTransfer: true,  // 可跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    },
    // 可跨店服务员2（机动人员）
    {
        id: 'emp-029',
        name: '赵三一',
        position: '服务员',
        skills: ['点餐', '收银', '迎宾'],
        phone: '13100011111',
        hireDate: '2026-01-15',
        status: 'active',
        storeId: 'store-002',
        canTransfer: true,  // 可跨店调配
        weeklyHours: 0,
        monthlyShifts: 0,
        preferences: {
            preferredShifts: [],
            avoidShifts: [],
            avoidDays: [],
            maxHoursPerWeek: 44
        }
    }
];

// 默认每日需求配置（按门店类型和班次）
// 不同门店类型有不同的人员需求
const defaultRequirements = {
    // 旗舰店需求较多
    'store-001': {
        weekday: {
            'shift-morning': { '服务员': 3, '厨师': 2 },
            'shift-evening': { '服务员': 3, '厨师': 2 }
        },
        weekend: {
            'shift-morning': { '服务员': 4, '厨师': 2 },
            'shift-evening': { '服务员': 4, '厨师': 2 }
        }
    },
    // 标准店中等需求
    'store-002': {
        weekday: {
            'shift-morning': { '服务员': 2, '厨师': 1 },
            'shift-evening': { '服务员': 2, '厨师': 1 }
        },
        weekend: {
            'shift-morning': { '服务员': 2, '厨师': 1 },
            'shift-evening': { '服务员': 3, '厨师': 1 }
        }
    },
    // 快餐店需求较少
    'store-003': {
        weekday: {
            'shift-morning': { '服务员': 1, '厨师': 1 },
            'shift-evening': { '服务员': 1, '厨师': 1 }
        },
        weekend: {
            'shift-morning': { '服务员': 2, '厨师': 1 },
            'shift-evening': { '服务员': 2, '厨师': 1 }
        }
    },
    // 通用配置（用于没有特定配置的门店）
    '_default': {
        weekday: {
            'shift-morning': { '服务员': 2, '厨师': 1 },
            'shift-evening': { '服务员': 2, '厨师': 1 }
        },
        weekend: {
            'shift-morning': { '服务员': 2, '厨师': 1 },
            'shift-evening': { '服务员': 3, '厨师': 1 }
        }
    }
};

// 默认设置
const defaultSettings = {
    currentStoreId: 'store-001',  // 当前选中的门店
    chainMode: true,               // 连锁模式开关
    storeName: '总店',             // 当前门店名称
    openTime: '09:00',
    closeTime: '22:00',
    hoursMode: 'weekly',       // 'weekly' 或 'period'
    maxWeeklyHours: 44,
    maxPeriodHours: 176,       // 月度工时（44 × 4周）
    minRestHours: 8,           // 减少到8小时
    maxConsecutiveDays: 6,
    minRestDays: 1,
    apiEndpoint: 'http://localhost:7012',
    timeout: 30,
    // 跨店调配设置
    crossStoreTransfer: {
        enabled: true,          // 是否允许跨店调配
        maxDistance: 10,        // 最大调配距离(km)
        extraPayRate: 1.2       // 跨店工作额外工资系数
    }
};

// 排班状态枚举
const SCHEDULE_STATUS = {
    DRAFT: 'draft',           // 草稿 - 正在排班
    PUBLISHED: 'published',   // 已发布 - 已公布给员工
    ACTIVE: 'active',         // 执行中 - 当前周正在执行
    ARCHIVED: 'archived'      // 已归档 - 历史记录
};

// 状态显示配置
const STATUS_CONFIG = {
    draft: { label: '草稿', icon: '📝', color: '#f59e0b' },
    published: { label: '已发布', icon: '📢', color: '#3b82f6' },
    active: { label: '执行中', icon: '⚡', color: '#10b981' },
    archived: { label: '已归档', icon: '📁', color: '#6b7280' }
};

// 应用状态
class AppState {
    constructor() {
        this.stores = this.loadFromStorage('stores', defaultStores);
        this.shifts = this.loadFromStorage('shifts', defaultShifts);
        this.employees = this.loadFromStorage('employees', defaultEmployees);
        this.requirements = this.loadFromStorage('requirements', defaultRequirements);
        this.settings = this.loadFromStorage('settings', defaultSettings);
        this.scheduleWeeks = this.loadFromStorage('scheduleWeeks', {}); // 周次排班数据（包含每周历史）
        this.assignments = [];
        this.unfilledRequirements = [];
        this.currentWeekStart = this.getWeekStart(new Date());
        this.currentView = 'schedule';
        this.schedulePeriod = 7; // 排班周期：7(1周)、14(2周)、'month'(月度)
        this.currentStoreId = this.settings.currentStoreId || 'store-001';
        
        // 初始化当前周状态
        this.initCurrentWeekStatus();
    }
    
    // ===== 门店管理方法 =====
    
    // 获取所有门店
    getAllStores() {
        return this.stores.filter(s => s.status === 'active');
    }
    
    // 判断是否为"全部门店"模式
    isAllStoresMode() {
        return this.currentStoreId === 'all';
    }
    
    // 获取当前门店（"全部"模式时返回null）
    getCurrentStore() {
        if (this.isAllStoresMode()) {
            return null;
        }
        return this.stores.find(s => s.id === this.currentStoreId) || this.stores[0];
    }
    
    // 切换门店（支持"all"选项）
    switchStore(storeId) {
        // 支持"全部"门店选项
        if (storeId === 'all') {
            this.currentStoreId = 'all';
            this.settings.currentStoreId = 'all';
            this.settings.storeName = '全部门店';
            this.saveToStorage('settings', this.settings);
            this.loadWeekSchedule();
            return true;
        }
        
        const store = this.stores.find(s => s.id === storeId);
        if (store) {
            this.currentStoreId = storeId;
            this.settings.currentStoreId = storeId;
            this.settings.storeName = store.name;
            this.settings.openTime = store.openTime;
            this.settings.closeTime = store.closeTime;
            this.saveToStorage('settings', this.settings);
            
            // 切换门店时重新加载排班数据
            this.loadWeekSchedule();
            return true;
        }
        return false;
    }
    
    // 获取当前门店的员工（包括可跨店调配的其他门店员工）
    // "全部"模式时返回所有活跃员工
    getCurrentStoreEmployees(includeTransferable = false) {
        // "全部"模式：返回所有活跃员工
        if (this.isAllStoresMode()) {
            return this.employees.filter(e => e.status === 'active');
        }
        
        // 默认启用跨店调配（如果设置中没有明确禁用）
        const transferEnabled = this.settings.crossStoreTransfer?.enabled !== false;
        if (includeTransferable && transferEnabled) {
            return this.employees.filter(e => 
                e.storeId === this.currentStoreId || 
                (e.canTransfer && e.status === 'active')
            );
        }
        return this.employees.filter(e => e.storeId === this.currentStoreId);
    }
    
    // 获取指定门店的员工
    getStoreEmployees(storeId) {
        return this.employees.filter(e => e.storeId === storeId);
    }
    
    // 获取可调配到当前门店的其他门店员工
    getTransferableEmployees() {
        // 默认启用跨店调配（如果设置中没有明确禁用）
        const transferEnabled = this.settings.crossStoreTransfer?.enabled !== false;
        if (!transferEnabled) return [];
        return this.employees.filter(e => 
            e.storeId !== this.currentStoreId && 
            e.canTransfer && 
            e.status === 'active'
        );
    }
    
    // 添加门店
    addStore(store) {
        store.id = 'store-' + Date.now();
        store.createdAt = new Date().toISOString().split('T')[0];
        this.stores.push(store);
        this.saveToStorage('stores', this.stores);
        return store;
    }
    
    // 更新门店
    updateStore(id, updates) {
        const index = this.stores.findIndex(s => s.id === id);
        if (index !== -1) {
            this.stores[index] = { ...this.stores[index], ...updates };
            this.saveToStorage('stores', this.stores);
            // 如果更新的是当前门店，同步更新设置
            if (id === this.currentStoreId) {
                this.settings.storeName = this.stores[index].name;
                this.saveToStorage('settings', this.settings);
            }
            return true;
        }
        return false;
    }
    
    // 删除门店（软删除）
    deleteStore(id) {
        if (id === this.currentStoreId) {
            console.warn('不能删除当前门店');
            return false;
        }
        const index = this.stores.findIndex(s => s.id === id);
        if (index !== -1) {
            this.stores[index].status = 'inactive';
            this.saveToStorage('stores', this.stores);
            return true;
        }
        return false;
    }
    
    // 获取门店统计信息
    getStoreStats(storeId) {
        const employees = this.getStoreEmployees(storeId);
        const activeEmployees = employees.filter(e => e.status === 'active');
        const positions = {};
        activeEmployees.forEach(e => {
            positions[e.position] = (positions[e.position] || 0) + 1;
        });
        return {
            totalEmployees: employees.length,
            activeEmployees: activeEmployees.length,
            positions,
            transferable: employees.filter(e => e.canTransfer && e.status === 'active').length
        };
    }
    
    // 获取门店排班键（包含门店ID）
    getStoreWeekKey(date) {
        const weekKey = this.getWeekKey(date);
        return `${this.currentStoreId}_${weekKey}`;
    }
    
    // 设置排班周期
    setSchedulePeriod(period) {
        this.schedulePeriod = period;
        // 如果是月度，重新计算起始日期为月初
        if (period === 'month') {
            const d = new Date(this.currentWeekStart);
            this.currentPeriodStart = new Date(d.getFullYear(), d.getMonth(), 1);
        } else {
            this.currentPeriodStart = new Date(this.currentWeekStart);
        }
    }
    
    // 获取周次键名 (格式: YYYY-WW)
    getWeekKey(date) {
        const d = new Date(date);
        const yearStart = new Date(d.getFullYear(), 0, 1);
        const weekNum = Math.ceil((((d - yearStart) / 86400000) + yearStart.getDay() + 1) / 7);
        return `${d.getFullYear()}-W${String(weekNum).padStart(2, '0')}`;
    }
    
    // 初始化当前周状态（使用门店隔离的键）
    initCurrentWeekStatus() {
        const storeWeekKey = this.getStoreWeekKey(this.currentWeekStart);
        if (!this.scheduleWeeks[storeWeekKey]) {
            const today = new Date();
            const weekStart = new Date(this.currentWeekStart);
            
            // 判断周次状态
            let status = SCHEDULE_STATUS.DRAFT;
            if (weekStart <= today && today <= new Date(weekStart.getTime() + 6 * 24 * 60 * 60 * 1000)) {
                status = SCHEDULE_STATUS.ACTIVE; // 当前周
            } else if (weekStart < today) {
                status = SCHEDULE_STATUS.ARCHIVED; // 过去的周
            }
            
            // 简单日期格式化
            const startDateStr = weekStart.toISOString().split('T')[0];
            this.scheduleWeeks[storeWeekKey] = {
                weekKey: storeWeekKey,
                storeId: this.currentStoreId,   // 记录门店ID
                startDate: startDateStr,
                status,
                assignments: [],
                history: [], // 该周的操作历史
                createdAt: null,
                publishedAt: null,
                archivedAt: null
            };
        }
    }
    
    // 获取当前周排班数据
    getCurrentWeekSchedule() {
        const storeWeekKey = this.getStoreWeekKey(this.currentWeekStart);
        this.initCurrentWeekStatus();
        return this.scheduleWeeks[storeWeekKey];
    }
    
    // 保存排班到当前周
    saveScheduleToWeek(assignments) {
        const storeWeekKey = this.getStoreWeekKey(this.currentWeekStart);
        this.initCurrentWeekStatus();
        this.scheduleWeeks[storeWeekKey].assignments = assignments;
        this.scheduleWeeks[storeWeekKey].storeId = this.currentStoreId;
        this.scheduleWeeks[storeWeekKey].createdAt = new Date().toISOString();
        this.saveToStorage('scheduleWeeks', this.scheduleWeeks);
    }
    
    // 发布排班
    publishSchedule() {
        const storeWeekKey = this.getStoreWeekKey(this.currentWeekStart);
        const schedule = this.scheduleWeeks[storeWeekKey];
        if (schedule && schedule.status === SCHEDULE_STATUS.DRAFT) {
            schedule.status = SCHEDULE_STATUS.PUBLISHED;
            schedule.publishedAt = new Date().toISOString();
            this.saveToStorage('scheduleWeeks', this.scheduleWeeks);
            return true;
        }
        return false;
    }
    
    // 归档排班（任何状态的周都可以归档）
    archiveSchedule(storeWeekKey) {
        const schedule = this.scheduleWeeks[storeWeekKey];
        if (schedule && schedule.status !== SCHEDULE_STATUS.ARCHIVED) {
            schedule.status = SCHEDULE_STATUS.ARCHIVED;
            schedule.archivedAt = new Date().toISOString();
            this.saveToStorage('scheduleWeeks', this.scheduleWeeks);
            return true;
        }
        return false;
    }
    
    // 解锁已发布的排班（恢复为草稿状态）
    // 注意：归档状态不能解锁
    unlockSchedule(storeWeekKey) {
        const schedule = this.scheduleWeeks[storeWeekKey];
        if (schedule && schedule.status === SCHEDULE_STATUS.PUBLISHED) {
            schedule.status = SCHEDULE_STATUS.DRAFT;
            schedule.unlockedAt = new Date().toISOString();
            this.saveToStorage('scheduleWeeks', this.scheduleWeeks);
            return true;
        }
        return false;
    }
    
    // 加载周排班数据
    loadWeekSchedule() {
        const schedule = this.getCurrentWeekSchedule();
        if (schedule && schedule.assignments.length > 0) {
            this.assignments = schedule.assignments;
            return true;
        }
        return false;
    }
    
    // 获取指定月份的所有排班数据（从当前门店已保存的周排班中）
    getMonthAssignments(year, month) {
        const monthEnd = new Date(year, month + 1, 0);
        const monthStartStr = `${year}-${String(month + 1).padStart(2, '0')}-01`;
        const monthEndStr = `${year}-${String(month + 1).padStart(2, '0')}-${monthEnd.getDate()}`;
        
        const allAssignments = [];
        
        // 遍历所有保存的周排班（只获取当前门店的）
        Object.keys(this.scheduleWeeks).forEach(storeWeekKey => {
            const schedule = this.scheduleWeeks[storeWeekKey];
            // 只包含当前门店的、已发布或已归档的排班
            if ((schedule.storeId === this.currentStoreId || storeWeekKey.startsWith(this.currentStoreId + '_')) &&
                (schedule.status === SCHEDULE_STATUS.PUBLISHED || 
                 schedule.status === SCHEDULE_STATUS.ARCHIVED)) {
                schedule.assignments.forEach(a => {
                    // 检查日期是否在指定月份内
                    if (a.date >= monthStartStr && a.date <= monthEndStr) {
                        allAssignments.push(a);
                    }
                });
            }
        });
        
        return allAssignments;
    }
    
    // 获取所有门店某月份的排班汇总（用于连锁管理视图）
    getAllStoresMonthSummary(year, month) {
        const monthEnd = new Date(year, month + 1, 0);
        const monthStartStr = `${year}-${String(month + 1).padStart(2, '0')}-01`;
        const monthEndStr = `${year}-${String(month + 1).padStart(2, '0')}-${monthEnd.getDate()}`;
        
        const summary = {};
        
        this.stores.forEach(store => {
            summary[store.id] = {
                store,
                totalShifts: 0,
                totalHours: 0,
                employeeCount: 0,
                crossStoreShifts: 0
            };
        });
        
        // 遍历所有保存的周排班
        Object.keys(this.scheduleWeeks).forEach(storeWeekKey => {
            const schedule = this.scheduleWeeks[storeWeekKey];
            const storeId = schedule.storeId || storeWeekKey.split('_')[0];
            
            if (schedule.status === SCHEDULE_STATUS.PUBLISHED || 
                schedule.status === SCHEDULE_STATUS.ARCHIVED) {
                schedule.assignments.forEach(a => {
                    if (a.date >= monthStartStr && a.date <= monthEndStr) {
                        if (summary[storeId]) {
                            summary[storeId].totalShifts++;
                            summary[storeId].totalHours += (a.hours || 0);
                            // 检查是否跨店排班
                            const emp = this.employees.find(e => e.id === a.employeeId);
                            if (emp && emp.storeId !== storeId) {
                                summary[storeId].crossStoreShifts++;
                            }
                        }
                    }
                });
            }
        });
        
        // 统计各门店员工数
        this.stores.forEach(store => {
            const activeEmps = this.employees.filter(e => e.storeId === store.id && e.status === 'active');
            summary[store.id].employeeCount = activeEmps.length;
        });
        
        return summary;
    }
    
    // 获取周状态
    getWeekStatus() {
        const schedule = this.getCurrentWeekSchedule();
        return schedule ? schedule.status : SCHEDULE_STATUS.DRAFT;
    }
    
    // 检查是否可以编辑排班
    canEditSchedule() {
        const status = this.getWeekStatus();
        return status === SCHEDULE_STATUS.DRAFT || status === SCHEDULE_STATUS.PUBLISHED;
    }

    loadFromStorage(key, defaultValue) {
        try {
            const saved = localStorage.getItem(`restaurant-scheduler-${key}`);
            return saved ? JSON.parse(saved) : defaultValue;
        } catch (e) {
            console.warn(`Failed to load ${key} from storage:`, e);
            return defaultValue;
        }
    }

    saveToStorage(key, value) {
        try {
            localStorage.setItem(`restaurant-scheduler-${key}`, JSON.stringify(value));
        } catch (e) {
            console.warn(`Failed to save ${key} to storage:`, e);
        }
    }

    getWeekStart(date) {
        const d = new Date(date);
        const day = d.getDay();
        const diff = d.getDate() - day + (day === 0 ? -6 : 1); // 周一为起始
        return new Date(d.setDate(diff));
    }

    getWeekDates() {
        const dates = [];
        const start = new Date(this.currentWeekStart);
        
        // 根据排班周期返回不同天数
        let days = 7;
        if (this.schedulePeriod === 14) {
            days = 14;
        } else if (this.schedulePeriod === 'month') {
            // 计算当月天数
            const year = start.getFullYear();
            const month = start.getMonth();
            days = new Date(year, month + 1, 0).getDate();
            // 从月初开始
            start.setDate(1);
        }
        
        for (let i = 0; i < days; i++) {
            const d = new Date(start);
            d.setDate(start.getDate() + i);
            dates.push(d);
        }
        return dates;
    }
    
    // 获取周期标签
    getPeriodLabel() {
        const dates = this.getWeekDates();
        if (dates.length === 0) return '';
        
        const start = dates[0];
        const end = dates[dates.length - 1];
        
        if (this.schedulePeriod === 'month') {
            return `${start.getFullYear()}年${start.getMonth() + 1}月`;
        } else {
            return `${start.getMonth() + 1}月${start.getDate()}日 - ${end.getMonth() + 1}月${end.getDate()}日`;
        }
    }

    prevWeek() {
        if (this.schedulePeriod === 'month') {
            this.currentWeekStart.setMonth(this.currentWeekStart.getMonth() - 1);
            this.currentWeekStart.setDate(1);
        } else {
            this.currentWeekStart.setDate(this.currentWeekStart.getDate() - this.schedulePeriod);
        }
        this.initCurrentWeekStatus();
        this.loadWeekSchedule();
    }

    nextWeek() {
        if (this.schedulePeriod === 'month') {
            this.currentWeekStart.setMonth(this.currentWeekStart.getMonth() + 1);
            this.currentWeekStart.setDate(1);
        } else {
            this.currentWeekStart.setDate(this.currentWeekStart.getDate() + this.schedulePeriod);
        }
        this.initCurrentWeekStatus();
        this.loadWeekSchedule();
    }

    goToToday() {
        this.currentWeekStart = this.getWeekStart(new Date());
        this.initCurrentWeekStatus();
        this.loadWeekSchedule();
    }

    // 员工操作
    addEmployee(employee) {
        employee.id = 'emp-' + Date.now();
        this.employees.push(employee);
        this.saveToStorage('employees', this.employees);
    }

    updateEmployee(id, updates) {
        const index = this.employees.findIndex(e => e.id === id);
        if (index !== -1) {
            this.employees[index] = { ...this.employees[index], ...updates };
            this.saveToStorage('employees', this.employees);
        }
    }

    deleteEmployee(id) {
        this.employees = this.employees.filter(e => e.id !== id);
        this.saveToStorage('employees', this.employees);
    }

    getEmployee(id) {
        return this.employees.find(e => e.id === id);
    }

    // 班次操作
    addShift(shift) {
        shift.id = 'shift-' + Date.now();
        this.shifts.push(shift);
        this.saveToStorage('shifts', this.shifts);
    }

    updateShift(id, updates) {
        const index = this.shifts.findIndex(s => s.id === id);
        if (index !== -1) {
            this.shifts[index] = { ...this.shifts[index], ...updates };
            this.saveToStorage('shifts', this.shifts);
        }
    }

    deleteShift(id) {
        this.shifts = this.shifts.filter(s => s.id !== id);
        this.saveToStorage('shifts', this.shifts);
    }

    getShift(id) {
        return this.shifts.find(s => s.id === id);
    }

    // 设置操作
    updateSettings(updates) {
        this.settings = { ...this.settings, ...updates };
        this.saveToStorage('settings', this.settings);
    }

    resetSettings() {
        this.settings = { ...defaultSettings };
        this.saveToStorage('settings', this.settings);
    }

    // 需求操作
    updateRequirements(requirements) {
        this.requirements = requirements;
        this.saveToStorage('requirements', this.requirements);
    }

    // 获取指定日期和门店的需求
    // 如果不指定门店ID，则使用当前门店或默认配置
    getRequirementsForDate(date, storeId = null) {
        const dayOfWeek = date.getDay();
        const isWeekend = dayOfWeek === 0 || dayOfWeek === 6;
        
        // 确定使用哪个门店的需求配置
        // 如果是"全部门店"模式且没有指定storeId，使用默认配置
        let targetStoreId = storeId;
        if (!targetStoreId) {
            targetStoreId = this.isAllStoresMode() ? 'store-001' : this.currentStoreId;
        }
        
        // 兼容旧格式（直接的weekday/weekend结构）
        if (this.requirements.weekday && !this.requirements['_default']) {
            // 旧格式数据，直接返回
            return isWeekend ? (this.requirements.weekend || {}) : (this.requirements.weekday || {});
        }
        
        // 获取门店特定的需求配置，如果没有则使用默认配置
        let storeReqs = this.requirements[targetStoreId];
        if (!storeReqs) {
            storeReqs = this.requirements['_default'];
        }
        
        if (!storeReqs) {
            // 如果还是没有，返回空的需求
            return {};
        }
        
        return isWeekend ? (storeReqs.weekend || {}) : (storeReqs.weekday || {});
    }

    // 重置所有数据
    resetAllData() {
        this.stores = [...defaultStores];
        this.shifts = [...defaultShifts];
        this.employees = [...defaultEmployees];
        this.requirements = { ...defaultRequirements };
        this.settings = { ...defaultSettings };
        this.scheduleWeeks = {}; // 重置所有周的数据（包含历史）
        this.currentStoreId = 'store-001';
        this.saveToStorage('stores', this.stores);
        this.saveToStorage('shifts', this.shifts);
        this.saveToStorage('employees', this.employees);
        this.saveToStorage('requirements', this.requirements);
        this.saveToStorage('settings', this.settings);
        this.saveToStorage('scheduleWeeks', this.scheduleWeeks);
    }

    // 换班历史操作 - 改为存储到当前周
    addHistoryRecord(record) {
        const schedule = this.getCurrentWeekSchedule();
        if (!schedule) return null;
        
        // 确保 history 字段存在
        if (!schedule.history) {
            schedule.history = [];
        }
        
        const historyRecord = {
            id: 'hist-' + Date.now(),
            timestamp: new Date().toISOString(),
            ...record
        };
        schedule.history.unshift(historyRecord); // 最新的在前面
        // 只保留最近100条记录
        if (schedule.history.length > 100) {
            schedule.history = schedule.history.slice(0, 100);
        }
        this.saveToStorage('scheduleWeeks', this.scheduleWeeks);
        return historyRecord;
    }

    // 获取当前周的历史记录
    get shiftHistory() {
        const schedule = this.getCurrentWeekSchedule();
        return schedule && schedule.history ? schedule.history : [];
    }

    clearHistory() {
        const schedule = this.getCurrentWeekSchedule();
        if (schedule) {
            schedule.history = [];
            this.saveToStorage('scheduleWeeks', this.scheduleWeeks);
        }
    }

    getHistoryByDate(dateStr) {
        return this.shiftHistory.filter(h => h.date === dateStr);
    }

    getHistoryByEmployee(employeeName) {
        return this.shiftHistory.filter(h => 
            h.employeeName === employeeName || 
            h.fromEmployee === employeeName || 
            h.toEmployee === employeeName
        );
    }
}

// 创建全局状态实例
const appState = new AppState();
