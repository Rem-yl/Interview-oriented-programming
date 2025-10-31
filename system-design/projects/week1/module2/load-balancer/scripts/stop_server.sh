#!/bin/bash

# 负载均衡器测试 - 服务器停止脚本
# 用途: 停止所有运行中的后端服务器实例

set -e

# PID 文件目录
PID_DIR="./pids"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  停止所有服务器实例${NC}"
echo -e "${YELLOW}========================================${NC}"

# 检查 PID 目录是否存在
if [ ! -d "$PID_DIR" ]; then
    echo -e "${YELLOW}没有找到运行中的服务器${NC}"
    exit 0
fi

# 停止所有服务器
STOPPED_COUNT=0
for pid_file in "$PID_DIR"/server-*.pid; do
    if [ -f "$pid_file" ]; then
        PID=$(cat "$pid_file")
        PORT=$(basename "$pid_file" .pid | sed 's/server-//')

        # 检查进程是否存在
        if ps -p $PID > /dev/null 2>&1; then
            # 优雅关闭 (SIGTERM)
            kill $PID 2>/dev/null || true

            # 等待进程结束（最多 5 秒）
            for i in {1..10}; do
                if ! ps -p $PID > /dev/null 2>&1; then
                    break
                fi
                sleep 0.5
            done

            # 如果还在运行，强制关闭 (SIGKILL)
            if ps -p $PID > /dev/null 2>&1; then
                echo -e "${YELLOW}⚠ 端口 $PORT: 优雅关闭失败，强制关闭...${NC}"
                kill -9 $PID 2>/dev/null || true
            fi

            echo -e "${GREEN}✓ 已停止服务器: 端口 $PORT (PID: $PID)${NC}"
            ((STOPPED_COUNT++))
        else
            echo -e "${YELLOW}⚠ 端口 $PORT: 进程不存在 (PID: $PID)${NC}"
        fi

        # 删除 PID 文件
        rm -f "$pid_file"
    fi
done

echo -e "\n${GREEN}已停止 $STOPPED_COUNT 个服务器实例${NC}"

# 清理空目录
if [ -d "$PID_DIR" ] && [ -z "$(ls -A $PID_DIR)" ]; then
    rmdir "$PID_DIR"
fi

echo -e "${YELLOW}========================================${NC}"
