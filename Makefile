.PHONY: build run docker docker-down clean

# 本地编译
build:
    go build -o video_feed .

# 本地运行
run:
    go run main.go

# Docker 一键启动（后台运行）
docker:
    docker compose up --build -d

# 停止并删除 Docker 容器
docker-down:
    docker compose down

# 清理编译产物
clean:
    rm -f video_feed