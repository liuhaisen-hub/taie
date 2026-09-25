APP_NAME := taie
VITE_PORT ?= 9245

.PHONY: help run build migrate bindings wire clean

help: ## 列出所有可用目标
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-8s\033[0m %s\n", $$1, $$2}'

run: ## 本地运行（dev 模式，前端热重载）
	wails3 dev -config ./build/config.yml -port $(VITE_PORT)

build: ## 打包生产版本（构建并生成 bin/$(APP_NAME).app）
	wails3 task package

migrate: ## 执行数据库初始化（执行 migrations/ 下的 SQL 脚本，可重复执行）
	go run . migrate

bindings: ## 修改服务方法后重新生成前端 bindings（TypeScript interface）
	wails3 generate bindings -ts

wire: ## 修改 provider 后重新生成 wire_gen.go
	wire

clean: ## 清理构建产物（bin/）
	rm -rf bin
