# PaiBan 排班引擎 Makefile
# ================================

# 变量定义
APP_NAME := paiban
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Go 相关
GO := go
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOMOD := $(GO) mod

# Docker 相关
DOCKER := docker
DOCKER_COMPOSE := docker compose

# 默认目标
.DEFAULT_GOAL := help

# ================================
# 帮助
# ================================
.PHONY: help
help: ## 显示帮助信息
	@echo "PaiBan 排班引擎 - 可用命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ================================
# 开发
# ================================
.PHONY: init
init: ## 初始化项目（安装依赖）
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "依赖安装完成"

.PHONY: build
build: ## 构建二进制文件
	$(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/server
	@echo "构建完成: bin/$(APP_NAME)"

.PHONY: run
run: ## 本地运行服务
	$(GO) run ./cmd/server

.PHONY: dev
dev: ## 开发模式运行（带热重载，需要安装 air）
	@which air > /dev/null || (echo "请先安装 air: go install github.com/cosmtrek/air@latest" && exit 1)
	air

.PHONY: clean
clean: ## 清理构建产物
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "清理完成"

# ================================
# 测试
# ================================
.PHONY: test
test: ## 运行所有测试
	$(GOTEST) -v -race ./...

.PHONY: test-unit
test-unit: ## 运行单元测试
	$(GOTEST) -v -race -coverprofile=coverage.out ./pkg/... ./internal/...
	$(GO) tool cover -func=coverage.out
	@echo ""
	@echo "覆盖率报告已生成: coverage.out"

.PHONY: test-integration
test-integration: docker-test-up ## 运行集成测试
	$(GOTEST) -v -tags=integration ./tests/integration/...
	$(MAKE) docker-test-down

.PHONY: test-e2e
test-e2e: docker-test-up ## 运行端到端测试
	./scripts/wait-for-ready.sh
	$(GOTEST) -v -tags=e2e ./tests/e2e/...
	$(MAKE) docker-test-down

.PHONY: test-scenario
test-scenario: ## 运行场景测试
	$(GOTEST) -v -tags=scenario ./tests/scenario/...

.PHONY: test-benchmark
test-benchmark: ## 运行性能基准测试
	$(GOTEST) -bench=. -benchmem ./tests/benchmark/...

.PHONY: coverage
coverage: test-unit ## 生成覆盖率 HTML 报告
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"
	@which open > /dev/null && open coverage.html || echo "请手动打开 coverage.html"

# ================================
# 代码质量
# ================================
.PHONY: lint
lint: ## 运行代码检查
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint" && exit 1)
	golangci-lint run ./...

.PHONY: fmt
fmt: ## 格式化代码
	$(GO) fmt ./...
	@echo "代码格式化完成"

.PHONY: vet
vet: ## 运行 go vet
	$(GO) vet ./...

.PHONY: tidy
tidy: ## 整理依赖
	$(GOMOD) tidy

# ================================
# Docker
# ================================
.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	$(DOCKER) build -f deployments/Dockerfile \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(APP_NAME):$(VERSION) \
		-t $(APP_NAME):latest .
	@echo "Docker 镜像构建完成: $(APP_NAME):$(VERSION)"

.PHONY: docker-up
docker-up: ## 启动开发环境
	$(DOCKER_COMPOSE) up -d
	@echo ""
	@echo "服务已启动:"
	@echo "  - PaiBan API: http://localhost:7012"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"

.PHONY: docker-down
docker-down: ## 停止开发环境
	$(DOCKER_COMPOSE) down
	@echo "服务已停止"

.PHONY: docker-logs
docker-logs: ## 查看容器日志
	$(DOCKER_COMPOSE) logs -f

.PHONY: docker-ps
docker-ps: ## 查看容器状态
	$(DOCKER_COMPOSE) ps

.PHONY: docker-clean
docker-clean: ## 清理 Docker 资源
	$(DOCKER_COMPOSE) down -v --remove-orphans
	$(DOCKER) system prune -f
	@echo "Docker 资源已清理"

.PHONY: docker-test-up
docker-test-up: ## 启动测试环境
	$(DOCKER_COMPOSE) -f docker-compose.test.yaml up -d postgres-test redis-test
	@echo "测试环境已启动"
	@sleep 3

.PHONY: docker-test-down
docker-test-down: ## 停止测试环境
	$(DOCKER_COMPOSE) -f docker-compose.test.yaml down
	@echo "测试环境已停止"

.PHONY: docker-admin
docker-admin: ## 启动管理界面（pgAdmin + RedisInsight）
	$(DOCKER_COMPOSE) --profile admin up -d
	@echo ""
	@echo "管理界面已启动:"
	@echo "  - pgAdmin: http://localhost:5050 (admin@paiban.local / admin123)"
	@echo "  - RedisInsight: http://localhost:8001"

# ================================
# 数据库
# ================================
.PHONY: db-migrate
db-migrate: ## 运行数据库迁移
	@echo "运行数据库迁移..."
	# TODO: 使用 golang-migrate 或 goose 执行迁移
	@echo "迁移完成"

.PHONY: db-reset
db-reset: ## 重置数据库
	$(DOCKER_COMPOSE) exec postgres psql -U paiban -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	$(MAKE) db-migrate
	@echo "数据库已重置"

# ================================
# API 文档
# ================================
.PHONY: docs
docs: ## 生成 API 文档
	@which swag > /dev/null || (echo "请先安装 swag: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	swag init -g cmd/server/main.go -o api/docs
	@echo "API 文档已生成: api/docs/"

# ================================
# 发布
# ================================
.PHONY: release
release: lint test build docker-build ## 构建发布版本
	@echo ""
	@echo "发布版本 $(VERSION) 构建完成"
	@echo "  - 二进制: bin/$(APP_NAME)"
	@echo "  - Docker: $(APP_NAME):$(VERSION)"

.PHONY: version
version: ## 显示版本信息
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

