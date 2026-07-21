# Video Feed - 短视频后端服务

基于 Go + Gin 构建的高并发短视频后端，涵盖用户鉴权、视频 Feed 流、实时热榜、点赞评论等核心功能。引入 Redis 缓存、RabbitMQ 异步削峰、gobreaker 熔断降级、singleflight 防击穿等稳定性策略，支持 Docker Compose 一键部署。

## 技术栈

| 层 | 技术 |
|---|------|
| 语言 | Go 1.21 |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7 |
| 消息队列 | RabbitMQ 3.9 |
| 前端 | Vue 3 + Vite |
| 容器化 | Docker + Docker Compose |

## 功能

- 用户注册/登录（JWT + bcrypt）
- 视频上传（本地存储 + 缓存更新 + MQ 异步任务）
- Feed 流（游标分页 + Redis ZSet 缓存 + singleflight 防击穿）
- 实时热榜（Redis ZSet + 异步重建 + 熔断降级）
- 点赞/取消（Redis Set 先行 + MQ 异步落盘 + 事务一致性）
- 评论系统（发表、游标分页、权限删除）
- 视频删除（DB + Redis + 文件联动清理）

## 高可用设计

| 机制 | 技术 | 说明 |
|------|------|------|
| 限流 | `x/time/rate` 令牌桶 | 登录 5 QPS，未登录 10 QPS |
| 熔断 | `gobreaker` | Redis 故障自动降级到 MySQL |
| 防击穿 | `singleflight` | 并发请求合并为一次 DB 查询 |
| 异步削峰 | RabbitMQ | 点赞落盘异步化 |
| 优雅停机 | `signal.Notify` | 15s 等待现有请求完成 |
| 健康检查 | `/health` | Docker healthcheck 自动恢复 |

## 项目结构

```
├── main.go
├── Dockerfile
├── docker-compose.yml
├── controllers/              # 业务控制器
│   ├── user.go               # 注册 / 登录
│   ├── video.go              # 上传视频
│   ├── feed.go               # Feed 流
│   ├── hot.go                # 热榜
│   ├── like.go               # 点赞
│   ├── comment.go            # 评论
│   └── delete_video.go
├── middleware/
│   ├── auth.go               # JWT 鉴权
│   └── rate_limit.go         # 限流
├── models/
│   ├── db.go                 # DB / Redis / 熔断器初始化
│   ├── mq.go                 # RabbitMQ 连接
│   ├── cache_rebuild.go      # 缓存重建
│   ├── user.go
│   ├── video.go
│   ├── like.go
│   └── comment.go
├── mq_task/
│   └── consumer.go           # MQ 消费者
├── routes/
│   └── routes.go
├── utils/
│   └── jwt.go
└── frontend/                 # Vue 3 前端
    └── src/
        ├── App.vue
        └── components/       # AuthCard / FeedList / HotList / VideoCard 等
```

## 快速开始

### Docker 部署（推荐）

```bash
# 一键启动 MySQL + Redis + RabbitMQ + App
docker compose up -d --build

# 访问
http://localhost:8085
```

### 停止

```bash
docker compose down
```

### 本地开发（需先启动 MySQL、Redis、RabbitMQ）

```bash
# 设置环境变量
$env:DB_HOST="localhost"
$env:DB_PASSWORD="123456"
$env:DB_NAME="video_feed"
$env:REDIS_ADDR="localhost:6379"
$env:JWT_SECRET="your_secret"
$env:RABBITMQ_URL="amqp://admin:password@localhost:5672/"

# 启动
go run main.go
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| DB_HOST | `mysql` | MySQL 地址 |
| DB_PORT | `3306` | MySQL 端口 |
| DB_USER | `root` | MySQL 用户名 |
| DB_PASSWORD | — | MySQL 密码（必填） |
| DB_NAME | `video_feed` | 数据库名 |
| REDIS_ADDR | `redis:6379` | Redis 地址 |
| JWT_SECRET | — | JWT 密钥（必填） |
| RABBITMQ_URL | `amqp://admin:password@rabbitmq:5672/` | MQ 连接地址 |

## API

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/register` | 注册 |
| POST | `/login` | 登录 |
| GET | `/health` | 健康检查 |

### 需认证（Header: `Authorization: Bearer <token>`）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/video/upload` | 上传视频 |
| DELETE | `/api/video/:id` | 删除视频 |
| GET | `/api/feed?last_id=0&limit=10` | Feed 流（游标分页） |
| GET | `/api/hot?limit=5` | 热榜 |
| POST | `/api/video/:id/like` | 点赞/取消 |
| POST | `/api/video/:id/comment` | 发表评论 |
| GET | `/api/video/:id/comments` | 评论列表（游标分页） |
| DELETE | `/api/comment/:id` | 删除评论 |

### 示例

```bash
# 注册
curl -X POST http://localhost:8085/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 登录
curl -X POST http://localhost:8085/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# Feed 流
curl "http://localhost:8085/api/feed?limit=10" \
  -H "Authorization: Bearer <token>"
```
