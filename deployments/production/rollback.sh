#!/bin/bash
# AI Gateway 回滚脚本

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_info "========================================="
log_info "AI Gateway 回滚脚本"
log_info "========================================="

cd "$(dirname "$0")"

# 检查是否有备份
if [ ! -d "backups" ]; then
    log_error "未找到备份目录"
    exit 1
fi

# 列出可用备份
log_info "可用备份："
ls -lt backups/ | grep "^d" | head -10

# 获取备份版本
read -p "请输入要回滚到的版本 (格式: YYYYMMDD-HHMMSS): " backup_version

backup_dir="backups/$backup_version"

if [ ! -d "$backup_dir" ]; then
    log_error "备份 $backup_version 不存在"
    exit 1
fi

# 确认回滚
log_warn "警告：即将回滚到 $backup_version"
log_warn "此操作将停止当前服务并使用备份版本启动"
read -p "确认回滚? [yes/NO]: " confirm

if [ "$confirm" != "yes" ]; then
    log_info "回滚已取消"
    exit 0
fi

# 停止当前服务
log_info "停止当前服务..."
docker-compose -f docker-compose.prod.yml down

# 恢复备份
log_info "恢复备份..."
cp -r "$backup_dir/*" ./

# 启动服务
log_info "启动服务..."
docker-compose -f docker-compose.prod.yml up -d

# 等待健康检查
log_info "等待服务启动..."
sleep 10

if curl -sf http://localhost:8080/health > /dev/null; then
    log_info "回滚成功！"
else
    log_error "回滚后服务启动失败，请手动检查"
    exit 1
fi

log_info "========================================="
log_info "回滚完成！"
log_info "========================================="
