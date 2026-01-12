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
        // 计算日期范围（需要在 buildEmployees 之前计算）
        const startDate = formatDate(weekDates[0]);
        const endDate = formatDate(weekDates[weekDates.length - 1]);
        
        const employees = this.buildEmployees(startDate, endDate);
        const shifts = this.buildShifts();
        const requirements = this.buildRequirements(weekDates);
        const constraints = this.buildConstraints();
        
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
     * 
     * @param {string} startDate - 排班起始日期（用于排除旧排班）
     * @param {string} endDate - 排班结束日期
     */
    buildEmployees(startDate = null, endDate = null) {
        // 获取员工：全部模式获取所有活跃员工，否则获取当前门店+可调配员工
        const employees = appState.isAllStoresMode()
            ? appState.employees.filter(e => e.status === 'active')
            : appState.getCurrentStoreEmployees(true).filter(e => e.status === 'active');
        
        // 计算每个员工各月已有班次数（排除当前排班日期范围内的旧排班）
        // 返回格式: { employeeName: { "2026-01": 5, "2026-02": 3 }, ... }
        const monthlyShiftCounts = this.getEmployeeMonthlyShiftCounts(startDate, endDate);
        
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
                monthly_shifts_counts: monthlyShiftCounts[e.name] || {},  // 各月已有班次数 { "YYYY-MM": count }
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
        const { hoursMode, maxWeeklyHours, maxPeriodHours, minRestHours, maxConsecutiveDays, minRestDays, maxShiftsPerMonth, monthlyMaxShifts } = appState.settings;
        
        const constraints = {
            hours_mode: hoursMode || 'weekly',
            min_rest_between_shifts: minRestHours,
            max_consecutive_days: maxConsecutiveDays,
            min_rest_days_per_week: minRestDays,
            max_shifts_per_month: maxShiftsPerMonth || 26  // 每月最大班次数（默认值）
        };
        
        // 每月单独设置的最大班次数限制（如果有配置）
        // 格式: { "2026-01": 20, "2026-02": 26, ... }
        if (monthlyMaxShifts && Object.keys(monthlyMaxShifts).length > 0) {
            constraints.monthly_max_shifts = monthlyMaxShifts;
        }
        
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
        
        // 获取当前排班日期范围
        const startDate = formatDate(weekDates[0]);
        const endDate = formatDate(weekDates[weekDates.length - 1]);
        
        // 跟踪每个员工各月已分配的班次数（用于每月最大班次数约束）
        // 格式: { employeeName: { "YYYY-MM": count, ... }, ... }
        // 排除当前排班日期范围内的旧排班（这些会被新排班覆盖）
        const employeeMonthlyShifts = this.getEmployeeMonthlyShiftCounts(startDate, endDate);
        console.log(`📊 各月已有班次统计（排除 ${startDate} 至 ${endDate}）:`, employeeMonthlyShifts);
        
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
            
            // 传递累计的月班次数
            const requestData = this.buildStoreScheduleRequest(weekDates, store, storeEmployees, storeRequirements, employeeMonthlyShifts);
            
            try {
                const result = await this.sendScheduleRequest(requestData);
                
                // 收集本店排班结果，添加门店信息
                result.assignments.forEach(a => {
                    // 为本店排班添加门店信息（所属门店=工作门店=当前门店）
                    a.storeId = store.id;
                    a.storeName = store.name;
                    a.storeCode = store.code || '';
                    a.workStoreId = store.id;
                    a.workStoreName = store.name;
                    a.workStoreCode = store.code || '';
                    
                    allAssignments.push(a);
                    
                    const key = `${a.employeeName}-${a.date}`;
                    employeeDayAssigned[key] = true;
                    
                    // 按月份累计班次数
                    if (a.date && a.employeeName) {
                        const month = a.date.substring(0, 7);
                        if (!employeeMonthlyShifts[a.employeeName]) {
                            employeeMonthlyShifts[a.employeeName] = {};
                        }
                        employeeMonthlyShifts[a.employeeName][month] = 
                            (employeeMonthlyShifts[a.employeeName][month] || 0) + 1;
                    }
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
                // 构建补充排班请求，传入已分配跟踪和月班次数
                const supplementResult = await this.generateSupplementSchedule(
                    weekDates, 
                    allUnfilledReqs, 
                    transferEmployees,
                    allAssignments,
                    employeeDayAssigned,  // 传入员工每天分配跟踪
                    employeeMonthlyShifts  // 传入员工月班次数
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
        
        // ===== 阶段3：智能调配 - 从超编门店转移到缺编门店 =====
        // 重新分析所有门店的超编和缺编情况
        const staffingAnalysis = this.analyzeAllStoreStaffing(weekDates, allAssignments);
        const overstaffed = staffingAnalysis.overstaffed;
        const understaffed = staffingAnalysis.understaffed;
        
        console.log(`📌 阶段3：智能调配分析 - 超编 ${overstaffed.length} 处，缺编 ${understaffed.length} 处`);
        
        if (overstaffed.length > 0 && understaffed.length > 0) {
            let transferCount = 0;
            
            // 尝试从超编门店调配员工到缺编门店
            for (const shortage of understaffed) {
                if (shortage.gap <= 0) continue; // 已满足
                
                // 寻找同日期、同班次、同岗位的超编（不同门店）
                const matchingOverstaffed = overstaffed.find(o => 
                    o.date === shortage.date && 
                    o.shiftId === shortage.shiftId && 
                    o.position === shortage.position &&
                    o.storeId !== shortage.storeId &&
                    o.over > 0
                );
                
                if (matchingOverstaffed) {
                    console.log(`    🔍 找到匹配：${matchingOverstaffed.storeName} ${matchingOverstaffed.position} 超${matchingOverstaffed.over}人 → ${shortage.storeName} 缺${shortage.gap}人`);
                    
                    // 找到可以调配的排班（从超编门店，且员工可调配）
                    const transferCandidate = allAssignments.find(a => {
                        const workStore = a.workStoreId || a.storeId;
                        const isInOverstaffedStore = workStore === matchingOverstaffed.storeId;
                        const isMatchingSlot = a.date === matchingOverstaffed.date &&
                                               a.shiftId === matchingOverstaffed.shiftId &&
                                               a.position === matchingOverstaffed.position;
                        const notYetTransferred = !a.transferred;
                        
                        // 检查员工是否可调配
                        const emp = appState.employees.find(e => e.name === a.employeeName);
                        const canTransfer = emp && emp.canTransfer;
                        
                        return isInOverstaffedStore && isMatchingSlot && notYetTransferred && canTransfer;
                    });
                    
                    if (transferCandidate) {
                        // 执行调配：更新工作门店
                        const targetStore = appState.stores.find(s => s.id === shortage.storeId);
                        const sourceStore = appState.stores.find(s => s.id === matchingOverstaffed.storeId);
                        
                        console.log(`    🔄 调配 ${transferCandidate.employeeName} 从 ${sourceStore?.name || '?'} → ${targetStore?.name || '?'}`);
                        
                        // 保存原始门店信息
                        if (!transferCandidate.originalStoreId) {
                            transferCandidate.originalStoreId = transferCandidate.workStoreId || transferCandidate.storeId;
                            transferCandidate.originalStoreCode = transferCandidate.workStoreCode || transferCandidate.storeCode;
                        }
                        
                        transferCandidate.workStoreId = shortage.storeId;
                        transferCandidate.workStoreName = targetStore?.name || '';
                        transferCandidate.workStoreCode = targetStore?.code || '';  // 修复：更新门店代码
                        transferCandidate.transferred = true;
                        transferCandidate.transferNote = `从${sourceStore?.name || '超编门店'}调配`;
                        
                        // 更新统计
                        matchingOverstaffed.over--;
                        shortage.gap--;
                        transferCount++;
                    } else {
                        console.log(`    ⚠️ 未找到可调配的员工（${matchingOverstaffed.storeName}）`);
                    }
                }
            }
            
            if (transferCount > 0) {
                console.log(`    ✅ 完成 ${transferCount} 次跨店调配`);
            }
            
            // 更新未满足需求列表
            const remainingUnfilled = understaffed.filter(s => s.gap > 0);
            allUnfilledReqs.length = 0;
            remainingUnfilled.forEach(s => {
                allUnfilledReqs.push({
                    date: s.date,
                    shiftId: s.shiftId,
                    shiftName: s.shiftName,
                    position: s.position,
                    storeId: s.storeId,
                    storeName: s.storeName,
                    required: s.required,
                    assigned: s.assigned,
                    reason: '人员不足'
                });
            });
        }
        
        // 计算总体统计
        const totalRequired = this.calculateTotalRequired(weekDates);
        // 满足率使用向下取整，避免99.x%被误显示为100%
        const satisfactionRate = totalRequired > 0 
            ? Math.floor((allAssignments.length / totalRequired) * 100)
            : 100;
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
     * @param {Array} weekDates - 排班日期
     * @param {Object} store - 门店对象
     * @param {Array} employees - 员工列表
     * @param {Array} requirements - 需求列表
     * @param {Object} employeeMonthlyShifts - 员工当月累计班次数 {employeeName: count}
     */
    buildStoreScheduleRequest(weekDates, store, employees, requirements, employeeMonthlyShifts = {}) {
        const shifts = this.buildShifts();
        const constraints = this.buildConstraints();
        
        const startDate = formatDate(weekDates[0]);
        const endDate = formatDate(weekDates[weekDates.length - 1]);
        
        return {
            org_id: '550e8400-e29b-41d4-a716-446655440000', // 固定UUID格式
            start_date: startDate,
            end_date: endDate,
            scenario: 'restaurant',
            employees: employees.map(e => this.buildEmployeeData(e, employeeMonthlyShifts[e.name] || {})),
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
     * @param {Object} e - 员工对象
     * @param {Object} monthlyShiftsCounts - 各月已有班次数 { "YYYY-MM": count, ... }
     */
    buildEmployeeData(e, monthlyShiftsCounts = {}) {
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
            monthly_shifts_counts: monthlyShiftsCounts,  // 各月已有班次数 { "YYYY-MM": count }
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
     * 分析各门店各时段的人员配置情况
     * 返回超编和缺编两个列表（用于智能调配）
     */
    analyzeAllStoreStaffing(weekDates, assignments) {
        const overstaffed = [];
        const understaffed = [];
        const stores = appState.getAllStores();
        const allPositions = ['厨师', '服务员', '收银员'];
        
        weekDates.forEach(date => {
            const dateStr = formatDate(date);
            
            stores.forEach(store => {
                appState.shifts.forEach(shift => {
                    // 获取需求
                    const dayReqs = appState.getRequirementsForDate(date, store.id);
                    const shiftReqs = dayReqs[shift.id] || {};
                    
                    // 统计实际分配（考虑 workStoreId）
                    const positionAssigned = {};
                    assignments.forEach(a => {
                        if (a.date === dateStr && a.shiftId === shift.id) {
                            const workStore = a.workStoreId || a.storeId;
                            if (workStore === store.id) {
                                const pos = a.position || '未知';
                                positionAssigned[pos] = (positionAssigned[pos] || 0) + 1;
                            }
                        }
                    });
                    
                    // 检查所有岗位的超编和缺编
                    allPositions.forEach(pos => {
                        const required = shiftReqs[pos] || 0;
                        const assigned = positionAssigned[pos] || 0;
                        
                        if (assigned > required && required > 0) {
                            // 超编
                            overstaffed.push({
                                date: dateStr,
                                shiftId: shift.id,
                                shiftName: shift.name,
                                position: pos,
                                storeId: store.id,
                                storeName: store.name,
                                required,
                                assigned,
                                over: assigned - required
                            });
                        } else if (assigned < required) {
                            // 缺编
                            understaffed.push({
                                date: dateStr,
                                shiftId: shift.id,
                                shiftName: shift.name,
                                position: pos,
                                storeId: store.id,
                                storeName: store.name,
                                required,
                                assigned,
                                gap: required - assigned
                            });
                        }
                    });
                });
            });
        });
        
        return { overstaffed, understaffed };
    }
    
    /**
     * 用机动人员补充排班
     * 按日期分别处理，确保每个员工每天只分配一次
     * @param {Object} employeeMonthlyShifts - 员工当月累计班次数
     */
    async generateSupplementSchedule(weekDates, unfilledReqs, transferEmployees, existingAssignments, employeeDayAssigned = {}, employeeMonthlyShifts = {}) {
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
            const scheduleMonth = date.substring(0, 7);
            
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
                employees: availableEmployees.map(e => this.buildEmployeeData(e, employeeMonthlyShifts[e.name] || {})),
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
                    
                    // 更新月班次数（用于后续约束检查）
                    if (a.date && a.employeeName) {
                        const assignMonth = a.date.substring(0, 7);
                        if (!employeeMonthlyShifts[a.employeeName]) {
                            employeeMonthlyShifts[a.employeeName] = {};
                        }
                        employeeMonthlyShifts[a.employeeName][assignMonth] = 
                            (employeeMonthlyShifts[a.employeeName][assignMonth] || 0) + 1;
                    }
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
     * 获取每个员工各月已有的班次数
     * 返回格式: { employeeName: { "2026-01": 5, "2026-02": 3 }, ... }
     * @param {string} excludeStartDate - 排除的日期范围起始（重新排班时，排除即将被覆盖的日期）
     * @param {string} excludeEndDate - 排除的日期范围结束
     */
    getEmployeeMonthlyShiftCounts(excludeStartDate = null, excludeEndDate = null) {
        // 结构: { employeeName: { "YYYY-MM": count, ... }, ... }
        const counts = {};
        
        (appState.assignments || []).forEach(a => {
            if (!a.date || !a.employeeName) return;
            
            // 如果指定了排除范围，跳过该范围内的排班（这些会被新排班覆盖）
            if (excludeStartDate && excludeEndDate) {
                if (a.date >= excludeStartDate && a.date <= excludeEndDate) {
                    return; // 跳过即将被覆盖的排班
                }
            }
            
            const month = a.date.substring(0, 7); // YYYY-MM
            if (!counts[a.employeeName]) {
                counts[a.employeeName] = {};
            }
            counts[a.employeeName][month] = (counts[a.employeeName][month] || 0) + 1;
        });
        
        return counts;
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
