#!/bin/bash

# 快速检查所有服务器状态

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║              服务器状态检查                                 ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# 检查端口是否监听
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1 ; then
        return 0  # 端口被监听
    else
        return 1  # 端口未监听
    fi
}

# 检查 HTTP 服务是否响应
check_http() {
    local url=$1
    if curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null | grep -q "200\|401"; then
        return 0  # 服务正常响应
    else
        return 1  # 服务无响应
    fi
}

# 打印表头
printf "${BLUE}%-20s %-10s %-12s %-15s %-10s${NC}\n" "方案" "端口" "端口状态" "HTTP状态" "PID"
echo "--------------------------------------------------------------------------------"

# 检查 Sticky Session
for port in 8081 8082 8083; do
    port_status="❌ 未监听"
    http_status="❌ 无响应"
    pid="-"

    if check_port $port; then
        port_status="${GREEN}✅ 监听中${NC}"
        pid=$(lsof -ti:$port)

        if check_http "http://localhost:$port/profile"; then
            http_status="${GREEN}✅ 正常${NC}"
        else
            http_status="${YELLOW}⚠️  异常${NC}"
        fi
    else
        port_status="${RED}❌ 未监听${NC}"
        http_status="${RED}❌ 无响应${NC}"
    fi

    printf "%-20s %-10s %-22s %-25s %-10s\n" "Sticky Session" "$port" "$port_status" "$http_status" "$pid"
done

# 检查 Redis Session
for port in 8091 8092 8093; do
    port_status="❌ 未监听"
    http_status="❌ 无响应"
    pid="-"

    if check_port $port; then
        port_status="${GREEN}✅ 监听中${NC}"
        pid=$(lsof -ti:$port)

        if check_http "http://localhost:$port/profile"; then
            http_status="${GREEN}✅ 正常${NC}"
        else
            http_status="${YELLOW}⚠️  异常${NC}"
        fi
    else
        port_status="${RED}❌ 未监听${NC}"
        http_status="${RED}❌ 无响应${NC}"
    fi

    printf "%-20s %-10s %-22s %-25s %-10s\n" "Redis Session" "$port" "$port_status" "$http_status" "$pid"
done

# 检查 JWT Token
for port in 8010 8011 8012; do
    port_status="❌ 未监听"
    http_status="❌ 无响应"
    pid="-"

    if check_port $port; then
        port_status="${GREEN}✅ 监听中${NC}"
        pid=$(lsof -ti:$port)

        if check_http "http://localhost:$port/profile"; then
            http_status="${GREEN}✅ 正常${NC}"
        else
            http_status="${YELLOW}⚠️  异常${NC}"
        fi
    else
        port_status="${RED}❌ 未监听${NC}"
        http_status="${RED}❌ 无响应${NC}"
    fi

    printf "%-20s %-10s %-22s %-25s %-10s\n" "JWT Token" "$port" "$port_status" "$http_status" "$pid"
done

# 检查 Redis
echo "--------------------------------------------------------------------------------"
if docker ps | grep -q redis; then
    redis_status="${GREEN}✅ 运行中${NC}"
    if redis-cli ping >/dev/null 2>&1; then
        redis_ping="${GREEN}✅ PONG${NC}"
    else
        redis_ping="${RED}❌ 无响应${NC}"
    fi
    printf "%-20s %-10s %-22s %-25s %-10s\n" "Redis (Docker)" "6379" "$redis_status" "$redis_ping" "docker"
else
    redis_status="${RED}❌ 已停止${NC}"
    redis_ping="${RED}❌ 无响应${NC}"
    printf "%-20s %-10s %-22s %-25s %-10s\n" "Redis (Docker)" "6379" "$redis_status" "$redis_ping" "-"
fi

echo ""

# 统计
sticky_count=$(lsof -Pi :8081,:8082,:8083 -sTCP:LISTEN -t 2>/dev/null | wc -l | tr -d ' ')
redis_count=$(lsof -Pi :8091,:8092,:8093 -sTCP:LISTEN -t 2>/dev/null | wc -l | tr -d ' ')
jwt_count=$(lsof -Pi :8010,:8011,:8012 -sTCP:LISTEN -t 2>/dev/null | wc -l | tr -d ' ')

echo -e "${BLUE}统计:${NC}"
echo "  Sticky Session: $sticky_count/3 运行中"
echo "  Redis Session:  $redis_count/3 运行中"
echo "  JWT Token:      $jwt_count/3 运行中"

echo ""
echo -e "${BLUE}快捷命令:${NC}"
echo "  启动所有服务器:     ./start_all_servers.sh"
echo "  停止所有服务器:     ./stop_all_servers.sh"
echo "  查看日志:           tail -f logs/sticky-server-1.log"
echo "  运行性能测试:       cd test-scripts && python performance_compare.py"
echo ""
