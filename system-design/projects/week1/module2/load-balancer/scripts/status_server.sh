#!/bin/bash

# 负载均衡器测试 - 服务器状态查看脚本
# 用途: 查看所有后端服务器的运行状态

# PID 文件目录
PID_DIR="./pids"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  后端服务器状态${NC}"
echo -e "${YELLOW}========================================${NC}"

# 检查 PID 目录是否存在
if [ ! -d "$PID_DIR" ] || [ -z "$(ls -A $PID_DIR 2>/dev/null)" ]; then
    echo -e "${YELLOW}没有找到运行中的服务器${NC}"
    echo -e "\n提示: 使用 ${GREEN}./start_server.sh${NC} 启动服务器"
    exit 0
fi

# 显示服务器状态
RUNNING_COUNT=0
TOTAL_COUNT=0

printf "\n%-8s %-10s %-10s %-20s\n" "端口" "PID" "状态" "健康检查"
printf "%-8s %-10s %-10s %-20s\n" "----" "----" "----" "----------"

for pid_file in "$PID_DIR"/server-*.pid; do
    if [ -f "$pid_file" ]; then
        PID=$(cat "$pid_file")
        PORT=$(basename "$pid_file" .pid | sed 's/server-//')
        ((TOTAL_COUNT++))

        # 检查进程状态
        if ps -p $PID > /dev/null 2>&1; then
            STATUS="${GREEN}运行中${NC}"

            # 健康检查
            if curl -s "http://localhost:$PORT/health" > /dev/null 2>&1; then
                HEALTH="${GREEN}✓ 正常${NC}"
                ((RUNNING_COUNT++))
            else
                HEALTH="${RED}✗ 失败${NC}"
            fi
        else
            STATUS="${RED}已停止${NC}"
            HEALTH="${RED}✗ 不可用${NC}"
        fi

        printf "%-8s %-10s %-20s %-30s\n" "$PORT" "$PID" "$(echo -e $STATUS)" "$(echo -e $HEALTH)"
    fi
done

echo -e "\n${YELLOW}========================================${NC}"
echo -e "总计: ${GREEN}$RUNNING_COUNT${NC}/$TOTAL_COUNT 运行中"
echo -e "${YELLOW}========================================${NC}"

# 显示管理提示
echo -e "\n${YELLOW}管理命令:${NC}"
echo -e "  查看日志: ${GREEN}tail -f ./logs/server-<端口>.log${NC}"
echo -e "  停止所有: ${GREEN}./stop_server.sh${NC}"
echo -e "  重启所有: ${GREEN}./stop_server.sh && ./start_server.sh${NC}"
