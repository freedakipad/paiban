/**
 * PaiBan API 控制台 - 应用逻辑
 * 智能排班引擎 API 测试工具
 * 
 * 依赖: scenarios.js (场景数据)
 */

// ========== 工具函数 ==========
function getShiftName(shifts, code) {
  if (!shifts) return code;
  const shift = shifts.find(s => s.code === code || s.type === code);
  return shift ? shift.name : code;
}

function daysBetween(start, end) {
  const s = new Date(start);
  const e = new Date(end);
  return Math.ceil((e - s) / (1000 * 60 * 60 * 24)) + 1;
}

// ========== 全局状态 ==========
let currentScenario = 'restaurant';
let isLoading = false;
let lastResponse = null;
let constraintTemplatesData = [];
let constraintLibraryData = [];
const editModeState = {};
const originalValues = {};

// ========== DOM 元素引用 ==========
let serverUrlInput, statusDot, statusText, requestBody, requestBizView;
let responseMeta, resultsSummary, responseTabs, businessView, responseOutput, sendBtn;

// ========== 初始化 ==========
document.addEventListener('DOMContentLoaded', function() {
  // 获取DOM元素
  serverUrlInput = document.getElementById('serverUrl');
  statusDot = document.getElementById('statusDot');
  statusText = document.getElementById('statusText');
  requestBody = document.getElementById('requestBody');
  requestBizView = document.getElementById('requestBizView');
  responseMeta = document.getElementById('responseMeta');
  resultsSummary = document.getElementById('resultsSummary');
  responseTabs = document.getElementById('responseTabs');
  businessView = document.getElementById('businessView');
  responseOutput = document.getElementById('responseOutput');
  sendBtn = document.getElementById('sendBtn');

  // 绑定场景卡片点击事件
  document.querySelectorAll('.scenario-card').forEach(card => {
    card.addEventListener('click', () => selectScenario(card.dataset.scenario));
  });

  // 绑定服务器URL变化事件
  serverUrlInput.addEventListener('change', checkServerStatus);

  // 初始化
  loadScenario('restaurant');
  checkServerStatus();
  setInterval(checkServerStatus, 30000);
});

// ========== 场景管理 ==========
function selectScenario(scenario) {
  currentScenario = scenario;
  document.querySelectorAll('.scenario-card').forEach(c => c.classList.remove('active'));
  document.querySelector(`.scenario-card.${scenario}`).classList.add('active');
  loadScenario(scenario);
}

function loadScenario(scenario) {
  const data = scenarioData[scenario];
  document.getElementById('requestMethod').textContent = data.method;
  document.getElementById('requestMethod').className = `method-badge ${data.method.toLowerCase()}`;
  document.getElementById('endpointMethod').textContent = data.method;
  document.getElementById('endpointPath').textContent = data.endpoint;
  requestBody.value = JSON.stringify(data.body, null, 2);
  renderRequestBizView(data.body, scenario);
}

// ========== 请求业务视图渲染 ==========
function renderRequestBizView(body, scenario) {
  const meta = scenarioMeta[scenario];
  const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];

  let html = `
    <div class="biz-section">
      <div class="biz-section-title">📌 基本信息</div>
      <div class="info-grid">
        <div class="info-item">
          <div class="info-label">场景类型</div>
          <div class="info-value" style="color: ${meta.color}">${meta.icon} ${meta.name}</div>
        </div>
        <div class="info-item">
          <div class="info-label">开始日期</div>
          <div class="info-value cyan">${body.start_date}</div>
        </div>
        <div class="info-item">
          <div class="info-label">结束日期</div>
          <div class="info-value cyan">${body.end_date}</div>
        </div>
        <div class="info-item">
          <div class="info-label">排班周期</div>
          <div class="info-value green">${daysBetween(body.start_date, body.end_date)} 天</div>
        </div>
      </div>
    </div>

    <div class="biz-section">
      <div class="biz-section-title">👥 参与员工 (${body.employees.length} 人)</div>
      <div class="item-grid">
        ${body.employees.map(e => {
          const prefs = e.preferences;
          let prefTags = '';
          if (prefs) {
            if (prefs.preferred_shifts?.length) prefTags += prefs.preferred_shifts.map(s => `<span class="item-tag" style="background: #238636; color: #fff;">偏好${getShiftName(body.shifts, s)}</span>`).join('');
            if (prefs.avoid_shifts?.length) prefTags += prefs.avoid_shifts.map(s => `<span class="item-tag" style="background: #da3633; color: #fff;">避免${getShiftName(body.shifts, s)}</span>`).join('');
            if (prefs.avoid_days?.length) prefTags += `<span class="item-tag" style="background: #da3633; color: #fff;">避免${prefs.avoid_days.map(d => weekDays[d]).join('/')}</span>`;
            if (prefs.max_hours_per_week) prefTags += `<span class="item-tag" style="background: #1f6feb; color: #fff;">≤${prefs.max_hours_per_week}h/周</span>`;
          }
          return `
          <div class="item-card">
            <div class="item-name">👤 ${e.name}</div>
            <div class="item-detail">${e.position || '员工'}</div>
            ${e.skills?.length ? `<div class="item-tags">${e.skills.map(s => `<span class="item-tag">${s}</span>`).join('')}</div>` : ''}
            ${prefTags ? `<div class="item-tags" style="margin-top: 4px;">${prefTags}</div>` : ''}
          </div>
        `}).join('')}
      </div>
    </div>

    <div class="biz-section">
      <div class="biz-section-title">🕐 班次设置 (${body.shifts.length} 个)</div>
      <div class="item-grid">
        ${body.shifts.map(s => `
          <div class="item-card shift">
            <div class="item-name">📅 ${s.name}</div>
            <div class="item-detail">${s.start_time} - ${s.end_time} (${Math.round(s.duration/60)}h)</div>
            <div class="item-tags"><span class="item-tag">${s.code}</span></div>
          </div>
        `).join('')}
      </div>
    </div>

    <div class="biz-section">
      <div class="biz-section-title">📋 排班需求 (${body.requirements.length} 条)</div>
      <div class="item-grid">
        ${body.requirements.slice(0, 12).map(r => {
          const d = new Date(r.date);
          const shift = body.shifts?.find(s => s.id === r.shift_id);
          const shiftName = shift ? shift.name : '';
          return `
            <div class="item-card requirement">
              <div class="item-name">${weekDays[d.getDay()]} ${r.date}</div>
              <div class="item-detail">${r.note || shiftName || '班次'}: ${r.min_employees} 人${r.position ? ` (${r.position})` : ''}</div>
              <div class="item-tags">
                ${shiftName ? `<span class="item-tag">${shiftName}</span>` : ''}
                <span class="item-tag">优先级 ${r.priority || 5}</span>
              </div>
            </div>
          `;
        }).join('')}
        ${body.requirements.length > 12 ? `<div class="item-card requirement" style="text-align: center; border-left-color: var(--text-muted);"><div class="item-detail">... 还有 ${body.requirements.length - 12} 条需求</div></div>` : ''}
      </div>
    </div>

    <div class="biz-section">
      <div class="biz-section-title">⚙️ 约束规则</div>
      <div class="item-grid">
        ${Object.entries(body.constraints || {}).map(([k, v]) => `
          <div class="item-card constraint">
            <div class="item-name">${constraintLabel(k)}</div>
            <div class="item-detail" style="color: var(--accent-green); font-weight: 600;">${formatConstraintValue(k, v)}</div>
          </div>
        `).join('')}
      </div>
    </div>

    ${body.options ? `
    <div class="biz-section">
      <div class="biz-section-title">🎛️ 计算选项</div>
      <div class="info-grid">
        ${body.options.timeout_seconds ? `<div class="info-item"><div class="info-label">超时时间</div><div class="info-value">${body.options.timeout_seconds} 秒</div></div>` : ''}
        ${body.options.optimization_level ? `<div class="info-item"><div class="info-label">优化级别</div><div class="info-value">${['', '快速', '平衡', '最优'][body.options.optimization_level]}</div></div>` : ''}
        ${body.options.respect_preferences !== undefined ? `<div class="info-item"><div class="info-label">考虑偏好</div><div class="info-value">${body.options.respect_preferences ? '是' : '否'}</div></div>` : ''}
      </div>
    </div>
    ` : ''}
  `;

  requestBizView.innerHTML = html;
}

function constraintLabel(key) {
  const labels = {
    max_hours_per_week: '最大周工时',
    min_rest_hours: '最小休息时间',
    max_consecutive_days: '最大连续工作天数',
    max_consecutive_nights: '最大连续夜班',
    max_orders_per_day: '每日最大订单',
    skill_match_required: '技能匹配',
    continuity_required: '服务连续性',
    max_patients_per_day: '每日最大患者数',
    rotation_pattern: '倒班模式'
  };
  return labels[key] || key;
}

function formatConstraintValue(key, value) {
  if (typeof value === 'boolean') return value ? '必须' : '不要求';
  if (key.includes('hours')) return `${value} 小时`;
  if (key.includes('days')) return `${value} 天`;
  return value;
}

// ========== Tab切换 ==========
function switchMainTab(tabName) {
  document.querySelectorAll('.main-tab').forEach(t => t.classList.remove('active'));
  document.querySelector(`.main-tab[data-main-tab="${tabName}"]`)?.classList.add('active');
  document.querySelectorAll('.main-tab-content').forEach(c => c.classList.remove('active'));
  document.getElementById(tabName + 'Content')?.classList.add('active');
}

function switchRequestTab(tabName) {
  document.querySelectorAll('#requestContent .tab').forEach(t => t.classList.remove('active'));
  document.querySelectorAll('#requestContent .tab-content').forEach(c => c.classList.remove('active'));
  document.querySelector(`#requestContent .tab[data-tab="${tabName}"]`)?.classList.add('active');
  document.getElementById(tabName + 'Tab')?.classList.add('active');
  if (tabName === 'reqJson') {
    try { JSON.parse(requestBody.value); } catch(e) { /* keep existing */ }
  }
}

function switchResponseTab(tabName) {
  document.querySelectorAll('#responseTabs .tab').forEach(t => t.classList.remove('active'));
  document.getElementById('resBizTab').classList.remove('active');
  document.getElementById('resJsonTab').classList.remove('active');
  document.querySelector(`#responseTabs .tab[data-tab="${tabName}"]`)?.classList.add('active');
  document.getElementById(tabName + 'Tab')?.classList.add('active');
}

// ========== API 请求 ==========
async function checkServerStatus() {
  const url = serverUrlInput.value;
  try {
    const response = await fetch(`${url}/health`);
    if (response.ok) { statusDot.classList.remove('offline'); statusText.textContent = '服务正常'; }
    else throw new Error();
  } catch (e) { statusDot.classList.add('offline'); statusText.textContent = '无法连接'; }
}

async function sendRequest() {
  if (isLoading) return;
  const url = serverUrlInput.value;
  const data = scenarioData[currentScenario];

  let body;
  try { body = JSON.parse(requestBody.value); }
  catch (e) { showError('JSON 格式错误: ' + e.message); return; }

  isLoading = true;
  sendBtn.disabled = true;
  sendBtn.innerHTML = '⏳ 请求中...';

  const startTime = performance.now();

  try {
    const response = await fetch(`${url}${data.endpoint}`, {
      method: data.method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
    const responseData = await response.json();
    lastResponse = responseData;
    showResponse(response.status, performance.now() - startTime, responseData);
  } catch (e) { showError('请求失败: ' + e.message); }
  finally {
    isLoading = false;
    sendBtn.disabled = false;
    sendBtn.innerHTML = '🚀 发送请求';
  }
}

// ========== 响应处理 ==========
function showResponse(status, time, data) {
  switchMainTab('response');
  
  const statusBadge = document.getElementById('responseStatusBadge');
  const indicator = document.getElementById('responseIndicator');
  if (status >= 200 && status < 300) {
    statusBadge.className = 'status-badge success';
    statusBadge.textContent = '✓';
    indicator.style.display = 'inline';
  } else {
    statusBadge.className = 'status-badge error';
    statusBadge.textContent = '✗';
    indicator.style.display = 'none';
  }

  responseMeta.style.display = 'flex';
  const statusEl = document.getElementById('responseStatus');
  statusEl.textContent = status;
  statusEl.className = 'response-meta-value ' + (status >= 200 && status < 300 ? 'success' : 'error');
  document.getElementById('responseTime').textContent = time.toFixed(0) + 'ms';
  document.getElementById('responseSize').textContent = formatBytes(JSON.stringify(data).length);

  if (data.assignments !== undefined) {
    resultsSummary.style.display = 'grid';
    responseTabs.style.display = 'flex';
    document.getElementById('summaryAssignments').textContent = data.assignments?.length || 0;
    document.getElementById('summaryFillRate').textContent = (data.statistics?.fill_rate?.toFixed(1) || '0') + '%';
    document.getElementById('summaryScore').textContent = data.constraint_result?.score?.toFixed(1) || '0';
    document.getElementById('summaryDuration').textContent = data.duration || '-';
    renderResponseBizView(data);
  } else {
    resultsSummary.style.display = 'none';
    responseTabs.style.display = 'none';
    businessView.innerHTML = renderGenericBizView(data);
  }

  responseOutput.innerHTML = syntaxHighlight(JSON.stringify(data, null, 2));
  switchResponseTab('resBiz');
}

function renderResponseBizView(data) {
  const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
  const msgClass = data.success ? (data.partial ? 'partial' : 'success') : 'error';
  const msgIcon = data.success ? (data.partial ? '⚠️' : '✅') : '❌';

  let html = `
    <div class="biz-section">
      <div class="biz-message ${msgClass}">
        <span style="font-size: 1.5rem;">${msgIcon}</span>
        <span>${data.message || (data.success ? '排班生成成功' : '排班失败')}</span>
      </div>
    </div>
  `;

  if (data.statistics) {
    const s = data.statistics;
    html += `
      <div class="biz-section">
        <div class="biz-section-title">📈 统计摘要</div>
        <div class="stats-grid">
          <div class="stat-card"><div class="stat-label">总排班数</div><div class="stat-value green">${s.total_assignments || 0}</div><div class="stat-desc">已安排的班次</div></div>
          <div class="stat-card"><div class="stat-label">需求满足</div><div class="stat-value cyan">${s.filled_requirements || 0} / ${s.total_requirements || 0}</div><div class="stat-desc">满足率 ${(s.fill_rate || 0).toFixed(1)}%</div></div>
          <div class="stat-card"><div class="stat-label">总工时</div><div class="stat-value orange">${s.total_hours || 0}h</div><div class="stat-desc">人均 ${(s.avg_hours_per_employee || 0).toFixed(1)}h</div></div>
          <div class="stat-card"><div class="stat-label">迭代次数</div><div class="stat-value purple">${s.iterations || 0}</div><div class="stat-desc">算法轮次</div></div>
        </div>
      </div>
    `;
  }

  if (data.constraint_result) {
    const cr = data.constraint_result;
    html += `
      <div class="biz-section">
        <div class="biz-section-title">✅ 约束检查</div>
        <div style="display: flex; align-items: center; gap: 1rem; margin-bottom: 1rem;">
          <div class="constraint-badge ${cr.is_valid ? 'valid' : 'invalid'}">${cr.is_valid ? '✓ 所有硬约束满足' : '✗ 存在约束违规'}</div>
          <span style="color: var(--text-secondary);">得分: <strong style="color: var(--accent-cyan);">${(cr.score || 0).toFixed(1)}</strong>/100</span>
        </div>
    `;
    
    if (cr.hard_violations && cr.hard_violations.length > 0) {
      html += renderViolations(cr.hard_violations, 'hard');
    }
    
    if (cr.soft_violations && cr.soft_violations.length > 0) {
      html += renderViolations(cr.soft_violations, 'soft');
    }
    
    html += `</div>`;
  }

  // 显示未满足的需求
  if (data.unfilled?.length > 0) {
    html += renderUnfilledRequirements(data.unfilled);
  }

  if (data.assignments?.length) {
    const byDay = {};
    data.assignments.forEach(a => { if (!byDay[a.date]) byDay[a.date] = []; byDay[a.date].push(a); });

    html += `<div class="biz-section"><div class="biz-section-title">📋 排班详情</div>`;
    Object.keys(byDay).sort().forEach(date => {
      const d = new Date(date);
      html += `
        <div class="day-group">
          <div class="day-header"><span class="day-badge">${weekDays[d.getDay()]}</span><span>${date}</span><span style="color: var(--text-muted);">(${byDay[date].length}个)</span></div>
          <div class="assignment-cards">
            ${byDay[date].map(a => `
              <div class="assignment-card">
                <div class="assignment-emp">👤 ${a.employee_name}</div>
                <div class="assignment-detail"><span class="assignment-shift">${a.shift_name}</span><span>${a.start_time}-${a.end_time}</span><span style="color:var(--accent-green)">${a.hours}h</span></div>
              </div>
            `).join('')}
          </div>
        </div>
      `;
    });
    html += '</div>';
  }

  businessView.innerHTML = html;
}

function renderViolations(violations, type) {
  const isHard = type === 'hard';
  const byType = {};
  violations.forEach(v => {
    const t = v.constraint_name || v.constraint_type;
    if (!byType[t]) byType[t] = [];
    byType[t].push(v);
  });

  let html = `
    <div class="violations-container ${isHard ? '' : 'soft'}">
      <div class="violations-header">
        <span class="violations-icon">${isHard ? '⚠️' : '💡'}</span>
        <span class="violations-title">${isHard ? '硬约束违规详情' : '软约束提醒'} (${violations.length}项)</span>
      </div>
      <div class="violations-list">
  `;
  
  Object.keys(byType).forEach(t => {
    const items = byType[t];
    html += `
      <div class="violation-group">
        <div class="violation-rule">
          <span class="violation-rule-icon">📋</span>
          <span class="violation-rule-name">${t}</span>
          <span class="violation-count">${items.length}项违规</span>
        </div>
        <div class="violation-items">
    `;
    
    items.forEach(v => {
      html += `
        <div class="violation-item ${isHard ? '' : 'soft'}">
          <span class="violation-severity">${v.severity === 'error' ? '❌' : '⚠️'}</span>
          <span class="violation-message">${v.message}</span>
          <span class="violation-penalty">-${v.penalty}分</span>
        </div>
      `;
    });
    
    html += `</div></div>`;
  });
  
  html += `</div></div>`;
  return html;
}

// 渲染未满足的需求
function renderUnfilledRequirements(unfilled) {
  const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
  
  // 按日期分组
  const byDate = {};
  unfilled.forEach(item => {
    if (!byDate[item.date]) byDate[item.date] = [];
    byDate[item.date].push(item);
  });
  
  let html = `
    <div class="biz-section">
      <div class="biz-section-title" style="color: var(--accent-red);">⚠️ 未满足的需求 (${unfilled.length}项)</div>
      <div class="unfilled-notice" style="background: rgba(248, 81, 73, 0.1); border: 1px solid rgba(248, 81, 73, 0.3); border-radius: 8px; padding: 1rem; margin-bottom: 1rem;">
        <div style="display: flex; align-items: center; gap: 0.5rem; color: var(--accent-red); margin-bottom: 0.5rem;">
          <span style="font-size: 1.2rem;">📋</span>
          <span style="font-weight: 600;">以下需求未能满足，可能原因：</span>
        </div>
        <ul style="color: var(--text-secondary); margin: 0; padding-left: 1.5rem; font-size: 0.9rem;">
          <li>可用员工不足</li>
          <li>员工技能不匹配</li>
          <li>员工工时已满</li>
          <li>存在约束冲突</li>
        </ul>
      </div>
      <div class="unfilled-list">
  `;
  
  Object.keys(byDate).sort().forEach(date => {
    const d = new Date(date);
    const items = byDate[date];
    
    html += `
      <div class="unfilled-day-group" style="margin-bottom: 1rem; padding: 0.75rem; background: rgba(248, 81, 73, 0.05); border-radius: 8px; border-left: 3px solid var(--accent-red);">
        <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
          <span class="day-badge" style="background: var(--accent-red); color: white;">${weekDays[d.getDay()]}</span>
          <span style="color: var(--text-primary); font-weight: 500;">${date}</span>
          <span style="color: var(--text-muted);">(${items.length}项未满足)</span>
        </div>
        <div style="display: flex; flex-wrap: wrap; gap: 0.5rem;">
    `;
    
    items.forEach(item => {
      const shiftName = item.shift_name || getShiftNameFromId(item.shift_id);
      const reason = item.reason || '无可用员工';
      const needed = item.needed || 1;
      const assigned = item.assigned || 0;
      
      html += `
        <div class="unfilled-item" style="background: var(--bg-tertiary); border: 1px solid rgba(248, 81, 73, 0.3); border-radius: 6px; padding: 0.5rem 0.75rem; font-size: 0.85rem;">
          <div style="display: flex; align-items: center; gap: 0.5rem;">
            <span style="color: var(--accent-red);">❌</span>
            <span style="color: var(--text-primary); font-weight: 500;">${shiftName}</span>
            ${item.position ? `<span style="color: var(--text-muted);">(${item.position})</span>` : ''}
          </div>
          <div style="color: var(--text-secondary); font-size: 0.8rem; margin-top: 0.25rem;">
            需要 ${needed} 人，仅分配 ${assigned} 人
            ${item.reason ? ` · <span style="color: var(--accent-orange);">${reason}</span>` : ''}
          </div>
        </div>
      `;
    });
    
    html += `</div></div>`;
  });
  
  html += `</div></div>`;
  return html;
}

// 根据ID获取班次名称
function getShiftNameFromId(shiftId) {
  // 尝试从当前场景数据中获取班次名称
  const data = scenarioData[currentScenario];
  if (data?.body?.shifts) {
    const shift = data.body.shifts.find(s => s.id === shiftId);
    if (shift) return shift.name;
  }
  // 如果找不到，返回简短ID
  return shiftId.substring(0, 8) + '...';
}

function renderGenericBizView(data) {
  if (data.status === 'ok') return `<div class="biz-section"><div class="biz-message success"><span style="font-size:1.5rem">✅</span><span>服务正常运行</span></div></div>`;
  if (data.templates) return renderConstraintTemplates(data.templates);
  if (data.library) return renderConstraintLibrary(data.library);
  if (data.version) return `<div class="biz-section"><div class="biz-section-title">📋 版本</div><div class="info-grid"><div class="info-item"><div class="info-label">版本</div><div class="info-value cyan">${data.version}</div></div></div></div>`;
  return `<div style="text-align: center; padding: 2rem; color: var(--text-muted);">💡 请切换到 JSON 查看详情</div>`;
}

function showError(message) {
  switchMainTab('response');
  
  const statusBadge = document.getElementById('responseStatusBadge');
  const indicator = document.getElementById('responseIndicator');
  statusBadge.className = 'status-badge error';
  statusBadge.textContent = '✗';
  indicator.style.display = 'none';

  responseMeta.style.display = 'flex';
  document.getElementById('responseStatus').textContent = 'ERROR';
  document.getElementById('responseStatus').className = 'response-meta-value error';
  resultsSummary.style.display = 'none';
  responseTabs.style.display = 'none';
  businessView.innerHTML = `<div class="biz-section"><div class="biz-message error"><span style="font-size:1.5rem">❌</span><span>${message}</span></div></div>`;
}

// ========== 约束模板 ==========
function renderConstraintTemplates(templates) {
  constraintTemplatesData = JSON.parse(JSON.stringify(templates));
  
  let html = `<div class="biz-section"><div class="biz-section-title">📐 约束模板列表 (${templates.length}个场景)</div>`;
  
  templates.forEach((t, idx) => {
    const scenarioIcons = {restaurant: '🍜', factory: '🏭', housekeeping: '🏠', nursing: '💊'};
    const icon = scenarioIcons[t.scenario] || '📋';
    const hardConstraints = (t.constraints || []).filter(c => c.type === 'hard');
    const softConstraints = (t.constraints || []).filter(c => c.type === 'soft');
    
    html += `
      <div class="template-card" id="template-${idx}">
        <div class="template-header" onclick="toggleTemplate(${idx})">
          <div class="template-icon">${icon}</div>
          <div class="template-info">
            <div class="template-name">${t.name}</div>
            <div class="template-desc">${t.description}</div>
          </div>
          <div class="template-toggle">
            <span class="template-count">${(t.constraints || []).length} 条约束</span>
            <span class="template-arrow" id="arrow-${idx}">▼</span>
          </div>
        </div>
        <div class="template-body" id="body-${idx}" style="display: none;">
          <div class="template-actions">
            <button class="template-action-btn edit" onclick="event.stopPropagation(); toggleEditMode(${idx})">
              <span id="editIcon-${idx}">✏️</span> <span id="editText-${idx}">编辑</span>
            </button>
            <button class="template-action-btn cancel" id="cancelBtn-${idx}" style="display: none;" onclick="event.stopPropagation(); cancelEditMode(${idx})">
              ❌ 取消
            </button>
            <button class="template-action-btn add" onclick="event.stopPropagation(); showAddFromLibrary(${idx}, '${t.scenario}')">
              📚 从约束库添加
            </button>
            <button class="template-action-btn apply" onclick="event.stopPropagation(); applyTemplate('${t.scenario}', ${idx})">
              ✅ 应用到当前配置
            </button>
          </div>
    `;
    
    if (hardConstraints.length > 0) {
      html += renderConstraintGroup(hardConstraints, idx, 'hard');
    }
    
    if (softConstraints.length > 0) {
      html += renderConstraintGroup(softConstraints, idx, 'soft');
    }
    
    html += `</div></div>`;
  });
  
  html += '</div>';
  return html;
}

function renderConstraintGroup(constraints, templateIdx, type) {
  const isHard = type === 'hard';
  let html = `
    <div class="constraint-group">
      <div class="constraint-group-title">
        <span class="constraint-type-badge ${type}">${isHard ? '硬约束' : '软约束'}</span>
        <span class="constraint-group-desc">${isHard ? '必须满足，违反将导致排班无效' : '尽量满足，用于优化排班质量'}</span>
      </div>
      <div class="constraint-list">
  `;
  
  constraints.forEach((c, cIdx) => {
    const fullIdx = `${templateIdx}-${type}-${cIdx}`;
    html += `
      <div class="constraint-item ${type === 'soft' ? 'soft' : ''}" id="constraint-${fullIdx}">
        <div class="constraint-name">${c.description}</div>
        <div class="constraint-meta">
          <span class="constraint-category">${c.category}</span>
          <span class="constraint-default" id="default-${fullIdx}">${c.default}</span>
          <input type="text" class="constraint-input" id="input-${fullIdx}" value="${c.default}" style="display:none;" data-name="${c.name}">
          <button class="constraint-delete-btn" onclick="event.stopPropagation(); deleteConstraint(${templateIdx}, '${type}', ${cIdx})" title="删除此约束">🗑️</button>
        </div>
      </div>
    `;
  });
  
  html += `</div></div>`;
  return html;
}

function toggleTemplate(idx) {
  const body = document.getElementById('body-' + idx);
  const arrow = document.getElementById('arrow-' + idx);
  if (body.style.display === 'none') {
    body.style.display = 'block';
    arrow.textContent = '▲';
  } else {
    body.style.display = 'none';
    arrow.textContent = '▼';
  }
}

function toggleEditMode(idx) {
  const isEditing = editModeState[idx];
  const card = document.getElementById('template-' + idx);
  const defaults = card.querySelectorAll('.constraint-default');
  const inputs = card.querySelectorAll('.constraint-input');
  const editIcon = document.getElementById('editIcon-' + idx);
  const editText = document.getElementById('editText-' + idx);
  const cancelBtn = document.getElementById('cancelBtn-' + idx);

  if (!isEditing) {
    originalValues[idx] = [];
    defaults.forEach(el => originalValues[idx].push(el.textContent));
    
    defaults.forEach(el => el.style.display = 'none');
    inputs.forEach(el => el.style.display = 'inline-block');
    editIcon.textContent = '💾';
    editText.textContent = '保存';
    cancelBtn.style.display = 'inline-flex';
    editModeState[idx] = true;
  } else {
    inputs.forEach((input, i) => {
      defaults[i].textContent = input.value;
      defaults[i].style.display = 'inline';
      input.style.display = 'none';
    });
    editIcon.textContent = '✏️';
    editText.textContent = '编辑';
    cancelBtn.style.display = 'none';
    editModeState[idx] = false;
    showToast('✅ 约束配置已保存');
  }
}

function cancelEditMode(idx) {
  const card = document.getElementById('template-' + idx);
  const defaults = card.querySelectorAll('.constraint-default');
  const inputs = card.querySelectorAll('.constraint-input');
  const editIcon = document.getElementById('editIcon-' + idx);
  const editText = document.getElementById('editText-' + idx);
  const cancelBtn = document.getElementById('cancelBtn-' + idx);

  if (originalValues[idx]) {
    defaults.forEach((el, i) => {
      el.textContent = originalValues[idx][i] || el.textContent;
      el.style.display = 'inline';
    });
    inputs.forEach((input, i) => {
      input.value = originalValues[idx][i] || input.value;
      input.style.display = 'none';
    });
  }

  editIcon.textContent = '✏️';
  editText.textContent = '编辑';
  cancelBtn.style.display = 'none';
  editModeState[idx] = false;
  showToast('🔄 已取消编辑');
}

function applyTemplate(scenario, idx) {
  const card = document.getElementById('template-' + idx);
  const inputs = card.querySelectorAll('.constraint-input');
  
  const constraintConfig = {};
  inputs.forEach(input => {
    const name = input.dataset.name;
    const value = input.value;
    const numMatch = value.match(/(\d+)/);
    if (numMatch) {
      constraintConfig[name] = parseInt(numMatch[1]);
    } else {
      constraintConfig[name] = value;
    }
  });

  try {
    const currentReq = JSON.parse(requestBody.value);
    if (!currentReq.constraints) currentReq.constraints = {};
    if (!currentReq.constraints.config) currentReq.constraints.config = {};
    
    Object.assign(currentReq.constraints.config, constraintConfig);
    currentReq.scenario = scenario;
    
    requestBody.value = JSON.stringify(currentReq, null, 2);
    loadScenario(scenario === 'nursing' ? 'nursing' : scenario);
    
    showToast(`已应用 ${constraintTemplatesData[idx].name} 到当前配置`);
    switchMainTab('request');
  } catch (e) {
    showToast('应用失败: ' + e.message, 'error');
  }
}

function deleteConstraint(templateIdx, type, constraintIdx) {
  const template = constraintTemplatesData[templateIdx];
  if (!template || !template.constraints) return;

  const constraints = template.constraints.filter(c => c.type === type);
  if (constraintIdx >= 0 && constraintIdx < constraints.length) {
    const targetConstraint = constraints[constraintIdx];
    const originalIdx = template.constraints.indexOf(targetConstraint);
    if (originalIdx > -1) {
      const deletedName = template.constraints[originalIdx].description;
      template.constraints.splice(originalIdx, 1);
      refreshTemplateDisplay(templateIdx);
      showToast(`已删除约束: ${deletedName}`);
    }
  }
}

function refreshTemplateDisplay(templateIdx) {
  const container = document.querySelector('.biz-section');
  if (container) {
    const newHtml = renderConstraintTemplates(constraintTemplatesData);
    container.outerHTML = newHtml;
    setTimeout(() => {
      const body = document.getElementById('body-' + templateIdx);
      const arrow = document.getElementById('arrow-' + templateIdx);
      if (body) {
        body.style.display = 'block';
        arrow.textContent = '▲';
      }
    }, 50);
  }
}

// ========== 约束库 ==========
async function showAddFromLibrary(templateIdx, scenario) {
  if (constraintLibraryData.length === 0) {
    const url = serverUrlInput.value;
    try {
      const response = await fetch(`${url}/api/v1/constraints/library`);
      const data = await response.json();
      constraintLibraryData = data.library || [];
    } catch(e) {
      showToast('获取约束库失败', 'error');
      return;
    }
  }

  const template = constraintTemplatesData[templateIdx];
  const existingNames = (template.constraints || []).map(c => c.name);
  const available = constraintLibraryData.filter(c => 
    c.scenarios.includes(scenario) && !existingNames.includes(c.name)
  );

  if (available.length === 0) {
    showToast('该场景没有更多可添加的约束');
    return;
  }

  const modal = document.createElement('div');
  modal.className = 'library-modal';
  modal.id = 'libraryModal';
  modal.innerHTML = `
    <div class="library-modal-content">
      <div class="library-modal-header">
        <span>📚 从约束库添加 - ${scenario}</span>
        <button onclick="closeLibraryModal()" class="library-modal-close">✕</button>
      </div>
      <div class="library-modal-body">
        ${available.map((c, i) => `
          <div class="library-select-item" onclick="selectConstraintFromLibrary(${templateIdx}, '${c.name}')">
            <div class="library-select-header">
              <span class="constraint-type-badge ${c.type}">${c.type === 'hard' ? '硬' : '软'}</span>
              <span class="library-select-name">${c.display_name}</span>
            </div>
            <div class="library-select-desc">${c.description}</div>
            <div class="library-select-params">
              ${c.params && c.params.length > 0 
                ? c.params.map(p => `<span class="param-badge">${p.name}: ${p.default}</span>`).join('') 
                : '<span class="param-badge">无参数</span>'}
            </div>
          </div>
        `).join('')}
      </div>
    </div>
  `;
  document.body.appendChild(modal);
}

function closeLibraryModal() {
  const modal = document.getElementById('libraryModal');
  if (modal) modal.remove();
}

function selectConstraintFromLibrary(templateIdx, constraintName) {
  const libConstraint = constraintLibraryData.find(c => c.name === constraintName);
  if (!libConstraint) return;

  const defaultValue = libConstraint.params && libConstraint.params.length > 0
    ? libConstraint.params.map(p => p.default).join(', ')
    : '启用';

  const newConstraint = {
    name: libConstraint.name,
    type: libConstraint.type,
    category: libConstraint.category,
    description: libConstraint.display_name,
    default: defaultValue
  };

  if (!constraintTemplatesData[templateIdx].constraints) {
    constraintTemplatesData[templateIdx].constraints = [];
  }
  constraintTemplatesData[templateIdx].constraints.push(newConstraint);

  closeLibraryModal();
  refreshTemplateDisplay(templateIdx);
  showToast(`已添加约束: ${libConstraint.display_name}`);
}

function renderConstraintLibrary(library) {
  const scenarioNames = {
    restaurant: '🍜 餐饮',
    factory: '🏭 工厂',
    housekeeping: '🏠 家政',
    nursing: '💊 长护险'
  };

  const grouped = {};
  library.forEach(c => {
    if (!grouped[c.category]) grouped[c.category] = [];
    grouped[c.category].push(c);
  });

  let html = `<div class="biz-section"><div class="biz-section-title">📚 约束库 - 后端支持的所有约束 (${library.length}项)</div>`;
  html += `<p style="color: var(--text-muted); margin-bottom: 1rem; font-size: 0.9rem;">💡 这些是后端实际支持的约束规则，可用于配置排班请求。</p>`;

  Object.entries(grouped).forEach(([category, constraints]) => {
    html += `
      <div class="library-category">
        <div class="library-category-header">${category} (${constraints.length})</div>
        <div class="library-constraints">
    `;
    constraints.forEach(c => {
      const scenarioTags = c.scenarios.map(s => `<span class="scenario-tag">${scenarioNames[s] || s}</span>`).join('');
      const paramList = c.params && c.params.length > 0 
        ? c.params.map(p => `<span class="param-badge" title="${p.description}">${p.name}: ${p.default}</span>`).join('')
        : '<span class="param-badge">无参数</span>';
      
      html += `
        <div class="library-item ${c.type}">
          <div class="library-item-header">
            <span class="constraint-type-badge ${c.type}">${c.type === 'hard' ? '硬' : '软'}</span>
            <span class="library-item-name">${c.display_name}</span>
            <code class="library-item-code">${c.name}</code>
          </div>
          <div class="library-item-desc">${c.description}</div>
          <div class="library-item-meta">
            <div class="scenario-tags">${scenarioTags}</div>
            <div class="param-list">${paramList}</div>
          </div>
        </div>
      `;
    });
    html += `</div></div>`;
  });

  html += '</div>';
  return html;
}

// ========== 快捷操作 ==========
async function checkHealth() {
  const url = serverUrlInput.value;
  const startTime = performance.now();
  try {
    const response = await fetch(`${url}/health`);
    const data = await response.json();
    lastResponse = data;
    showResponse(response.status, performance.now() - startTime, data);
  } catch(e) { showError('健康检查失败: ' + e.message); }
}

async function getVersion() {
  const url = serverUrlInput.value;
  const startTime = performance.now();
  try {
    const response = await fetch(`${url}/version`);
    const data = await response.json();
    lastResponse = data;
    showResponse(response.status, performance.now() - startTime, data);
  } catch(e) { showError('获取版本失败: ' + e.message); }
}

async function getConstraintTemplates() {
  const url = serverUrlInput.value;
  const startTime = performance.now();
  try {
    const response = await fetch(`${url}/api/v1/constraints/templates`);
    const data = await response.json();
    lastResponse = data;
    showResponse(response.status, performance.now() - startTime, data);
  } catch(e) { showError('获取模板失败: ' + e.message); }
}

async function getConstraintLibrary() {
  const url = serverUrlInput.value;
  const startTime = performance.now();
  try {
    const response = await fetch(`${url}/api/v1/constraints/library`);
    const data = await response.json();
    constraintLibraryData = data.library || [];
    lastResponse = data;
    showResponse(response.status, performance.now() - startTime, data);
    switchMainTab('response');
  } catch(e) { showError('获取约束库失败: ' + e.message); }
}

// ========== 工具函数 ==========
function formatRequest() { 
  try { requestBody.value = JSON.stringify(JSON.parse(requestBody.value), null, 2); } 
  catch(e) { alert('JSON错误: ' + e.message); } 
}

function resetRequest() { 
  loadScenario(currentScenario); 
}

function clearResponse() {
  const statusBadge = document.getElementById('responseStatusBadge');
  const indicator = document.getElementById('responseIndicator');
  statusBadge.className = 'status-badge waiting';
  statusBadge.textContent = '⏳';
  indicator.style.display = 'none';

  responseMeta.style.display = 'none';
  resultsSummary.style.display = 'none';
  responseTabs.style.display = 'none';
  businessView.innerHTML = `<div class="response-placeholder"><div class="response-placeholder-icon">📡</div><span>选择场景并发送请求</span></div>`;
  responseOutput.innerHTML = '';
  lastResponse = null;
}

function copyResponse() { 
  if (lastResponse) { 
    navigator.clipboard.writeText(JSON.stringify(lastResponse, null, 2)); 
    alert('已复制'); 
  } 
}

function syntaxHighlight(json) {
  return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
    m => `<span class="${/^"/.test(m) ? (/:$/.test(m) ? 'json-key' : 'json-string') : /true|false/.test(m) ? 'json-boolean' : /null/.test(m) ? 'json-null' : 'json-number'}">${m}</span>`);
}

function formatBytes(bytes) { 
  return bytes < 1024 ? bytes + ' B' : bytes < 1048576 ? (bytes/1024).toFixed(1) + ' KB' : (bytes/1048576).toFixed(1) + ' MB'; 
}

function showToast(message, type = 'success') {
  const toast = document.createElement('div');
  toast.className = 'toast ' + type;
  toast.innerHTML = `<span>${type === 'success' ? '✅' : '❌'}</span> ${message}`;
  document.body.appendChild(toast);
  
  setTimeout(() => toast.classList.add('show'), 10);
  setTimeout(() => {
    toast.classList.remove('show');
    setTimeout(() => toast.remove(), 300);
  }, 2000);
}
