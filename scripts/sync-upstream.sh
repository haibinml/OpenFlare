#!/usr/bin/env bash
# 从 Wavelet 上游同步 Cordis 内核与平台插件到本仓库。
#
# 只覆盖与上游同构的三个目录：backend/core、backend/pkg、backend/plugins。
# 永不触碰下游自有内容：backend/OpenFlare（业务插件）、backend/share（共享层）、
# backend/cmd 与 backend/main.go（本仓库自持的装配根）。
#
# 用法：
#   scripts/sync-upstream.sh [上游 backend 路径]
#   scripts/sync-upstream.sh --check     # 仅报告差异，不写入
set -euo pipefail

CHECK=0
SRC="/Users/ryan/Code/Go/Wavelet/backend"
if [[ "${1:-}" == "--check" ]]; then
	CHECK=1
	shift
fi
if [[ -n "${1:-}" ]]; then
	SRC="$1"
fi

if [[ ! -d "$SRC/core" ]]; then
	echo "error: 上游 backend 路径无效：$SRC" >&2
	exit 1
fi

DST="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/backend"

RSYNC_OPTS=(-a --delete)
if [[ "$CHECK" == "1" ]]; then
	RSYNC_OPTS+=(-n -i)
	echo "==> dry-run：仅列出将变更的文件"
fi

# 上游工作区里被 gitignore 的运行期产物绝不允许带进本仓库（曾经误提交
# uploads/diskcache 缓存块与 driver_http/dist 前端构建物共 1000+ 文件）。
EXCLUDES=(--exclude 'uploads' --exclude 'dist' --exclude 'data' --exclude '*.db' --exclude '.DS_Store')

for dir in core pkg plugins; do
	echo "==> sync $dir"
	rsync "${RSYNC_OPTS[@]}" "${EXCLUDES[@]}" "$SRC/$dir/" "$DST/$dir/" | sed "s|^|$dir/|"
done

PATCHES="$DST/OpenFlare/upstream-patches.md"

if [[ "$CHECK" == "1" ]]; then
	echo "==> 检查完成（未写入）。以上 >f/<f 行即为与上游的差异。"
else
	echo "==> 同步完成。"
	if [[ -f "$PATCHES" ]]; then
		echo "注意：以下上游文件带有本地补丁，同步后请确认补丁仍在（cd backend && go build ./... && go test ./core/...）："
		sed -n 's/^- `\(.*\)`$/\t\1/p' "$PATCHES"
	fi
fi
