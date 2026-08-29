#!/bin/sh
set -eu

# execute this first
# go install github.com/swaggo/swag/cmd/swag@latest

# 仅扫描 OpenFlare 自身代码：上游 core/ 与 plugins/ 尚未挂载进装配根（P4 接入），
# 且其包名（如 model、disk）与下游注解里的裸名引用冲突，扫描会导致 ParseComment 失败。
(
	cd backend
	swag init -o docs --parseDependency --parseInternal --exclude plugins,core
)
cp backend/docs/swagger.json backend/docs/swagger.yaml docs/
