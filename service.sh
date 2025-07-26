#!/bin/bash



SERVICE_NAME="share_file-linux-amd64.bin"

# 运行用户 特殊情况下允许改为root
RUN_USER="share_file"
#=============

# 服务描述
SERVICE_DESCRIPTION="linux代理"
# Go 程序路径
PROGRAM_PATH="$PWD/$SERVICE_NAME"
# Systemd 服务文件路径
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"
# 启动参数
COMMAND=""

# 安装服务
install_service() {
    # 检查用户是否存在
    if id "$RUN_USER" &>/dev/null; then
        echo "用户 '$RUN_USER' 存在。"
    else
        echo "用户 '$RUN_USER' 不存在。准备创建"
        # 创建系统用户（无家目录，禁止登录）
        if useradd --system \
                   --no-create-home \
                   --shell /sbin/nologin \
                   "$RUN_USER"; then
            echo "[SUCCESS] 已成功创建系统用户 '$RUN_USER'"
            echo "用户信息："
            grep "^$RUN_USER:" /etc/passwd
        else
            echo "[ERROR] 创建用户 '$RUN_USER' 失败！" >&2
            exit 1
        fi
    fi

    # 检查是否已安装
    if [ -f "$SERVICE_FILE" ]; then
        echo "Service already installed."
        exit 1
    fi

    # 检查 Go 程序是否存在
    if [ ! -f "$PROGRAM_PATH" ]; then
        echo "Go program not found at $PROGRAM_PATH."
        exit 1
    fi

    # 创建 systemd 服务文件
    cat <<EOF | sudo tee "$SERVICE_FILE" > /dev/null
[Unit]
Description=$SERVICE_DESCRIPTION
After=network.target

[Service]
ExecStart=$PROGRAM_PATH $COMMAND
Restart=always
User=$RUN_USER
Group=$RUN_USER
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF
    # 设置权限
    chmod 644 "$SERVICE_FILE"
    # 重新加载 systemd 配置
    sudo systemctl daemon-reload

    # 启用并启动服务
    sudo systemctl enable "$SERVICE_NAME"
    sudo systemctl start "$SERVICE_NAME"

    echo "Service installed and started successfully."
}

# 卸载服务
uninstall_service() {
    # 检查服务是否已安装
    if [ ! -f "$SERVICE_FILE" ]; then
        echo "Service not installed."
        exit 1
    fi

    # 停止并禁用服务
    sudo systemctl stop "$SERVICE_NAME"
    sudo systemctl disable "$SERVICE_NAME"

    # 删除服务文件
    sudo rm "$SERVICE_FILE"
    sudo systemctl daemon-reload

    # 删除 Go 程序
#    sudo rm "$PROGRAM_PATH"

    echo "Service uninstalled successfully."
}

# 启动服务
start_service() {
    if [ ! -f "$SERVICE_FILE" ]; then
        echo "Service not installed."
        exit 1
    fi

    sudo systemctl start "$SERVICE_NAME"
    echo "Service started successfully."
}

# 停止服务
stop_service() {
    if [ ! -f "$SERVICE_FILE" ]; then
        echo "Service not installed."
        exit 1
    fi

    sudo systemctl stop "$SERVICE_NAME"
    echo "Service stopped successfully."
}

# 查看服务状态
status_service() {
    if [ ! -f "$SERVICE_FILE" ]; then
        echo "Service not installed."
        exit 1
    fi

    sudo systemctl status "$SERVICE_NAME"
}


# 主逻辑
case "$1" in
    install)
        install_service
        ;;
    uninstall)
        uninstall_service
        ;;
    start)
        start_service
        ;;
    stop)
        stop_service
        ;;
    status)
        status_service
        ;;
    *)
        show_help
        ;;
esac