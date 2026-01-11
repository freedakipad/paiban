# PaiBan Docker 快速部署

## 🚀 快速开始

### 一键启动

```bash
# 克隆并启动（请替换为实际仓库地址）
git clone <your-repo-url>
cd paiban
docker compose up -d

# 验证服务
curl http://localhost:7012/health
# 返回: {"status":"ok","timestamp":"..."}
```

### 服务端口

| 服务 | 端口 | 访问地址 |
|------|------|----------|
| **PaiBan API** | 7012 | http://localhost:7012 |
| PostgreSQL | 5432 | localhost:5432 |
| Redis | 6379 | localhost:6379 |

## 📋 常用命令

```bash
# 启动
docker compose up -d

# 停止
docker compose down

# 查看日志
docker compose logs -f paiban

# 重启
docker compose restart paiban

# 查看状态
docker compose ps
```

## 🔧 自定义配置

### 修改端口

编辑 `docker-compose.yaml`：

```yaml
services:
  paiban:
    ports:
      - "8080:7012"  # 改为 8080
```

### 环境变量

```bash
# 创建 .env 文件
cat > .env << EOF
DB_PASSWORD=your-password
API_RATE_LIMIT=200
APP_LOG_LEVEL=debug
EOF

# 启动时自动读取
docker compose up -d
```

## 🏭 生产部署

### 构建镜像

```bash
docker build -f deployments/Dockerfile -t paiban:latest .
```

### 生产配置

```yaml
# docker-compose.prod.yaml
services:
  paiban:
    image: paiban:latest
    restart: always
    environment:
      APP_ENV: production
      DB_HOST: your-db-host
      DB_PASSWORD: ${DB_PASSWORD}
      REDIS_HOST: your-redis-host
    ports:
      - "7012:7012"
    deploy:
      resources:
        limits:
          memory: 1G
```

```bash
docker compose -f docker-compose.prod.yaml up -d
```

## 🌐 前端部署

### 使用 Nginx

```bash
# 添加到 docker-compose.yaml
cat >> docker-compose.yaml << 'EOF'

  frontend:
    image: nginx:alpine
    ports:
      - "8888:80"
    volumes:
      - ./frontend:/usr/share/nginx/html:ro
    depends_on:
      - paiban
EOF

docker compose up -d frontend
```

访问：http://localhost:8888

## 🔍 故障排除

```bash
# 服务未启动
docker compose logs paiban

# 端口冲突
lsof -i :7012

# 数据库连接
docker compose exec postgres pg_isready -U paiban

# 完全重置
docker compose down -v
docker compose up -d --build
```

## 📊 管理工具（可选）

```bash
# 启动管理界面
docker compose --profile admin up -d

# pgAdmin: http://localhost:5050
# 账号: admin@paiban.local / admin123

# RedisInsight: http://localhost:8001
```

---

📖 完整文档：[docs/docker-deploy.md](docs/docker-deploy.md)
