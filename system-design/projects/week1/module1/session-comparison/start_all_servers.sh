#!/bin/bash

# 会话管理方案 - 批量启动所有服务器
# 用法: ./start_all_servers.sh [sticky|redis|jwt|all]

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

# 日志目录
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"

# PID 文件目录
PID_DIR="$SCRIPT_DIR/pids"
mkdir -p "$PID_DIR"

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

# 检查端口是否被占用
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1 ; then
        return 0  # 端口被占用
    else
        return 1  # 端口空闲
    fi
}

# 等待服务器启动
wait_for_server() {
    local port=$1
    local max_attempts=30
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if check_port $port; then
            return 0
        fi
        sleep 0.5
        attempt=$((attempt + 1))
    done

    return 1
}

# 启动 Redis
start_redis() {
    log_info "检查 Redis..."

    # 检查 Redis 是否已运行
    if docker ps | grep -q redis; then
        log_success "Redis 已经在运行"
        return 0
    fi

    # 检查是否有停止的 Redis 容器
    if docker ps -a | grep -q redis; then
        log_info "启动已存在的 Redis 容器..."
        docker start redis >/dev/null 2>&1
    else
        log_info "创建并启动 Redis 容器..."
        docker run -d --name redis -p 6379:6379 redis:alpine >/dev/null 2>&1
    fi

    # 验证 Redis
    sleep 2
    if redis-cli ping >/dev/null 2>&1; then
        log_success "Redis 启动成功 (端口 6379)"
    else
        log_error "Redis 启动失败"
        return 1
    fi
}

# 启动 Sticky Session 服务器
start_sticky_session() {
    log_info "启动 Sticky Session 服务器..."

    cd "$SCRIPT_DIR/sticky-session"

    # 检查 go.mod 是否存在
    if [ ! -f "go.mod" ]; then
        log_error "go.mod 不存在，请先运行 go mod init"
        return 1
    fi

    # 启动 3 个服务器
    local ports=(8081 8082 8083)
    local server_ids=("server-1" "server-2" "server-3")

    for i in {0..2}; do
        local port=${ports[$i]}
        local server_id=${server_ids[$i]}

        # 检查端口是否已被占用
        if check_port $port; then
            log_warning "端口 $port 已被占用，跳过"
            continue
        fi

        log_info "  启动 $server_id (端口 $port)..."

        # 后台启动服务器
        PORT=$port SERVER_ID=$server_id \
            nohup go run main.go \
            > "$LOG_DIR/sticky-$server_id.log" 2>&1 &

        local pid=$!
        echo $pid > "$PID_DIR/sticky-$server_id.pid"

        # 等待服务器启动
        if wait_for_server $port; then
            log_success "  $server_id 启动成功 (PID: $pid)"
        else
            log_error "  $server_id 启动失败"
            return 1
        fi
    done

    cd "$SCRIPT_DIR"
}

# 启动 Redis Session 服务器
start_redis_session() {
    log_info "启动 Redis Session 服务器..."

    # 先确保 Redis 运行
    start_redis || return 1

    cd "$SCRIPT_DIR/redis-session"

    # 检查 go.mod 是否存在
    if [ ! -f "go.mod" ]; then
        log_error "go.mod 不存在，请先运行 go mod init"
        return 1
    fi

    # 启动 3 个服务器
    local ports=(8091 8092 8093)
    local server_ids=("server-1" "server-2" "server-3")

    for i in {0..2}; do
        local port=${ports[$i]}
        local server_id=${server_ids[$i]}

        # 检查端口是否已被占用
        if check_port $port; then
            log_warning "端口 $port 已被占用，跳过"
            continue
        fi

        log_info "  启动 $server_id (端口 $port)..."

        # 后台启动服务器
        PORT=$port SERVERID=$server_id \
            nohup go run main.go \
            > "$LOG_DIR/redis-$server_id.log" 2>&1 &

        local pid=$!
        echo $pid > "$PID_DIR/redis-$server_id.pid"

        # 等待服务器启动
        if wait_for_server $port; then
            log_success "  $server_id 启动成功 (PID: $pid)"
        else
            log_error "  $server_id 启动失败"
            return 1
        fi
    done

    cd "$SCRIPT_DIR"
}

# 启动 JWT Token 服务器
start_jwt_token() {
    log_info "启动 JWT Token 服务器..."

    cd "$SCRIPT_DIR/jwt-token"

    # 检查 go.mod 是否存在
    if [ ! -f "go.mod" ]; then
        log_error "go.mod 不存在，请先运行 go mod init"
        return 1
    fi

    # 启动 3 个服务器
    local ports=(8010 8011 8012)
    local server_ids=("server-1" "server-2" "server-3")

    for i in {0..2}; do
        local port=${ports[$i]}
        local server_id=${server_ids[$i]}

        # 检查端口是否已被占用
        if check_port $port; then
            log_warning "端口 $port 已被占用，跳过"
            continue
        fi

        log_info "  启动 $server_id (端口 $port)..."

        # 后台启动服务器
        PORT=$port SERVERID=$server_id \
            nohup go run main.go \
            > "$LOG_DIR/jwt-$server_id.log" 2>&1 &

        local pid=$!
        echo $pid > "$PID_DIR/jwt-$server_id.pid"

        # 等待服务器启动
        if wait_for_server $port; then
            log_success "  $server_id 启动成功 (PID: $pid)"
        else
            log_error "  $server_id 启动失败"
            return 1
        fi
    done

    cd "$SCRIPT_DIR"
}

# 显示服务器状态
show_status() {
    echo ""
    log_info "服务器状态:"
    echo ""
    printf "${BLUE}%-20s %-10s %-10s %-10s${NC}\n" "方案" "端口" "状态" "PID"
    echo "------------------------------------------------------------"

    # Sticky Session
    for port in 8081 8082 8083; do
        if check_port $port; then
            pid=$(lsof -ti:$port)
            printf "%-20s %-10s ${GREEN}%-10s${NC} %-10s\n" "Sticky Session" "$port" "运行中" "$pid"
        else
            printf "%-20s %-10s ${RED}%-10s${NC} %-10s\n" "Sticky Session" "$port" "已停止" "-"
        fi
    done

    # Redis Session
    for port in 8091 8092 8093; do
        if check_port $port; then
            pid=$(lsof -ti:$port)
            printf "%-20s %-10s ${GREEN}%-10s${NC} %-10s\n" "Redis Session" "$port" "运行中" "$pid"
        else
            printf "%-20s %-10s ${RED}%-10s${NC} %-10s\n" "Redis Session" "$port" "已停止" "-"
        fi
    done

    # JWT Token
    for port in 8010 8011 8012; do
        if check_port $port; then
            pid=$(lsof -ti:$port)
            printf "%-20s %-10s ${GREEN}%-10s${NC} %-10s\n" "JWT Token" "$port" "运行中" "$pid"
        else
            printf "%-20s %-10s ${RED}%-10s${NC} %-10s\n" "JWT Token" "$port" "已停止" "-"
        fi
    done

    # Redis
    if docker ps | grep -q redis; then
        printf "%-20s %-10s ${GREEN}%-10s${NC} %-10s\n" "Redis" "6379" "运行中" "docker"
    else
        printf "%-20s %-10s ${RED}%-10s${NC} %-10s\n" "Redis" "6379" "已停止" "-"
    fi

    echo ""
    log_info "日志目录: $LOG_DIR"
    log_info "PID 目录: $PID_DIR"
    echo ""
}

# 主函数
main() {
    local mode=${1:-all}

    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║         会话管理方案 - 批量启动服务器                      ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""

    case $mode in
        sticky)
            start_sticky_session
            ;;
        redis)
            start_redis_session
            ;;
        jwt)
            start_jwt_token
            ;;
        all)
            start_sticky_session
            echo ""
            start_redis_session
            echo ""
            start_jwt_token
            ;;
        status)
            show_status
            exit 0
            ;;
        *)
            log_error "未知模式: $mode"
            echo ""
            echo "用法: $0 [sticky|redis|jwt|all|status]"
            echo ""
            echo "选项:"
            echo "  sticky  - 只启动 Sticky Session 服务器"
            echo "  redis   - 只启动 Redis Session 服务器"
            echo "  jwt     - 只启动 JWT Token 服务器"
            echo "  all     - 启动所有服务器 (默认)"
            echo "  status  - 显示服务器状态"
            echo ""
            exit 1
            ;;
    esac

    # 显示状态
    show_status

    log_success "所有服务器启动完成！"
    echo ""
    log_info "查看日志:"
    echo "  tail -f $LOG_DIR/sticky-server-1.log"
    echo "  tail -f $LOG_DIR/redis-server-1.log"
    echo "  tail -f $LOG_DIR/jwt-server-1.log"
    echo ""
    log_info "停止所有服务器:"
    echo "  ./stop_all_servers.sh"
    echo ""
}

# 运行主函数
main "$@"
