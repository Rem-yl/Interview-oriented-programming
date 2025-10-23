#!/bin/bash

# 会话管理方案 - 批量停止所有服务器
# 用法: ./stop_all_servers.sh [sticky|redis|jwt|all]

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# PID 文件目录
PID_DIR="$SCRIPT_DIR/pids"

# 打印带颜色的消息
log_info() {
    echo -e "${BLUE}ℹ${NC}  $1"
}

log_success() {
    echo -e "${GREEN}✅${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠️${NC}  $1"
}

log_error() {
    echo -e "${RED}❌${NC} $1"
}

# 停止指定端口的进程
stop_by_port() {
    local port=$1
    local service_name=$2

    local pid=$(lsof -ti:$port 2>/dev/null)

    if [ -z "$pid" ]; then
        log_warning "$service_name (端口 $port) 未运行"
        return 0
    fi

    log_info "停止 $service_name (端口 $port, PID: $pid)..."
    kill $pid 2>/dev/null || kill -9 $pid 2>/dev/null

    # 等待进程结束
    local max_attempts=10
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if ! ps -p $pid > /dev/null 2>&1; then
            log_success "$service_name 已停止"
            return 0
        fi
        sleep 0.5
        attempt=$((attempt + 1))
    done

    log_error "$service_name 停止失败"
    return 1
}

# 停止 PID 文件中的进程
stop_by_pid_file() {
    local pid_file=$1
    local service_name=$2

    if [ ! -f "$pid_file" ]; then
        return 0
    fi

    local pid=$(cat "$pid_file")

    if ! ps -p $pid > /dev/null 2>&1; then
        rm -f "$pid_file"
        return 0
    fi

    log_info "停止 $service_name (PID: $pid)..."
    kill $pid 2>/dev/null || kill -9 $pid 2>/dev/null

    # 等待进程结束
    sleep 1

    if ! ps -p $pid > /dev/null 2>&1; then
        log_success "$service_name 已停止"
        rm -f "$pid_file"
    else
        log_error "$service_name 停止失败"
    fi
}

# 停止 Sticky Session 服务器
stop_sticky_session() {
    log_info "停止 Sticky Session 服务器..."

    local ports=(8081 8082 8083)
    local server_ids=("server-1" "server-2" "server-3")

    for i in {0..2}; do
        local port=${ports[$i]}
        local server_id=${server_ids[$i]}

        stop_by_port $port "Sticky Session $server_id"

        # 删除 PID 文件
        rm -f "$PID_DIR/sticky-$server_id.pid"
    done
}

# 停止 Redis Session 服务器
stop_redis_session() {
    log_info "停止 Redis Session 服务器..."

    local ports=(8091 8092 8093)
    local server_ids=("server-1" "server-2" "server-3")

    for i in {0..2}; do
        local port=${ports[$i]}
        local server_id=${server_ids[$i]}

        stop_by_port $port "Redis Session $server_id"

        # 删除 PID 文件
        rm -f "$PID_DIR/redis-$server_id.pid"
    done
}

# 停止 JWT Token 服务器
stop_jwt_token() {
    log_info "停止 JWT Token 服务器..."

    local ports=(8010 8011 8012)
    local server_ids=("server-1" "server-2" "server-3")

    for i in {0..2}; do
        local port=${ports[$i]}
        local server_id=${server_ids[$i]}

        stop_by_port $port "JWT Token $server_id"

        # 删除 PID 文件
        rm -f "$PID_DIR/jwt-$server_id.pid"
    done
}

# 停止 Redis
stop_redis() {
    log_info "停止 Redis..."

    if docker ps | grep -q redis; then
        docker stop redis >/dev/null 2>&1
        log_success "Redis 已停止"
    else
        log_warning "Redis 未运行"
    fi
}

# 清理所有 Go 进程 (危险操作，慎用)
force_cleanup() {
    log_warning "强制清理所有 Go 进程..."

    # 查找所有 go run 进程
    local pids=$(ps aux | grep "go run main.go" | grep -v grep | awk '{print $2}')

    if [ -z "$pids" ]; then
        log_info "没有找到运行中的 Go 进程"
        return 0
    fi

    for pid in $pids; do
        log_info "杀死进程 $pid"
        kill -9 $pid 2>/dev/null || true
    done

    log_success "清理完成"
}

# 主函数
main() {
    local mode=${1:-all}

    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║         会话管理方案 - 批量停止服务器                      ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""

    case $mode in
        sticky)
            stop_sticky_session
            ;;
        redis)
            stop_redis_session
            ;;
        jwt)
            stop_jwt_token
            ;;
        all)
            stop_sticky_session
            echo ""
            stop_redis_session
            echo ""
            stop_jwt_token
            ;;
        redis-docker)
            stop_redis
            ;;
        force)
            force_cleanup
            ;;
        *)
            log_error "未知模式: $mode"
            echo ""
            echo "用法: $0 [sticky|redis|jwt|all|redis-docker|force]"
            echo ""
            echo "选项:"
            echo "  sticky       - 只停止 Sticky Session 服务器"
            echo "  redis        - 只停止 Redis Session 服务器"
            echo "  jwt          - 只停止 JWT Token 服务器"
            echo "  all          - 停止所有服务器 (默认)"
            echo "  redis-docker - 停止 Redis Docker 容器"
            echo "  force        - 强制杀死所有 Go 进程 (慎用)"
            echo ""
            exit 1
            ;;
    esac

    echo ""
    log_success "停止操作完成！"
    echo ""

    # 提示检查状态
    log_info "检查服务器状态:"
    echo "  ./start_all_servers.sh status"
    echo ""
}

# 运行主函数
main "$@"
