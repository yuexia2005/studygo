#构建阶段（builder）
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

#复制所有源代码到容器内
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o video_feed .


#运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /root
#从构建阶段复制二进制文件
COPY --from=builder /app/video_feed .
#创建上传目录（视频文件存储）
RUN mkdir -p uploads

EXPOSE 8085

CMD ["./video_feed"]
