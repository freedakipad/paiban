/**
 * 餐饮门店智能排班系统 - API 集成
 */

class ScheduleAPI {
    constructor() {
        this.baseUrl = appState.settings.apiEndpoint;
        this.timeout = appState.settings.timeout * 1000;
    }

    updateConfig() {
        this.baseUrl = appState.settings.apiEndpoint;
        this.timeout = appState.settings.timeout * 1000;
    }

    /**
     * 测试API连接
     */
    async testConnection() {
        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 5000);
            
            const response = await fetch(`${this.baseUrl}/health`, {
                method: 'GET',
                signal: controller.signal
            });
            
            clearTimeout(timeoutId);
            
            if (response.ok) {
                const data = await response.json();
                return { success: true, data };
            }
            return { success: false, error: `HTTP ${response.status}` };
        } catch (error) {
            return { success: false, error: error.message };
        }
    }

    /**
     * 构建排班请求数据
     */
    buildScheduleRequest(weekDates) {
        const employees = this.buildEmployees();
        const shifts = this.buildShifts();
        const requirements = this.buildRequirements(weekDates);
        const constraints = this.buildConstraints();
        
        // 计算日期范围
        const startDate = formatDate(weekDates[0]);
        const endDate = formatDate(weekDates[weekDates.length - 1]);
        
        // 根据排班天数动态调整超时时间
        const days = weekDates.length;
        let timeout = appState.settings.timeout;
        if (days > 14) {
            timeout = Math.max(timeout, 60); // 月度排班至少60秒
        } else if (days > 7) {
            timeout = Math.max(timeout, 45); // 2周排班至少45秒
        }
        
        return {
            org_id: '550e8400-e29b-41d4-a716-446655440000', // 固定的演示组织ID
            start_date: startDate,
            end_date: endDate,
            scenario: 'restaurant',
            employees,
            shifts,
            requirements,
            constraints,
            options: {
                timeout_seconds: timeout,
                optimization_level: 2, // 平衡模式
                respect_preferences: true
            }
        };
    }

    /**
     * 构建员工数据
     * "全部门店"模式：返回所有活跃员工，并标记其所属门店和优先级
     * 单店模式：返回当前门店员工+可调配员工
     * 
     * 排班优先级策略：
     * - 本店员工优先用于本店需求（priority_weight: 100）
     * - 可调配员工用于其他门店需求（priority_weight: 50）
     * - 不可调配员工只能用于本店（can_transfer: false）
     */
    buildEmployees() {
        // 获取员工：全部模式获取所有活跃员工，否则获取当前门店+可调配员工
        const employees = appState.isAllStoresMode()
            ? appState.employees.filter(e => e.status === 'active')
            : appState.getCurrentStoreEmployees(true).filter(e => e.status === 'active');
        
        return employees.map(e => {
            // 为员工生成唯一ID，同时保存本地ID映射
            const uuid = generateUUID();
            this._employeeUUIDs = this._employeeUUIDs || {};
            this._employeeUUIDs[e.id] = uuid;
            
            return {
                id: uuid,
                name: e.name,
                position: e.position,
                skills: e.skills || [],
                status: e.status,
                store_id: e.storeId,           // 员工所属门店
                can_transfer: e.canTransfer || false,    // 是否可跨店调配（默认不可）
                home_store_priority: 100,      // 本店需求优先级
                transfer_priority: e.canTransfer ? 50 : 0,  // 跨店调配优先级
                preferences: e.preferences ? {
                    preferred_shifts: (e.preferences.preferredShifts || []).map(sid => {
                        const shift = appState.getShift(sid);
                        return shift ? this.getShiftUUID(sid) : null;
                    }).filter(Boolean),
                    avoid_shifts: (e.preferences.avoidShifts || []).map(sid => {
                        const shift = appState.getShift(sid);
                        return shift ? this.getShiftUUID(sid) : null;
                    }).filter(Boolean),
                    avoid_days: e.preferences.avoidDays || [],
                    max_hours_per_week: e.preferences.maxHoursPerWeek || 44
                } : undefined
            };
        });
    }

    /**
     * 构建班次数据
     */
    buildShifts() {
        return appState.shifts.map(s => ({
            id: this.getShiftUUID(s.id),
            name: s.name,
            code: s.code,
            start_time: s.startTime,
            end_time: s.endTime
        }));
    }

    // 存储班次ID映射
    _shiftUUIDs = {};

    getShiftUUID(localId) {
        if (!this._shiftUUIDs[localId]) {
            this._shiftUUIDs[localId] = generateUUID();
        }
        return this._shiftUUIDs[localId];
    }

    getLocalShiftId(uuid) {
        for (const [localId, u] of Object.entries(this._shiftUUIDs)) {
            if (u === uuid) return localId;
        }
        return null;
    }

    /**
     * 构建需求数据
     * "全部门店"模式：为每个门店生成独立的需求，带有store_id标识
     * 每个门店有自己的需求配置，根据门店类型差异化
     */
    buildRequirements(weekDates) {
        const requirements = [];
        
        // 获取需要排班的门店列表
        const stores = appState.isAllStoresMode() 
            ? appState.getAllStores()
            : [appState.getCurrentStore()].filter(Boolean);
        
        stores.forEach(store => {
            weekDates.forEach(date => {
                const dateStr = formatDate(date);
                // 为每个门店获取其特定的需求配置
                const dayReqs = appState.getRequirementsForDate(date, store.id);
                
                appState.shifts.forEach(shift => {
                    const shiftReqs = dayReqs[shift.id];
                    if (!shiftReqs) return;
                    
                    Object.entries(shiftReqs).forEach(([position, count]) => {
                        if (count > 0) {
                            requirements.push({
                                id: generateUUID(),
                                date: dateStr,
                                shift_id: this.getShiftUUID(shift.id),
                                store_id: store.id,      // 门店ID
                                store_name: store.name,  // 门店名称（便于调试）
                                position: position,
                                min_employees: count,
                                priority: position === '厨师' ? 9 : 8,
                                note: `${store.name} ${getDayName(date)} ${shift.name} - ${position}`
                            });
                        }
                    });
                });
            });
        });
        
        return requirements;
    }

    /**
     * 构建约束数据 - 后端期望map[string]interface{}格式
     */
    buildConstraints() {
        const { hoursMode, maxWeeklyHours, maxPeriodHours, minRestHours, maxConsecutiveDays, minRestDays, maxShiftsPerMonth } = appState.settings;
        
        const constraints = {
            hours_mode: hoursMode || 'weekly',
            min_rest_between_shifts: minRestHours,
            max_consecutive_days: maxConsecutiveDays,
            min_rest_days_per_week: minRestDays
        };
        
        // 根据工时模式设置相应参数
        if (hoursMode === 'period') {
            constraints.max_hours_per_period = maxPeriodHours || 176;
            constraints.max_hours_per_week = 999; // 禁用周工时约束
        } else {
            constraints.max_weekly_hours = maxWeeklyHours;
            constraints.max_hours_per_week = maxWeeklyHours;
        }
        
        // 多门店联合排班模式：防止员工在同一时间段被分配到不同门店
        if (appState.isAllStoresMode()) {
            constraints.multi_store_mode = true;
            constraints.prevent_duplicate_assignments = true;  // 员工同一时间只能分配一次
            constraints.prefer_home_store = true;              // 优先本店员工
            constraints.use_transfer_as_backup = true;         // 机动人员作为补充
        }
        
        return constraints;
    }

    /**
     * 生成排班
     * 多门店模式采用分阶段策略：
     * 阶段1：每个门店用本店员工排班
     * 阶段2：用可调配员工填补未满足需求
     */
    async generateSchedule(weekDates) {
        // 如果是"全部门店"模式，使用分阶段排班策略
        if (appState.isAllStoresMode()) {
            return this.generateMultiStoreSchedule(weekDates);
        }
        
        // 单店模式使用常规排班
        return this.generateSingleSchedule(weekDates);
    }
    
    /**
     * 多门店分阶段排班
     * 阶段1：每个门店独立排班（只用本店员工）
     * 阶段2：用可调配员工填补所有未满足需求
     */
    async generateMultiStoreSchedule(weekDates) {
        console.log('🏢 启动多门店分阶段排班策略');
        
        const allAssignments = [];
        const allUnfilledReqs = [];
        const allViolations = [];
        let totalStats = { fulfillmentRate: 0, assignmentCount: 0, avgScore: 0, totalRequired: 0 };
        
        // 跟踪每个员工每天的分配情况（防止同一员工同一天多班）
        const employeeDayAssigned = {}; // key: employeeName-date, value: true
        
        // ===== 阶段1：每个门店用本店员工独立排班 =====
        console.log('📌 阶段1：各门店本店员工排班');
        const stores = appState.getAllStores();
        
        for (const store of stores) {
            console.log(`  🏪 正在为 ${store.name} 排班...`);
            
            // 只用本店员工，且排除当天已分配的员工
            const storeEmployees = appState.employees.filter(
                e => e.status === 'active' && e.storeId === store.id
            );
            
            // 本店需求
            const storeRequirements = this.buildRequirementsForStore(weekDates, store);
            
            if (storeEmployees.length === 0 || storeRequirements.length === 0) {
                console.log(`    ⚠️ ${store.name} 无员工或无需求，跳过`);
                allUnfilledReqs.push(...storeRequirements.map(r => ({
                    date: r.date,
                    shiftId: this.getLocalShiftId(r.shift_id) || r.shift_id,
                    position: r.position,
                    required: r.min_employees,
                    assigned: 0,
                    reason: '无本店员工',
                    storeId: store.id,
                    storeName: store.name
                })));
                continue;
            }
            
            const requestData = this.buildStoreScheduleRequest(weekDates, store, storeEmployees, storeRequirements);
            
            try {
                const result = await this.sendScheduleRequest(requestData);
                
                // 收集本店排班结果
                allAssignments.push(...result.assignments);
                
                // 更新员工每天分配跟踪
                result.assignments.forEach(a => {
                    const key = `${a.employeeName}-${a.date}`;
                    employeeDayAssigned[key] = true;
                });
                
                allUnfilledReqs.push(...result.unfilledRequirements.map(u => ({
                    ...u,
                    storeId: store.id,
                    storeName: store.name
                })));
                allViolations.push(...(result.constraintViolations || []));
                
                console.log(`    ✅ ${store.name} 排班完成: ${result.assignments.length} 班次`);
            } catch (error) {
                console.error(`    ❌ ${store.name} 排班失败:`, error.message);
            }
        }
        
        // ===== 阶段2：用可调配员工填补未满足需求 =====
        if (allUnfilledReqs.length > 0) {
            console.log(`📌 阶段2：机动人员补充排班（${allUnfilledReqs.length} 个未满足需求）`);
            
            // 获取所有可调配员工
            const transferEmployees = appState.employees.filter(
                e => e.status === 'active' && e.canTransfer
            );
            
            if (transferEmployees.length > 0) {
                // 构建补充排班请求，传入已分配跟踪
                const supplementResult = await this.generateSupplementSchedule(
                    weekDates, 
                    allUnfilledReqs, 
                    transferEmployees,
                    allAssignments,
                    employeeDayAssigned  // 传入员工每天分配跟踪
                );
                
                if (supplementResult.assignments.length > 0) {
                    console.log(`    ✅ 机动人员补充了 ${supplementResult.assignments.length} 班次`);
                    allAssignments.push(...supplementResult.assignments);
                    
                    // 更新未满足需求列表
                    allUnfilledReqs.length = 0;
                    allUnfilledReqs.push(...supplementResult.unfilledRequirements);
                }
            } else {
                console.log('    ⚠️ 无可调配员工，无法补充');
            }
        }
        
        // 计算总体统计
        const totalRequired = this.calculateTotalRequired(weekDates);
        const satisfactionRate = Math.round((allAssignments.length / totalRequired) * 100);
        const avgScore = allAssignments.length > 0 
            ? Math.round(allAssignments.reduce((sum, a) => sum + (a.score || 0), 0) / allAssignments.length)
            : 0;
        
        console.log(`🎯 多门店排班完成: ${allAssignments.length}/${totalRequired} 班次, 满足率 ${satisfactionRate}%`);
        
        return {
            success: true,
            message: allUnfilledReqs.length > 0 
                ? `生成了部分排班方案，存在${allUnfilledReqs.length}个未满足的需求`
                : '排班成功完成',
            assignments: allAssignments,
            unfilledRequirements: allUnfilledReqs,
            constraintViolations: allViolations,
            statistics: {
                satisfactionRate,
                assignmentCount: allAssignments.length,
                avgScore,
                totalRequired
            }
        };
    }
    
    /**
     * 为单个门店构建排班请求
     */
    buildStoreScheduleRequest(weekDates, store, employees, requirements) {
        const shifts = this.buildShifts();
        const constraints = this.buildConstraints();
        
        const startDate = formatDate(weekDates[0]);
        const endDate = formatDate(weekDates[weekDates.length - 1]);
        
        return {
            org_id: '550e8400-e29b-41d4-a716-446655440000', // 固定UUID格式
            start_date: startDate,
            end_date: endDate,
            scenario: 'restaurant',
            employees: employees.map(e => this.buildEmployeeData(e)),
            shifts,
            requirements,
            constraints,
            options: {
                timeout_seconds: Math.max(15, Math.round(appState.settings.timeout / 3)),
                optimization_level: 2,
                respect_preferences: true
            }
        };
    }
    
    /**
     * 构建单个员工数据
     */
    buildEmployeeData(e) {
        const uuid = generateUUID();
        this._employeeUUIDs = this._employeeUUIDs || {};
        this._employeeUUIDs[e.id] = uuid;
        
        return {
            id: uuid,
            name: e.name,
            position: e.position,
            skills: e.skills || [],
            status: e.status,
            store_id: e.storeId,
            can_transfer: e.canTransfer || false,
            preferences: e.preferences ? {
                preferred_shifts: (e.preferences.preferredShifts || []).map(sid => {
                    const shift = appState.getShift(sid);
                    return shift ? this.getShiftUUID(sid) : null;
                }).filter(Boolean),
                avoid_shifts: (e.preferences.avoidShifts || []).map(sid => {
                    const shift = appState.getShift(sid);
                    return shift ? this.getShiftUUID(sid) : null;
                }).filter(Boolean),
                avoid_days: e.preferences.avoidDays || [],
                max_hours_per_week: e.preferences.maxHoursPerWeek || 44
            } : undefined
        };
    }
    
    /**
     * 为单个门店构建需求
     */
    buildRequirementsForStore(weekDates, store) {
        const requirements = [];
        
        weekDates.forEach(date => {
            const dateStr = formatDate(date);
            const dayReqs = appState.getRequirementsForDate(date, store.id);
            
            appState.shifts.forEach(shift => {
                const shiftReqs = dayReqs[shift.id];
                if (!shiftReqs) return;
                
                Object.entries(shiftReqs).forEach(([position, count]) => {
                    if (count > 0) {
                        requirements.push({
                            id: generateUUID(),
                            date: dateStr,
                            shift_id: this.getShiftUUID(shift.id),
                            store_id: store.id,
                            store_name: store.name,
                            position: position,
                            min_employees: count,
                            priority: position === '厨师' ? 9 : 8,
                            note: `${store.name} ${getDayName(date)} ${shift.name} - ${position}`
                        });
                    }
                });
            });
        });
        
        return requirements;
    }
    
    /**
     * 用机动人员补充排班
     * 按日期分别处理，确保每个员工每天只分配一次
     */
    async generateSupplementSchedule(weekDates, unfilledReqs, transferEmployees, existingAssignments, employeeDayAssigned = {}) {
        const allSupplementAssignments = [];
        const remainingUnfilled = [];
        
        // 按日期分组未满足需求
        const reqsByDate = {};
        unfilledReqs.forEach(u => {
            const shortage = (u.required || 1) - (u.assigned || 0);
            if (shortage > 0) {
                if (!reqsByDate[u.date]) reqsByDate[u.date] = [];
                reqsByDate[u.date].push({
                    ...u,
                    shortage
                });
            }
        });
        
        // 按日期逐个处理
        for (const date of Object.keys(reqsByDate).sort()) {
            const dateReqs = reqsByDate[date];
            
            // 过滤当天可用的员工（排除已分配的）
            const availableEmployees = transferEmployees.filter(e => {
                const key = `${e.name}-${date}`;
                return !employeeDayAssigned[key];
            });
            
            if (availableEmployees.length === 0) {
                // 当天无可用员工，记录为未满足
                remainingUnfilled.push(...dateReqs.map(r => ({
                    date: r.date,
                    shiftId: r.shiftId,
                    position: r.position,
                    required: r.required,
                    assigned: r.assigned,
                    storeId: r.storeId,
                    storeName: r.storeName,
                    reason: '无可用机动人员'
                })));
                continue;
            }
            
            // 构建当天的补充需求
            const daySupplementReqs = dateReqs.map(u => ({
                id: generateUUID(),
                date: u.date,
                shift_id: this.getShiftUUID(u.shiftId) || u.shiftId,
                store_id: u.storeId,
                store_name: u.storeName,
                position: u.position,
                min_employees: u.shortage,
                priority: u.position === '厨师' ? 9 : 8,
                note: `补充: ${u.storeName} ${u.date} - ${u.position}`
            }));
            
            const requestData = {
                org_id: '550e8400-e29b-41d4-a716-446655440000',
                start_date: date,
                end_date: date,
                scenario: 'restaurant',
                employees: availableEmployees.map(e => this.buildEmployeeData(e)),
                shifts: this.buildShifts(),
                requirements: daySupplementReqs,
                constraints: this.buildConstraints(),
                options: {
                    timeout_seconds: 10,
                    optimization_level: 2,
                    respect_preferences: true
                }
            };
            
            try {
                const result = await this.sendScheduleRequest(requestData);
                
                // 收集结果并更新跟踪，添加工作门店信息
                result.assignments.forEach(a => {
                    // 为补充排班添加工作门店信息（从需求中获取）
                    const req = dateReqs.find(r => r.position === a.position);
                    if (req) {
                        a.workStoreId = req.storeId;
                        a.workStoreName = req.storeName;
                        const workStore = appState.stores.find(s => s.id === req.storeId);
                        a.workStoreCode = workStore?.code || '';
                    }
                    allSupplementAssignments.push(a);
                    const key = `${a.employeeName}-${a.date}`;
                    employeeDayAssigned[key] = true;
                });
                
                // 收集未满足需求
                if (result.unfilledRequirements) {
                    remainingUnfilled.push(...result.unfilledRequirements.map(u => ({
                        ...u,
                        storeId: dateReqs[0]?.storeId,
                        storeName: dateReqs[0]?.storeName
                    })));
                }
            } catch (error) {
                console.error(`补充排班失败 (${date}):`, error.message);
                remainingUnfilled.push(...dateReqs.map(r => ({
                    date: r.date,
                    shiftId: r.shiftId,
                    position: r.position,
                    required: r.required,
                    assigned: r.assigned,
                    storeId: r.storeId,
                    storeName: r.storeName,
                    reason: '排班请求失败'
                })));
            }
        }
        
        return {
            assignments: allSupplementAssignments,
            unfilledRequirements: remainingUnfilled
        };
    }
    
    /**
     * 计算总需求数
     */
    calculateTotalRequired(weekDates) {
        let total = 0;
        const stores = appState.isAllStoresMode() 
            ? appState.getAllStores() 
            : [appState.getCurrentStore()].filter(Boolean);
        
        stores.forEach(store => {
            weekDates.forEach(date => {
                const dayReqs = appState.getRequirementsForDate(date, store.id);
                appState.shifts.forEach(shift => {
                    const shiftReqs = dayReqs[shift.id];
                    if (shiftReqs) {
                        Object.values(shiftReqs).forEach(count => {
                            total += count;
                        });
                    }
                });
            });
        });
        
        return total;
    }
    
    /**
     * 发送排班请求到后端
     */
    async sendScheduleRequest(requestData) {
        const days = requestData.requirements.length / 4; // 粗略估计天数
        let httpTimeout = this.timeout;
        if (days > 14) {
            httpTimeout = Math.max(httpTimeout, 65000);
        } else if (days > 7) {
            httpTimeout = Math.max(httpTimeout, 50000);
        }
        
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), httpTimeout);
        
        try {
            const response = await fetch(`${this.baseUrl}/api/v1/schedule/generate`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(requestData),
                signal: controller.signal
            });
            
            clearTimeout(timeoutId);
            
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `HTTP ${response.status}`);
            }
            
            const data = await response.json();
            return this.processScheduleResponse(data, requestData);
            
        } catch (error) {
            clearTimeout(timeoutId);
            if (error.name === 'AbortError') {
                throw new Error('请求超时');
            }
            throw error;
        }
    }
    
    /**
     * 单店排班（常规模式）
     */
    async generateSingleSchedule(weekDates) {
        const requestData = this.buildScheduleRequest(weekDates);
        
        // 调试输出
        console.log('排班请求数据:', JSON.stringify(requestData, null, 2));
        
        // 根据排班天数动态调整HTTP超时时间
        const days = weekDates.length;
        let httpTimeout = this.timeout;
        if (days > 14) {
            httpTimeout = Math.max(httpTimeout, 65000); // 月度排班65秒
        } else if (days > 7) {
            httpTimeout = Math.max(httpTimeout, 50000); // 2周排班50秒
        }
        
        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), httpTimeout);
            
            const response = await fetch(`${this.baseUrl}/api/v1/schedule/generate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(requestData),
                signal: controller.signal
            });
            
            clearTimeout(timeoutId);
            
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `HTTP ${response.status}`);
            }
            
            const data = await response.json();
            return this.processScheduleResponse(data, requestData);
            
        } catch (error) {
            if (error.name === 'AbortError') {
                throw new Error('请求超时，请检查排班引擎是否正常运行');
            }
            throw error;
        }
    }

    /**
     * 处理排班响应
     */
    processScheduleResponse(response, request) {
        const assignments = (response.assignments || []).map(a => {
            // 查找对应的班次
            const localShiftId = this.getLocalShiftId(a.shift_id);
            const shift = localShiftId ? appState.getShift(localShiftId) : null;
            
            // 查找对应的员工
            const emp = request.employees.find(e => e.id === a.employee_id);
            
            // 查找员工所属门店（从本地员工数据获取）
            // 优先通过员工名称匹配本地员工
            const localEmp = appState.employees.find(e => e.name === (a.employee_name || emp?.name));
            const storeId = localEmp?.storeId || emp?.store_id || null;
            const store = storeId ? appState.stores.find(s => s.id === storeId) : null;
            
            // 使用本地员工ID（如果找到），否则使用后端返回的ID
            const localEmployeeId = localEmp?.id || a.employee_id;
            
            // 工作门店（从后端返回或请求中获取）
            const workStoreId = a.work_store_id || a.store_id || request.store_id || storeId;
            const workStore = workStoreId ? appState.stores.find(s => s.id === workStoreId) : store;
            
            return {
                id: a.id,
                employeeId: localEmployeeId,  // 使用本地员工ID以便统计匹配
                employeeName: a.employee_name || (emp ? emp.name : '未知'),
                shiftId: localShiftId || a.shift_id,
                shiftName: a.shift_name || (shift ? shift.name : '未知'),
                date: a.date,
                startTime: a.start_time,
                endTime: a.end_time,
                position: a.position,
                hours: a.hours,
                score: a.score,
                scoreDetail: a.score_detail,
                storeId: storeId,                          // 员工所属门店ID
                storeName: store?.name || '未知门店',       // 员工所属门店名称
                storeCode: store?.code || '',              // 员工所属门店代码
                workStoreId: workStoreId,                  // 工作门店ID
                workStoreName: workStore?.name || '未知门店', // 工作门店名称
                workStoreCode: workStore?.code || ''       // 工作门店代码
            };
        });
        
        const unfilledRequirements = (response.unfilled || response.unfilled_requirements || []).map(u => ({
            date: u.date,
            shiftId: this.getLocalShiftId(u.shift_id) || u.shift_id,
            shiftName: u.shift_name,
            position: u.position,
            required: u.required || u.needed || 1,
            assigned: u.assigned || 0,
            storeName: u.store_name || '',
            storeId: u.store_id || '',
            reason: u.reason
        }));
        
        // 提取约束违反信息
        const constraintViolations = [];
        if (response.constraint_result?.hard_violations) {
            response.constraint_result.hard_violations.forEach(v => {
                constraintViolations.push({
                    type: 'hard',
                    constraintType: v.constraint_type,
                    constraintName: v.constraint_name,
                    message: v.message,
                    severity: v.severity || 'error'
                });
            });
        }
        if (response.constraint_result?.soft_violations) {
            response.constraint_result.soft_violations.forEach(v => {
                constraintViolations.push({
                    type: 'soft',
                    constraintType: v.constraint_type,
                    constraintName: v.constraint_name,
                    message: v.message,
                    severity: v.severity || 'warning'
                });
            });
        }
        
        return {
            success: response.success,
            message: response.message,
            assignments,
            unfilledRequirements,
            constraintViolations,
            staffingSuggestions: response.suggestions || [],  // 补员建议
            statistics: {
                totalAssignments: assignments.length,
                totalHours: response.statistics?.total_hours || assignments.reduce((sum, a) => sum + a.hours, 0),
                fulfillmentRate: response.statistics?.fulfillment_rate || 
                    (request.requirements.length > 0 
                        ? Math.round((assignments.length / request.requirements.length) * 100) 
                        : 100),
                averageScore: response.statistics?.average_score || 
                    (assignments.length > 0 
                        ? Math.round(assignments.reduce((sum, a) => sum + (a.score || 0), 0) / assignments.length) 
                        : 0),
                violations: response.statistics?.violations || [],
                constraintScore: response.constraint_result?.score
            },
            computeTime: response.compute_time_ms
        };
    }

    /**
     * 根据每月最大班次数限制过滤排班
     * 考虑已有排班，确保每个员工每月总班次不超过限制
     */
    filterByMonthlyShiftLimit(newAssignments, maxShiftsPerMonth) {
        // 统计每个员工每月已有的班次数（从现有排班中）
        const employeeMonthlyShifts = {};
        
        // 先统计现有排班
        (appState.assignments || []).forEach(a => {
            const month = a.date.substring(0, 7); // YYYY-MM
            const key = `${a.employeeName}-${month}`;
            employeeMonthlyShifts[key] = (employeeMonthlyShifts[key] || 0) + 1;
        });
        
        // 过滤新排班，确保不超过限制
        const filtered = [];
        newAssignments.forEach(a => {
            const month = a.date.substring(0, 7);
            const key = `${a.employeeName}-${month}`;
            const currentCount = employeeMonthlyShifts[key] || 0;
            
            if (currentCount < maxShiftsPerMonth) {
                filtered.push(a);
                employeeMonthlyShifts[key] = currentCount + 1;
            } else {
                console.log(`⚠️ 过滤排班: ${a.employeeName} 在 ${month} 已有 ${currentCount} 班，超出限制 ${maxShiftsPerMonth}`);
            }
        });
        
        return filtered;
    }
    
    /**
     * 计算员工当前周期的班次数（用于均衡分配）
     */
    getEmployeeShiftCounts(weekDates) {
        const counts = {};
        const startDate = formatDate(weekDates[0]);
        const endDate = formatDate(weekDates[weekDates.length - 1]);
        
        // 统计当前周期内的排班
        (appState.assignments || []).forEach(a => {
            if (a.date >= startDate && a.date <= endDate) {
                counts[a.employeeName] = (counts[a.employeeName] || 0) + 1;
            }
        });
        
        return counts;
    }
    
    /**
     * 对员工列表按班次数排序（班次少的优先）
     */
    sortEmployeesByWorkload(employees, shiftCounts) {
        return [...employees].sort((a, b) => {
            const countA = shiftCounts[a.name] || 0;
            const countB = shiftCounts[b.name] || 0;
            return countA - countB; // 班次少的排前面
        });
    }

    /**
     * 验证排班
     */
    async validateSchedule(assignments) {
        const requestData = {
            employees: this.buildEmployees(),
            shifts: this.buildShifts(),
            assignments: assignments.map(a => ({
                employee_id: a.employeeId,
                shift_id: a.shiftId,
                date: a.date
            })),
            constraints: this.buildConstraints()
        };
        
        try {
            const response = await fetch(`${this.baseUrl}/api/v1/schedule/validate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(requestData)
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            throw error;
        }
    }
}

// 创建全局API实例
const scheduleAPI = new ScheduleAPI();
