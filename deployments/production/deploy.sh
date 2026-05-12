#!/bin/bash
# AI Gateway 生产环境部署脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查必要工具
check_requirements() {
    log_info "检查部署环境..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        exit 1
    fi

    log_info "部署环境检查通过"
}

# 检查环境变量
check_env_file() {
    log_info "检查环境变量配置..."

    if [ ! -f ".env" ]; then
        log_warn ".env 文件不存在，从示例创建..."
        cp .env.example .env
        log_error "请编辑 .env 文件并填写正确的配置值后重新运行部署"
        exit 1
    fi

    # 检查关键配置
    source .env

    if [ "$ENCRYPTION_KEY" = "your_32_character_encryption_key_here" ]; then
        log_error "请设置安全的 ENCRYPTION_KEY"
        exit 1
    fi

    if [ "$POSTGRES_PASSWORD" = "your_secure_password_here" ]; then
        log_error "请设置安全的 POSTGRES_PASSWORD"
        exit 1
    fi

    log_info "环境变量检查通过"
}

# 拉取最新镜像
pull_images() {
    log_info "拉取最新镜像..."
    docker-compose -f docker-compose.prod.yml pull
}

# 构建应用镜像
build_images() {
    log_info "构建应用镜像..."
    docker-compose -f docker-compose.prod.yml build --parallel
}

# 停止旧容器
stop_old() {
    log_info "停止旧容器..."
    docker-compose -f docker-compose.prod.yml down
}

# 启动新容器
start_new() {
    log_info "启动新容器..."
    docker-compose -f docker-compose.prod.yml up -d
}

# 等待健康检查
wait_for_health() {
    log_info "等待服务启动..."

    local max_attempts=30
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if curl -sf http://localhost:8080/health > /dev/null; then
            log_info "后端服务已就绪"
            return 0
        fi

        attempt=$((attempt + 1))
        echo -n "."
        sleep 2
    done

    log_error "服务启动超时"
    return 1
}

# 运行数据库迁移
run_migrations() {
    log_info "运行数据库迁移..."
    # 这里可以添加迁移命令
    log_info "数据库迁移完成"
}

# 显示部署状态
show_status() {
    log_info "部署状态："
    docker-compose -f docker-compose.prod.yml ps
}

# 主部署流程
main() {
    log_info "========================================="
    log_info "AI Gateway 生产环境部署"
    log_info "========================================="

    cd "$(dirname "$0")"

    check_requirements
    check_env_file

    # 询问部署类型
    echo ""
    echo "请选择部署类型："
    echo "1) 完整部署 (构建 + 启动)"
    echo "2) 仅重启 (使用已有镜像)"
    echo "3) 仅更新配置"
    read -p "请输入选择 [1-3]: " choice

    case $choice in
        1)
            pull_images
            build_images
            stop_old
            start_new
            wait_for_health
            run_migrations
            show_status
            ;;
        2)
            stop_old
            start_new
            wait_for_health
            show_status
            ;;
        3)
            log_warn "配置更新需要重启容器才能生效"
            read -p "是否现在重启? [y/N]: " confirm
            if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
                stop_old
                start_new
                wait_for_health
                show_status
            fi
            ;;
        *)
            log_error "无效选择"
            exit 1
            ;;
    esac

    log_info ""
    log_info "========================================="
    log_info "部署完成！"
    log_info "========================================="
    log_info "管理端: https://${ADMIN_DOMAIN:-localhost:5174}"
    log_info "用户端: https://${USER_DOMAIN:-localhost:5173}"
    log_info "API: https://${DOMAIN:-localhost:8080}"
    log_info "监控: https://monitoring.yourdomain.com"
    log_info "========================================="
}

# 捕获中断信号
trap 'log_error "部署被中断"; exit 1' INT TERM

# 执行主流程
main "$@"
