# Video Feed - 短视频后端服务

基于 Go 构建的短视频后端服务，涵盖账号、视频发布、Feed 流、热榜、点赞、评论等核心功能，引入 Redis 缓存、RabbitMQ 异步处理、熔断降级与热点防击穿等方案，通过 Docker Compose 一键部署。

## 技术栈

| 层 | 技术 |
|---|------|
| 语言 | Go 1.21 |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7 |
| 消息队列 | RabbitMQ |
| 容器化 | Docker + Docker Compose |

## 功能清单

- 用户注册 / 登录（JWT 鉴权，bcrypt 密码加密）
- 视频上传（本地文件存储 + 数据库记录 + 缓存更新 + MQ 异步任务）
- Feed 流（游标分页 + Redis ZSet 缓存 + singleflight 防击穿）
- 热榜（Redis ZSet 排序 + 异步重建 + 熔断降级）
- 点赞 / 取消点赞（Redis Set + MQ 异步落盘 + 事务保证一致性）
- 评论系统（发表、游标分页查询、权限校验删除）
- 视频删除（数据库 + Redis + 本地文件联动清理）

## 高可用设计

| 机制 | 实现 | 说明 |
|------|------|------|
| 限流 | `golang.org/x/time/rate` | 已登录按用户 5 QPS，未登录按 IP 10 QPS |
| 熔断降级 | `gobreaker` | Redis 故障时自动降级到数据库 |
| 热点防击穿 | `singleflight` | 同参数并发请求合并为一次数据库查询 |
| 异步解耦 | RabbitMQ | 点赞落盘异步化，提升接口响应 |
| 优雅停机 | `signal.Notify` + `srv.Shutdown` | 等待现有请求完成再退出 |
| 健康检查 | `/health` 端点 | 配合 Docker `healthcheck` 自动恢复 |

## 项目结构

```
├── main.go                  # 入口，优雅停机
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── models/                  # 数据模型与连接初始化（DB / Redis / MQ / CircuitBreaker）
│   ├── db.go
│   ├── mq.go
│   ├── user.go
│   ├── video.go
│   ├── like.go
│   └── comment.go
├── controllers/             # 业务控制器
│   ├── user.go              # 注册 / 登录
│   ├── video.go             # 上传视频
│   ├── feed.go              # Feed 流
│   ├── hot.go               # 热榜
│   ├── like.go              # 点赞
│   ├── comment.go           # 评论
│   └── delete_video.go      # 删除视频
├── middleware/               # 中间件
│   ├── auth.go              # JWT 鉴权
│   └── rate_limit.go        # 限流
├── mq_task/                 # MQ 消费者
│   └── consumer.go
├── utils/                   # 工具
│   └── jwt.go
├── routes/                  # 路由
│   └── routes.go
└── videofeed.html           # 前端页面
```

## 快速开始

### Docker 部署（推荐）

```bash
# 克隆项目
git clone https://github.com/yuexia2005/studygo
cd video_feed

# 一键启动（MySQL + Redis + RabbitMQ + App）
docker compose up -d --build

# 访问
http://localhost:8085
```

### 停止服务

```bash
docker compose down
```

### 本地运行（需要先启动 MySQL 和 Redis）

```bash
# 设置环境变量
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=video_feed
export REDIS_ADDR=localhost:6379
export JWT_SECRET=your_jwt_secret
export RABBITMQ_URL=amqp://admin:password@localhost:5672/

# 运行
go run main.go
```

## 环境变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `DB_HOST` | MySQL 主机 | `mysql` |
| `DB_PORT` | MySQL 端口 | `3306` |
| `DB_USER` | MySQL 用户名 | `root` |
| `DB_PASSWORD` | MySQL 密码 | `your_password` |
| `DB_NAME` | 数据库名 | `video_feed` |
| `REDIS_ADDR` | Redis 地址 | `redis:6379` |
| `JWT_SECRET` | JWT 签名密钥 | `your_secret_key` |
| `RABBITMQ_URL` | RabbitMQ 连接地址 | `amqp://admin:password@rabbitmq:5672/` |

## API 概览

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/register` | 用户注册 |
| POST | `/login` | 用户登录 |
| GET | `/health` | 健康检查 |

### 需认证接口（Header: `Authorization: Bearer <token>`）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/video/upload` | 上传视频 |
| DELETE | `/api/video/:id` | 删除视频 |
| GET | `/api/feed` | Feed 流（游标分页：`?last_id=0&limit=10`） |
| GET | `/api/hot` | 热榜（`?limit=20`） |
| POST | `/api/video/:id/like` | 点赞 / 取消点赞 |
| POST | `/api/video/:id/comment` | 发表评论 |
| GET | `/api/video/:id/comments` | 评论列表（游标分页：`?last_id=0&limit=5`） |
| DELETE | `/api/comment/:id` | 删除评论 |

### 注册

```bash
curl -X POST http://localhost:8085/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

### 登录

```bash
curl -X POST http://localhost:8085/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

### Feed 流

```bash
curl http://localhost:8085/api/feed?limit=10 \
  -H "Authorization: Bearer <token>"
```

## License

MIT
