# ===== 前端构建阶段 =====
FROM node:18-alpine AS frontend-builder

# npm 国内镜像
RUN npm config set registry https://registry.npmmirror.com

WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm install

COPY frontend/ .
RUN npm run build


# ===== Go 构建阶段 =====
FROM golang:1.21-alpine AS builder

# Go 模块代理走国内源（七牛云）
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

#复制所有源代码到容器内
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o video_feed .


# ===== 运行阶段 =====
FROM alpine:latest

# Alpine 包管理器走阿里云镜像
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk --no-cache add ca-certificates tzdata curl

WORKDIR /root
#从构建阶段复制 Go 二进制文件
COPY --from=builder /app/video_feed .
#从构建阶段复制 Vue3 前端编译产物
COPY --from=frontend-builder /frontend/dist ./dist
#创建上传目录（视频文件存储）
RUN mkdir -p uploads

EXPOSE 8085

CMD ["./video_feed"]
