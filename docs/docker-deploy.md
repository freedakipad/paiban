# Docker 部署指南

本文档详细介绍如何使用 Docker 和 Docker Compose 部署 PaiBan 排班引擎。

## 目录

- [快速开始](#快速开始)
- [开发环境部署](#开发环境部署)
- [生产环境部署](#生产环境部署)
- [前端控制台部署](#前端控制台部署)
- [Kubernetes 部署](#kubernetes-部署)
- [常用命令](#常用命令)
- [故障排除](#故障排除)

## 快速开始

### 最简部署（仅后端服务）

```bash
# 1. 克隆项目（请替换为实际仓库地址）
git clone <your-repo-url>
cd paiban

# 2. 启动所有服务
docker compose up -d

# 3. 检查服务状态
docker compose ps

# 4. 测试服务
curl http://localhost:7012/health
```

### 服务端口

| 服务 | 端口 | 描述 |
|------|------|------|
| PaiBan API | 7012 | 排班引擎 API |
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存 |
| pgAdmin | 5050 | 数据库管理 (可选) |
| RedisInsight | 8001 | Redis 管理 (可选) |

## 开发环境部署

### 启动基础服务

```bash
# 仅启动数据库和缓存（用于本地开发）
docker compose up -d postgres redis

# 检查服务状态
docker compose ps
```

### 启动完整服务

```bash
# 启动所有基础服务 + PaiBan 后端
docker compose up -d

# 查看日志
docker compose logs -f paiban
```

### 启动管理界面

```bash
# 包含 pgAdmin 和 RedisInsight
docker compose --profile admin up -d

# pgAdmin 访问: http://localhost:5050
# 账号: admin@paiban.local / admin123

# RedisInsight 访问: http://localhost:8001
```

### 开发模式热重载

本地开发时，建议只启动数据库和缓存，本地运行 Go 服务：

```bash
# 1. 启动依赖服务
docker compose up -d postgres redis

# 2. 本地编译运行
go build -o bin/paiban cmd/server/main.go
./bin/paiban

# 或使用 air 热重载
go install github.com/air-verse/air@latest
air
```

## 生产环境部署

### 创建生产配置

创建 `docker-compose.prod.yaml`：

```yaml
# docker-compose.prod.yaml
services:
  paiban:
    image: paiban/paiban:latest
    container_name: paiban-server
    restart: always
    environment:
      APP_ENV: production
      APP_PORT: 7012
      APP_LOG_LEVEL: info
      DB_HOST: ${DB_HOST}
      DB_PORT: ${DB_PORT}
      DB_NAME: ${DB_NAME}
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_SSL_MODE: require
      REDIS_HOST: ${REDIS_HOST}
      REDIS_PORT: ${REDIS_PORT}
      API_RATE_LIMIT: 100
      API_TIMEOUT: 30s
    ports:
      - "7012:7012"
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:7012/health"]
      interval: 30s
      timeout: 5s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 256M
```

### 构建生产镜像

```bash
# 构建带版本标签的镜像
docker build \
  -f deployments/Dockerfile \
  --build-arg VERSION=v1.0.0 \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  -t paiban/paiban:v1.0.0 \
  -t paiban/paiban:latest \
  .

# 推送到镜像仓库
docker push paiban/paiban:v1.0.0
docker push paiban/paiban:latest
```

### 生产环境启动

```bash
# 创建环境变量文件
cat > .env.prod << EOF
DB_HOST=your-db-host
DB_PORT=5432
DB_NAME=paiban
DB_USER=paiban
DB_PASSWORD=your-secure-password
REDIS_HOST=your-redis-host
REDIS_PORT=6379
EOF

# 使用生产配置启动
docker compose -f docker-compose.prod.yaml --env-file .env.prod up -d
```

## 前端控制台部署

### 方式一：使用 Nginx

创建 `frontend/nginx.conf`：

```nginx
server {
    listen 80;
    server_name localhost;
    
    root /usr/share/nginx/html;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    # 代理 API 请求到后端
    location /api/ {
        proxy_pass http://paiban:7012;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

添加到 `docker-compose.yaml`：

```yaml
services:
  # ... 其他服务

  frontend:
    image: nginx:alpine
    container_name: paiban-frontend
    restart: unless-stopped
    ports:
      - "8888:80"
    volumes:
      - ./frontend:/usr/share/nginx/html:ro
      - ./frontend/nginx.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
      - paiban
    networks:
      - paiban-network
```

### 方式二：独立 Docker 构建

创建 `frontend/Dockerfile`：

```dockerfile
FROM nginx:alpine

COPY index.html /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

构建并运行：

```bash
cd frontend
docker build -t paiban/frontend:latest .
docker run -d -p 8888:80 --name paiban-frontend paiban/frontend:latest
```

## Kubernetes 部署

项目包含完整的 K8s 配置，位于 `deployments/k8s/` 目录。

### 部署步骤

```bash
# 1. 创建命名空间
kubectl apply -f deployments/k8s/namespace.yaml

# 2. 创建 ConfigMap 和 Secret
kubectl apply -f deployments/k8s/configmap.yaml
kubectl apply -f deployments/k8s/secret.yaml

# 3. 部署服务
kubectl apply -f deployments/k8s/deployment.yaml
kubectl apply -f deployments/k8s/service.yaml

# 4. 配置 Ingress（可选）
kubectl apply -f deployments/k8s/ingress.yaml

# 5. 配置 HPA 自动扩缩容（可选）
kubectl apply -f deployments/k8s/hpa.yaml

# 6. 配置 PDB 可用性保障（可选）
kubectl apply -f deployments/k8s/pdb.yaml
```

### 查看部署状态

```bash
# 查看 Pod 状态
kubectl get pods -n paiban

# 查看日志
kubectl logs -f deployment/paiban -n paiban

# 查看服务
kubectl get svc -n paiban
```

## 常用命令

### Docker Compose

```bash
# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 重启特定服务
docker compose restart paiban

# 查看日志
docker compose logs -f paiban

# 查看服务状态
docker compose ps

# 进入容器
docker compose exec paiban sh

# 重新构建镜像
docker compose build --no-cache paiban

# 清理未使用的资源
docker system prune -a
```

### 数据库管理

```bash
# 进入 PostgreSQL
docker compose exec postgres psql -U paiban -d paiban

# 备份数据库
docker compose exec postgres pg_dump -U paiban paiban > backup.sql

# 恢复数据库
docker compose exec -T postgres psql -U paiban paiban < backup.sql
```

### Redis 管理

```bash
# 进入 Redis CLI
docker compose exec redis redis-cli

# 查看所有键
docker compose exec redis redis-cli KEYS "*"

# 清空缓存
docker compose exec redis redis-cli FLUSHALL
```

## 故障排除

### 服务无法启动

```bash
# 查看详细日志
docker compose logs paiban

# 检查容器状态
docker compose ps -a

# 检查网络
docker network ls
docker network inspect paiban-network
```

### 数据库连接失败

```bash
# 检查 PostgreSQL 是否就绪
docker compose exec postgres pg_isready -U paiban

# 测试连接
docker compose exec paiban sh -c 'nc -zv postgres 5432'
```

### 端口冲突

```bash
# 检查端口占用
lsof -i :7012
lsof -i :5432
lsof -i :6379

# 修改端口映射
# 在 docker-compose.yaml 中修改 ports 配置
```

### 清理重建

```bash
# 停止并删除所有容器、网络、卷
docker compose down -v

# 删除所有镜像
docker compose down --rmi all

# 重新构建并启动
docker compose up -d --build
```

## 环境变量参考

| 变量 | 默认值 | 描述 |
|------|--------|------|
| APP_ENV | development | 运行环境 |
| APP_PORT | 7012 | 服务端口 |
| APP_LOG_LEVEL | info | 日志级别 |
| DB_HOST | postgres | 数据库主机 |
| DB_PORT | 5432 | 数据库端口 |
| DB_NAME | paiban | 数据库名称 |
| DB_USER | paiban | 数据库用户 |
| DB_PASSWORD | - | 数据库密码 |
| DB_SSL_MODE | disable | SSL 模式 |
| REDIS_HOST | redis | Redis 主机 |
| REDIS_PORT | 6379 | Redis 端口 |
| REDIS_DB | 0 | Redis 数据库索引 |
| API_RATE_LIMIT | 100 | 速率限制 |
| API_TIMEOUT | 30s | 请求超时 |
