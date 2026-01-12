/**
 * 餐饮门店智能排班系统 - 工具函数
 */

// 日期格式化
function formatDate(date, format = 'YYYY-MM-DD') {
    const d = new Date(date);
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    
    return format
        .replace('YYYY', year)
        .replace('MM', month)
        .replace('DD', day);
}

// 获取星期几
function getDayName(date) {
    const days = ['日', '一', '二', '三', '四', '五', '六'];
    return '周' + days[new Date(date).getDay()];
}

// 获取完整星期名
function getFullDayName(date) {
    const days = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六'];
    return days[new Date(date).getDay()];
}

// 计算两个时间之间的小时数
function calculateHours(startTime, endTime) {
    const [sh, sm] = startTime.split(':').map(Number);
    const [eh, em] = endTime.split(':').map(Number);
    let hours = (eh * 60 + em - sh * 60 - sm) / 60;
    if (hours < 0) hours += 24; // 跨天
    return hours;
}

// 生成UUID
function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0;
        const v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

// 获取周信息
function getWeekInfo(date) {
    const d = new Date(date);
    const year = d.getFullYear();
    const firstDay = new Date(year, 0, 1);
    const pastDays = (d - firstDay) / 86400000;
    const weekNumber = Math.ceil((pastDays + firstDay.getDay() + 1) / 7);
    return `${year}年第${weekNumber}周`;
}

// 获取月份信息
function getMonthInfo(date) {
    const d = new Date(date);
    return `${d.getFullYear()}年${d.getMonth() + 1}月`;
}

// 判断是否是今天
function isToday(date) {
    const today = new Date();
    const d = new Date(date);
    return d.toDateString() === today.toDateString();
}

// 判断是否是周末
function isWeekend(date) {
    const day = new Date(date).getDay();
    return day === 0 || day === 6;
}

// 获取员工首字母头像
function getAvatarLetter(name) {
    return name ? name.charAt(0) : '?';
}

// 根据班次类型获取CSS类名
function getShiftClass(shiftId) {
    if (shiftId.includes('morning')) return 'shift-morning';
    if (shiftId.includes('afternoon')) return 'shift-afternoon';
    if (shiftId.includes('evening')) return 'shift-evening';
    if (shiftId.includes('split')) return 'shift-split';
    return '';
}

// 获取评分等级
function getScoreLevel(score) {
    if (score >= 90) return 'high';
    if (score >= 70) return 'medium';
    return 'low';
}

// 获取门店CSS类名（用于门店标识颜色）
function getStoreClass(storeCode) {
    if (!storeCode) return '';
    const code = storeCode.toUpperCase();
    switch (code) {
        case 'HQ': return 'store-hq';   // 总店 - 红色
        case 'WJ': return 'store-wj';   // 望京 - 蓝色
        case 'GM': return 'store-gm';   // 国贸 - 绿色
        default: return '';
    }
}

// 获取岗位CSS类名（用于岗位标签颜色）
function getPositionClass(position) {
    if (!position) return '';
    if (position.includes('服务员')) return 'pos-waiter';
    if (position.includes('厨师')) return 'pos-chef';
    if (position.includes('收银')) return 'pos-cashier';
    return '';
}

// 计算满足率
function calculateFulfillmentRate(assignments, requirements) {
    if (!requirements || requirements.length === 0) return 100;
    const totalRequired = requirements.reduce((sum, r) => sum + (r.min_employees || 1), 0);
    const fulfilled = assignments.length;
    return Math.round((fulfilled / totalRequired) * 100);
}

// 显示Toast通知
function showToast(message, type = 'info', duration = 3000) {
    const container = document.getElementById('toastContainer');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    
    const icons = {
        success: '✅',
        warning: '⚠️',
        error: '❌',
        info: 'ℹ️'
    };
    
    toast.innerHTML = `
        <span class="toast-icon">${icons[type]}</span>
        <span class="toast-message">${message}</span>
        <button class="toast-close" onclick="this.parentElement.remove()">×</button>
    `;
    
    container.appendChild(toast);
    
    setTimeout(() => {
        toast.style.animation = 'slideInRight 0.3s ease reverse';
        setTimeout(() => toast.remove(), 300);
    }, duration);
}

// 显示确认对话框
function showConfirm(title, message) {
    return new Promise((resolve) => {
        const result = confirm(`${title}\n\n${message}`);
        resolve(result);
    });
}

// 打开弹窗
function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.add('active');
        document.body.style.overflow = 'hidden';
    }
}

// 关闭弹窗
function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.remove('active');
        document.body.style.overflow = '';
    }
}

// 关闭所有弹窗
function closeAllModals() {
    document.querySelectorAll('.modal.active').forEach(modal => {
        modal.classList.remove('active');
    });
    document.body.style.overflow = '';
}

// 获取表单数据
function getFormData(formId) {
    const form = document.getElementById(formId);
    if (!form) return {};
    
    const data = {};
    const inputs = form.querySelectorAll('input, select, textarea');
    inputs.forEach(input => {
        if (input.id) {
            if (input.type === 'checkbox') {
                data[input.id] = input.checked;
            } else if (input.type === 'number') {
                data[input.id] = Number(input.value);
            } else {
                data[input.id] = input.value;
            }
        }
    });
    return data;
}

// 设置表单数据
function setFormData(data) {
    Object.keys(data).forEach(key => {
        const input = document.getElementById(key);
        if (input) {
            if (input.type === 'checkbox') {
                input.checked = data[key];
            } else {
                input.value = data[key];
            }
        }
    });
}

// 清空表单
function clearForm(formSelector) {
    const form = document.querySelector(formSelector);
    if (form) {
        form.querySelectorAll('input, select, textarea').forEach(input => {
            if (input.type === 'checkbox') {
                input.checked = false;
            } else if (input.type !== 'hidden') {
                input.value = '';
            }
        });
    }
}

// 防抖函数
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// 节流函数
function throttle(func, limit) {
    let inThrottle;
    return function(...args) {
        if (!inThrottle) {
            func.apply(this, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

// 深拷贝
function deepClone(obj) {
    return JSON.parse(JSON.stringify(obj));
}

// 数组去重
function unique(arr, key) {
    if (key) {
        const seen = new Set();
        return arr.filter(item => {
            const k = item[key];
            if (seen.has(k)) return false;
            seen.add(k);
            return true;
        });
    }
    return [...new Set(arr)];
}

// 根据键分组
function groupBy(arr, key) {
    return arr.reduce((groups, item) => {
        const k = typeof key === 'function' ? key(item) : item[key];
        (groups[k] = groups[k] || []).push(item);
        return groups;
    }, {});
}

// 排序
function sortBy(arr, key, desc = false) {
    return [...arr].sort((a, b) => {
        const va = typeof key === 'function' ? key(a) : a[key];
        const vb = typeof key === 'function' ? key(b) : b[key];
        if (va < vb) return desc ? 1 : -1;
        if (va > vb) return desc ? -1 : 1;
        return 0;
    });
}

// 格式化时间范围
function formatTimeRange(start, end) {
    return `${start} - ${end}`;
}

// 解析时间字符串为分钟
function parseTimeToMinutes(timeStr) {
    const [h, m] = timeStr.split(':').map(Number);
    return h * 60 + m;
}

// 检查时间冲突
function hasTimeConflict(shift1, shift2) {
    const s1Start = parseTimeToMinutes(shift1.startTime);
    const s1End = parseTimeToMinutes(shift1.endTime);
    const s2Start = parseTimeToMinutes(shift2.startTime);
    const s2End = parseTimeToMinutes(shift2.endTime);
    
    return !(s1End <= s2Start || s2End <= s1Start);
}

// 获取颜色的亮度
function getLuminance(hexColor) {
    const hex = hexColor.replace('#', '');
    const r = parseInt(hex.substr(0, 2), 16) / 255;
    const g = parseInt(hex.substr(2, 2), 16) / 255;
    const b = parseInt(hex.substr(4, 2), 16) / 255;
    return 0.299 * r + 0.587 * g + 0.114 * b;
}

// 根据背景色获取合适的文字颜色
function getContrastColor(hexColor) {
    return getLuminance(hexColor) > 0.5 ? '#000000' : '#ffffff';
}

// 格式化数字（添加千分位）
function formatNumber(num) {
    return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

// 生成颜色（基于字符串）
function stringToColor(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    const h = hash % 360;
    return `hsl(${h}, 70%, 50%)`;
}
