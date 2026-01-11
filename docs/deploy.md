# PaiBan 部署指南

本文档介绍如何部署 PaiBan 排班引擎服务。

## 目录

- [环境要求](#环境要求)
- [快速启动](#快速启动)
- [生产部署](#生产部署)
- [前端部署](#前端部署)
- [服务管理](#服务管理)
- [配置说明](#配置说明)
- [故障排除](#故障排除)

## 环境要求

- **Go 1.23+** - 编译和运行
- **PostgreSQL 15+** - 数据存储（可选）
- **Redis 6+** - 缓存（可选）

> 注：PostgreSQL 和 Redis 为可选依赖，服务可在无数据库模式下运行。

## 快速启动

### 一键启动

```bash
# 克隆项目
git clone https://github.com/freedakipad/paiban.git
cd paiban

# 快速启动
./scripts/quick-start.sh
```

### 手动启动

```bash
# 1. 编译
go build -o bin/paiban cmd/server/main.go

# 2. 运行
./bin/paiban

# 3. 验证
curl http://localhost:7012/health
```

服务默认端口：`7012`

## 生产部署

### 1. 编译发布版本

```bash
# 编译优化版本
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=v1.0.0" \
    -trimpath \
    -o bin/paiban \
    cmd/server/main.go

# 查看文件大小
ls -lh bin/paiban
```

### 2. 创建配置文件

```bash
# 复制默认配置
cp configs/app.yaml /etc/paiban/app.yaml

# 编辑配置
vim /etc/paiban/app.yaml
```

### 3. 配置 Systemd 服务

创建服务文件 `/etc/systemd/system/paiban.service`：

```ini
[Unit]
Description=PaiBan Scheduling Engine
After=network.target

[Service]
Type=simple
User=paiban
Group=paiban
WorkingDirectory=/opt/paiban
ExecStart=/opt/paiban/bin/paiban
Restart=always
RestartSec=5

# 环境变量
Environment=APP_ENV=production
Environment=APP_PORT=7012
Environment=APP_LOG_LEVEL=info

# 安全限制
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/paiban

[Install]
WantedBy=multi-user.target
```

### 4. 启动服务

```bash
# 创建用户
sudo useradd -r -s /bin/false paiban

# 复制文件
sudo mkdir -p /opt/paiban/bin
sudo cp bin/paiban /opt/paiban/bin/
sudo cp -r configs /opt/paiban/

# 设置权限
sudo chown -R paiban:paiban /opt/paiban

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable paiban
sudo systemctl start paiban

# 查看状态
sudo systemctl status paiban
```

### 5. 配置 Nginx 反向代理（可选）

```nginx
server {
    listen 80;
    server_name paiban.example.com;

    location / {
        proxy_pass http://127.0.0.1:7012;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Request-ID $request_id;
    }
}
```

## 前端部署

### 方式一：静态文件服务

```bash
# 使用 Python
cd frontend
python3 -m http.server 8888

# 或使用 Node.js
npx serve -p 8888 frontend/
```

### 方式二：Nginx 托管

```nginx
server {
    listen 8888;
    root /opt/paiban/frontend;
    index index.html;

    location /api/ {
        proxy_pass http://127.0.0.1:7012;
    }
}
```

## 服务管理

### 使用 Systemd

```bash
# 启动
sudo systemctl start paiban

# 停止
sudo systemctl stop paiban

# 重启
sudo systemctl restart paiban

# 查看日志
sudo journalctl -u paiban -f
```

### 直接运行

```bash
# 前台运行
./bin/paiban

# 后台运行
nohup ./bin/paiban > /var/log/paiban/app.log 2>&1 &

# 停止
pkill -f 'bin/paiban'
```

## 配置说明

### 环境变量

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `APP_ENV` | development | 运行环境 |
| `APP_PORT` | 7012 | 服务端口 |
| `APP_LOG_LEVEL` | info | 日志级别 |
| `DB_HOST` | localhost | 数据库主机 |
| `DB_PORT` | 5432 | 数据库端口 |
| `DB_NAME` | paiban | 数据库名称 |
| `DB_USER` | paiban | 数据库用户 |
| `DB_PASSWORD` | - | 数据库密码 |
| `REDIS_HOST` | localhost | Redis 主机 |
| `REDIS_PORT` | 6379 | Redis 端口 |
| `API_RATE_LIMIT` | 100 | 速率限制 |
| `API_TIMEOUT` | 30s | 请求超时 |

### 配置文件

`configs/app.yaml`:

```yaml
app:
  env: production
  port: 7012
  log_level: info

database:
  host: localhost
  port: 5432
  name: paiban
  user: paiban
  password: your-password

redis:
  host: localhost
  port: 6379
  db: 0

scheduler:
  max_workers: 4
  default_timeout: 30s

api:
  rate_limit: 100
  timeout: 30s
```

## 故障排除

### 服务无法启动

```bash
# 检查端口占用
lsof -i :7012

# 查看日志
journalctl -u paiban -n 50

# 检查配置
./bin/paiban --config /etc/paiban/app.yaml --check
```

### 健康检查

```bash
# 基本健康检查
curl http://localhost:7012/health

# 详细信息
curl http://localhost:7012/api/v1/
```

### 性能调优

```bash
# 增加文件描述符限制
ulimit -n 65535

# 优化 Go 运行时
export GOMAXPROCS=4
export GOGC=100
```

## 升级指南

### 平滑升级

```bash
# 1. 编译新版本
go build -o bin/paiban.new cmd/server/main.go

# 2. 备份旧版本
cp bin/paiban bin/paiban.bak

# 3. 替换二进制
mv bin/paiban.new bin/paiban

# 4. 重启服务
sudo systemctl restart paiban

# 5. 验证
curl http://localhost:7012/health
```

### 回滚

```bash
# 还原备份
mv bin/paiban.bak bin/paiban
sudo systemctl restart paiban
```
