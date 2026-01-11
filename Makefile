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

.PHONY: build-linux
build-linux: ## 构建 Linux 二进制文件
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -trimpath -o bin/$(APP_NAME)-linux ./cmd/server
	@echo "Linux 构建完成: bin/$(APP_NAME)-linux"

.PHONY: run
run: ## 本地运行服务
	$(GO) run ./cmd/server

.PHONY: start
start: build ## 构建并启动服务
	./bin/$(APP_NAME)

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
test-integration: ## 运行集成测试
	$(GOTEST) -v -tags=integration ./tests/integration/...

.PHONY: test-e2e
test-e2e: ## 运行端到端测试
	$(GOTEST) -v -tags=e2e ./tests/e2e/...

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

.PHONY: check
check: fmt vet lint test ## 运行所有检查

# ================================
# 数据库
# ================================
.PHONY: db-migrate
db-migrate: ## 运行数据库迁移
	@echo "运行数据库迁移..."
	# TODO: 使用 golang-migrate 或 goose 执行迁移
	@echo "迁移完成"

# ================================
# API 文档
# ================================
.PHONY: docs
docs: ## 生成 API 文档
	@which swag > /dev/null || (echo "请先安装 swag: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	swag init -g cmd/server/main.go -o api/docs
	@echo "API 文档已生成: api/docs/"

# ================================
# 前端
# ================================
.PHONY: frontend
frontend: ## 启动前端控制台
	@echo "启动前端控制台: http://localhost:8888"
	cd frontend && python3 -m http.server 8888

# ================================
# 发布
# ================================
.PHONY: release
release: lint test build build-linux ## 构建发布版本
	@echo ""
	@echo "发布版本 $(VERSION) 构建完成"
	@echo "  - macOS: bin/$(APP_NAME)"
	@echo "  - Linux: bin/$(APP_NAME)-linux"

.PHONY: version
version: ## 显示版本信息
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

# ================================
# 部署
# ================================
.PHONY: install
install: build ## 安装到系统目录
	sudo mkdir -p /opt/paiban/bin
	sudo cp bin/$(APP_NAME) /opt/paiban/bin/
	sudo cp -r configs /opt/paiban/
	@echo "安装完成: /opt/paiban/"

.PHONY: deploy
deploy: build-linux ## 部署（需要配置目标服务器）
	@echo "请使用 scp 或其他工具将 bin/$(APP_NAME)-linux 部署到目标服务器"
	@echo "参考文档: docs/deploy.md"