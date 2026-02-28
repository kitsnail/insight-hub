.PHONY: build run test clean fmt lint help

# 变量
BINARY_NAME := insight-hub
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := bin
MAIN_PATH := ./cmd/insight-hub

# 默认目标
.DEFAULT_GOAL := help

## build: 编译项目
build:
	@echo "编译 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

## run: 运行项目
run:
	go run $(MAIN_PATH)

## run-with-config: 使用指定配置运行
run-with-config:
	go run $(MAIN_PATH) -config ~/.insight-hub/config.yaml

## test: 运行测试
test:
	go test -v -race ./...

## test-cover: 运行测试并生成覆盖率报告
test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"

## clean: 清理编译产物
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "清理完成"

## fmt: 格式化代码
fmt:
	go fmt ./...

## lint: 代码检查
lint:
	@which golangci-lint > /dev/null || (echo "请安装 golangci-lint" && exit 1)
	golangci-lint run ./...

## tidy: 整理依赖
tidy:
	go mod tidy

## deps: 安装依赖
deps:
	go mod download

## db-init: 初始化数据库
db-init:
	@echo "初始化数据库..."
	@mkdir -p ~/.insight-hub/data
	@cp config.example.yaml ~/.insight-hub/config.yaml 2>/dev/null || true
	@echo "数据库初始化完成"

## help: 显示帮助信息
help:
	@echo "Insight Hub - 个人知识管理工具"
	@echo ""
	@echo "使用方法:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
