# PaiBan 排班引擎 - 启动指南

## 快速启动

### 1. 启动后端服务 (端口 7012)

```bash
cd /Users/freedak/Documents/go-new/paiban
go run cmd/server/main.go
```

或在后台运行：
```bash
go run cmd/server/main.go 2>&1 &
```

### 2. 启动餐饮门店前端 (端口 8080)

```bash
cd /Users/freedak/Documents/go-new/paiban/restaurant-scheduler
python3 -m http.server 8080
```

或在后台运行：
```bash
python3 -m http.server 8080 2>&1 &
```

### 3. 访问地址

- **后端 API**: http://localhost:7012
- **健康检查**: http://localhost:7012/health
- **餐饮门店前端**: http://localhost:8080

## 端口管理

### 查看端口占用
```bash
lsof -i :7012 :8080
```

### 清空端口
```bash
# 清空后端端口
lsof -ti:7012 | xargs kill -9 2>/dev/null

# 清空前端端口
lsof -ti:8080 | xargs kill -9 2>/dev/null

# 一键清空所有
lsof -ti:7012 | xargs kill -9 2>/dev/null; lsof -ti:8080 | xargs kill -9 2>/dev/null
```

## 一键启动脚本

```bash
#!/bin/bash
# 清空端口
lsof -ti:7012 | xargs kill -9 2>/dev/null
lsof -ti:8080 | xargs kill -9 2>/dev/null

# 等待端口释放
sleep 1

# 启动后端
cd /Users/freedak/Documents/go-new/paiban
go run cmd/server/main.go 2>&1 &

# 等待后端启动
sleep 3

# 启动前端
cd /Users/freedak/Documents/go-new/paiban/restaurant-scheduler
python3 -m http.server 8080 2>&1 &

echo "✅ 服务已启动"
echo "   后端: http://localhost:7012"
echo "   前端: http://localhost:8080"
```

## 服务验证

### 检查后端健康状态
```bash
curl http://localhost:7012/health
```

### 检查前端
```bash
curl -I http://localhost:8080/
```

## 常见问题

### 端口被占用
```bash
# 查看占用端口的进程
lsof -i :7012
lsof -i :8080

# 强制杀掉进程
kill -9 <PID>
```

### 后端启动失败
1. 确保 Go 环境已安装
2. 确保在项目根目录下运行
3. 检查依赖: `go mod tidy`

### 前端启动失败
1. 确保 Python3 已安装
2. 确保在 restaurant-scheduler 目录下运行
