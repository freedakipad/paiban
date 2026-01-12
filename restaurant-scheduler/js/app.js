/**
 * 餐饮门店智能排班系统 - 主应用
 */

// 评分项中文映射表
const SCORE_LABEL_MAP = {
    'skill_match': '技能匹配',
    'distance': '路程距离',
    'preference': '员工偏好',
    'workload_balance': '工时均衡',
    'continuity': '连续性',
    'reasons': '其他因素'
};

// 获取评分项中文名称
function getScoreLabel(key) {
    return SCORE_LABEL_MAP[key] || key;
}

/**
 * 格式化员工显示名称：姓名（手机号）
 * @param {Object} emp - 员工对象，包含name和phone属性
 * @returns {string} 格式化后的员工显示名称
 */
function formatEmployeeName(emp) {
    if (!emp) return '未知员工';
    return emp.name || '未知';
}

/**
 * 根据员工姓名获取格式化的显示名称
 * @param {string} employeeName - 员工姓名
 * @returns {string} 格式化后的员工显示名称
 */
function formatEmployeeNameByName(employeeName) {
    const emp = appState.employees.find(e => e.name === employeeName);
    return formatEmployeeName(emp);
}

// 切换班次折叠
function toggleShiftCollapse(shiftId) {
    if (!appState.shiftCollapsed) {
        appState.shiftCollapsed = {};
    }
    appState.shiftCollapsed[shiftId] = !appState.shiftCollapsed[shiftId];
    renderScheduleGrid();
}

// 测算排班缺口
function checkScheduleGaps() {
    const weekDates = appState.getWeekDates();
    const gaps = [];
    const isAllMode = appState.isAllStoresMode();
    const stores = isAllMode ? appState.getAllStores() : [appState.getCurrentStore()].filter(Boolean);
    
    // 遍历每天、每个班次检查缺口
    weekDates.forEach(date => {
        const dateStr = formatDate(date);
        appState.shifts.forEach(shift => {
            // 获取该日期该班次的所有排班
            const shiftAssignments = appState.assignments.filter(
                a => a.date === dateStr && a.shiftId === shift.id
            );
            
            // 统计每个岗位已分配人数
            const positionAssigned = {};
            shiftAssignments.forEach(a => {
                const pos = a.position || '未知';
                positionAssigned[pos] = (positionAssigned[pos] || 0) + 1;
            });
            
            // 汇总所有门店该日期该班次的需求
            const totalReqByPosition = {};
            stores.forEach(store => {
                const dayReqs = appState.getRequirementsForDate(date, store.id);
                const shiftReqs = dayReqs[shift.id] || {};
                Object.entries(shiftReqs).forEach(([pos, count]) => {
                    totalReqByPosition[pos] = (totalReqByPosition[pos] || 0) + count;
                });
            });
            
            // 检查每个岗位的缺口
            Object.entries(totalReqByPosition).forEach(([pos, required]) => {
                const assigned = positionAssigned[pos] || 0;
                if (assigned < required) {
                    gaps.push({
                        date: dateStr,
                        shiftId: shift.id,
                        shiftName: shift.name,
                        position: pos,
                        required,
                        assigned,
                        gap: required - assigned
                    });
                }
            });
        });
    });
    
    // 高亮显示有缺口的单元格
    document.querySelectorAll('.grid-cell.has-gap').forEach(cell => {
        cell.classList.remove('has-gap');
    });
    
    if (gaps.length === 0) {
        showToast('✅ 测算完成：所有班次已满足需求！', 'success');
        return;
    }
    
    // 高亮有缺口的单元格
    gaps.forEach(g => {
        const cell = document.querySelector(`.grid-cell[data-date="${g.date}"][data-shift="${g.shiftId}"]`);
        if (cell) {
            cell.classList.add('has-gap');
        }
    });
    
    // 显示缺口汇总
    const totalGap = gaps.reduce((sum, g) => sum + g.gap, 0);
    const gapSummary = gaps.slice(0, 3).map(g => 
        `${g.date.slice(5)} ${g.shiftName} ${g.position}缺${g.gap}人`
    ).join('；');
    const moreText = gaps.length > 3 ? `...等${gaps.length}处` : '';
    
    showToast(`⚠️ 发现${totalGap}个缺口：${gapSummary}${moreText}`, 'warning');
}

// 切换评分明细展开/折叠
function toggleScoreDetail(header) {
    const breakdown = header.parentElement;
    const isCollapsed = breakdown.classList.contains('collapsed');
    
    if (isCollapsed) {
        breakdown.classList.remove('collapsed');
        header.querySelector('.toggle-icon').textContent = '▼';
    } else {
        breakdown.classList.add('collapsed');
        header.querySelector('.toggle-icon').textContent = '▶';
    }
}

// 防止重复初始化标志
let appInitialized = false;

// DOM Ready - 仅在脚本加载完成时作为备用
document.addEventListener('DOMContentLoaded', () => {
    // 延迟检查，因为动态加载可能还在进行中
    setTimeout(() => {
        if (!appInitialized) {
            initApp();
        }
    }, 100);
});

/**
 * 初始化应用
 */
function initApp() {
    if (appInitialized) return;
    appInitialized = true;
    // 初始化导航
    initNavigation();
    
    // 初始化门店选择器
    initStoreSelector();
    
    // 初始化排班表
    initScheduleView();
    
    // 初始化员工管理
    initEmployeeView();
    
    // 初始化班次设置
    initShiftView();
    
    // 初始化设置
    initSettingsView();
    
    // 初始化弹窗
    initModals();
    
    // 初始化智能排班按钮
    initGenerateButton();
    
    // 初始化排班操作按钮
    initScheduleActions();
    
    // 加载当前周排班数据
    appState.loadWeekSchedule();
    
    // 渲染初始视图
    renderScheduleGrid();
    renderEmployeeGrid();
    renderShiftList();
    
    // 更新周状态显示
    updateWeekStatus();
    
    // 初始化历史记录数量
    updateHistoryCount();
    
    // 更新门店显示
    updateStoreDisplay();
    
    console.log('🍜 餐饮门店智能排班系统已启动');
}

/* ========================================
   门店选择器
   ======================================== */

function initStoreSelector() {
    const selectorBtn = document.getElementById('storeSelectorBtn');
    const selector = document.getElementById('storeSelector');
    const dropdown = document.getElementById('storeDropdown');
    
    if (!selectorBtn || !selector) return;
    
    // 点击按钮切换下拉菜单
    selectorBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        selector.classList.toggle('open');
        if (selector.classList.contains('open')) {
            renderStoreList();
        }
    });
    
    // 点击页面其他地方关闭下拉菜单
    document.addEventListener('click', () => {
        selector.classList.remove('open');
    });
    
    // 阻止下拉菜单点击事件冒泡
    if (dropdown) {
        dropdown.addEventListener('click', (e) => {
            e.stopPropagation();
        });
    }
}

// 渲染门店列表
function renderStoreList() {
    const listEl = document.getElementById('storeList');
    const statsEl = document.getElementById('chainStats');
    if (!listEl) return;
    
    const stores = appState.getAllStores();
    const currentStoreId = appState.currentStoreId;
    const totalEmployees = appState.employees.filter(e => e.status === 'active').length;
    
    let html = '';
    
    // 添加"全部门店"选项（放在最前面）
    const isAllActive = currentStoreId === 'all';
    html += `
        <div class="store-item ${isAllActive ? 'active' : ''}" onclick="switchStore('all')">
            <span class="store-item-icon">🏢</span>
            <div class="store-item-info">
                <div class="store-item-name">
                    全部门店
                    <span class="store-type-badge" style="color: #ef4444">联合排班</span>
                </div>
                <div class="store-item-detail">为所有门店统一排班，避免人员冲突</div>
            </div>
            <div class="store-item-stats">${totalEmployees}人</div>
        </div>
    `;
    
    // 分隔线
    html += `<div class="store-list-divider"></div>`;
    
    // 各门店选项
    stores.forEach(store => {
        const isActive = store.id === currentStoreId;
        const storeType = STORE_TYPES[store.type] || STORE_TYPES.standard;
        const stats = appState.getStoreStats(store.id);
        
        html += `
            <div class="store-item ${isActive ? 'active' : ''}" onclick="switchStore('${store.id}')">
                <span class="store-item-icon">${storeType.icon}</span>
                <div class="store-item-info">
                    <div class="store-item-name">
                        ${store.name}
                        <span class="store-type-badge" style="color: ${storeType.color}">${storeType.label}</span>
                    </div>
                    <div class="store-item-detail">${store.address || ''}</div>
                </div>
                <div class="store-item-stats">${stats.activeEmployees}人</div>
            </div>
        `;
    });
    
    listEl.innerHTML = html;
    
    // 更新统计信息
    if (statsEl) {
        statsEl.innerHTML = `
            <span>${stores.length} 家门店</span>
            <span>•</span>
            <span>${totalEmployees} 名员工</span>
        `;
    }
}

// 切换门店
function switchStore(storeId) {
    if (appState.switchStore(storeId)) {
        // 更新UI显示
        updateStoreDisplay();
        
        // 关闭下拉菜单
        document.getElementById('storeSelector').classList.remove('open');
        
        // 重新渲染所有视图
        renderScheduleGrid();
        renderEmployeeGrid();
        renderShiftList();  // 更新班次页签的需求配置
        updateWeekStatus();
        
        // 显示提示
        const storeName = storeId === 'all' ? '全部门店' : appState.getCurrentStore().name;
        showToast(`已切换到 ${storeName}`, 'success');
    }
}

// 更新门店显示
function updateStoreDisplay() {
    const nameEl = document.getElementById('currentStoreName');
    const codeEl = document.getElementById('currentStoreCode');
    
    // 支持"全部门店"模式
    if (appState.isAllStoresMode()) {
        if (nameEl) nameEl.textContent = '全部门店';
        if (codeEl) codeEl.textContent = '(联合)';
        return;
    }
    
    const store = appState.getCurrentStore();
    if (!store) return;
    
    if (nameEl) nameEl.textContent = store.name;
    if (codeEl) codeEl.textContent = `(${store.code})`;
}

// 显示门店管理弹窗
function showStoreManagement() {
    const stores = appState.stores;
    
    const content = `
        <div class="store-management-modal">
            <div class="modal-header-title">
                <h3>🏢 门店管理</h3>
                <button class="modal-close" onclick="closeStoreManagement()">×</button>
            </div>
            <div class="modal-body">
                <div class="store-management-list">
                    ${stores.map(store => {
                        const storeType = STORE_TYPES[store.type] || STORE_TYPES.standard;
                        const stats = appState.getStoreStats(store.id);
                        const isActive = store.status === 'active';
                        return `
                            <div class="store-management-item ${isActive ? '' : 'inactive'}">
                                <div class="store-mgmt-icon">${storeType.icon}</div>
                                <div class="store-mgmt-info">
                                    <div class="store-mgmt-name">${store.name} <span class="store-code">${store.code}</span></div>
                                    <div class="store-mgmt-detail">${store.address || '无地址'}</div>
                                    <div class="store-mgmt-stats">
                                        👥 ${stats.activeEmployees}人 | 📍 ${store.type ? storeType.label : '标准店'} | 
                                        ${isActive ? '✅ 营业中' : '❌ 已停业'}
                                    </div>
                                </div>
                                <div class="store-mgmt-actions">
                                    <button class="btn-sm" onclick="editStore('${store.id}')">✏️ 编辑</button>
                                    ${store.id !== appState.currentStoreId ? 
                                        `<button class="btn-sm danger" onclick="toggleStoreStatus('${store.id}')">${isActive ? '停业' : '恢复'}</button>` 
                                        : ''}
                                </div>
                            </div>
                        `;
                    }).join('')}
                </div>
                <div class="store-add-section">
                    <button class="btn-add-store" onclick="showAddStoreForm()">➕ 新增门店</button>
                </div>
            </div>
        </div>
    `;
    
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay store-management-overlay';
    overlay.innerHTML = content;
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) closeStoreManagement();
    });
    
    document.body.appendChild(overlay);
}

// 关闭门店管理弹窗
function closeStoreManagement() {
    const overlay = document.querySelector('.store-management-overlay');
    if (overlay) overlay.remove();
}

// 编辑门店
function editStore(storeId) {
    const store = appState.stores.find(s => s.id === storeId);
    if (!store) return;
    
    const content = `
        <div class="store-edit-modal">
            <div class="modal-header-title">
                <h3>✏️ 编辑门店</h3>
                <button class="modal-close" onclick="closeStoreEdit()">×</button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>门店名称</label>
                    <input type="text" id="editStoreName" value="${store.name}" placeholder="输入门店名称">
                </div>
                <div class="form-group">
                    <label>门店代码</label>
                    <input type="text" id="editStoreCode" value="${store.code}" placeholder="如: HQ, WJ">
                </div>
                <div class="form-group">
                    <label>门店地址</label>
                    <input type="text" id="editStoreAddress" value="${store.address || ''}" placeholder="输入门店地址">
                </div>
                <div class="form-group">
                    <label>门店类型</label>
                    <select id="editStoreType">
                        <option value="flagship" ${store.type === 'flagship' ? 'selected' : ''}>🏪 旗舰店</option>
                        <option value="standard" ${store.type === 'standard' ? 'selected' : ''}>🏬 标准店</option>
                        <option value="express" ${store.type === 'express' ? 'selected' : ''}>🍱 快餐店</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>联系电话</label>
                    <input type="text" id="editStorePhone" value="${store.phone || ''}" placeholder="门店电话">
                </div>
                <div class="form-row-group">
                    <div class="form-group half">
                        <label>营业开始</label>
                        <input type="time" id="editStoreOpenTime" value="${store.openTime}">
                    </div>
                    <div class="form-group half">
                        <label>营业结束</label>
                        <input type="time" id="editStoreCloseTime" value="${store.closeTime}">
                    </div>
                </div>
                <div class="form-group">
                    <label>座位数</label>
                    <input type="number" id="editStoreCapacity" value="${store.capacity || 50}" min="10" max="500">
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn-secondary" onclick="closeStoreEdit()">取消</button>
                <button class="btn-primary" onclick="saveStoreEdit('${storeId}')">保存</button>
            </div>
        </div>
    `;
    
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay store-edit-overlay';
    overlay.innerHTML = content;
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) closeStoreEdit();
    });
    
    document.body.appendChild(overlay);
}

// 关闭门店编辑弹窗
function closeStoreEdit() {
    const overlay = document.querySelector('.store-edit-overlay');
    if (overlay) overlay.remove();
}

// 保存门店编辑
function saveStoreEdit(storeId) {
    const updates = {
        name: document.getElementById('editStoreName').value.trim(),
        code: document.getElementById('editStoreCode').value.trim().toUpperCase(),
        address: document.getElementById('editStoreAddress').value.trim(),
        type: document.getElementById('editStoreType').value,
        phone: document.getElementById('editStorePhone').value.trim(),
        openTime: document.getElementById('editStoreOpenTime').value,
        closeTime: document.getElementById('editStoreCloseTime').value,
        capacity: parseInt(document.getElementById('editStoreCapacity').value) || 50
    };
    
    if (!updates.name) {
        showToast('门店名称不能为空', 'error');
        return;
    }
    
    if (appState.updateStore(storeId, updates)) {
        showToast('门店信息已更新', 'success');
        closeStoreEdit();
        closeStoreManagement();
        updateStoreDisplay();
        renderStoreList();
    }
}

// 切换门店状态（营业/停业）
function toggleStoreStatus(storeId) {
    const store = appState.stores.find(s => s.id === storeId);
    if (!store) return;
    
    const newStatus = store.status === 'active' ? 'inactive' : 'active';
    const action = newStatus === 'active' ? '恢复营业' : '停业';
    
    if (confirm(`确定要将 ${store.name} ${action} 吗？`)) {
        appState.updateStore(storeId, { status: newStatus });
        showToast(`${store.name} 已${action}`, 'info');
        closeStoreManagement();
        showStoreManagement(); // 刷新列表
    }
}

// 显示新增门店表单
function showAddStoreForm() {
    const content = `
        <div class="store-edit-modal">
            <div class="modal-header-title">
                <h3>➕ 新增门店</h3>
                <button class="modal-close" onclick="closeStoreEdit()">×</button>
            </div>
            <div class="modal-body">
                <div class="form-group">
                    <label>门店名称 *</label>
                    <input type="text" id="newStoreName" placeholder="如: 朝阳分店">
                </div>
                <div class="form-group">
                    <label>门店代码 *</label>
                    <input type="text" id="newStoreCode" placeholder="如: CY" maxlength="4">
                </div>
                <div class="form-group">
                    <label>门店地址</label>
                    <input type="text" id="newStoreAddress" placeholder="输入门店地址">
                </div>
                <div class="form-group">
                    <label>门店类型</label>
                    <select id="newStoreType">
                        <option value="standard">🏬 标准店</option>
                        <option value="flagship">🏪 旗舰店</option>
                        <option value="express">🍱 快餐店</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>联系电话</label>
                    <input type="text" id="newStorePhone" placeholder="门店电话">
                </div>
                <div class="form-row-group">
                    <div class="form-group half">
                        <label>营业开始</label>
                        <input type="time" id="newStoreOpenTime" value="09:00">
                    </div>
                    <div class="form-group half">
                        <label>营业结束</label>
                        <input type="time" id="newStoreCloseTime" value="22:00">
                    </div>
                </div>
                <div class="form-group">
                    <label>座位数</label>
                    <input type="number" id="newStoreCapacity" value="50" min="10" max="500">
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn-secondary" onclick="closeStoreEdit()">取消</button>
                <button class="btn-primary" onclick="saveNewStore()">创建门店</button>
            </div>
        </div>
    `;
    
    const overlay = document.createElement('div');
    overlay.className = 'modal-overlay store-edit-overlay';
    overlay.innerHTML = content;
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) closeStoreEdit();
    });
    
    document.body.appendChild(overlay);
}

// 保存新门店
function saveNewStore() {
    const name = document.getElementById('newStoreName').value.trim();
    const code = document.getElementById('newStoreCode').value.trim().toUpperCase();
    
    if (!name || !code) {
        showToast('门店名称和代码不能为空', 'error');
        return;
    }
    
    // 检查代码是否重复
    if (appState.stores.find(s => s.code === code)) {
        showToast('门店代码已存在', 'error');
        return;
    }
    
    const newStore = {
        name,
        code,
        address: document.getElementById('newStoreAddress').value.trim(),
        type: document.getElementById('newStoreType').value,
        phone: document.getElementById('newStorePhone').value.trim(),
        openTime: document.getElementById('newStoreOpenTime').value,
        closeTime: document.getElementById('newStoreCloseTime').value,
        capacity: parseInt(document.getElementById('newStoreCapacity').value) || 50,
        status: 'active',
        manager: '',
        location: { lat: 39.9, lng: 116.4 } // 默认北京坐标
    };
    
    appState.addStore(newStore);
    showToast(`门店 ${name} 创建成功`, 'success');
    closeStoreEdit();
    closeStoreManagement();
    showStoreManagement(); // 刷新列表
}

function updateHistoryCount() {
    const countEl = document.getElementById('historyCount');
    if (countEl) {
        countEl.textContent = appState.shiftHistory.length;
    }
}

/* ========================================
   排班状态管理
   ======================================== */

function initScheduleActions() {
    document.getElementById('btnSave').addEventListener('click', saveSchedule);
    document.getElementById('btnPublish').addEventListener('click', publishSchedule);
    document.getElementById('btnArchive').addEventListener('click', archiveSchedule);
    document.getElementById('btnUnlock').addEventListener('click', unlockSchedule);
}

// 检查是否可以归档当前周（必须在下周一或之后）
function canArchiveWeek() {
    const today = new Date();
    const weekStart = new Date(appState.currentWeekStart);
    // 下周一 = 当前周开始 + 7天
    const nextMonday = new Date(weekStart);
    nextMonday.setDate(nextMonday.getDate() + 7);
    nextMonday.setHours(0, 0, 0, 0);
    
    return today >= nextMonday;
}

// 更新门店筛选下拉框
function updateScheduleStoreFilter() {
    const filterSelect = document.getElementById('scheduleStoreFilter');
    const isAllMode = appState.isAllStoresMode();
    
    if (isAllMode) {
        // 显示门店筛选
        filterSelect.style.display = 'inline-block';
        
        // 更新选项列表
        let html = '<option value="all">📍 全部门店</option>';
        appState.stores.forEach(store => {
            const selected = appState.scheduleViewStoreFilter === store.id ? 'selected' : '';
            html += `<option value="${store.id}" ${selected}>${store.icon || '🏪'} ${store.name}</option>`;
        });
        filterSelect.innerHTML = html;
        
        // 初始化筛选值
        if (!appState.scheduleViewStoreFilter) {
            appState.scheduleViewStoreFilter = 'all';
        }
    } else {
        // 隐藏门店筛选
        filterSelect.style.display = 'none';
        appState.scheduleViewStoreFilter = null;
    }
}

// 更新周状态显示
function updateWeekStatus() {
    const status = appState.getWeekStatus();
    const config = STATUS_CONFIG[status] || STATUS_CONFIG.draft;
    const statusEl = document.getElementById('weekStatus');
    
    if (statusEl) {
        statusEl.className = `week-status ${status}`;
        statusEl.innerHTML = `
            <span class="status-icon">${config.icon}</span>
            <span class="status-text">${config.label}</span>
        `;
    }
    
    // 更新按钮状态
    updateActionButtons(status);
}

// 更新操作按钮状态
function updateActionButtons(status) {
    const saveBtn = document.getElementById('btnSave');
    const publishBtn = document.getElementById('btnPublish');
    const archiveBtn = document.getElementById('btnArchive');
    const unlockBtn = document.getElementById('btnUnlock');
    const generateBtn = document.getElementById('btnGenerate');
    
    // 默认隐藏解锁按钮
    unlockBtn.style.display = 'none';
    
    // 根据状态启用/禁用按钮
    switch (status) {
        case SCHEDULE_STATUS.DRAFT:
            saveBtn.disabled = false;
            publishBtn.disabled = appState.assignments.length === 0;
            archiveBtn.disabled = appState.assignments.length === 0; // 有排班才能归档
            generateBtn.disabled = false;
            break;
        case SCHEDULE_STATUS.PUBLISHED:
            // 发布后只读，需解锁才能修改
            saveBtn.disabled = true;
            publishBtn.disabled = true;
            archiveBtn.disabled = !canArchiveWeek(); // 必须在下周一或之后
            generateBtn.disabled = true;
            // 显示解锁按钮
            unlockBtn.style.display = 'inline-flex';
            break;
        case SCHEDULE_STATUS.ACTIVE:
            saveBtn.disabled = false;
            publishBtn.disabled = true;
            archiveBtn.disabled = appState.assignments.length === 0; // 有排班才能归档
            generateBtn.disabled = false; // 允许微调（重新生成）
            break;
        case SCHEDULE_STATUS.ARCHIVED:
            // 归档是终态，不可解锁
            saveBtn.disabled = true;
            publishBtn.disabled = true;
            archiveBtn.disabled = true;
            generateBtn.disabled = true;
            // 归档后不显示解锁按钮
            break;
        default:
            saveBtn.disabled = false;
            publishBtn.disabled = true;
            archiveBtn.disabled = true;
    }
}

// 保存排班
function saveSchedule() {
    if (appState.assignments.length === 0) {
        showToast('没有排班数据需要保存', 'warning');
        return;
    }
    
    appState.saveScheduleToWeek(appState.assignments);
    showToast('排班已保存', 'success');
    updateWeekStatus();
}

// 发布排班
function publishSchedule() {
    if (appState.assignments.length === 0) {
        showToast('请先生成排班', 'warning');
        return;
    }
    
    const confirmed = confirm('确定要发布本周排班吗？发布后员工将收到通知。');
    if (!confirmed) return;
    
    // 先保存
    appState.saveScheduleToWeek(appState.assignments);
    
    // 发布
    if (appState.publishSchedule()) {
        showToast('排班已发布！员工将收到通知。', 'success');
        updateWeekStatus();
        renderScheduleGrid(); // 重新渲染使卡片变为只读
        
        // 记录历史
        appState.addHistoryRecord({
            type: 'publish',
            action: '发布排班',
            description: `发布了 ${appState.getWeekKey(appState.currentWeekStart)} 的排班表`
        });
        updateHistoryCount();
    } else {
        showToast('发布失败，请检查排班状态', 'error');
    }
}

// 归档当前周排班（必须在下周一或之后）
function archiveSchedule() {
    const weekKey = appState.getWeekKey(appState.currentWeekStart);
    
    if (!canArchiveWeek()) {
        showToast('归档必须在下周一或之后进行', 'warning');
        return;
    }
    
    if (appState.assignments.length === 0) {
        showToast('没有排班数据可归档', 'warning');
        return;
    }
    
    const confirmed = confirm(`确定要归档 ${weekKey} 的排班记录吗？归档后将永久锁定，无法修改。`);
    if (!confirmed) return;
    
    // 先保存
    appState.saveScheduleToWeek(appState.assignments);
    
    if (appState.archiveSchedule(weekKey)) {
        showToast(`${weekKey} 排班已归档（永久锁定）`, 'success');
        
        // 记录历史
        appState.addHistoryRecord({
            type: 'archive',
            action: '归档排班',
            description: `归档了 ${weekKey} 的排班表（永久锁定）`
        });
        updateHistoryCount();
        updateWeekStatus();
        renderScheduleGrid();
    } else {
        showToast('归档失败', 'error');
    }
}

// 解锁已发布的排班（归档状态不能解锁）
function unlockSchedule() {
    const weekKey = appState.getWeekKey(appState.currentWeekStart);
    const storeWeekKey = appState.getStoreWeekKey(appState.currentWeekStart);
    const status = appState.getWeekStatus();
    
    if (status === SCHEDULE_STATUS.ARCHIVED) {
        showToast('归档后的排班无法解锁', 'error');
        return;
    }
    
    if (status !== SCHEDULE_STATUS.PUBLISHED) {
        showToast('只有已发布的排班才能解锁', 'warning');
        return;
    }
    
    const confirmed = confirm(`确定要解锁 ${weekKey} 的排班表吗？解锁后可以重新编辑。`);
    if (!confirmed) return;
    
    if (appState.unlockSchedule(storeWeekKey)) {
        showToast(`${weekKey} 排班已解锁，现在可以编辑`, 'success');
        
        // 记录历史
        appState.addHistoryRecord({
            type: 'unlock',
            action: '解锁排班',
            description: `解锁了 ${weekKey} 的排班表`
        });
        updateHistoryCount();
        updateWeekStatus();
        renderScheduleGrid();
    } else {
        showToast('解锁失败', 'error');
    }
}

/* ========================================
   导航功能
   ======================================== */

function initNavigation() {
    const navBtns = document.querySelectorAll('.nav-btn');
    
    navBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const tabId = btn.dataset.tab;
            
            // 更新按钮状态
            navBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            
            // 切换内容区
            document.querySelectorAll('.tab-content').forEach(content => {
                content.classList.remove('active');
            });
            document.getElementById(`tab-${tabId}`).classList.add('active');
            
            appState.currentView = tabId;
            
            // 切换到员工页签时自动刷新所有员工的排班信息
            if (tabId === 'employees') {
                refreshAllEmployeeSchedules();
            }
        });
    });
}

/**
 * 刷新所有员工的排班信息
 */
function refreshAllEmployeeSchedules() {
    // 重新从 localStorage 加载周排班数据
    const savedWeeks = localStorage.getItem('restaurant-scheduler-scheduleWeeks');
    if (savedWeeks) {
        try {
            appState.scheduleWeeks = JSON.parse(savedWeeks);
        } catch (e) {
            console.warn('Failed to reload scheduleWeeks:', e);
        }
    }
    
    // 重新渲染员工表格
    renderEmployeeGrid();
}

/* ========================================
   排班表视图
   ======================================== */

function initScheduleView() {
    // 排班周期选择
    document.getElementById('periodSelect').addEventListener('change', (e) => {
        const period = e.target.value === 'month' ? 'month' : parseInt(e.target.value);
        appState.setSchedulePeriod(period);
        
        // 当选择月度排班时，自动切换到月度工时计算模式
        if (period === 'month') {
            appState.updateSettings({ hoursMode: 'period' });
            console.log('已自动切换到月度工时计算模式');
        } else if (period === 7 || period === 14) {
            // 选择周度排班时，使用周度工时计算模式
            appState.updateSettings({ hoursMode: 'weekly' });
            console.log('已自动切换到周度工时计算模式');
        }
        
        renderScheduleGrid();
        renderEmployeeGrid(); // 更新员工统计标签
        updateWeekStatus();
        updateHistoryCount();
    });
    
    // 周导航
    document.getElementById('prevWeek').addEventListener('click', () => {
        appState.prevWeek();
        renderScheduleGrid();
        updateWeekStatus();
        updateHistoryCount(); // 更新当前周的历史数量
    });
    
    document.getElementById('nextWeek').addEventListener('click', () => {
        appState.nextWeek();
        renderScheduleGrid();
        updateWeekStatus();
        updateHistoryCount(); // 更新当前周的历史数量
    });
    
    document.getElementById('todayBtn').addEventListener('click', () => {
        appState.goToToday();
        renderScheduleGrid();
        updateWeekStatus();
        updateHistoryCount(); // 更新当前周的历史数量
    });
    
    // 视图切换
    document.querySelectorAll('.view-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.view-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            // TODO: 实现日视图
        });
    });
    
    // 关闭未满足面板
    document.getElementById('closeUnfilled').addEventListener('click', () => {
        document.getElementById('unfilledPanel').style.display = 'none';
    });
    
    // 门店筛选
    document.getElementById('scheduleStoreFilter').addEventListener('change', (e) => {
        appState.scheduleViewStoreFilter = e.target.value;
        renderScheduleGrid();
    });
}

function renderScheduleGrid() {
    const grid = document.getElementById('scheduleGrid');
    const weekDates = appState.getWeekDates();
    const status = appState.getWeekStatus();
    // 发布和归档状态都是只读
    const isReadOnly = status === SCHEDULE_STATUS.ARCHIVED || status === SCHEDULE_STATUS.PUBLISHED;
    
    // 更新周期标签
    document.getElementById('weekLabel').textContent = appState.getPeriodLabel();
    
    // 更新门店筛选下拉框
    updateScheduleStoreFilter();
    
    // 动态设置网格列数 - 所有周期使用标准视图，横向滚动
    grid.style.gridTemplateColumns = `120px repeat(${weekDates.length}, minmax(120px, 1fr))`;
    
    // 构建表头
    let html = '<div class="grid-cell grid-header"></div>';
    weekDates.forEach(date => {
        const todayClass = isToday(date) ? 'today' : '';
        html += `
            <div class="grid-cell grid-header ${todayClass}">
                <div class="day-name">${getDayName(date)}</div>
                <div class="day-date">${date.getDate()}</div>
            </div>
        `;
    });
    
    // 初始化折叠状态（如果不存在）
    if (!appState.shiftCollapsed) {
        appState.shiftCollapsed = {};
    }
    
    // 构建每个班次的行
    appState.shifts.forEach(shift => {
        const isCollapsed = appState.shiftCollapsed[shift.id] || false;
        const collapseIcon = isCollapsed ? '▶' : '▼';
        
        // 行标签（可点击折叠）
        html += `
            <div class="grid-cell row-label ${isCollapsed ? 'collapsed' : ''}" onclick="toggleShiftCollapse('${shift.id}')">
                <span class="collapse-icon">${collapseIcon}</span>
                <div class="shift-name">${shift.name}</div>
                <div class="shift-time">${shift.startTime}-${shift.endTime}</div>
            </div>
        `;
        
        // 如果折叠，显示空单元格
        if (isCollapsed) {
            weekDates.forEach(() => {
                html += `<div class="grid-cell collapsed-cell"></div>`;
            });
            return; // 跳过该班次的详细内容
        }
        
        // 每天的单元格
        weekDates.forEach(date => {
            const dateStr = formatDate(date);
            let cellAssignments = appState.assignments.filter(
                a => a.date === dateStr && a.shiftId === shift.id
            );
            
            // 在全部门店模式下，按门店顺序排序
            const isAllMode = appState.isAllStoresMode();
            if (isAllMode) {
                // 获取门店排序顺序
                const storeOrder = appState.stores.reduce((map, store, idx) => {
                    map[store.id] = idx;
                    return map;
                }, {});
                
                // 按门店顺序排序
                cellAssignments = cellAssignments.sort((a, b) => {
                    const orderA = storeOrder[a.storeId] ?? 999;
                    const orderB = storeOrder[b.storeId] ?? 999;
                    if (orderA !== orderB) return orderA - orderB;
                    // 同一门店内按岗位排序（厨师优先）
                    if (a.position !== b.position) {
                        return a.position === '厨师' ? -1 : 1;
                    }
                    return 0;
                });
                
                // 如果有门店筛选，应用筛选
                if (appState.scheduleViewStoreFilter && appState.scheduleViewStoreFilter !== 'all') {
                    cellAssignments = cellAssignments.filter(a => a.storeId === appState.scheduleViewStoreFilter);
                }
            }
            
            // 获取该日期该班次的需求 - 考虑门店筛选
            const filterStoreId = appState.scheduleViewStoreFilter;
            let totalRequired = 0;
            
            if (isAllMode && filterStoreId && filterStoreId !== 'all') {
                // 有门店筛选时，只计算该门店的需求
                const dayReqs = appState.getRequirementsForDate(date, filterStoreId);
                const shiftReqs = dayReqs[shift.id] || {};
                totalRequired = Object.values(shiftReqs).reduce((sum, n) => sum + n, 0);
            } else if (isAllMode) {
                // 全部门店模式，汇总所有门店的需求
                appState.getAllStores().forEach(store => {
                    const dayReqs = appState.getRequirementsForDate(date, store.id);
                    const shiftReqs = dayReqs[shift.id] || {};
                    totalRequired += Object.values(shiftReqs).reduce((sum, n) => sum + n, 0);
                });
            } else {
                // 单门店模式
                const dayReqs = appState.getRequirementsForDate(date);
                const shiftReqs = dayReqs[shift.id] || {};
                totalRequired = Object.values(shiftReqs).reduce((sum, n) => sum + n, 0);
            }
            
            const assigned = cellAssignments.length;
            
            html += `<div class="grid-cell" data-date="${dateStr}" data-shift="${shift.id}">`;
            
            // 显示排班卡片
            cellAssignments.forEach(a => {
                const shiftClass = getShiftClass(a.shiftId);
                // 发布后也能点击查看详情，但不能操作
                const clickHandler = `onclick="showAssignmentDetail('${a.id}', ${isReadOnly}); event.stopPropagation();"`;
                const readOnlyClass = isReadOnly ? 'archived' : '';
                
                // 获取门店信息和岗位样式
                const storeName = a.storeName || '';
                const storeCode = a.storeCode || '';
                const positionClass = getPositionClass(a.position);
                const storeClass = getStoreClass(storeCode);
                
                // 获取员工手机号（完整）
                const emp = appState.employees.find(e => e.name === a.employeeName);
                const phone = emp?.phone || '';
                
                // 紧凑卡片：姓名 + 岗位色签 + 门店色签 + 完整手机号
                html += `
                    <div class="assignment-card compact ${shiftClass} ${readOnlyClass}" data-id="${a.id}" ${clickHandler}>
                        <span class="card-name">${formatEmployeeNameByName(a.employeeName)}</span>
                        <span class="card-tag ${positionClass}">${a.position || ''}</span>
                        ${storeCode ? `<span class="card-tag ${storeClass}">${storeCode}</span>` : ''}
                        ${phone ? `<span class="card-phone">${phone}</span>` : ''}
                    </div>
                `;
            });
            
            // 只读状态（发布/归档）不显示添加按钮
            if (!isReadOnly) {
                // 计算每个岗位的缺口
                const positionGaps = calculatePositionGaps(date, shift.id, cellAssignments, filterStoreId, isAllMode);
                const hasGaps = positionGaps.length > 0;
                
                if (hasGaps) {
                    // 按岗位显示缺口
                    const gapText = positionGaps.map(g => `${g.position}${g.gap}人`).join('，');
                    html += `<div class="requirement-indicator unfilled" onclick="openManualAssign('${dateStr}', '${shift.id}'); event.stopPropagation();">+ 还缺${gapText}</div>`;
                } else if (totalRequired > 0) {
                    html += `<div class="add-assignment-btn" onclick="openManualAssign('${dateStr}', '${shift.id}'); event.stopPropagation();">+</div>`;
                } else {
                    html += `<div class="add-assignment-btn" onclick="openManualAssign('${dateStr}', '${shift.id}'); event.stopPropagation();">+</div>`;
                }
            }
            
            html += '</div>';
        });
    });
    
    grid.innerHTML = html;
    
    // 更新统计
    updateScheduleStats();
}

function updateScheduleStats() {
    // 获取当前门店筛选器的值
    const filterStoreId = document.getElementById('scheduleStoreFilter')?.value || 'all';
    
    // 根据筛选器过滤排班数据
    let filteredAssignments = appState.assignments;
    if (filterStoreId !== 'all' && appState.isAllStoresMode()) {
        filteredAssignments = appState.assignments.filter(a => a.storeId === filterStoreId);
    }
    
    // 更新总班次
    document.getElementById('statTotal').textContent = filteredAssignments.length;
    
    // 更新平均评分
    if (filteredAssignments.length > 0) {
        const avgScore = filteredAssignments.reduce((sum, a) => sum + (a.score || 0), 0) / filteredAssignments.length;
        document.getElementById('statScore').textContent = Math.round(avgScore);
    } else {
        document.getElementById('statScore').textContent = '--';
    }
    
    // 计算满足率 - 根据筛选器计算
    const weekDates = appState.getWeekDates();
    const rate = calculateSatisfactionRateForStore(weekDates, filteredAssignments, filterStoreId);
    document.getElementById('statRate').textContent = `${rate}%`;
}

function showAssignmentDetail(assignmentId) {
    currentAssignmentId = assignmentId;
    const assignment = appState.assignments.find(a => a.id === assignmentId);
    if (!assignment) return;
    
    const detail = document.getElementById('assignmentDetail');
    const shift = appState.getShift(assignment.shiftId);
    
    let scoreDetailHtml = '';
    if (assignment.scoreDetail) {
        let detailItems = '';
        Object.entries(assignment.scoreDetail).forEach(([key, val]) => {
            const valNum = typeof val === 'number' ? val : 0;
            detailItems += `<li><span>${getScoreLabel(key)}</span><span>${valNum.toFixed(1)}分</span></li>`;
        });
        scoreDetailHtml = `
            <div class="score-breakdown collapsed">
                <div class="score-breakdown-header" onclick="toggleScoreDetail(this)">
                    <span>📊 评分明细</span>
                    <span class="toggle-icon">▶</span>
                </div>
                <ul class="score-breakdown-list">${detailItems}</ul>
            </div>`;
    }
    
    const manualBadge = assignment.isManual ? '<span class="manual-badge">手动</span>' : '';
    
    detail.innerHTML = `
        <div class="assignment-detail-grid">
            <div class="detail-item">
                <label>员工</label>
                <span>${formatEmployeeNameByName(assignment.employeeName)} ${manualBadge}</span>
            </div>
            <div class="detail-item">
                <label>岗位</label>
                <span>${assignment.position || '未指定'}</span>
            </div>
            <div class="detail-item">
                <label>日期</label>
                <span>${assignment.date} ${getDayName(assignment.date)}</span>
            </div>
            <div class="detail-item">
                <label>班次</label>
                <span>${assignment.shiftName || shift?.name || '未知'}</span>
            </div>
            <div class="detail-item">
                <label>时间</label>
                <span>${assignment.startTime} - ${assignment.endTime}</span>
            </div>
            <div class="detail-item">
                <label>工时</label>
                <span>${assignment.hours} 小时</span>
            </div>
            ${assignment.score ? `
            <div class="detail-item full-width">
                <label>综合评分</label>
                <span class="score-badge ${getScoreLevel(assignment.score || 0)}">${Math.round(assignment.score || 0)} 分</span>
            </div>
            ` : ''}
        </div>
        ${scoreDetailHtml}
    `;
    
    // 根据当前周状态显示/隐藏操作按钮
    const schedule = appState.getCurrentWeekSchedule();
    const status = schedule ? schedule.status : SCHEDULE_STATUS.DRAFT;
    const isReadOnly = status === SCHEDULE_STATUS.PUBLISHED || status === SCHEDULE_STATUS.ARCHIVED;
    
    const swapBtn = document.getElementById('swapAssignment');
    const removeBtn = document.getElementById('removeAssignment');
    if (swapBtn) swapBtn.style.display = isReadOnly ? 'none' : '';
    if (removeBtn) removeBtn.style.display = isReadOnly ? 'none' : '';
    
    openModal('assignmentModal');
}

function showUnfilledRequirements(forceShow = false) {
    const panel = document.getElementById('unfilledPanel');
    const list = document.getElementById('unfilledList');
    const analysisContent = document.getElementById('analysisContent');
    const solutionContent = document.getElementById('solutionContent');
    const unfilledSection = document.getElementById('unfilledSection');
    
    // 确保 unfilledRequirements 存在
    if (!appState.unfilledRequirements) {
        appState.unfilledRequirements = [];
    }
    
    // 如果没有排班数据，不显示
    if (!appState.assignments || appState.assignments.length === 0) {
        panel.style.display = 'none';
        return;
    }
    
    // 生成分析报告
    const analysis = analyzeSchedulingProblem();
    
    // 渲染分析内容 - 分组显示
    let problemsHtml = '';
    
    // 按类别分组问题
    const summaryItems = analysis.problems.filter(p => p.category === 'summary');
    const violationItems = analysis.problems.filter(p => p.category === 'violation');
    const gapItems = analysis.problems.filter(p => p.category === 'gap');
    const storeItems = analysis.problems.filter(p => p.category === 'store');
    const otherItems = analysis.problems.filter(p => !p.category);
    
    // 概览部分
    if (summaryItems.length > 0) {
        problemsHtml += `<div class="analysis-group"><div class="analysis-group-title">📈 人力概览</div>`;
        summaryItems.forEach(p => {
            problemsHtml += `<div class="analysis-item ${p.severity}"><span class="icon">${p.icon}</span><span>${p.message}</span></div>`;
        });
        problemsHtml += '</div>';
    }
    
    // 约束违规部分
    if (violationItems.length > 0) {
        problemsHtml += `<div class="analysis-group"><div class="analysis-group-title">⚠️ 约束违规 (${violationItems.length}项)</div>`;
        violationItems.slice(0, 3).forEach(p => {
            problemsHtml += `<div class="analysis-item ${p.severity}"><span class="icon">${p.icon}</span><span>${p.message}</span></div>`;
        });
        if (violationItems.length > 3) {
            problemsHtml += `<div class="analysis-item info"><span class="icon">...</span><span>还有 ${violationItems.length - 3} 条违规</span></div>`;
        }
        problemsHtml += '</div>';
    }
    
    // 岗位缺口部分
    if (gapItems.length > 0) {
        problemsHtml += `<div class="analysis-group"><div class="analysis-group-title">👥 岗位缺口</div>`;
        gapItems.forEach(p => {
            problemsHtml += `<div class="analysis-item ${p.severity}"><span class="icon">${p.icon}</span><span>${p.message}</span></div>`;
        });
        problemsHtml += '</div>';
    }
    
    // 门店详情部分
    if (storeItems.length > 0) {
        problemsHtml += `<div class="analysis-group collapsible"><div class="analysis-group-title" onclick="this.parentElement.classList.toggle('expanded')">🏪 门店详情 <span class="expand-icon">▶</span></div><div class="analysis-group-content">`;
        storeItems.forEach(p => {
            problemsHtml += `<div class="analysis-item ${p.severity}"><span class="icon">${p.icon}</span><span>${p.message}</span></div>`;
        });
        problemsHtml += '</div></div>';
    }
    
    // 其他问题
    if (otherItems.length > 0) {
        otherItems.forEach(p => {
            problemsHtml += `<div class="analysis-item ${p.severity}"><span class="icon">${p.icon}</span><span>${p.message}</span></div>`;
        });
    }
    
    analysisContent.innerHTML = problemsHtml;
    
    // 渲染解决方案 - 分优先级显示
    const primarySolutions = analysis.solutions.filter(s => s.type === 'primary');
    const otherSolutions = analysis.solutions.filter(s => s.type !== 'primary');
    
    let solutionsHtml = '';
    if (primarySolutions.length > 0) {
        solutionsHtml += `<div class="solution-group"><div class="solution-group-title">🎯 推荐操作</div>`;
        primarySolutions.forEach(s => {
            solutionsHtml += `<div class="solution-item ${s.type}"><span class="icon">${s.icon}</span><span>${s.message}</span></div>`;
        });
        solutionsHtml += '</div>';
    }
    if (otherSolutions.length > 0) {
        solutionsHtml += `<div class="solution-group"><div class="solution-group-title">💡 其他建议</div>`;
        otherSolutions.forEach(s => {
            solutionsHtml += `<div class="solution-item ${s.type}"><span class="icon">${s.icon}</span><span>${s.message}</span></div>`;
        });
        solutionsHtml += '</div>';
    }
    
    solutionContent.innerHTML = solutionsHtml;
    
    // 渲染未满足明细
    let html = '';
    if (appState.unfilledRequirements.length > 0) {
        // 只显示前10条，避免列表过长
        const displayItems = appState.unfilledRequirements.slice(0, 10);
        displayItems.forEach(u => {
            const shift = appState.getShift(u.shiftId);
            html += `
                <div class="unfilled-item">
                    <span class="unfilled-icon">❌</span>
                    <div class="unfilled-info">
                        <div class="unfilled-title">${u.date} ${u.shiftName || shift?.name || ''} - ${u.position || ''}</div>
                        <div class="unfilled-detail">需要 ${u.required} 人，已排 ${u.assigned} 人，缺 ${u.required - u.assigned} 人</div>
                        ${u.reason ? `<div class="unfilled-reason">${u.reason}</div>` : ''}
                    </div>
                </div>
            `;
        });
        
        if (appState.unfilledRequirements.length > 10) {
            html += `<div class="unfilled-item" style="justify-content: center; color: var(--text-muted);">
                ... 还有 ${appState.unfilledRequirements.length - 10} 条未满足需求
            </div>`;
        }
    } else {
        html = '<div class="unfilled-item" style="justify-content: center; color: var(--success-color);">✅ 所有排班需求已满足</div>';
    }
    
    list.innerHTML = html;
    
    // 更新报告标题和面板样式
    const panelTitle = panel.querySelector('.panel-title');
    const hasProblems = appState.unfilledRequirements.length > 0 || (appState.constraintViolations && appState.constraintViolations.length > 0);
    
    if (panelTitle) {
        if (hasProblems) {
            panelTitle.textContent = '⚠️ 排班分析报告';
            panel.classList.add('has-problems');
            panel.classList.remove('success');
        } else {
            panelTitle.textContent = '✅ 排班分析报告';
            panel.classList.add('success');
            panel.classList.remove('has-problems');
        }
    }
    
    // 更新未满足需求部分标题
    if (unfilledSection) {
        const sectionTitle = unfilledSection.querySelector('.section-title');
        if (sectionTitle) {
            if (appState.unfilledRequirements.length > 0) {
                sectionTitle.textContent = '📋 未满足明细';
                unfilledSection.style.display = 'block';
            } else {
                sectionTitle.textContent = '📋 排班状态';
                unfilledSection.style.display = 'block';
            }
        }
    }
    
    panel.style.display = 'block';
}

/**
 * 分析排班问题并生成报告
 * 在"全部门店"模式下分析所有门店，单店模式只分析当前门店
 */
function analyzeSchedulingProblem() {
    const problems = [];
    const solutions = [];
    
    // 获取当前排班周期信息
    let periodDays = 7;
    if (appState.schedulePeriod === 'month') {
        // 计算当月天数
        const d = new Date(appState.currentWeekStart);
        periodDays = new Date(d.getFullYear(), d.getMonth() + 1, 0).getDate();
    } else if (appState.schedulePeriod) {
        periodDays = appState.schedulePeriod;
    }
    const periodName = periodDays === 7 ? '周' : (periodDays === 14 ? '双周' : '月');
    const isAllMode = appState.isAllStoresMode();
    const modeLabel = isAllMode ? '全部门店' : (appState.getCurrentStore()?.name || '当前门店');
    
    // 统计需求
    const totalUnfilled = appState.unfilledRequirements?.length || 0;
    const totalMissing = (appState.unfilledRequirements || []).reduce((sum, u) => sum + (u.required - u.assigned), 0);
    const totalAssigned = appState.assignments?.length || 0;
    
    // 计算满足率 - 使用与页面统计相同的方法，基于需求和分配数
    const weekDates = appState.getWeekDates();
    let satisfactionRate = calculateSatisfactionRateForStore(weekDates, appState.assignments || [], isAllMode ? 'all' : appState.currentStoreId);
    if (isNaN(satisfactionRate)) satisfactionRate = 100;
    
    // 按岗位分析缺口
    const positionGaps = {};
    (appState.unfilledRequirements || []).forEach(u => {
        const pos = u.position || '未知';
        if (!positionGaps[pos]) positionGaps[pos] = 0;
        positionGaps[pos] += (u.required - u.assigned);
    });
    
    // 员工分析 - 根据模式筛选员工
    // 在"全部门店"模式使用所有活跃员工，单店模式使用当前门店+可调配员工
    const relevantEmployees = isAllMode 
        ? appState.employees.filter(e => e.status === 'active')
        : appState.getCurrentStoreEmployees(true).filter(e => e.status === 'active');
    const allEmployees = isAllMode ? appState.employees : appState.getCurrentStoreEmployees(true);
    const activeEmployees = relevantEmployees;
    const positionCounts = {};
    activeEmployees.forEach(e => {
        const pos = e.position || '未知';
        if (!positionCounts[pos]) positionCounts[pos] = 0;
        positionCounts[pos]++;
    });
    
    // 计算每周理论最大班次
    const maxHoursPerWeek = appState.settings?.maxWeeklyHours || 44;
    const avgShiftHours = 5;
    const maxShiftsPerPersonPerWeek = Math.floor(maxHoursPerWeek / avgShiftHours);
    const weeks = Math.ceil(periodDays / 7);
    const theoreticalMaxShifts = activeEmployees.length * maxShiftsPerPersonPerWeek * weeks;
    
    // 计算实际需求班次
    const totalRequiredShifts = totalAssigned + totalMissing;
    
    // 计算利用率
    const utilizationRate = theoreticalMaxShifts > 0 ? Math.round((totalAssigned / theoreticalMaxShifts) * 100) : 0;
    
    // 人力状况摘要 - 始终显示
    const waiterCount = positionCounts['服务员'] || 0;
    const chefCount = positionCounts['厨师'] || 0;
    const totalActive = activeEmployees.length;
    const totalAll = allEmployees.length;
    const inactiveCount = totalAll - totalActive;
    
    // 显示当前分析的门店范围
    if (isAllMode) {
        problems.push({
            icon: '🏢',
            message: `联合排班：${appState.getAllStores().length} 家门店，${totalActive} 名在职员工（服务员 ${waiterCount}人，厨师 ${chefCount}人）`,
            severity: 'info',
            category: 'summary'
        });
    }
    
    // 添加人力概况（始终显示）
    if (satisfactionRate === 100) {
        if (utilizationRate < 50) {
            problems.push({
                icon: '✅',
                message: `人力充裕：利用率 ${utilizationRate}%，排班弹性良好`,
                severity: 'success',
                category: 'summary'
            });
        } else if (utilizationRate < 80) {
            problems.push({
                icon: '✅',
                message: `人力适中：利用率 ${utilizationRate}%`,
                severity: 'success',
                category: 'summary'
            });
        } else {
            problems.push({
                icon: '⚡',
                message: `人力紧张：利用率 ${utilizationRate}%，建议储备后备人员`,
                severity: 'warning',
                category: 'summary'
            });
        }
    } else {
        problems.push({
            icon: '📊',
            message: `满足率 ${satisfactionRate}%，利用率 ${utilizationRate}%`,
            severity: satisfactionRate < 80 ? 'warning' : 'info',
            category: 'summary'
        });
    }
    
    // 如果有离职员工，显示提示
    if (inactiveCount > 0) {
        problems.push({
            icon: 'ℹ️',
            message: `${inactiveCount} 名员工处于离职状态`,
            severity: 'info',
            category: 'summary'
        });
    }
    
    // 显示约束违反信息
    const violations = appState.constraintViolations || [];
    if (violations.length > 0) {
        const hardViolations = violations.filter(v => v.type === 'hard');
        const softViolations = violations.filter(v => v.type === 'soft');
        
        // 硬约束违反
        if (hardViolations.length > 0) {
            // 显示每个硬约束违反的详情（简化消息）
            hardViolations.forEach(v => {
                // 简化消息格式
                let msg = v.message;
                // 尝试提取关键信息
                const match = msg.match(/员工\s*(\S+)\s*连续工作\s*(\d+)\s*天/);
                if (match) {
                    msg = `${match[1]} 连续工作 ${match[2]} 天（限制 6 天）`;
                }
                problems.push({
                    icon: '⛔',
                    message: msg,
                    severity: 'error',
                    category: 'violation'
                });
            });
        }
        
        // 软约束违反
        if (softViolations.length > 0) {
            softViolations.forEach(v => {
                problems.push({
                    icon: '⚡',
                    message: `${v.constraintName || v.constraintType}: ${v.message}`,
                    severity: 'warning',
                    category: 'violation'
                });
            });
        }
    }
    
    // 问题1: 总体资源不足
    if (totalRequiredShifts > theoreticalMaxShifts) {
        const shortage = totalRequiredShifts - theoreticalMaxShifts;
        problems.push({
            icon: '🚨',
            message: `产能不足：需 ${totalRequiredShifts} 班次，最大产能 ${theoreticalMaxShifts} 班次，缺口 ${shortage}`,
            severity: 'critical',
            category: 'gap'
        });
    }
    
    // 问题2: 按岗位分析
    for (const [pos, gap] of Object.entries(positionGaps)) {
        const available = positionCounts[pos] || 0;
        if (gap > 0) {
            problems.push({
                icon: pos === '厨师' ? '👨‍🍳' : '🧑‍💼',
                message: `${pos}缺 ${gap} 班次（现有 ${available} 人）`,
                severity: gap > available * maxShiftsPerPersonPerWeek ? 'critical' : 'warning',
                category: 'gap'
            });
        }
    }
    
    // ========== 门店级别精确分析 ==========
    // 只有当满足率不是100%时才显示门店缺口（因为跨店员工可以满足任何门店需求）
    const storeAnalysis = analyzeByStore(periodDays, maxShiftsPerPersonPerWeek, weeks);
    
    // 只有当有实际未满足需求时才显示门店缺口
    if (satisfactionRate < 100 && storeAnalysis.storeGaps.length > 0) {
        storeAnalysis.storeGaps.forEach(sg => {
            sg.positionGaps.forEach(pg => {
                problems.push({
                    icon: '🏪',
                    message: `${sg.storeName}：${pg.position} 缺 ${pg.gap} 班次`,
                    severity: 'info',
                    category: 'store'
                });
            });
        });
    }
    
    // 解决方案 - 只有当满足率不是100%时才显示增员建议
    // 方案1: 按门店精确建议
    if (satisfactionRate < 100 && storeAnalysis.recommendations.length > 0) {
        storeAnalysis.recommendations.forEach(rec => {
            solutions.push({
                icon: '🎯',
                message: `${rec.storeName}：+${rec.count}${rec.position}（利用率→${rec.utilizationAfter}%）`,
                type: 'primary'
            });
        });
        
        // 总结
        solutions.push({
            icon: '📈',
            message: `共增 ${storeAnalysis.totalNewHires} 人，预计利用率 ${storeAnalysis.projectedUtilization}%`,
            type: 'primary'
        });
    } else if (satisfactionRate < 100) {
        // 方案2: 增加人手总数
        const neededEmployees = Math.ceil((totalRequiredShifts - theoreticalMaxShifts) / (maxShiftsPerPersonPerWeek * weeks));
        if (neededEmployees > 0) {
            solutions.push({
                icon: '➕',
                message: `增加约 ${neededEmployees} 名员工满足${periodName}需求`,
                type: 'primary'
            });
        }
        
        // 方案3: 按岗位建议
        for (const [pos, gap] of Object.entries(positionGaps)) {
            const needed = Math.ceil(gap / (maxShiftsPerPersonPerWeek * weeks));
            if (needed > 0) {
                solutions.push({
                    icon: '🎯',
                    message: `${pos}：+${needed}人或临时调配`,
                    type: 'primary'
                });
            }
        }
    }
    
    // 方案3: 缩短排班周期
    if (periodDays > 7 && satisfactionRate < 70) {
        solutions.push({
            icon: '📆',
            message: `尝试使用1周排班，便于灵活调整`,
            type: 'primary'
        });
    }
    
    // 方案4: 调整约束
    if (satisfactionRate < 50) {
        solutions.push({
            icon: '⚙️',
            message: `考虑临时放宽工时限制或休息要求`,
            type: 'secondary'
        });
    }
    
    // 方案5: 使用临时工/兼职
    if (totalMissing > 20) {
        solutions.push({
            icon: '🤝',
            message: `建议招聘临时工或兼职人员补充`,
            type: 'success'
        });
    }
    
    // 方案6: 针对约束违规的建议
    if (violations.length > 0) {
        // 分析违规类型并提供针对性建议
        const hasConsecutiveDaysViolation = violations.some(v => 
            v.constraintType === 'max_consecutive_days');
        const hasMaxHoursViolation = violations.some(v => 
            v.constraintType === 'max_hours_per_week' || v.constraintType === 'max_hours');
        const hasRestViolation = violations.some(v => 
            v.constraintType === 'min_rest_between_shifts');
        
        if (hasConsecutiveDaysViolation) {
            // 分析哪个岗位缺人
            const violationPositions = {};
            violations.filter(v => v.constraintType === 'max_consecutive_days').forEach(v => {
                // 从消息中提取员工名，然后找到其岗位
                const empName = v.message.match(/员工\s*(\S+)/)?.[1];
                if (empName) {
                    const emp = activeEmployees.find(e => e.name === empName);
                    if (emp) {
                        violationPositions[emp.position] = (violationPositions[emp.position] || 0) + 1;
                    }
                }
            });
            
            for (const [pos, count] of Object.entries(violationPositions)) {
                solutions.push({
                    icon: '👥',
                    message: `${pos}岗位人手不足：建议增加至少 1 名${pos}，以避免连续工作超限`,
                    type: 'primary'
                });
            }
            
            solutions.push({
                icon: '⚙️',
                message: `临时方案：可在设置中增加"最大连续工作天数"限制（当前：${appState.settings?.maxConsecutiveDays || 6}天）`,
                type: 'secondary'
            });
        }
        
        if (hasMaxHoursViolation) {
            solutions.push({
                icon: '⏰',
                message: `员工工时超限：建议增加人手分担工作量，或在设置中调整每周最大工时`,
                type: 'warning'
            });
        }
        
        if (hasRestViolation) {
            solutions.push({
                icon: '😴',
                message: `休息时间不足：建议调整班次时间或增加人手，确保员工有足够休息`,
                type: 'warning'
            });
        }
        
        // 如果有硬约束违规但没有具体建议，给出通用建议
        const hardViolations = violations.filter(v => v.type === 'hard');
        if (hardViolations.length > 0 && solutions.length === 0) {
            solutions.push({
                icon: '💡',
                message: `存在硬约束违规：请检查员工数量是否充足，或调整约束参数`,
                type: 'warning'
            });
        }
    }
    
    // 如果仍然没有解决方案但排班成功，给出正面反馈
    if (solutions.length === 0 && satisfactionRate === 100) {
        solutions.push({
            icon: '✅',
            message: `当前排班配置良好，无需调整`,
            type: 'success'
        });
    }
    
    return { problems, solutions };
}

/**
 * 按门店分析人力缺口并生成精确招聘建议
 * @param {number} periodDays - 排班周期天数
 * @param {number} maxShiftsPerPersonPerWeek - 每人每周最大班次
 * @param {number} weeks - 周数
 * @returns {Object} { storeGaps: [], recommendations: [], totalNewHires: number, projectedUtilization: number }
 */
function analyzeByStore(periodDays, maxShiftsPerPersonPerWeek, weeks) {
    const isAllMode = appState.isAllStoresMode();
    const stores = isAllMode ? appState.getAllStores() : [appState.getCurrentStore()];
    const storeGaps = [];
    const recommendations = [];
    let totalNewHires = 0;
    
    stores.forEach(store => {
        if (!store) return;
        
        // 获取该门店的员工
        const storeEmployees = appState.employees.filter(e => 
            e.storeId === store.id && e.status === 'active'
        );
        
        // 按岗位统计当前人数
        const positionCounts = {};
        storeEmployees.forEach(e => {
            const pos = e.position || '未知';
            positionCounts[pos] = (positionCounts[pos] || 0) + 1;
        });
        
        // 计算该门店的需求
        const weekDates = appState.getWeekDates();
        const positionDemand = {};
        
        weekDates.forEach(date => {
            const dayReqs = appState.getRequirementsForDate(date, store.id);
            appState.shifts.forEach(shift => {
                const shiftReqs = dayReqs[shift.id] || {};
                Object.entries(shiftReqs).forEach(([pos, count]) => {
                    if (count > 0) {
                        positionDemand[pos] = (positionDemand[pos] || 0) + count;
                    }
                });
            });
        });
        
        // 计算该门店的已分配
        const positionAssigned = {};
        appState.assignments.filter(a => a.storeId === store.id).forEach(a => {
            const pos = a.position || '未知';
            positionAssigned[pos] = (positionAssigned[pos] || 0) + 1;
        });
        
        // 分析每个岗位的缺口
        const positionGaps = [];
        const maxCapacityPerPerson = maxShiftsPerPersonPerWeek * weeks;
        
        Object.entries(positionDemand).forEach(([pos, demand]) => {
            const available = positionCounts[pos] || 0;
            const assigned = positionAssigned[pos] || 0;
            const gap = demand - assigned;
            
            if (gap > 0) {
                positionGaps.push({
                    position: pos,
                    demand: demand,
                    assigned: assigned,
                    gap: gap,
                    available: available
                });
                
                // 计算需要增加的人数
                const currentCapacity = available * maxCapacityPerPerson;
                const neededExtra = demand - currentCapacity;
                
                if (neededExtra > 0) {
                    const newHiresNeeded = Math.ceil(neededExtra / maxCapacityPerPerson);
                    
                    // 计算增员后利用率
                    const newTotal = available + newHiresNeeded;
                    const newCapacity = newTotal * maxCapacityPerPerson;
                    const utilizationAfter = Math.round((demand / newCapacity) * 100);
                    
                    recommendations.push({
                        storeId: store.id,
                        storeName: store.name,
                        position: pos,
                        count: newHiresNeeded,
                        currentCount: available,
                        utilizationAfter: utilizationAfter,
                        currentUtilization: currentCapacity > 0 ? Math.round((assigned / currentCapacity) * 100) : 0
                    });
                    
                    totalNewHires += newHiresNeeded;
                }
            }
        });
        
        if (positionGaps.length > 0) {
            storeGaps.push({
                storeId: store.id,
                storeName: store.name,
                positionGaps: positionGaps
            });
        }
    });
    
    // 计算增员后的整体利用率
    const totalAssigned = appState.assignments?.length || 0;
    const currentEmployees = appState.employees.filter(e => e.status === 'active').length;
    const currentCapacity = currentEmployees * maxShiftsPerPersonPerWeek * weeks;
    const newCapacity = (currentEmployees + totalNewHires) * maxShiftsPerPersonPerWeek * weeks;
    
    // 计算增员后的预期排班班次（假设能填满当前缺口）
    const currentGaps = storeGaps.reduce((sum, sg) => 
        sum + sg.positionGaps.reduce((s, pg) => s + pg.gap, 0), 0
    );
    const projectedAssigned = totalAssigned + currentGaps;
    
    // 预计利用率 = 预期排班班次 / 新增产能
    const projectedUtilization = newCapacity > 0 ? Math.min(Math.round((projectedAssigned / newCapacity) * 100), 100) : 0;
    
    return {
        storeGaps,
        recommendations,
        totalNewHires,
        projectedUtilization,
        currentUtilization: currentCapacity > 0 ? Math.round((totalAssigned / currentCapacity) * 100) : 0
    };
}

/* ========================================
   员工管理视图
   ======================================== */

// 员工日历当前查看月份
let employeeCalendarMonth = new Date();

function initEmployeeView() {
    // 初始化月份显示
    updateCalendarMonthDisplay();
    
    // 月份导航
    document.getElementById('prevMonth').addEventListener('click', () => {
        employeeCalendarMonth.setMonth(employeeCalendarMonth.getMonth() - 1);
        updateCalendarMonthDisplay();
        renderEmployeeGrid();
    });
    
    document.getElementById('nextMonth').addEventListener('click', () => {
        employeeCalendarMonth.setMonth(employeeCalendarMonth.getMonth() + 1);
        updateCalendarMonthDisplay();
        renderEmployeeGrid();
    });
    
    // 添加员工
    document.getElementById('addEmployee').addEventListener('click', () => {
        document.getElementById('employeeModalTitle').textContent = '添加员工';
        document.getElementById('employeeId').value = '';
        clearEmployeeForm();
        openModal('employeeModal');
    });
    
    // 保存员工
    document.getElementById('saveEmployee').addEventListener('click', saveEmployee);
    
    // 筛选
    document.getElementById('filterPosition').addEventListener('change', renderEmployeeGrid);
    document.getElementById('filterStatus').addEventListener('change', renderEmployeeGrid);
    document.getElementById('searchEmployee').addEventListener('input', debounce(renderEmployeeGrid, 300));
}

function updateCalendarMonthDisplay() {
    const display = document.getElementById('calendarMonthDisplay');
    if (display) {
        const year = employeeCalendarMonth.getFullYear();
        const month = employeeCalendarMonth.getMonth() + 1;
        display.textContent = `${year}年${month}月`;
    }
}

function renderEmployeeGrid() {
    const grid = document.getElementById('employeeGrid');
    const positionFilter = document.getElementById('filterPosition').value;
    const statusFilter = document.getElementById('filterStatus').value;
    const searchTerm = document.getElementById('searchEmployee').value.toLowerCase();
    
    // 获取所有需要显示的员工（包括离职员工）
    // "全部门店"模式：显示所有员工；单店模式：当前门店员工
    let employees;
    if (appState.isAllStoresMode()) {
        employees = [...appState.employees]; // 所有员工（包括离职）
    } else {
        employees = appState.employees.filter(e => e.storeId === appState.currentStoreId);
    }
    
    // 应用筛选
    if (positionFilter) {
        employees = employees.filter(e => e.position === positionFilter);
    }
    if (statusFilter) {
        employees = employees.filter(e => e.status === statusFilter);
    } else {
        // 默认不筛选状态时，显示所有员工（包括离职）
    }
    if (searchTerm) {
        employees = employees.filter(e => e.name.toLowerCase().includes(searchTerm));
    }
    
    // 按门店分组
    const hasMultipleStores = appState.stores && appState.stores.length > 1;
    const isAllMode = appState.isAllStoresMode();
    
    if (hasMultipleStores) {
        if (isAllMode) {
            // "全部门店"模式：按门店顺序分组显示所有员工
            const groupedEmps = {};
            appState.stores.forEach(store => {
                groupedEmps[store.id] = [];
            });
            
            employees.forEach(emp => {
                const storeId = emp.storeId || 'unknown';
                if (!groupedEmps[storeId]) {
                    groupedEmps[storeId] = [];
                }
                groupedEmps[storeId].push(emp);
            });
            
            employees = [];
            appState.stores.forEach(store => {
                if (groupedEmps[store.id]) {
                    employees = employees.concat(groupedEmps[store.id]);
                }
            });
        } else {
            // 单店模式：当前门店员工 + 其他门店可调配员工（按门店分组）
            const currentStoreEmps = employees.filter(e => e.storeId === appState.currentStoreId);
            const otherStoreEmps = employees.filter(e => e.storeId !== appState.currentStoreId);
            
            // 按门店ID分组其他门店员工
            const groupedOtherEmps = {};
            otherStoreEmps.forEach(emp => {
                const storeId = emp.storeId || 'unknown';
                if (!groupedOtherEmps[storeId]) {
                    groupedOtherEmps[storeId] = [];
                }
                groupedOtherEmps[storeId].push(emp);
            });
            
            employees = currentStoreEmps;
            // 将其他门店员工按门店顺序追加
            appState.stores.forEach(store => {
                if (store.id !== appState.currentStoreId && groupedOtherEmps[store.id]) {
                    employees = employees.concat(groupedOtherEmps[store.id]);
                }
            });
        }
    }
    
    // 获取选定月份的所有排班数据（已发布和已归档）
    const year = employeeCalendarMonth.getFullYear();
    const month = employeeCalendarMonth.getMonth();
    const monthAssignments = appState.getMonthAssignments(year, month);
    
    // 计算每个员工在该月的排班统计
    const empStats = {};
    monthAssignments.forEach(a => {
        const empName = a.employeeName;
        if (!empStats[empName]) {
            empStats[empName] = { shifts: 0, hours: 0 };
        }
        empStats[empName].shifts++;
        empStats[empName].hours += a.hours;
    });
    
    let html = '';
    let currentGroupStoreId = null;
    
    employees.forEach(emp => {
        // 添加门店分组标题（当有多门店时）
        if (hasMultipleStores && emp.storeId !== currentGroupStoreId) {
            currentGroupStoreId = emp.storeId;
            const store = appState.stores.find(s => s.id === currentGroupStoreId);
            const storeType = store ? (STORE_TYPES[store.type] || STORE_TYPES.standard) : STORE_TYPES.standard;
            const isCurrentStore = currentGroupStoreId === appState.currentStoreId;
            
            // "全部门店"模式下，所有门店都平等显示
            let groupLabel;
            if (isAllMode) {
                groupLabel = `${storeType.icon} ${store ? store.name : '门店'}`;
            } else {
                groupLabel = isCurrentStore 
                    ? `📍 ${store ? store.name : '当前门店'}（本店员工）` 
                    : `${storeType.icon} ${store ? store.name : '其他门店'}（可调配员工）`;
            }
            
            html += `
                <div class="employee-group-header ${isAllMode ? 'all-stores' : (isCurrentStore ? 'current-store' : 'other-store')}">
                    <span class="group-label">${groupLabel}</span>
                    <span class="group-count">${employees.filter(e => e.storeId === currentGroupStoreId).length}人</span>
                </div>
            `;
        }
        const stats = empStats[emp.name] || { shifts: 0, hours: 0 };
        const positionIcon = getPositionIcon(emp.position);
        
        // 获取员工所属门店信息（有多个门店时显示门店标识）
        const empStore = appState.stores?.find(s => s.id === emp.storeId);
        let storeBadge = '';
        if (hasMultipleStores && empStore) {
            // "全部门店"模式或非当前门店的员工显示门店标识
            if (isAllMode || emp.storeId !== appState.currentStoreId) {
                storeBadge = `<span class="store-badge" title="${empStore.name}">${empStore.code}</span>`;
            }
        }
        
        // 获取该员工在选定月份的排班情况
        const empAssignments = monthAssignments.filter(a => a.employeeName === emp.name);
        
        // 生成日历视图
        const calendarHtml = generateEmployeeCalendar(empAssignments, emp.id);
        
        html += `
            <div class="employee-card-large">
                <div class="employee-card-header">
                    <div class="employee-avatar" style="background: ${stringToColor(emp.name)}">${getAvatarLetter(emp.name)}</div>
                    <div class="employee-info">
                        <div class="employee-name-row">
                            <h4>${formatEmployeeName(emp)}</h4>
                            <button class="btn-edit-emp" onclick="event.stopPropagation(); editEmployee('${emp.id}')">✏️</button>
                        </div>
                        <span class="employee-position">${positionIcon} ${emp.position} ${storeBadge}</span>
                        <div class="employee-skills-inline">
                            ${(emp.skills || []).map(s => `<span class="skill-tag-small">${s}</span>`).join('')}
                        </div>
                    </div>
                    <div class="employee-summary">
                        <span class="summary-item"><strong>${stats.shifts}</strong> 班次</span>
                        <span class="summary-item"><strong>${stats.hours}</strong> 工时</span>
                        <span class="summary-item status-${emp.status}">${emp.status === 'active' ? '✅在职' : '❌离职'}</span>
                        ${emp.canTransfer ? '<span class="summary-item transfer-badge" title="可跨店调配">🔄</span>' : ''}
                    </div>
                </div>
                <div class="employee-calendar" id="emp-calendar-${emp.id}">
                    ${calendarHtml}
                </div>
            </div>
        `;
    });
    
    if (employees.length === 0) {
        html = '<div class="empty-state"><p>暂无员工数据</p></div>';
    }
    
    grid.innerHTML = html;
}

/**
 * 生成员工日历视图
 */
function generateEmployeeCalendar(empAssignments, empId) {
    // 获取班次缩写映射
    const shiftCodeMap = {};
    appState.shifts.forEach(s => {
        shiftCodeMap[s.id] = { code: s.code || s.name.charAt(0), name: s.name };
    });
    
    // 按日期分组排班
    const assignmentsByDate = {};
    empAssignments.forEach(a => {
        if (!assignmentsByDate[a.date]) {
            assignmentsByDate[a.date] = [];
        }
        assignmentsByDate[a.date].push(a);
    });
    
    // 使用选定的月份
    const year = employeeCalendarMonth.getFullYear();
    const month = employeeCalendarMonth.getMonth();
    const firstDayOfMonth = new Date(year, month, 1);
    const lastDayOfMonth = new Date(year, month + 1, 0);
    
    // 生成日历 HTML
    let html = '<div class="emp-calendar-grid">';
    
    // 日历头部 - 星期
    const dayNames = ['日', '一', '二', '三', '四', '五', '六'];
    html += '<div class="calendar-header">';
    dayNames.forEach(d => {
        html += `<div class="calendar-day-name">${d}</div>`;
    });
    html += '</div>';
    
    // 补齐月初空白
    const startPadding = firstDayOfMonth.getDay();
    
    html += '<div class="calendar-body">';
    
    // 添加月初空白
    for (let i = 0; i < startPadding; i++) {
        html += '<div class="calendar-cell empty"></div>';
    }
    
    // 添加日期单元格
    for (let d = 1; d <= lastDayOfMonth.getDate(); d++) {
        const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
        const dayAssignments = assignmentsByDate[dateStr] || [];
        const isToday = dateStr === formatDate(new Date());
        const isWeekend = (startPadding + d - 1) % 7 === 0 || (startPadding + d - 1) % 7 === 6;
        
        let cellClass = 'calendar-cell';
        if (isToday) cellClass += ' today';
        if (isWeekend) cellClass += ' weekend';
        if (dayAssignments.length > 0) cellClass += ' has-shift';
        
        html += `<div class="${cellClass}">`;
        html += `<div class="calendar-date">${d}</div>`;
        
        if (dayAssignments.length > 0) {
            html += '<div class="calendar-shifts">';
            dayAssignments.forEach(a => {
                const shift = shiftCodeMap[a.shiftId] || { code: '?', name: '未知' };
                // 添加点击事件显示班次详情，使用Base64编码避免特殊字符问题
                const assignmentData = btoa(encodeURIComponent(JSON.stringify({
                    date: a.date,
                    shiftId: a.shiftId,
                    shiftName: shift.name,
                    employeeName: a.employeeName,
                    hours: a.hours,
                    score: a.score
                })));
                html += `<span class="shift-badge clickable" title="${shift.name}" onclick="showCalendarShiftDetail('${assignmentData}')">${shift.code}</span>`;
            });
            html += '</div>';
        }
        
        html += '</div>';
    }
    
    html += '</div>';  // 关闭 calendar-body
    html += '</div>';  // 关闭 emp-calendar-grid
    return html;
}

function getPositionIcon(position) {
    const icons = {
        '服务员': '👤',
        '厨师': '👨‍🍳',
        '收银员': '💰',
        '店长': '👔'
    };
    return icons[position] || '👤';
}

/**
 * 显示班次详情弹窗（点击日历中的班次标签时触发）
 */
function showCalendarShiftDetail(encodedData) {
    try {
        // 先用atob解码Base64，再用decodeURIComponent解码URI编码
        const data = JSON.parse(decodeURIComponent(atob(encodedData)));
        
        // 格式化日期
        const dateObj = new Date(data.date);
        const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
        const formattedDate = `${data.date} ${weekDays[dateObj.getDay()]}`;
        
        // 获取班次详情
        const shift = appState.shifts.find(s => s.id === data.shiftId);
        const shiftTime = shift ? `${shift.startTime} - ${shift.endTime}` : '未知时间';
        
        // 创建弹窗内容
        const content = `
            <div class="assignment-detail-modal">
                <div class="detail-header">
                    <h3>📅 排班详情</h3>
                    <button class="modal-close" onclick="this.closest('.modal-overlay').remove()">×</button>
                </div>
                <div class="detail-content">
                    <div class="detail-row">
                        <span class="detail-label">👤 员工</span>
                        <span class="detail-value">${formatEmployeeNameByName(data.employeeName)}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">📆 日期</span>
                        <span class="detail-value">${formattedDate}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">🕐 班次</span>
                        <span class="detail-value">${data.shiftName}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">⏰ 时间</span>
                        <span class="detail-value">${shiftTime}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">📊 工时</span>
                        <span class="detail-value">${data.hours} 小时</span>
                    </div>
                    ${data.score ? `
                    <div class="detail-row">
                        <span class="detail-label">⭐ 评分</span>
                        <span class="detail-value">${data.score} 分</span>
                    </div>
                    ` : ''}
                </div>
            </div>
        `;
        
        // 创建遮罩层
        const overlay = document.createElement('div');
        overlay.className = 'modal-overlay';
        overlay.innerHTML = content;
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) overlay.remove();
        });
        
        document.body.appendChild(overlay);
    } catch (e) {
        console.error('Failed to show assignment detail:', e);
        showToast('无法显示班次详情', 'error');
    }
}

function editEmployee(id) {
    const emp = appState.getEmployee(id);
    if (!emp) return;
    
    document.getElementById('employeeModalTitle').textContent = '编辑员工';
    document.getElementById('employeeId').value = emp.id;
    document.getElementById('empName').value = emp.name;
    document.getElementById('empPosition').value = emp.position;
    document.getElementById('empPhone').value = emp.phone || '';
    document.getElementById('empHireDate').value = emp.hireDate || '';
    document.getElementById('empStatus').value = emp.status || 'active';
    
    // 填充门店选择
    populateStoreSelect();
    document.getElementById('empStore').value = emp.storeId || appState.currentStoreId;
    
    // 设置跨店调配
    document.getElementById('empCanTransfer').checked = emp.canTransfer || false;
    
    // 设置技能
    document.querySelectorAll('#empSkills input[type="checkbox"]').forEach(cb => {
        cb.checked = (emp.skills || []).includes(cb.value);
    });
    
    openModal('employeeModal');
}

// 填充门店选择下拉框
function populateStoreSelect() {
    const select = document.getElementById('empStore');
    if (!select) return;
    
    select.innerHTML = '';
    appState.stores.forEach(store => {
        if (store.status === 'active') {
            const opt = document.createElement('option');
            opt.value = store.id;
            opt.textContent = `${store.name} (${store.code})`;
            select.appendChild(opt);
        }
    });
}

function clearEmployeeForm() {
    document.getElementById('empName').value = '';
    document.getElementById('empPosition').value = '服务员';
    document.getElementById('empPhone').value = '';
    document.getElementById('empHireDate').value = '';
    document.getElementById('empStatus').value = 'active';
    
    // 填充门店选择并默认选中当前门店
    populateStoreSelect();
    document.getElementById('empStore').value = appState.currentStoreId;
    
    // 默认可跨店调配为否
    document.getElementById('empCanTransfer').checked = false;
    
    document.querySelectorAll('#empSkills input[type="checkbox"]').forEach(cb => {
        cb.checked = false;
    });
}

function saveEmployee() {
    const id = document.getElementById('employeeId').value;
    const name = document.getElementById('empName').value.trim();
    const position = document.getElementById('empPosition').value;
    const phone = document.getElementById('empPhone').value.trim();
    const hireDate = document.getElementById('empHireDate').value;
    const status = document.getElementById('empStatus').value;
    const storeId = document.getElementById('empStore').value;
    const canTransfer = document.getElementById('empCanTransfer').checked;
    
    if (!name) {
        showToast('请输入员工姓名', 'warning');
        return;
    }
    
    // 获取选中的技能
    const skills = [];
    document.querySelectorAll('#empSkills input[type="checkbox"]:checked').forEach(cb => {
        skills.push(cb.value);
    });
    
    const employeeData = {
        name,
        position,
        skills,
        phone,
        hireDate,
        status,
        storeId,
        canTransfer
    };
    
    if (id) {
        appState.updateEmployee(id, employeeData);
        showToast('员工信息已更新', 'success');
    } else {
        appState.addEmployee(employeeData);
        showToast('员工添加成功', 'success');
    }
    
    closeModal('employeeModal');
    renderEmployeeGrid();
}

/* ========================================
   班次设置视图
   ======================================== */

function initShiftView() {
    // 添加班次
    document.getElementById('addShift').addEventListener('click', () => {
        document.getElementById('shiftModalTitle').textContent = '添加班次';
        document.getElementById('shiftId').value = '';
        clearShiftForm();
        openModal('shiftModal');
    });
    
    // 保存班次
    document.getElementById('saveShift').addEventListener('click', saveShift);
}

function renderShiftList() {
    const list = document.getElementById('shiftList');
    
    let html = '';
    appState.shifts.forEach(shift => {
        const shiftClass = getShiftTypeClass(shift.id);
        html += `
            <div class="shift-card ${shiftClass}" onclick="editShift('${shift.id}')">
                <div class="shift-color" style="background: ${shift.color}">${shift.code}</div>
                <div class="shift-info">
                    <h4>${shift.name}</h4>
                    <div class="shift-time">${shift.startTime} - ${shift.endTime}</div>
                </div>
                <div class="shift-hours">${shift.hours}h</div>
            </div>
        `;
    });
    
    list.innerHTML = html;
    
    // 渲染需求配置
    renderRequirementsConfig();
}

function getShiftTypeClass(shiftId) {
    if (shiftId.includes('morning')) return 'morning';
    if (shiftId.includes('afternoon')) return 'afternoon';
    if (shiftId.includes('evening')) return 'evening';
    if (shiftId.includes('split')) return 'split';
    return '';
}

// 切换需求配置的门店
function switchRequirementsStore(storeId) {
    appState.requirementsStoreId = storeId;
    renderRequirementsConfig();
    // 关闭下拉
    const list = document.getElementById('reqStoreList');
    if (list) list.classList.remove('show');
}

// 切换需求配置门店下拉显示
function toggleReqStoreDropdown() {
    const list = document.getElementById('reqStoreList');
    if (list) list.classList.toggle('show');
}

function renderRequirementsConfig() {
    const grid = document.getElementById('requirementsGrid');
    const positions = ['服务员', '厨师'];
    
    // 检查是否是全部门店模式
    const isAllMode = appState.isAllStoresMode();
    
    let html = '';
    
    if (isAllMode) {
        // 全部门店模式：显示所有门店需求总和
        html += `
            <div class="req-section req-store-hint">
                <span class="hint-icon">🏢</span>
                <span class="hint-text">全部门店需求总和（只读，切换到单店模式可编辑）</span>
            </div>
        `;
        
        // 计算所有门店的需求总和
        const totalWeekday = {};
        const totalWeekend = {};
        const totalHoliday = {};
        
        appState.stores.forEach(store => {
            const storeReqs = getStoreRequirements(store.id);
            appState.shifts.forEach(shift => {
                if (!totalWeekday[shift.id]) totalWeekday[shift.id] = {};
                if (!totalWeekend[shift.id]) totalWeekend[shift.id] = {};
                if (!totalHoliday[shift.id]) totalHoliday[shift.id] = {};
                
                positions.forEach(pos => {
                    totalWeekday[shift.id][pos] = (totalWeekday[shift.id][pos] || 0) + (storeReqs.weekday?.[shift.id]?.[pos] || 0);
                    totalWeekend[shift.id][pos] = (totalWeekend[shift.id][pos] || 0) + (storeReqs.weekend?.[shift.id]?.[pos] || 0);
                    totalHoliday[shift.id][pos] = (totalHoliday[shift.id][pos] || 0) + (storeReqs.holiday?.[shift.id]?.[pos] || storeReqs.weekend?.[shift.id]?.[pos] || 0);
                });
            });
        });
        
        // 工作日需求总和
        html += '<div class="req-section"><h4>📅 工作日需求（总和）</h4></div>';
        appState.shifts.forEach(shift => {
            html += `
                <div class="requirement-config readonly">
                    <div class="req-header">
                        <span class="req-title">${shift.name}</span>
                    </div>
                    <div class="req-inputs">
                        ${positions.map(pos => `
                            <div class="req-row">
                                <label>${pos}</label>
                                <span class="req-value">${totalWeekday[shift.id]?.[pos] || 0}</span>
                                <span>人</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        });
        
        // 周末需求总和
        html += '<div class="req-section"><h4>🎉 周末需求（总和）</h4></div>';
        appState.shifts.forEach(shift => {
            html += `
                <div class="requirement-config readonly">
                    <div class="req-header">
                        <span class="req-title">${shift.name}</span>
                    </div>
                    <div class="req-inputs">
                        ${positions.map(pos => `
                            <div class="req-row">
                                <label>${pos}</label>
                                <span class="req-value">${totalWeekend[shift.id]?.[pos] || 0}</span>
                                <span>人</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        });
        
        // 法定假日需求总和
        html += '<div class="req-section"><h4>🎌 法定假日需求（总和）</h4></div>';
        appState.shifts.forEach(shift => {
            html += `
                <div class="requirement-config readonly">
                    <div class="req-header">
                        <span class="req-title">${shift.name}</span>
                    </div>
                    <div class="req-inputs">
                        ${positions.map(pos => `
                            <div class="req-row">
                                <label>${pos}</label>
                                <span class="req-value">${totalHoliday[shift.id]?.[pos] || 0}</span>
                                <span>人</span>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        });
        
        grid.innerHTML = html;
        return;
    }
    
    // 单店模式：显示当前门店的需求配置（可编辑）
    const selectedStoreId = appState.currentStoreId || 'store-001';
    const storeReqs = getStoreRequirements(selectedStoreId);
    const selectedStore = appState.stores.find(s => s.id === selectedStoreId);
    const storeName = selectedStore?.name || '门店';
    
    html += `
        <div class="req-section req-store-hint">
            <span class="hint-icon">💡</span>
            <span class="hint-text">当前配置：<strong>${storeName}</strong>（使用顶部门店选择器切换）</span>
        </div>
    `;
    
    // 工作日需求
    html += '<div class="req-section"><h4>📅 工作日需求</h4></div>';
    appState.shifts.forEach(shift => {
        html += `
            <div class="requirement-config">
                <div class="req-header">
                    <span class="req-title">${shift.name}</span>
                </div>
                <div class="req-inputs">
                    ${positions.map(pos => `
                        <div class="req-row">
                            <label>${pos}</label>
                            <input type="number" min="0" max="10" 
                                value="${storeReqs.weekday?.[shift.id]?.[pos] || 0}"
                                onchange="updateRequirement('${selectedStoreId}', 'weekday', '${shift.id}', '${pos}', this.value)">
                            <span>人</span>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    });
    
    // 周末需求
    html += '<div class="req-section"><h4>🎉 周末需求</h4></div>';
    appState.shifts.forEach(shift => {
        html += `
            <div class="requirement-config">
                <div class="req-header">
                    <span class="req-title">${shift.name}</span>
                </div>
                <div class="req-inputs">
                    ${positions.map(pos => `
                        <div class="req-row">
                            <label>${pos}</label>
                            <input type="number" min="0" max="10" 
                                value="${storeReqs.weekend?.[shift.id]?.[pos] || 0}"
                                onchange="updateRequirement('${selectedStoreId}', 'weekend', '${shift.id}', '${pos}', this.value)">
                            <span>人</span>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    });
    
    // 法定假日需求
    html += '<div class="req-section"><h4>🎌 法定假日需求</h4></div>';
    appState.shifts.forEach(shift => {
        html += `
            <div class="requirement-config">
                <div class="req-header">
                    <span class="req-title">${shift.name}</span>
                </div>
                <div class="req-inputs">
                    ${positions.map(pos => `
                        <div class="req-row">
                            <label>${pos}</label>
                            <input type="number" min="0" max="10" 
                                value="${storeReqs.holiday?.[shift.id]?.[pos] || storeReqs.weekend?.[shift.id]?.[pos] || 0}"
                                onchange="updateRequirement('${selectedStoreId}', 'holiday', '${shift.id}', '${pos}', this.value)">
                            <span>人</span>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    });
    
    grid.innerHTML = html;
}

// 获取指定门店的需求配置
function getStoreRequirements(storeId) {
    // 支持新格式和旧格式
    if (appState.requirements[storeId]) {
        return appState.requirements[storeId];
    }
    if (appState.requirements['_default']) {
        return appState.requirements['_default'];
    }
    // 兼容旧格式
    if (appState.requirements.weekday) {
        return appState.requirements;
    }
    return { weekday: {}, weekend: {} };
}

function updateRequirement(storeId, dayType, shiftId, position, value) {
    // 确保门店配置存在
    if (!appState.requirements[storeId]) {
        appState.requirements[storeId] = { weekday: {}, weekend: {} };
    }
    if (!appState.requirements[storeId][dayType]) {
        appState.requirements[storeId][dayType] = {};
    }
    if (!appState.requirements[storeId][dayType][shiftId]) {
        appState.requirements[storeId][dayType][shiftId] = {};
    }
    appState.requirements[storeId][dayType][shiftId][position] = parseInt(value) || 0;
    appState.saveToStorage('requirements', appState.requirements);
}

function editShift(id) {
    const shift = appState.getShift(id);
    if (!shift) return;
    
    document.getElementById('shiftModalTitle').textContent = '编辑班次';
    document.getElementById('shiftId').value = shift.id;
    document.getElementById('shiftName').value = shift.name;
    document.getElementById('shiftCode').value = shift.code;
    document.getElementById('shiftStart').value = shift.startTime;
    document.getElementById('shiftEnd').value = shift.endTime;
    document.getElementById('shiftColor').value = shift.color;
    
    openModal('shiftModal');
}

function clearShiftForm() {
    document.getElementById('shiftName').value = '';
    document.getElementById('shiftCode').value = '';
    document.getElementById('shiftStart').value = '09:00';
    document.getElementById('shiftEnd').value = '17:00';
    document.getElementById('shiftColor').value = '#f59e0b';
}

function saveShift() {
    const id = document.getElementById('shiftId').value;
    const name = document.getElementById('shiftName').value.trim();
    const code = document.getElementById('shiftCode').value.trim();
    const startTime = document.getElementById('shiftStart').value;
    const endTime = document.getElementById('shiftEnd').value;
    const color = document.getElementById('shiftColor').value;
    
    if (!name) {
        showToast('请输入班次名称', 'warning');
        return;
    }
    
    const hours = calculateHours(startTime, endTime);
    
    const shiftData = {
        name,
        code: code || name.charAt(0),
        startTime,
        endTime,
        color,
        hours
    };
    
    if (id) {
        appState.updateShift(id, shiftData);
        showToast('班次已更新', 'success');
    } else {
        appState.addShift(shiftData);
        showToast('班次添加成功', 'success');
    }
    
    closeModal('shiftModal');
    renderShiftList();
}

/* ========================================
   设置视图
   ======================================== */

function initSettingsView() {
    // 加载当前设置
    loadSettings();
    
    // 工时模式切换
    document.getElementById('hoursMode').addEventListener('change', (e) => {
        updateHoursModeDisplay(e.target.value);
    });
    
    // 测试连接
    document.getElementById('testConnection').addEventListener('click', async () => {
        const statusEl = document.getElementById('connectionStatus');
        statusEl.textContent = '连接中...';
        statusEl.className = 'connection-status';
        
        scheduleAPI.updateConfig();
        const result = await scheduleAPI.testConnection();
        
        if (result.success) {
            statusEl.textContent = '✅ 连接成功';
            statusEl.className = 'connection-status success';
        } else {
            statusEl.textContent = `❌ 连接失败: ${result.error}`;
            statusEl.className = 'connection-status error';
        }
    });
    
    // 保存设置
    document.getElementById('saveSettings').addEventListener('click', () => {
        const settings = {
            storeName: document.getElementById('storeName').value,
            openTime: document.getElementById('openTime').value,
            closeTime: document.getElementById('closeTime').value,
            hoursMode: document.getElementById('hoursMode').value,
            maxWeeklyHours: parseInt(document.getElementById('maxWeeklyHours').value),
            maxPeriodHours: parseInt(document.getElementById('maxPeriodHours').value),
            minRestHours: parseInt(document.getElementById('minRestHours').value),
            maxConsecutiveDays: parseInt(document.getElementById('maxConsecutiveDays').value),
            minRestDays: parseInt(document.getElementById('minRestDays').value),
            apiEndpoint: document.getElementById('apiEndpoint').value,
            timeout: parseInt(document.getElementById('timeout').value)
        };
        
        appState.updateSettings(settings);
        scheduleAPI.updateConfig();
        
        // 更新门店名称显示
        document.querySelector('.store-name').textContent = `🏪 ${settings.storeName}`;
        
        showToast('设置已保存', 'success');
    });
    
    // 重置设置
    document.getElementById('resetSettings').addEventListener('click', async () => {
        const confirmed = await showConfirm('重置设置', '确定要恢复默认设置吗？');
        if (confirmed) {
            appState.resetSettings();
            loadSettings();
            showToast('设置已重置', 'info');
        }
    });
}

function updateHoursModeDisplay(mode) {
    const weeklyRow = document.getElementById('weeklyHoursRow');
    const periodRow = document.getElementById('periodHoursRow');
    
    if (mode === 'period') {
        weeklyRow.style.display = 'none';
        periodRow.style.display = 'flex';
    } else {
        weeklyRow.style.display = 'flex';
        periodRow.style.display = 'none';
    }
}

function loadSettings() {
    const s = appState.settings;
    document.getElementById('storeName').value = s.storeName;
    document.getElementById('openTime').value = s.openTime;
    document.getElementById('closeTime').value = s.closeTime;
    document.getElementById('hoursMode').value = s.hoursMode || 'weekly';
    document.getElementById('maxWeeklyHours').value = s.maxWeeklyHours;
    document.getElementById('maxPeriodHours').value = s.maxPeriodHours || 176;
    document.getElementById('minRestHours').value = s.minRestHours;
    document.getElementById('maxConsecutiveDays').value = s.maxConsecutiveDays;
    document.getElementById('minRestDays').value = s.minRestDays;
    document.getElementById('apiEndpoint').value = s.apiEndpoint;
    document.getElementById('timeout').value = s.timeout;
    
    // 更新工时模式显示
    updateHoursModeDisplay(s.hoursMode || 'weekly');
    
    // 更新门店名称显示
    document.querySelector('.store-name').textContent = `🏪 ${s.storeName}`;
}

/* ========================================
   智能排班
   ======================================== */

function initGenerateButton() {
    document.getElementById('btnGenerate').addEventListener('click', generateSchedule);
}

async function generateSchedule() {
    const btn = document.getElementById('btnGenerate');
    const originalText = btn.innerHTML;
    
    btn.innerHTML = '<span>⏳ 排班中...</span>';
    btn.disabled = true;
    
    try {
        const weekDates = appState.getWeekDates();
        const result = await scheduleAPI.generateSchedule(weekDates);
        
        // 只要有排班结果就显示，不论 success 是否为 true
        if (result.assignments && result.assignments.length > 0) {
            appState.assignments = result.assignments;
            
            // 保存约束违反信息
            appState.constraintViolations = result.constraintViolations || [];
            
            // 保存补员建议
            appState.staffingSuggestions = result.staffingSuggestions || [];
            
            // 优先使用后端返回的未满足需求，如果没有则前端计算
            if (result.unfilledRequirements && result.unfilledRequirements.length > 0) {
                appState.unfilledRequirements = result.unfilledRequirements;
            } else {
                // 后端返回空或没有，使用前端计算
                appState.unfilledRequirements = calculateUnfilledRequirements(weekDates, result.assignments);
            }
            console.log('未满足需求:', appState.unfilledRequirements);
            
            renderScheduleGrid();
            renderEmployeeGrid(); // 更新员工统计
            
            // 检查满足率
            const satisfactionRate = result.statistics?.satisfactionRate || 
                calculateSatisfactionRate(weekDates, result.assignments);
            
            // 不再自动弹出分析报告，用户可点击"排班报告"按钮查看
            // showUnfilledRequirements();
            
            // Toast 消息 - 根据是否有约束违反决定消息类型
            if (result.constraintViolations && result.constraintViolations.length > 0) {
                showToast(`排班完成但存在约束违规，共 ${result.assignments.length} 个班次，请查看分析报告`, 'warning');
            } else if (satisfactionRate < 100 || appState.unfilledRequirements.length > 0) {
                showToast(`排班完成，满足率 ${satisfactionRate}%，请查看分析报告`, 'warning');
            } else {
                const avgScore = result.statistics?.averageScore || result.statistics?.avgScore || 
                    Math.round(result.assignments.reduce((sum, a) => sum + (a.score || 0), 0) / result.assignments.length) || 0;
                showToast(`排班成功！共 ${result.assignments.length} 个班次，平均评分 ${avgScore} 分`, 'success');
            }
            
            console.log('排班结果:', result);
        } else {
            // 真正的失败：没有任何排班结果
            showToast('排班失败：无法生成任何排班，请检查员工数量和设置', 'error');
        }
        
    } catch (error) {
        console.error('排班失败:', error);
        showToast(`排班失败: ${error.message}`, 'error');
    } finally {
        btn.innerHTML = originalText;
        btn.disabled = false;
    }
}

/**
 * 计算未满足的排班需求（按日期+班次+岗位统计，不区分门店）
 * 因为跨店员工可以满足任何门店的需求
 */
function calculateUnfilledRequirements(weekDates, assignments) {
    const unfilled = [];
    
    // 格式化日期为字符串 YYYY-MM-DD
    const formatDateStr = (d) => {
        const date = new Date(d);
        return date.getFullYear() + '-' + 
               String(date.getMonth() + 1).padStart(2, '0') + '-' + 
               String(date.getDate()).padStart(2, '0');
    };
    
    const isAllMode = appState.isAllStoresMode();
    const stores = isAllMode ? appState.getAllStores() : [appState.getCurrentStore()].filter(Boolean);
    
    // 遍历每天、每个班次
    weekDates.forEach(dateObj => {
        const dateStr = formatDateStr(dateObj);
        
        appState.shifts.forEach(shift => {
            // 获取该日期该班次的所有排班（不区分门店）
            const shiftAssignments = assignments.filter(
                a => a.date === dateStr && a.shiftId === shift.id
            );
            
            // 统计每个岗位已分配人数
            const positionAssigned = {};
            shiftAssignments.forEach(a => {
                const pos = a.position || '未知';
                positionAssigned[pos] = (positionAssigned[pos] || 0) + 1;
            });
            
            // 汇总所有门店该日期该班次的需求
            const totalReqByPosition = {};
            stores.forEach(store => {
                const dayReqs = appState.getRequirementsForDate(dateObj, store.id);
                const shiftReqs = dayReqs[shift.id] || {};
                Object.entries(shiftReqs).forEach(([pos, count]) => {
                    totalReqByPosition[pos] = (totalReqByPosition[pos] || 0) + count;
                });
            });
            
            // 计算每个岗位的缺口
            Object.entries(totalReqByPosition).forEach(([pos, required]) => {
                if (required > 0) {
                    const assigned = positionAssigned[pos] || 0;
                    if (assigned < required) {
                        unfilled.push({
                            date: dateStr,
                            shiftId: shift.id,
                            shiftName: shift.name,
                            position: pos,
                            required,
                            assigned,
                            storeId: '',
                            storeName: '全部门店',
                            reason: assigned === 0 ? '无可用员工' : '员工不足'
                        });
                    }
                }
            });
        });
    });
    
    return unfilled;
}

/**
 * 计算每个岗位的缺口
 * @param {Date} date - 日期
 * @param {string} shiftId - 班次ID
 * @param {Array} cellAssignments - 该班次已分配的排班
 * @param {string} filterStoreId - 门店筛选ID
 * @param {boolean} isAllMode - 是否全部门店模式
 * @returns {Array} 缺口数组，如 [{position: '厨师', gap: 1}, {position: '服务员', gap: 2}]
 */
function calculatePositionGaps(date, shiftId, cellAssignments, filterStoreId, isAllMode) {
    const gaps = [];
    
    // 收集所有需求
    const positionReqs = {};
    
    if (filterStoreId && filterStoreId !== 'all') {
        // 有门店筛选时，只计算该门店的需求
        const dayReqs = appState.getRequirementsForDate(date, filterStoreId);
        const shiftReqs = dayReqs[shiftId] || {};
        Object.entries(shiftReqs).forEach(([pos, count]) => {
            positionReqs[pos] = (positionReqs[pos] || 0) + count;
        });
    } else if (isAllMode) {
        // 全部门店模式，汇总所有门店的需求
        appState.getAllStores().forEach(store => {
            const dayReqs = appState.getRequirementsForDate(date, store.id);
            const shiftReqs = dayReqs[shiftId] || {};
            Object.entries(shiftReqs).forEach(([pos, count]) => {
                positionReqs[pos] = (positionReqs[pos] || 0) + count;
            });
        });
    } else {
        // 单门店模式
        const dayReqs = appState.getRequirementsForDate(date);
        const shiftReqs = dayReqs[shiftId] || {};
        Object.entries(shiftReqs).forEach(([pos, count]) => {
            positionReqs[pos] = (positionReqs[pos] || 0) + count;
        });
    }
    
    // 统计每个岗位已分配人数
    const positionAssigned = {};
    cellAssignments.forEach(a => {
        const pos = a.position || '未知';
        positionAssigned[pos] = (positionAssigned[pos] || 0) + 1;
    });
    
    // 计算缺口（按岗位优先级排序：厨师 > 服务员 > 其他）
    const positionOrder = ['厨师', '服务员', '收银员'];
    const sortedPositions = Object.keys(positionReqs).sort((a, b) => {
        const aIdx = positionOrder.indexOf(a);
        const bIdx = positionOrder.indexOf(b);
        return (aIdx === -1 ? 999 : aIdx) - (bIdx === -1 ? 999 : bIdx);
    });
    
    sortedPositions.forEach(pos => {
        const required = positionReqs[pos] || 0;
        const assigned = positionAssigned[pos] || 0;
        const gap = required - assigned;
        if (gap > 0) {
            gaps.push({ position: pos, gap: gap });
        }
    });
    
    return gaps;
}

/**
 * 获取某日某班次某岗位的需求人数
 * @param {Date} date - 日期
 * @param {string} shiftId - 班次ID
 * @param {string} position - 岗位
 * @param {string} storeId - 门店ID（可选，默认使用当前门店）
 */
function getRequiredCount(date, shiftId, position, storeId = null) {
    // 使用 appState.getRequirementsForDate() 获取实际配置的需求
    const dayReqs = appState.getRequirementsForDate(date, storeId);
    if (!dayReqs || !dayReqs[shiftId]) return 0;
    
    return dayReqs[shiftId][position] || 0;
}

/**
 * 计算满足率 - 支持多门店模式
 */
function calculateSatisfactionRate(weekDates, assignments) {
    let totalRequired = 0;
    let totalAssigned = 0;
    
    // 所有岗位
    const allPositions = ['服务员', '厨师', '收银员'];
    
    // 确定要统计的门店
    const isAllStoresMode = appState.isAllStoresMode();
    const storesToCheck = isAllStoresMode ? appState.getAllStores() : [appState.getCurrentStore()];
    
    weekDates.forEach(date => {
        const dateStr = formatDate(date);
        
        storesToCheck.forEach(store => {
            appState.shifts.forEach(shift => {
                allPositions.forEach(pos => {
                    const required = getRequiredCount(date, shift.id, pos, store.id);
                    if (required > 0) {
                        totalRequired += required;
                        // 统计该日期班次岗位已分配人数（考虑门店过滤）
                        let matchingAssignments;
                        if (isAllStoresMode) {
                            // 多门店模式：按门店匹配
                            matchingAssignments = assignments.filter(a => 
                                a.date === dateStr && 
                                a.shiftId === shift.id && 
                                a.position === pos &&
                                a.storeId === store.id
                            );
                        } else {
                            // 单门店模式
                            matchingAssignments = assignments.filter(a => 
                                a.date === dateStr && a.shiftId === shift.id && a.position === pos
                            );
                        }
                        totalAssigned += Math.min(matchingAssignments.length, required);
                    }
                });
            });
        });
    });
    
    console.log(`📊 满足率计算: 需求 ${totalRequired}, 已分配 ${totalAssigned}, 满足率 ${totalRequired > 0 ? Math.round((totalAssigned / totalRequired) * 100) : 100}%`);
    return totalRequired > 0 ? Math.round((totalAssigned / totalRequired) * 100) : 100;
}

/**
 * 计算指定门店的满足率 - 用于门店筛选器
 * @param {Array} weekDates - 排班日期数组
 * @param {Array} assignments - 排班结果（已按门店筛选）
 * @param {string} storeId - 门店ID ('all' 表示全部)
 */
function calculateSatisfactionRateForStore(weekDates, assignments, storeId) {
    let totalRequired = 0;
    let totalAssigned = 0;
    
    // 所有岗位
    const allPositions = ['服务员', '厨师', '收银员'];
    
    // 确定要统计的门店
    let storesToCheck;
    if (storeId === 'all') {
        // 全部门店模式
        if (appState.isAllStoresMode()) {
            storesToCheck = appState.getAllStores();
        } else {
            storesToCheck = [appState.getCurrentStore()].filter(Boolean);
        }
    } else {
        // 指定单个门店 - 使用 stores.find 而非 getStore
        const store = appState.stores.find(s => s.id === storeId);
        storesToCheck = store ? [store] : [];
    }
    
    weekDates.forEach(date => {
        const dateStr = formatDate(date);
        
        storesToCheck.forEach(store => {
            appState.shifts.forEach(shift => {
                allPositions.forEach(pos => {
                    const required = getRequiredCount(date, shift.id, pos, store.id);
                    if (required > 0) {
                        totalRequired += required;
                        // 统计该日期班次岗位已分配人数（不限制门店，因为跨店员工也算）
                        const matchingAssignments = assignments.filter(a => 
                            a.date === dateStr && 
                            a.shiftId === shift.id && 
                            a.position === pos
                        );
                        totalAssigned += Math.min(matchingAssignments.length, required);
                    }
                });
            });
        });
    });
    
    return totalRequired > 0 ? Math.round((totalAssigned / totalRequired) * 100) : 100;
}

/**
 * 分析人力状况
 * @param {Array} weekDates - 排班日期数组
 * @param {Array} assignments - 排班结果
 * @param {number} satisfactionRate - 满足率
 * @returns {Object} 分析结果
 */
function analyzeStaffStatus(weekDates, assignments, satisfactionRate) {
    // 获取在职员工数
    const activeEmployees = appState.employees.filter(e => e.status === 'active');
    const totalEmployees = activeEmployees.length;
    const waiterCount = activeEmployees.filter(e => e.position === '服务员').length;
    const chefCount = activeEmployees.filter(e => e.position === '厨师').length;
    
    // 计算理论需求工时
    let totalRequiredHours = 0;
    weekDates.forEach(date => {
        appState.shifts.forEach(shift => {
            ['服务员', '厨师'].forEach(pos => {
                const required = getRequiredCount(date, shift.id, pos);
                if (required > 0) {
                    totalRequiredHours += required * shift.hours;
                }
            });
        });
    });
    
    // 计算可用工时（按月度工时计算，约176小时/人/月，周度约44小时/人/周）
    const periodDays = weekDates.length;
    const isMonthly = periodDays > 14;
    const maxHoursPerPerson = isMonthly ? appState.settings.maxPeriodHours || 176 : appState.settings.maxWeeklyHours || 44;
    const weeksInPeriod = Math.ceil(periodDays / 7);
    const availableHoursPerPerson = isMonthly ? maxHoursPerPerson : maxHoursPerPerson * weeksInPeriod;
    const totalAvailableHours = totalEmployees * availableHoursPerPerson;
    
    // 计算实际分配工时
    const actualAssignedHours = assignments.reduce((sum, a) => {
        const shift = appState.shifts.find(s => s.id === a.shiftId);
        return sum + (shift ? shift.hours : 0);
    }, 0);
    
    // 利用率
    const utilizationRate = totalAvailableHours > 0 ? Math.round((actualAssignedHours / totalAvailableHours) * 100) : 0;
    
    // 生成分析消息
    let message = '';
    let status = 'normal'; // normal, surplus, shortage
    
    if (satisfactionRate < 100) {
        // 人力不足
        status = 'shortage';
        const shortageRatio = 100 - satisfactionRate;
        if (shortageRatio > 20) {
            message = `⚠️ 人力严重不足！缺口约 ${shortageRatio}%，建议增加 ${Math.ceil(totalEmployees * shortageRatio / 100)} 名员工`;
        } else {
            message = `⚠️ 人力略有不足，缺口 ${shortageRatio}%，可通过调整班次或增员解决`;
        }
    } else if (utilizationRate < 50) {
        // 人力富裕
        status = 'surplus';
        message = `✅ 人力充裕（利用率 ${utilizationRate}%），${totalEmployees}人可轻松覆盖需求`;
    } else if (utilizationRate < 80) {
        // 人力适中
        status = 'normal';
        message = `✅ 人力适中（利用率 ${utilizationRate}%），排班弹性良好`;
    } else {
        // 人力刚好够用
        status = 'tight';
        message = `⚡ 人力紧张（利用率 ${utilizationRate}%），建议储备后备人员`;
    }
    
    return {
        status,
        message,
        totalEmployees,
        waiterCount,
        chefCount,
        totalRequiredHours,
        totalAvailableHours,
        actualAssignedHours,
        utilizationRate
    };
}

/* ========================================
   弹窗管理
   ======================================== */

function initModals() {
    // 关闭弹窗
    document.querySelectorAll('.modal-close, .modal-overlay').forEach(el => {
        el.addEventListener('click', closeAllModals);
    });
    
    // ESC 关闭
    document.addEventListener('keydown', e => {
        if (e.key === 'Escape') {
            closeAllModals();
        }
    });
    
    // 阻止内容区点击冒泡
    document.querySelectorAll('.modal-content').forEach(content => {
        content.addEventListener('click', e => e.stopPropagation());
    });
    
    // 手动调班相关事件
    initManualAssignEvents();
}

/* ========================================
   手动调班功能
   ======================================== */

let currentAssignmentId = null; // 当前选中的排班ID

function initManualAssignEvents() {
    // 移除排班
    document.getElementById('removeAssignment').addEventListener('click', () => {
        if (currentAssignmentId) {
            removeAssignment(currentAssignmentId);
        }
    });
    
    // 换班按钮
    document.getElementById('swapAssignment').addEventListener('click', () => {
        if (currentAssignmentId) {
            openSwapModal(currentAssignmentId);
        }
    });
    
    // 确认手动添加
    document.getElementById('confirmManualAssign').addEventListener('click', confirmManualAssign);
    
    // 确认换班
    document.getElementById('confirmSwap').addEventListener('click', confirmSwap);
    
    // 员工选择变更时检查冲突
    document.getElementById('manualEmployee').addEventListener('change', checkEmployeeConflict);
}

// 打开手动添加排班弹窗
function openManualAssign(date, shiftId) {
    const shift = appState.getShift(shiftId);
    
    document.getElementById('manualDate').value = date;
    document.getElementById('manualShiftId').value = shiftId;
    document.getElementById('manualDateDisplay').textContent = `${date} ${getDayName(date)}`;
    document.getElementById('manualShiftDisplay').textContent = shift ? `${shift.name} (${shift.startTime}-${shift.endTime})` : shiftId;
    
    // 填充可用员工列表
    const employeeSelect = document.getElementById('manualEmployee');
    employeeSelect.innerHTML = '<option value="">-- 请选择员工 --</option>';
    
    // 获取当天已排班的员工
    const assignedEmployees = appState.assignments
        .filter(a => a.date === date)
        .map(a => a.employeeName);
    
    appState.employees
        .filter(e => e.status === 'active')
        .forEach(emp => {
            const isAssigned = assignedEmployees.includes(emp.name);
            const opt = document.createElement('option');
            opt.value = emp.id;
            opt.textContent = `${formatEmployeeName(emp)} (${emp.position})${isAssigned ? ' ⚠️' : ''}`;
            opt.dataset.position = emp.position;
            employeeSelect.appendChild(opt);
        });
    
    document.getElementById('employeeConflictWarning').style.display = 'none';
    openModal('manualAssignModal');
}

// 检查员工冲突
function checkEmployeeConflict() {
    const empId = document.getElementById('manualEmployee').value;
    const date = document.getElementById('manualDate').value;
    
    if (!empId) {
        document.getElementById('employeeConflictWarning').style.display = 'none';
        return;
    }
    
    const emp = appState.getEmployee(empId);
    if (!emp) return;
    
    // 检查该员工在同一天是否已有排班
    const hasConflict = appState.assignments.some(a => a.date === date && a.employeeName === emp.name);
    
    document.getElementById('employeeConflictWarning').style.display = hasConflict ? 'block' : 'none';
    
    // 自动设置岗位
    document.getElementById('manualPosition').value = emp.position;
}

// 确认手动添加排班
function confirmManualAssign() {
    const empId = document.getElementById('manualEmployee').value;
    const date = document.getElementById('manualDate').value;
    const shiftId = document.getElementById('manualShiftId').value;
    const position = document.getElementById('manualPosition').value;
    
    if (!empId) {
        showToast('请选择员工', 'warning');
        return;
    }
    
    const emp = appState.getEmployee(empId);
    const shift = appState.getShift(shiftId);
    
    if (!emp || !shift) {
        showToast('员工或班次数据错误', 'error');
        return;
    }
    
    // 创建新的排班
    const newAssignment = {
        id: generateUUID(),
        employeeId: empId,
        employeeName: emp.name,
        shiftId: shiftId,
        shiftName: shift.name,
        date: date,
        startTime: shift.startTime,
        endTime: shift.endTime,
        position: position,
        hours: shift.hours,
        score: null, // 手动添加的不计算评分
        isManual: true
    };
    
    appState.assignments.push(newAssignment);
    
    // 记录历史
    appState.addHistoryRecord({
        type: 'add',
        action: '添加排班',
        employeeName: emp.name,
        date: date,
        shiftName: shift.name,
        shiftId: shiftId,
        position: position,
        description: `添加 ${formatEmployeeName(emp)} 到 ${date} ${shift.name}`
    });
    
    closeModal('manualAssignModal');
    renderScheduleGrid();
    renderEmployeeGrid();
    renderShiftHistory(); // 更新历史面板
    showToast(`已添加 ${formatEmployeeName(emp)} 到 ${date} ${shift.name}`, 'success');
}

// 移除排班
function removeAssignment(assignmentId) {
    const assignment = appState.assignments.find(a => a.id === assignmentId);
    if (!assignment) return;
    
    const confirmMsg = `确定要移除 ${formatEmployeeNameByName(assignment.employeeName)} 在 ${assignment.date} ${assignment.shiftName} 的排班吗？`;
    
    if (confirm(confirmMsg)) {
        // 记录历史
        appState.addHistoryRecord({
            type: 'remove',
            action: '移除排班',
            employeeName: assignment.employeeName,
            date: assignment.date,
            shiftName: assignment.shiftName,
            shiftId: assignment.shiftId,
            position: assignment.position,
            description: `移除 ${formatEmployeeNameByName(assignment.employeeName)} 在 ${assignment.date} ${assignment.shiftName} 的排班`
        });
        
        appState.assignments = appState.assignments.filter(a => a.id !== assignmentId);
        closeAllModals();
        renderScheduleGrid();
        renderEmployeeGrid();
        renderShiftHistory(); // 更新历史面板
        showToast(`已移除 ${formatEmployeeNameByName(assignment.employeeName)} 的排班`, 'info');
    }
}

// 打开换班弹窗
function openSwapModal(assignmentId) {
    const assignment = appState.assignments.find(a => a.id === assignmentId);
    if (!assignment) return;
    
    document.getElementById('swapFromId').value = assignmentId;
    
    // 显示原排班信息
    document.getElementById('swapFromCard').innerHTML = `
        <div class="swap-emp-name">${formatEmployeeNameByName(assignment.employeeName)}</div>
        <div class="swap-emp-detail">${assignment.date} ${assignment.shiftName}</div>
        <div class="swap-emp-position">${assignment.position || ''}</div>
    `;
    
    // 填充可交换的员工
    const swapSelect = document.getElementById('swapToEmployee');
    swapSelect.innerHTML = '<option value="">-- 选择要交换的员工 --</option>';
    
    /**
     * 换班条件：
     * 1. 不能是自己
     * 2. 岗位相同（服务员换服务员，厨师换厨师）
     * 3. 排班不同（不同日期 或 不同班次）
     */
    const validSwapOptions = appState.assignments.filter(a => {
        // 条件1：不能是自己
        if (a.id === assignmentId) return false;
        if (a.employeeName === assignment.employeeName) return false;
        
        // 条件2：岗位相同
        if (a.position !== assignment.position) return false;
        
        // 条件3：排班不同（不同日期 或 不同班次）
        const isDifferentSchedule = (a.date !== assignment.date) || (a.shiftId !== assignment.shiftId);
        if (!isDifferentSchedule) return false;
        
        return true;
    });
    
    // 按日期和班次排序
    validSwapOptions.sort((a, b) => {
        if (a.date !== b.date) return a.date.localeCompare(b.date);
        return a.shiftName.localeCompare(b.shiftName);
    });
    
    if (validSwapOptions.length === 0) {
        swapSelect.innerHTML = '<option value="">-- 无可交换的排班（需相同岗位、不同排班）--</option>';
    } else {
        validSwapOptions.forEach(a => {
            const opt = document.createElement('option');
            opt.value = a.id;
            opt.textContent = `${formatEmployeeNameByName(a.employeeName)} (${a.position}) - ${a.date} ${a.shiftName}`;
            swapSelect.appendChild(opt);
        });
    }
    
    closeModal('assignmentModal');
    openModal('swapModal');
}

// 确认换班
function confirmSwap() {
    const fromId = document.getElementById('swapFromId').value;
    const toId = document.getElementById('swapToEmployee').value;
    
    if (!toId) {
        showToast('请选择要交换的排班', 'warning');
        return;
    }
    
    const fromAssignment = appState.assignments.find(a => a.id === fromId);
    const toAssignment = appState.assignments.find(a => a.id === toId);
    
    if (!fromAssignment || !toAssignment) {
        showToast('排班数据错误', 'error');
        return;
    }
    
    // 保存交换前的信息用于历史记录
    const originalFromEmployee = fromAssignment.employeeName;
    const originalToEmployee = toAssignment.employeeName;
    
    // 交换员工信息
    const tempEmpId = fromAssignment.employeeId;
    const tempEmpName = fromAssignment.employeeName;
    const tempPosition = fromAssignment.position;
    
    fromAssignment.employeeId = toAssignment.employeeId;
    fromAssignment.employeeName = toAssignment.employeeName;
    fromAssignment.position = toAssignment.position;
    fromAssignment.isManual = true;
    fromAssignment.score = null;
    
    toAssignment.employeeId = tempEmpId;
    toAssignment.employeeName = tempEmpName;
    toAssignment.position = tempPosition;
    toAssignment.isManual = true;
    toAssignment.score = null;
    
    // 记录历史
    appState.addHistoryRecord({
        type: 'swap',
        action: '换班',
        fromEmployee: originalFromEmployee,
        toEmployee: originalToEmployee,
        fromDate: fromAssignment.date,
        toDate: toAssignment.date,
        fromShift: fromAssignment.shiftName,
        toShift: toAssignment.shiftName,
        description: `${originalFromEmployee} (${fromAssignment.date} ${fromAssignment.shiftName}) ⇄ ${originalToEmployee} (${toAssignment.date} ${toAssignment.shiftName})`
    });
    
    closeAllModals();
    renderScheduleGrid();
    renderEmployeeGrid();
    renderShiftHistory(); // 更新历史面板
    showToast(`已完成 ${formatEmployeeNameByName(fromAssignment.employeeName)} 和 ${formatEmployeeNameByName(toAssignment.employeeName)} 的换班`, 'success');
}

// 显示排班详情（排班表中点击班次卡片时触发）
// isReadOnly: 是否只读模式（发布后查看详情但不能操作）
function showAssignmentDetail(assignmentId, isReadOnly = false) {
    currentAssignmentId = assignmentId;
    const assignment = appState.assignments.find(a => a.id === assignmentId);
    if (!assignment) return;
    
    const detail = document.getElementById('assignmentDetail');
    const shift = appState.getShift(assignment.shiftId);
    
    // 根据只读模式控制操作按钮显示
    const actionButtons = document.querySelectorAll('#assignmentModal .modal-footer button');
    actionButtons.forEach(btn => {
        if (btn.textContent.includes('移除') || btn.textContent.includes('换班')) {
            btn.style.display = isReadOnly ? 'none' : '';
        }
    });
    
    let scoreDetailHtml = '';
    if (assignment.scoreDetail) {
        let detailItems = '';
        Object.entries(assignment.scoreDetail).forEach(([key, val]) => {
            const valNum = typeof val === 'number' ? val : 0;
            detailItems += `<li><span>${getScoreLabel(key)}</span><span>${valNum.toFixed(1)}分</span></li>`;
        });
        scoreDetailHtml = `
            <div class="score-breakdown collapsed">
                <div class="score-breakdown-header" onclick="toggleScoreDetail(this)">
                    <span>📊 评分明细</span>
                    <span class="toggle-icon">▶</span>
                </div>
                <ul class="score-breakdown-list">${detailItems}</ul>
            </div>`;
    }
    
    const manualBadge = assignment.isManual ? '<span class="manual-badge">手动</span>' : '';
    
    detail.innerHTML = `
        <div class="assignment-detail-grid">
            <div class="detail-item">
                <label>员工</label>
                <span>${formatEmployeeNameByName(assignment.employeeName)} ${manualBadge}</span>
            </div>
            <div class="detail-item">
                <label>岗位</label>
                <span>${assignment.position || '未指定'}</span>
            </div>
            <div class="detail-item">
                <label>所属门店</label>
                <span>${assignment.storeName || '未知'}</span>
            </div>
            <div class="detail-item">
                <label>联系电话</label>
                <span>${(() => { const e = appState.employees.find(emp => emp.name === assignment.employeeName); return e?.phone || '未知'; })()}</span>
            </div>
            <div class="detail-item">
                <label>日期</label>
                <span>${assignment.date} ${getDayName(assignment.date)}</span>
            </div>
            <div class="detail-item">
                <label>班次</label>
                <span>${assignment.shiftName || shift?.name || '未知'}</span>
            </div>
            <div class="detail-item">
                <label>时间</label>
                <span>${assignment.startTime} - ${assignment.endTime}</span>
            </div>
            <div class="detail-item">
                <label>工时</label>
                <span>${assignment.hours} 小时</span>
            </div>
            ${assignment.score ? `
            <div class="detail-item">
                <label>综合评分</label>
                <span class="score-badge ${getScoreLevel(assignment.score || 0)}">${Math.round(assignment.score || 0)} 分</span>
            </div>
            ` : ''}
        </div>
        ${scoreDetailHtml}
    `;
    
    openModal('assignmentModal');
}

// 添加样式
const additionalStyles = `
<style>
.assignment-detail-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
}

.detail-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.detail-item.full-width {
    grid-column: span 2;
}

.detail-item label {
    font-size: 12px;
    color: var(--text-muted);
}

.detail-item span {
    font-size: 15px;
    font-weight: 500;
}

.score-badge {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 20px;
    font-size: 14px;
    font-weight: 600;
}

.score-badge.high {
    background: rgba(16, 185, 129, 0.2);
    color: #10b981;
}

.score-badge.medium {
    background: rgba(245, 158, 11, 0.2);
    color: #f59e0b;
}

.score-badge.low {
    background: rgba(239, 68, 68, 0.2);
    color: #ef4444;
}

.score-breakdown {
    margin-top: 20px;
    padding-top: 16px;
    border-top: 1px solid rgba(255,255,255,0.1);
}

.score-breakdown h4 {
    font-size: 14px;
    margin-bottom: 12px;
    color: var(--text-secondary);
}

.score-breakdown ul {
    list-style: none;
}

.score-breakdown li {
    display: flex;
    justify-content: space-between;
    padding: 8px 0;
    border-bottom: 1px solid rgba(255,255,255,0.05);
}

.empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 200px;
    color: var(--text-muted);
    font-size: 16px;
}

.req-section {
    grid-column: 1 / -1;
    padding: 12px 0 8px;
}

.req-section h4 {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-secondary);
}

.unfilled-reason {
    font-size: 11px;
    color: var(--warning);
    margin-top: 4px;
}

/* 添加排班按钮 */
.add-assignment-btn {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: rgba(255,255,255,0.1);
    border: 1px dashed rgba(255,255,255,0.3);
    color: var(--text-muted);
    font-size: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s;
    margin: 4px auto 0;
}

.add-assignment-btn:hover {
    background: var(--primary);
    border-color: var(--primary);
    color: white;
    transform: scale(1.1);
}

.requirement-indicator.unfilled {
    cursor: pointer;
}

.requirement-indicator.unfilled:hover {
    background: rgba(232, 90, 79, 0.2);
    border-color: var(--primary);
}

/* 手动排班弹窗 */
.form-value {
    font-size: 15px;
    font-weight: 500;
    color: var(--text-primary);
}

.warning-box {
    background: rgba(245, 158, 11, 0.15);
    border: 1px solid rgba(245, 158, 11, 0.4);
    border-radius: var(--radius-sm);
    padding: 10px 14px;
    color: var(--warning);
    font-size: 13px;
    margin-top: 12px;
}

/* 换班弹窗 */
.swap-info {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px 0;
}

.swap-card {
    flex: 1;
    background: var(--bg-tertiary);
    border-radius: var(--radius-md);
    padding: 16px;
    text-align: center;
}

.swap-arrow {
    font-size: 24px;
    color: var(--secondary);
}

.swap-emp-name {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 6px;
}

.swap-emp-detail {
    font-size: 13px;
    color: var(--text-secondary);
}

.swap-emp-position {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 4px;
}

.swap-select {
    width: 100%;
    padding: 10px;
    background: var(--bg-secondary);
    border: 1px solid rgba(255,255,255,0.1);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-size: 14px;
}

/* 手动标记 */
.manual-badge {
    display: inline-block;
    padding: 2px 8px;
    background: var(--secondary);
    color: #1a1a2e;
    font-size: 10px;
    font-weight: 600;
    border-radius: 10px;
    margin-left: 6px;
    vertical-align: middle;
}
</style>
`;

document.head.insertAdjacentHTML('beforeend', additionalStyles);

/* ========================================
   换班历史功能
   ======================================== */

function renderShiftHistory() {
    const container = document.getElementById('historyList');
    if (!container) return;
    
    const history = appState.shiftHistory;
    
    if (history.length === 0) {
        container.innerHTML = '<div class="history-empty">暂无操作记录</div>';
        document.getElementById('historyCount').textContent = '0';
        return;
    }
    
    document.getElementById('historyCount').textContent = history.length;
    
    let html = '';
    history.forEach(record => {
        const time = formatHistoryTime(record.timestamp);
        const typeIcon = getHistoryTypeIcon(record.type);
        const typeClass = record.type;
        
        html += `
            <div class="history-item ${typeClass}">
                <div class="history-icon">${typeIcon}</div>
                <div class="history-content">
                    <div class="history-action">${record.action}</div>
                    <div class="history-desc">${record.description}</div>
                    <div class="history-time">${time}</div>
                </div>
            </div>
        `;
    });
    
    container.innerHTML = html;
}

function getHistoryTypeIcon(type) {
    switch (type) {
        case 'add': return '➕';
        case 'remove': return '🗑️';
        case 'swap': return '🔄';
        default: return '📝';
    }
}

function formatHistoryTime(timestamp) {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now - date;
    
    // 小于1分钟
    if (diff < 60000) {
        return '刚刚';
    }
    // 小于1小时
    if (diff < 3600000) {
        return `${Math.floor(diff / 60000)} 分钟前`;
    }
    // 小于24小时
    if (diff < 86400000) {
        return `${Math.floor(diff / 3600000)} 小时前`;
    }
    // 小于7天
    if (diff < 604800000) {
        return `${Math.floor(diff / 86400000)} 天前`;
    }
    // 超过7天显示日期
    return `${date.getMonth() + 1}月${date.getDate()}日 ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
}

function toggleHistoryPanel() {
    const panel = document.getElementById('historyPanel');
    const overlay = document.getElementById('historyOverlay');
    const isVisible = panel.classList.contains('active');
    
    if (isVisible) {
        panel.classList.remove('active');
        overlay.classList.remove('active');
    } else {
        renderShiftHistory();
        panel.classList.add('active');
        overlay.classList.add('active');
    }
}

function clearShiftHistory() {
    if (confirm('确定要清空所有操作历史吗？')) {
        appState.clearHistory();
        renderShiftHistory();
        showToast('历史记录已清空', 'info');
    }
}

// 历史面板样式
const historyStyles = `
<style>
/* 历史记录按钮 */
.history-toggle-btn {
    position: relative;
    padding: 8px 16px;
    background: var(--bg-tertiary);
    border: 1px solid rgba(255,255,255,0.1);
    border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
    color: var(--text-secondary);
    font-size: 13px;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    gap: 6px;
}

.history-toggle-btn:hover {
    background: var(--bg-secondary);
    color: var(--text-primary);
    border-color: var(--secondary);
}

.history-badge {
    background: var(--primary);
    color: white;
    font-size: 11px;
    padding: 2px 6px;
    border-radius: 10px;
    font-weight: 600;
}

/* 历史面板 */
.history-panel {
    position: fixed;
    right: -360px;
    top: 0;
    width: 360px;
    height: 100vh;
    background: var(--bg-secondary);
    border-left: 1px solid rgba(255,255,255,0.1);
    z-index: 1000;
    transition: right 0.3s ease;
    display: flex;
    flex-direction: column;
    box-shadow: -5px 0 20px rgba(0,0,0,0.3);
}

.history-panel.active {
    right: 0;
}

.history-header {
    padding: 20px;
    border-bottom: 1px solid rgba(255,255,255,0.1);
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: var(--bg-tertiary);
}

.history-header h3 {
    font-size: 16px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 8px;
}

.history-actions {
    display: flex;
    gap: 8px;
}

.history-close-btn,
.history-clear-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px 8px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    transition: all 0.2s;
}

.history-close-btn:hover {
    color: var(--text-primary);
    background: rgba(255,255,255,0.1);
}

.history-clear-btn:hover {
    color: var(--danger);
    background: rgba(239, 68, 68, 0.1);
}

.history-list {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
}

.history-empty {
    text-align: center;
    color: var(--text-muted);
    padding: 40px 20px;
    font-size: 14px;
}

.history-item {
    display: flex;
    gap: 12px;
    padding: 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-md);
    margin-bottom: 8px;
    border-left: 3px solid var(--text-muted);
    transition: all 0.2s;
}

.history-item:hover {
    background: rgba(255,255,255,0.05);
}

.history-item.add {
    border-left-color: var(--success);
}

.history-item.remove {
    border-left-color: var(--danger);
}

.history-item.swap {
    border-left-color: var(--secondary);
}

.history-icon {
    font-size: 18px;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255,255,255,0.05);
    border-radius: 50%;
    flex-shrink: 0;
}

.history-content {
    flex: 1;
    min-width: 0;
}

.history-action {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 4px;
}

.history-desc {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.4;
    word-break: break-all;
}

.history-time {
    font-size: 11px;
    color: var(--text-muted);
    margin-top: 6px;
}

/* 历史面板遮罩 */
.history-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0,0,0,0.5);
    z-index: 999;
    opacity: 0;
    visibility: hidden;
    transition: all 0.3s;
}

.history-overlay.active {
    opacity: 1;
    visibility: visible;
}
</style>
`;

document.head.insertAdjacentHTML('beforeend', historyStyles);

// ========================================
// 排班报告面板
// ========================================

function toggleReportPanel() {
    const panel = document.getElementById('reportPanel');
    const overlay = document.getElementById('reportOverlay');
    const isVisible = panel.classList.contains('active');
    
    if (isVisible) {
        panel.classList.remove('active');
        overlay.classList.remove('active');
    } else {
        renderReportPanel();
        panel.classList.add('active');
        overlay.classList.add('active');
    }
}

function renderReportPanel() {
    const container = document.getElementById('reportContent');
    if (!container) return;
    
    const assignments = appState.assignments || [];
    const violations = appState.constraintViolations || [];
    const unfilledRequirements = appState.unfilledRequirements || [];
    
    if (assignments.length === 0) {
        container.innerHTML = '<div class="report-empty">暂无排班数据，请先生成排班</div>';
        return;
    }
    
    // 使用 analyzeSchedulingProblem 获取完整分析报告
    const analysis = analyzeSchedulingProblem();
    
    // 计算统计数据
    const totalShifts = assignments.length;
    const hardViolations = violations.filter(v => v.type === 'hard');
    const softViolations = violations.filter(v => v.type === 'soft');
    
    // 计算满足率
    const weekDates = appState.getWeekDates();
    const isAllMode = appState.isAllStoresMode();
    let satisfactionRate = calculateSatisfactionRateForStore(weekDates, assignments, isAllMode ? 'all' : appState.currentStoreId);
    if (isNaN(satisfactionRate)) satisfactionRate = 100;
    
    let html = '';
    
    // 概览卡片
    html += `
        <div class="report-section">
            <div class="report-section-title">📈 排班概览</div>
            <div class="report-overview-cards">
                <div class="report-card">
                    <div class="report-card-value">${totalShifts}</div>
                    <div class="report-card-label">总班次</div>
                </div>
                <div class="report-card ${satisfactionRate < 100 ? 'warning' : 'success'}">
                    <div class="report-card-value">${satisfactionRate}%</div>
                    <div class="report-card-label">满足率</div>
                </div>
                <div class="report-card ${hardViolations.length > 0 ? 'error' : 'success'}">
                    <div class="report-card-value">${hardViolations.length}</div>
                    <div class="report-card-label">硬约束违规</div>
                </div>
                <div class="report-card ${softViolations.length > 0 ? 'warning' : 'success'}">
                    <div class="report-card-value">${softViolations.length}</div>
                    <div class="report-card-label">软约束违规</div>
                </div>
            </div>
        </div>
    `;
    
    // 问题分析（来自 analyzeSchedulingProblem）
    if (analysis.problems && analysis.problems.length > 0) {
        // 按类别分组问题
        const summaryItems = analysis.problems.filter(p => p.category === 'summary');
        const violationItems = analysis.problems.filter(p => p.category === 'violation');
        const gapItems = analysis.problems.filter(p => p.category === 'gap');
        const storeItems = analysis.problems.filter(p => p.category === 'store');
        
        // 人力概览
        if (summaryItems.length > 0) {
            html += `<div class="report-section"><div class="report-section-title">📊 人力概览</div><div class="report-analysis-list">`;
            summaryItems.forEach(p => {
                html += `<div class="report-analysis-item ${p.severity}"><span class="report-analysis-icon">${p.icon}</span><span class="report-analysis-text">${p.message}</span></div>`;
            });
            html += '</div></div>';
        }
        
        // 门店详情（岗位缺口已合并到这里）
        if (storeItems.length > 0) {
            html += `<div class="report-section"><div class="report-section-title">🏪 门店详情 (${storeItems.length})</div><div class="report-analysis-list">`;
            storeItems.forEach(p => {
                html += `<div class="report-analysis-item ${p.severity}"><span class="report-analysis-icon">${p.icon}</span><span class="report-analysis-text">${p.message}</span></div>`;
            });
            html += '</div></div>';
        }
    }
    
    // 硬约束违规明细
    if (hardViolations.length > 0) {
        html += `
            <div class="report-section">
                <div class="report-section-title error">⛔ 硬约束违规明细 (${hardViolations.length})</div>
                <div class="report-violation-list">
        `;
        hardViolations.forEach((v, index) => {
            const constraintName = v.constraintName || v.constraintType || '约束';
            html += `
                <div class="report-violation-item error">
                    <div class="report-violation-index">${index + 1}</div>
                    <div class="report-violation-content">
                        <div class="report-violation-type">${constraintName}</div>
                        <div class="report-violation-message">${v.message}</div>
                    </div>
                </div>
            `;
        });
        html += '</div></div>';
    }
    
    // 软约束违规明细
    if (softViolations.length > 0) {
        html += `
            <div class="report-section">
                <div class="report-section-title warning">⚠️ 软约束违规明细 (${softViolations.length})</div>
                <div class="report-violation-list">
        `;
        softViolations.forEach((v, index) => {
            const constraintName = v.constraintName || v.constraintType || '约束';
            html += `
                <div class="report-violation-item warning">
                    <div class="report-violation-index">${index + 1}</div>
                    <div class="report-violation-content">
                        <div class="report-violation-type">${constraintName}</div>
                        <div class="report-violation-message">${v.message}</div>
                    </div>
                </div>
            `;
        });
        html += '</div></div>';
    }
    
    // 未满足需求明细 - 按日期和门店分组显示
    if (unfilledRequirements.length > 0) {
        // 按日期+门店+班次分组
        const grouped = {};
        unfilledRequirements.forEach(req => {
            const key = `${req.date}-${req.storeName || ''}-${req.shiftName || ''}`;
            if (!grouped[key]) {
                grouped[key] = {
                    date: req.date,
                    storeName: req.storeName || '',
                    shiftName: req.shiftName || '',
                    positions: []
                };
            }
            grouped[key].positions.push({
                position: req.position,
                required: req.required,
                assigned: req.assigned,
                shortage: req.required - req.assigned
            });
        });
        
        // 按日期排序
        const sortedGroups = Object.values(grouped).sort((a, b) => {
            if (a.date !== b.date) return a.date.localeCompare(b.date);
            if (a.storeName !== b.storeName) return a.storeName.localeCompare(b.storeName);
            return a.shiftName.localeCompare(b.shiftName);
        });
        
        html += `
            <div class="report-section">
                <div class="report-section-title warning">📋 未满足需求明细 (${sortedGroups.length}条)</div>
                <div class="report-unfilled-list">
        `;
        sortedGroups.forEach((group, index) => {
            // 生成岗位缺口摘要
            const positionSummary = group.positions
                .map(p => `${p.position}缺${p.shortage}人`)
                .join('、');
            
            html += `
                <div class="report-unfilled-item">
                    <div class="report-unfilled-index">${index + 1}</div>
                    <div class="report-unfilled-content">
                        <div class="report-unfilled-info">
                            <span class="report-unfilled-day">${group.date}</span>
                            ${group.storeName ? `<span class="report-unfilled-store">${group.storeName}</span>` : ''}
                            <span class="report-unfilled-shift">${group.shiftName}</span>
                        </div>
                        <div class="report-unfilled-gap">${positionSummary}</div>
                    </div>
                </div>
            `;
        });
        html += '</div></div>';
    }
    
    // 解决方案（来自 analyzeSchedulingProblem）
    if (analysis.solutions && analysis.solutions.length > 0) {
        const primarySolutions = analysis.solutions.filter(s => s.type === 'primary');
        const otherSolutions = analysis.solutions.filter(s => s.type !== 'primary');
        
        if (primarySolutions.length > 0) {
            html += `<div class="report-section"><div class="report-section-title">🎯 推荐操作</div><div class="report-solution-list">`;
            primarySolutions.forEach(s => {
                html += `<div class="report-solution-item primary"><span class="report-solution-icon">${s.icon}</span><span class="report-solution-text">${s.message}</span></div>`;
            });
            html += '</div></div>';
        }
        
        if (otherSolutions.length > 0) {
            html += `<div class="report-section"><div class="report-section-title">💡 其他建议</div><div class="report-solution-list">`;
            otherSolutions.forEach(s => {
                html += `<div class="report-solution-item ${s.type}"><span class="report-solution-icon">${s.icon}</span><span class="report-solution-text">${s.message}</span></div>`;
            });
            html += '</div></div>';
        }
    }
    
    // 员工工作量统计 - 支持多种ID格式匹配
    const employeeShiftCounts = {};
    assignments.forEach(a => {
        const empId = a.employeeId || a.employee_id || a.empId;
        if (empId) {
            // 只计数一次，使用字符串形式的ID
            const idKey = String(empId);
            employeeShiftCounts[idKey] = (employeeShiftCounts[idKey] || 0) + 1;
        }
    });
    
    const employees = appState.employees.filter(e => e.status !== 'inactive');
    const sortedEmployees = employees
        .map(emp => ({
            ...emp,
            shiftCount: employeeShiftCounts[emp.id] || employeeShiftCounts[String(emp.id)] || 0
        }))
        .filter(emp => emp.shiftCount > 0)
        .sort((a, b) => b.shiftCount - a.shiftCount);
    
    if (sortedEmployees.length > 0) {
        html += `
            <div class="report-section">
                <div class="report-section-title">📊 员工工作量</div>
                <div class="report-employee-list">
        `;
        
        sortedEmployees.forEach(emp => {
            const maxShifts = appState.settings?.maxShiftsPerWeek || 6;
            const percentage = Math.min(100, Math.round((emp.shiftCount / maxShifts) * 100));
            const statusClass = percentage > 100 ? 'error' : percentage > 80 ? 'warning' : '';
            
            html += `
                <div class="report-employee-item">
                    <div class="report-employee-info">
                        <span class="report-employee-name">${emp.name}</span>
                        <span class="report-employee-position">${emp.position}</span>
                    </div>
                    <div class="report-employee-stats">
                        <div class="report-employee-bar">
                            <div class="report-employee-bar-fill ${statusClass}" style="width: ${percentage}%"></div>
                        </div>
                        <span class="report-employee-count">${emp.shiftCount} 班</span>
                    </div>
                </div>
            `;
        });
        html += '</div></div>';
    }
    
    // 补员建议（来自后端算法分析）
    const staffingSuggestions = appState.staffingSuggestions || [];
    if (staffingSuggestions.length > 0) {
        html += `
            <div class="report-section">
                <div class="report-section-title">👥 补员建议</div>
                <div class="report-staffing-list">
        `;
        staffingSuggestions.forEach(s => {
            const icon = s.type === 'shortage' ? '📢' : s.type === 'overwork' ? '⚠️' : '💡';
            html += `
                <div class="report-staffing-item ${s.type}">
                    <span class="report-staffing-icon">${icon}</span>
                    <div class="report-staffing-content">
                        ${s.position ? `<span class="report-staffing-position">${s.position}</span>` : ''}
                        <span class="report-staffing-reason">${s.reason}</span>
                        ${s.current_num !== undefined ? `<span class="report-staffing-nums">当前: ${s.current_num}人 → 建议: ${s.suggest_num}人</span>` : ''}
                    </div>
                </div>
            `;
        });
        html += '</div></div>';
    }
    
    // 如果没有任何问题，显示成功信息
    if (hardViolations.length === 0 && softViolations.length === 0 && unfilledRequirements.length === 0 && satisfactionRate >= 100) {
        html += `
            <div class="report-section">
                <div class="report-success">
                    <div class="report-success-icon">✅</div>
                    <div class="report-success-text">排班完美！没有任何违规或未满足的需求</div>
                </div>
            </div>
        `;
    }
    
    // AI 建议区块
    html += `
        <div class="report-section">
            <div class="report-section-title">🤖 AI 建议</div>
            <div class="report-ai-container">
                <button class="report-ai-btn" onclick="generateAIAdvice()">
                    <span class="ai-btn-icon">✨</span>
                    <span class="ai-btn-text">获取 AI 建议</span>
                </button>
                <div class="report-ai-content" id="aiAdviceContent"></div>
            </div>
        </div>
    `;
    
    container.innerHTML = html;
}

// 豆包 AI 配置
const DOUBAO_API_CONFIG = {
    url: 'https://ark.cn-beijing.volces.com/api/v3/chat/completions',
    apiKey: '9fd8383f-5776-4366-855d-c6f40e867940',
    model: 'doubao-seed-1-6-251015'
};

// 生成 AI 建议
async function generateAIAdvice() {
    const container = document.getElementById('aiAdviceContent');
    if (!container) return;
    
    // 显示加载状态
    container.innerHTML = '<div class="ai-loading"><span class="ai-loading-icon">⏳</span> AI 正在分析排班数据...</div>';
    
    try {
        // 准备排班数据摘要
        const reportData = prepareReportDataForAI();
        
        // 调用豆包 AI API
        const aiResponse = await callDoubaoAI(reportData);
        
        // 渲染 AI 建议
        renderDoubaoAIAdvice(container, aiResponse);
    } catch (error) {
        console.error('AI 建议生成失败:', error);
        // 失败时使用本地规则分析
        container.innerHTML = '<div class="ai-error">⚠️ AI 服务暂时不可用，使用本地分析...</div>';
        setTimeout(() => {
            const advice = analyzeAndGenerateAdvice();
            renderAIAdvice(container, advice);
        }, 500);
    }
}

// 准备发送给 AI 的排班数据摘要
function prepareReportDataForAI() {
    const assignments = appState.assignments || [];
    const violations = appState.constraintViolations || [];
    const unfilledRequirements = appState.unfilledRequirements || [];
    const employees = appState.employees.filter(e => e.status !== 'inactive');
    
    const hardViolations = violations.filter(v => v.type === 'hard');
    const softViolations = violations.filter(v => v.type === 'soft');
    
    // 计算员工工作量 - 支持多种ID格式匹配
    const employeeShiftCounts = {};
    assignments.forEach(a => {
        // 尝试多种ID格式
        const empId = a.employeeId || a.employee_id || a.empId;
        if (empId) {
            employeeShiftCounts[empId] = (employeeShiftCounts[empId] || 0) + 1;
            // 同时用字符串格式存储
            employeeShiftCounts[String(empId)] = (employeeShiftCounts[String(empId)] || 0) + 1;
        }
    });
    
    // 员工工作量统计 - 尝试多种ID格式匹配
    const employeeWorkload = employees.map(e => {
        const shifts = employeeShiftCounts[e.id] || employeeShiftCounts[String(e.id)] || 0;
        return {
            name: e.name,
            position: e.position,
            shifts: shifts
        };
    }).sort((a, b) => b.shifts - a.shifts);
    
    // 计算满足率
    const weekDates = appState.getWeekDates();
    const isAllMode = appState.isAllStoresMode();
    let satisfactionRate = calculateSatisfactionRateForStore(weekDates, assignments, isAllMode ? 'all' : appState.currentStoreId);
    if (isNaN(satisfactionRate)) satisfactionRate = 100;
    
    // 诊断信息
    const activeEmployees = employeeWorkload.filter(e => e.shifts > 0);
    const idleEmployees = employeeWorkload.filter(e => e.shifts === 0);
    const shiftsArray = activeEmployees.map(e => e.shifts);
    const maxShifts = shiftsArray.length > 0 ? Math.max(...shiftsArray) : 0;
    const minShifts = shiftsArray.length > 0 ? Math.min(...shiftsArray) : 0;
    
    // 检查岗位不匹配情况
    const positionCounts = {};
    employees.forEach(e => {
        positionCounts[e.position] = (positionCounts[e.position] || 0) + 1;
    });
    const idlePositions = {};
    idleEmployees.forEach(e => {
        idlePositions[e.position] = (idlePositions[e.position] || 0) + 1;
    });
    let positionMismatch = '';
    for (const [pos, count] of Object.entries(idlePositions)) {
        if (count > 0) {
            positionMismatch += `${pos}有${count}人闲置; `;
        }
    }
    
    return {
        summary: {
            totalShifts: assignments.length,
            satisfactionRate: satisfactionRate,
            totalEmployees: employees.length,
            hardViolationsCount: hardViolations.length,
            softViolationsCount: softViolations.length,
            unfilledCount: unfilledRequirements.length
        },
        hardViolations: hardViolations.slice(0, 10).map(v => ({
            type: v.constraintType || v.constraintName || '约束',
            message: v.message
        })),
        softViolations: softViolations.slice(0, 10).map(v => ({
            type: v.constraintType || v.constraintName || '约束',
            message: v.message
        })),
        unfilledRequirements: unfilledRequirements.slice(0, 10).map(r => ({
            date: r.date,
            shift: r.shiftName,
            position: r.position,
            required: r.required,
            assigned: r.assigned
        })),
        employeeWorkload: employeeWorkload.slice(0, 15),
        diagnostics: {
            activeCount: activeEmployees.length,
            idleCount: idleEmployees.length,
            maxShifts: maxShifts,
            minShifts: minShifts,
            positionMismatch: positionMismatch.trim()
        }
    };
}

// 调用豆包 AI API
async function callDoubaoAI(reportData) {
    const prompt = `你是一个专业的餐饮行业排班顾问。请根据以下排班报告数据，给出专业、具体、可操作的建议。

## 排班报告数据

### 概览
- 总班次: ${reportData.summary.totalShifts}
- 满足率: ${reportData.summary.satisfactionRate}%
- 在职员工数: ${reportData.summary.totalEmployees}
- 硬约束违规: ${reportData.summary.hardViolationsCount} 条
- 软约束违规: ${reportData.summary.softViolationsCount} 条
- 未满足需求: ${reportData.summary.unfilledCount} 条

### 硬约束违规明细
${reportData.hardViolations.length > 0 ? reportData.hardViolations.map(v => `- [${v.type}] ${v.message}`).join('\n') : '无'}

### 软约束违规明细
${reportData.softViolations.length > 0 ? reportData.softViolations.map(v => `- [${v.type}] ${v.message}`).join('\n') : '无'}

### 未满足需求
${reportData.unfilledRequirements.length > 0 ? reportData.unfilledRequirements.map(r => `- ${r.date} ${r.shift} ${r.position}: 需${r.required}人，已排${r.assigned}人`).join('\n') : '无'}

### 员工工作量（按班次排序）
${reportData.employeeWorkload.map(e => `- ${e.name}(${e.position}): ${e.shifts}班`).join('\n')}

### 工作量分布诊断
- 有班次的员工: ${reportData.diagnostics.activeCount}人
- 无班次的员工: ${reportData.diagnostics.idleCount}人
- 最高班次: ${reportData.diagnostics.maxShifts}班
- 最低班次(有排班): ${reportData.diagnostics.minShifts}班
- 班次差异: ${reportData.diagnostics.maxShifts - reportData.diagnostics.minShifts}班
${reportData.diagnostics.positionMismatch ? `- 岗位不匹配: ${reportData.diagnostics.positionMismatch}` : ''}

## 请给出建议

请从以下几个方面给出建议（使用markdown格式）：
1. **总体评价**：对当前排班状况的整体评估
2. **问题原因分析**：分析为什么会出现工作量分配不均（如"忙的忙死，闲的闲死"）的情况，可能的原因包括：员工技能/岗位不匹配、可用时间设置问题、约束冲突等
3. **紧急问题**：需要立即处理的问题（如有）
4. **优化建议**：如何改善当前排班，包括短期和长期措施
5. **人员配置建议**：是否需要调整人员配置或培训

请保持建议简洁实用，重点分析问题根因。`;

    const response = await fetch(DOUBAO_API_CONFIG.url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${DOUBAO_API_CONFIG.apiKey}`
        },
        body: JSON.stringify({
            model: DOUBAO_API_CONFIG.model,
            max_completion_tokens: 2000,
            messages: [
                {
                    role: 'user',
                    content: prompt
                }
            ]
        })
    });
    
    if (!response.ok) {
        throw new Error(`API 请求失败: ${response.status}`);
    }
    
    const data = await response.json();
    return data.choices?.[0]?.message?.content || '无法获取 AI 建议';
}

// 渲染豆包 AI 建议
function renderDoubaoAIAdvice(container, content) {
    // 简单的 markdown 转 HTML
    let html = content
        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
        .replace(/\n\n/g, '</p><p>')
        .replace(/\n- /g, '</li><li>')
        .replace(/\n(\d+)\. /g, '</li><li>')
        .replace(/^- /, '<li>')
        .replace(/^(\d+)\. /, '<li>');
    
    // 处理列表
    if (html.includes('<li>')) {
        html = html.replace(/<li>/g, '<li class="ai-list-item">');
        html = '<ul class="ai-list">' + html + '</li></ul>';
    }
    
    html = '<div class="ai-doubao-response"><p>' + html + '</p></div>';
    
    container.innerHTML = html;
}

function analyzeAndGenerateAdvice() {
    const assignments = appState.assignments || [];
    const violations = appState.constraintViolations || [];
    const unfilledRequirements = appState.unfilledRequirements || [];
    const employees = appState.employees.filter(e => e.status !== 'inactive');
    
    const hardViolations = violations.filter(v => v.type === 'hard');
    const softViolations = violations.filter(v => v.type === 'soft');
    
    // 计算员工工作量 - 支持多种ID格式匹配
    const employeeShiftCounts = {};
    assignments.forEach(a => {
        const empId = a.employeeId || a.employee_id || a.empId;
        if (empId) {
            employeeShiftCounts[empId] = (employeeShiftCounts[empId] || 0) + 1;
            employeeShiftCounts[String(empId)] = (employeeShiftCounts[String(empId)] || 0) + 1;
        }
    });
    
    // 分析工作量分布
    const shiftCounts = Object.values(employeeShiftCounts);
    const maxShifts = Math.max(...shiftCounts, 0);
    const minShifts = Math.min(...shiftCounts.filter(c => c > 0), 0);
    const avgShifts = shiftCounts.length > 0 ? (shiftCounts.reduce((a, b) => a + b, 0) / shiftCounts.length).toFixed(1) : 0;
    
    // 找出工作量过高和过低的员工
    const getShiftCount = (emp) => employeeShiftCounts[emp.id] || employeeShiftCounts[String(emp.id)] || 0;
    const overloadedEmployees = employees.filter(e => getShiftCount(e) > 5);
    const underutilizedEmployees = employees.filter(e => getShiftCount(e) < 2 && getShiftCount(e) > 0);
    const idleEmployees = employees.filter(e => getShiftCount(e) === 0);
    
    // 按岗位分析
    const positionStats = {};
    employees.forEach(e => {
        if (!positionStats[e.position]) {
            positionStats[e.position] = { total: 0, assigned: 0, shifts: 0 };
        }
        positionStats[e.position].total++;
        if (getShiftCount(e) > 0) {
            positionStats[e.position].assigned++;
            positionStats[e.position].shifts += getShiftCount(e);
        }
    });
    
    // 生成建议
    const advice = {
        summary: '',
        suggestions: [],
        warnings: [],
        optimizations: []
    };
    
    // 总体评估
    if (hardViolations.length === 0 && unfilledRequirements.length === 0) {
        advice.summary = '当前排班整体良好，所有硬约束均已满足。';
    } else if (hardViolations.length > 0) {
        advice.summary = `当前排班存在 ${hardViolations.length} 个硬约束违规，需要优先处理。`;
    } else {
        advice.summary = `当前排班有 ${unfilledRequirements.length} 个需求未满足，建议调整人员配置。`;
    }
    
    // 硬约束违规建议
    if (hardViolations.length > 0) {
        const consecutiveViolations = hardViolations.filter(v => v.constraintType === 'max_consecutive_days');
        const hoursViolations = hardViolations.filter(v => v.constraintType?.includes('hours'));
        
        if (consecutiveViolations.length > 0) {
            advice.warnings.push({
                icon: '⚠️',
                title: '连续工作天数超限',
                content: `有 ${consecutiveViolations.length} 名员工连续工作天数超过限制。建议：增加相应岗位人手，或调整排班让员工有休息日。`
            });
        }
        
        if (hoursViolations.length > 0) {
            advice.warnings.push({
                icon: '⏰',
                title: '工时超限',
                content: `有员工周工时超过限制。建议：合理分配工作量，避免单一员工承担过多班次。`
            });
        }
    }
    
    // 工作量分布建议
    if (maxShifts - minShifts > 3 && shiftCounts.length > 1) {
        advice.optimizations.push({
            icon: '⚖️',
            title: '工作量分布不均',
            content: `员工班次差异较大（最多 ${maxShifts} 班，最少 ${minShifts} 班）。建议：重新平衡工作量，让每位员工的班次更均匀。`
        });
    }
    
    if (overloadedEmployees.length > 0) {
        const names = overloadedEmployees.slice(0, 3).map(e => e.name).join('、');
        advice.optimizations.push({
            icon: '😓',
            title: '部分员工工作量过大',
            content: `${names}${overloadedEmployees.length > 3 ? '等' : ''} 班次较多，可能影响工作质量和员工满意度。建议适当减少其班次。`
        });
    }
    
    if (idleEmployees.length > 0) {
        const names = idleEmployees.slice(0, 3).map(e => e.name).join('、');
        advice.suggestions.push({
            icon: '💤',
            title: '有员工未被排班',
            content: `${names}${idleEmployees.length > 3 ? '等 ' + idleEmployees.length + ' 人' : ''} 本周未安排任何班次。如非休假，建议合理安排其工作。`
        });
    }
    
    // 岗位配置建议
    for (const [position, stats] of Object.entries(positionStats)) {
        if (stats.assigned < stats.total * 0.5) {
            advice.suggestions.push({
                icon: '👥',
                title: `${position}岗位利用率低`,
                content: `${position}共 ${stats.total} 人，仅 ${stats.assigned} 人被排班。建议检查是否有员工可用性问题或需求配置是否合理。`
            });
        }
    }
    
    // 未满足需求建议
    if (unfilledRequirements.length > 0) {
        const positionGaps = {};
        unfilledRequirements.forEach(req => {
            const pos = req.position || '未知';
            positionGaps[pos] = (positionGaps[pos] || 0) + (req.required - req.assigned);
        });
        
        for (const [pos, gap] of Object.entries(positionGaps)) {
            advice.suggestions.push({
                icon: '📋',
                title: `${pos}人手不足`,
                content: `${pos}岗位共缺 ${gap} 个班次。建议：增加${pos}人员，或调整现有员工的可用时间。`
            });
        }
    }
    
    // 如果一切正常，给出正面建议
    if (advice.warnings.length === 0 && advice.suggestions.length === 0 && advice.optimizations.length === 0) {
        advice.suggestions.push({
            icon: '🎉',
            title: '排班状态良好',
            content: '当前排班配置合理，工作量分布均匀，没有发现明显问题。建议保持现有配置。'
        });
    }
    
    return advice;
}

function renderAIAdvice(container, advice) {
    let html = '';
    
    // 总结
    html += `<div class="ai-summary">${advice.summary}</div>`;
    
    // 警告（优先显示）
    if (advice.warnings.length > 0) {
        html += '<div class="ai-advice-group warnings">';
        advice.warnings.forEach(w => {
            html += `
                <div class="ai-advice-item warning">
                    <div class="ai-advice-header">
                        <span class="ai-advice-icon">${w.icon}</span>
                        <span class="ai-advice-title">${w.title}</span>
                    </div>
                    <div class="ai-advice-content">${w.content}</div>
                </div>
            `;
        });
        html += '</div>';
    }
    
    // 优化建议
    if (advice.optimizations.length > 0) {
        html += '<div class="ai-advice-group optimizations">';
        advice.optimizations.forEach(o => {
            html += `
                <div class="ai-advice-item optimization">
                    <div class="ai-advice-header">
                        <span class="ai-advice-icon">${o.icon}</span>
                        <span class="ai-advice-title">${o.title}</span>
                    </div>
                    <div class="ai-advice-content">${o.content}</div>
                </div>
            `;
        });
        html += '</div>';
    }
    
    // 一般建议
    if (advice.suggestions.length > 0) {
        html += '<div class="ai-advice-group suggestions">';
        advice.suggestions.forEach(s => {
            html += `
                <div class="ai-advice-item suggestion">
                    <div class="ai-advice-header">
                        <span class="ai-advice-icon">${s.icon}</span>
                        <span class="ai-advice-title">${s.title}</span>
                    </div>
                    <div class="ai-advice-content">${s.content}</div>
                </div>
            `;
        });
        html += '</div>';
    }
    
    container.innerHTML = html;
}

function calculateTotalRequiredShifts() {
    // 计算总需求班次数
    let total = 0;
    const requirements = appState.requirements || {};
    const periodDays = appState.settings?.schedulePeriod === 'two_weeks' ? 14 : 7;
    const weekdayCount = 5;
    const weekendCount = 2;
    
    // requirements 是对象格式: { storeId: { weekday: { shiftId: { position: count } }, weekend: {...} } }
    for (const storeId in requirements) {
        const storeReq = requirements[storeId];
        if (storeReq && typeof storeReq === 'object') {
            // 工作日需求
            if (storeReq.weekday) {
                for (const shiftId in storeReq.weekday) {
                    const shiftReq = storeReq.weekday[shiftId];
                    for (const position in shiftReq) {
                        const count = shiftReq[position] || 0;
                        total += count * weekdayCount * (periodDays / 7);
                    }
                }
            }
            // 周末需求
            if (storeReq.weekend) {
                for (const shiftId in storeReq.weekend) {
                    const shiftReq = storeReq.weekend[shiftId];
                    for (const position in shiftReq) {
                        const count = shiftReq[position] || 0;
                        total += count * weekendCount * (periodDays / 7);
                    }
                }
            }
        }
    }
    
    return total || appState.assignments?.length || 0;
}

// 排班报告面板样式
const reportStyles = `
<style>
/* 按钮组 */
.btn-group {
    display: inline-flex;
}

/* 排班报告按钮 */
.report-toggle-btn {
    position: relative;
    padding: 8px 16px;
    background: var(--bg-tertiary);
    border: 1px solid rgba(255,255,255,0.1);
    border-radius: var(--radius-sm) 0 0 var(--radius-sm);
    border-right: none;
    color: var(--text-secondary);
    font-size: 13px;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    gap: 6px;
}

.report-toggle-btn:hover {
    background: var(--bg-secondary);
    color: var(--text-primary);
    border-color: var(--primary);
}

/* 排班报告面板 */
.report-panel {
    position: fixed;
    right: -420px;
    top: 0;
    width: 420px;
    height: 100vh;
    background: var(--bg-secondary);
    border-left: 1px solid rgba(255,255,255,0.1);
    z-index: 1000;
    transition: right 0.3s ease;
    display: flex;
    flex-direction: column;
    box-shadow: -5px 0 20px rgba(0,0,0,0.3);
}

.report-panel.active {
    right: 0;
}

.report-header {
    padding: 20px;
    border-bottom: 1px solid rgba(255,255,255,0.1);
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: var(--bg-tertiary);
}

.report-header h3 {
    font-size: 16px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0;
}

.report-close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px 8px;
    border-radius: var(--radius-sm);
    font-size: 18px;
    transition: all 0.2s;
}

.report-close-btn:hover {
    color: var(--text-primary);
    background: rgba(255,255,255,0.1);
}

.report-content {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
}

.report-empty {
    text-align: center;
    color: var(--text-muted);
    padding: 40px 20px;
    font-size: 14px;
}

/* 报告区块 */
.report-section {
    margin-bottom: 20px;
}

.report-section-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid rgba(255,255,255,0.1);
}

.report-section-title.error {
    color: var(--danger);
}

.report-section-title.warning {
    color: var(--warning);
}

/* 概览卡片 */
.report-overview-cards {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
}

.report-card {
    background: var(--bg-tertiary);
    border-radius: var(--radius-md);
    padding: 16px;
    text-align: center;
    border: 1px solid rgba(255,255,255,0.05);
}

.report-card.success {
    border-color: rgba(34, 197, 94, 0.3);
}

.report-card.warning {
    border-color: rgba(234, 179, 8, 0.3);
}

.report-card.error {
    border-color: rgba(239, 68, 68, 0.3);
}

.report-card-value {
    font-size: 24px;
    font-weight: 700;
    color: var(--text-primary);
}

.report-card.success .report-card-value {
    color: var(--success);
}

.report-card.warning .report-card-value {
    color: var(--warning);
}

.report-card.error .report-card-value {
    color: var(--danger);
}

.report-card-label {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 4px;
}

/* 岗位统计 */
.report-position-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.report-position-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
}

.report-position-name {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
}

.report-position-stats {
    display: flex;
    gap: 8px;
}

.report-stat-badge {
    font-size: 11px;
    padding: 3px 8px;
    background: var(--primary);
    color: white;
    border-radius: 10px;
}

.report-stat-badge.secondary {
    background: var(--secondary);
}

/* 违规列表 */
.report-violation-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.report-violation-item {
    display: flex;
    gap: 12px;
    padding: 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    border-left: 3px solid var(--text-muted);
}

.report-violation-item.error {
    border-left-color: var(--danger);
    background: rgba(239, 68, 68, 0.05);
}

.report-violation-item.warning {
    border-left-color: var(--warning);
    background: rgba(234, 179, 8, 0.05);
}

.report-violation-index {
    width: 24px;
    height: 24px;
    background: rgba(255,255,255,0.1);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-muted);
    flex-shrink: 0;
}

.report-violation-content {
    flex: 1;
    min-width: 0;
}

.report-violation-type {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    margin-bottom: 4px;
}

.report-violation-message {
    font-size: 13px;
    color: var(--text-primary);
    line-height: 1.4;
    word-break: break-word;
}

/* 未满足需求列表 */
.report-unfilled-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.report-unfilled-item {
    display: flex;
    gap: 12px;
    padding: 10px 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    border-left: 3px solid var(--warning);
}

.report-unfilled-index {
    width: 20px;
    height: 20px;
    background: rgba(234, 179, 8, 0.2);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    font-weight: 600;
    color: var(--warning);
    flex-shrink: 0;
}

.report-unfilled-content {
    flex: 1;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.report-unfilled-info {
    display: flex;
    gap: 8px;
    align-items: center;
}

.report-unfilled-day {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
}

.report-unfilled-shift {
    font-size: 11px;
    color: var(--text-secondary);
    padding: 2px 6px;
    background: rgba(255,255,255,0.1);
    border-radius: 4px;
}

.report-unfilled-store {
    font-size: 11px;
    color: var(--warning);
    font-weight: 500;
    padding: 2px 6px;
    background: rgba(234, 179, 8, 0.15);
    border-radius: 4px;
}

.report-unfilled-position {
    font-size: 11px;
    color: var(--primary);
}

.report-unfilled-gap {
    font-size: 12px;
    font-weight: 600;
    color: var(--danger);
}

/* 员工工作量 */
.report-employee-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-height: 300px;
    overflow-y: auto;
}

.report-employee-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
}

.report-employee-info {
    display: flex;
    gap: 8px;
    align-items: center;
}

.report-employee-name {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
}

.report-employee-position {
    font-size: 11px;
    color: var(--text-muted);
}

.report-employee-stats {
    display: flex;
    align-items: center;
    gap: 10px;
}

.report-employee-bar {
    width: 60px;
    height: 6px;
    background: rgba(255,255,255,0.1);
    border-radius: 3px;
    overflow: hidden;
}

.report-employee-bar-fill {
    height: 100%;
    background: var(--primary);
    border-radius: 3px;
    transition: width 0.3s;
}

.report-employee-bar-fill.warning {
    background: var(--warning);
}

.report-employee-bar-fill.error {
    background: var(--danger);
}

.report-employee-count {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    min-width: 40px;
    text-align: right;
}

/* 分析列表 */
.report-analysis-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.report-analysis-item {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 10px 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    border-left: 3px solid var(--text-muted);
}

.report-analysis-item.success {
    border-left-color: var(--success);
}

.report-analysis-item.warning {
    border-left-color: var(--warning);
}

.report-analysis-item.error, .report-analysis-item.critical {
    border-left-color: var(--danger);
}

.report-analysis-item.info {
    border-left-color: var(--primary);
}

.report-analysis-icon {
    font-size: 14px;
    flex-shrink: 0;
}

.report-analysis-text {
    font-size: 13px;
    color: var(--text-primary);
    line-height: 1.4;
}

/* 解决方案列表 */
.report-solution-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.report-solution-item {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 10px 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    border-left: 3px solid var(--primary);
}

.report-solution-item.primary {
    border-left-color: var(--primary);
    background: rgba(99, 102, 241, 0.05);
}

.report-solution-item.secondary {
    border-left-color: var(--secondary);
}

.report-solution-item.warning {
    border-left-color: var(--warning);
}

.report-solution-item.success {
    border-left-color: var(--success);
}

.report-solution-icon {
    font-size: 14px;
    flex-shrink: 0;
}

.report-solution-text {
    font-size: 13px;
    color: var(--text-primary);
    line-height: 1.4;
}

/* 成功提示 */
.report-success {
    text-align: center;
    padding: 30px 20px;
    background: rgba(34, 197, 94, 0.1);
    border-radius: var(--radius-md);
    border: 1px solid rgba(34, 197, 94, 0.2);
}

.report-success-icon {
    font-size: 40px;
    margin-bottom: 12px;
}

.report-success-text {
    font-size: 14px;
    color: var(--success);
    font-weight: 500;
}

/* AI 建议区块 */
.report-ai-container {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.report-ai-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 12px 20px;
    background: linear-gradient(135deg, var(--primary), #8b5cf6);
    border: none;
    border-radius: var(--radius-md);
    color: white;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s;
    box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.report-ai-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(99, 102, 241, 0.4);
}

.report-ai-btn:active {
    transform: translateY(0);
}

.ai-btn-icon {
    font-size: 16px;
}

.report-ai-content {
    min-height: 0;
}

.ai-loading {
    text-align: center;
    padding: 20px;
    color: var(--text-secondary);
    font-size: 13px;
}

.ai-loading-icon {
    margin-right: 8px;
    animation: pulse 1.5s infinite;
}

@keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
}

.ai-summary {
    padding: 12px 16px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-md);
    font-size: 14px;
    color: var(--text-primary);
    line-height: 1.5;
    margin-bottom: 12px;
    border-left: 3px solid var(--primary);
}

.ai-advice-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 12px;
}

.ai-advice-item {
    padding: 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    border-left: 3px solid var(--text-muted);
}

.ai-advice-item.warning {
    border-left-color: var(--danger);
    background: rgba(239, 68, 68, 0.05);
}

.ai-advice-item.optimization {
    border-left-color: var(--warning);
    background: rgba(234, 179, 8, 0.05);
}

.ai-advice-item.suggestion {
    border-left-color: var(--primary);
    background: rgba(99, 102, 241, 0.05);
}

.ai-advice-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
}

.ai-advice-icon {
    font-size: 14px;
}

.ai-advice-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
}

.ai-advice-content {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.5;
}

/* 豆包 AI 响应样式 */
.ai-doubao-response {
    font-size: 13px;
    color: var(--text-primary);
    line-height: 1.6;
}

.ai-doubao-response p {
    margin-bottom: 12px;
}

.ai-doubao-response strong {
    color: var(--primary);
    font-weight: 600;
}

.ai-doubao-response .ai-list {
    list-style: none;
    padding: 0;
    margin: 8px 0;
}

.ai-doubao-response .ai-list-item {
    padding: 8px 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    margin-bottom: 6px;
    border-left: 3px solid var(--primary);
}

.ai-error {
    text-align: center;
    padding: 16px;
    color: var(--warning);
    font-size: 13px;
    background: rgba(234, 179, 8, 0.1);
    border-radius: var(--radius-sm);
}

/* 补员建议样式 */
.report-staffing-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.report-staffing-item {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 12px;
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    border-left: 3px solid var(--primary);
}

.report-staffing-item.shortage {
    border-left-color: var(--warning);
    background: rgba(234, 179, 8, 0.1);
}

.report-staffing-item.overwork {
    border-left-color: var(--error);
    background: rgba(239, 68, 68, 0.1);
}

.report-staffing-icon {
    font-size: 18px;
    flex-shrink: 0;
}

.report-staffing-content {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.report-staffing-position {
    font-weight: 600;
    color: var(--text-primary);
    font-size: 13px;
}

.report-staffing-reason {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.5;
}

.report-staffing-nums {
    font-size: 12px;
    color: var(--primary);
    font-weight: 500;
    padding: 4px 8px;
    background: rgba(59, 130, 246, 0.1);
    border-radius: var(--radius-xs);
    display: inline-block;
    margin-top: 4px;
}

/* 排班报告遮罩 */
.report-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0,0,0,0.5);
    z-index: 999;
    opacity: 0;
    visibility: hidden;
    transition: all 0.3s;
}

.report-overlay.active {
    opacity: 1;
    visibility: visible;
}
</style>
`;

document.head.insertAdjacentHTML('beforeend', reportStyles);
