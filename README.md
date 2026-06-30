# 🔗 URL Shortener

一个生产级短链接服务，从零构建，开源可本地部署。

**技术栈：Go + Vue 3 + PostgreSQL + Redis + Docker Compose**

## ✨ 功能

- **短链生成**：加密安全随机短码（Base62, 7位, ~3.5万亿空间），支持自定义短码
- **高性能跳转**：三层查询（布隆过滤器 → Redis → PostgreSQL），302/301 可选
- **数据分析**：实时点击趋势图、地理分布、设备/浏览器分布、来源域名排名
- **密码保护**：链接密码验证，HTML 密码输入页
- **智能分发**：规则引擎，根据国家/设备/OS/浏览器等条件分发到不同目标
- **限流保护**：Redis Lua 令牌桶，全局 + 按 IP 限流，Redis 不可用时自动降级
- **开放 API**：JWT 鉴权 + API Key 管理
- **管理面板**：Vue 3 SPA，数据看板 ECharts 图表，链接 CRUD，二维码生成

## 🚀 快速开始

```bash
# 克隆项目
git clone <repo-url>
cd url-shortener

# 一键启动
docker compose up -d

# 打开浏览器
# 管理面板: http://localhost:3000
# API: http://localhost:8080
```

## 🛠️ 本地开发

### 后端

```bash
# 确保 PostgreSQL 和 Redis 运行
# 修改 .env 中数据库和 Redis 连接信息

go run ./cmd/server
# Server starting on 0.0.0.0:8080
```

### 前端

```bash
cd web
npm install
npm run dev
# http://localhost:5173
```

### 运行测试

```bash
go test ./... -v -cover
```

## 📡 API 文档

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/links` | 创建短链接 |
| GET | `/:short_code` | 跳转到原始 URL |

### 认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 用户登录 |

### 管理接口（需 JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/links` | 链接列表（分页） |
| GET | `/api/v1/links/:code` | 链接详情 |
| PATCH | `/api/v1/links/:code` | 更新链接 |
| DELETE | `/api/v1/links/:code` | 删除链接 |
| GET | `/api/v1/links/:code/stats` | 链接统计 |

### Dashboard（需 JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/dashboard/overview` | 总览数据 |
| GET | `/api/v1/dashboard/trend` | 点击趋势 |
| GET | `/api/v1/dashboard/geo` | 地理分布 |
| GET | `/api/v1/dashboard/devices` | 设备分布 |

### 创建链接示例

```bash
# 基础创建
curl -X POST http://localhost:8080/api/v1/links \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/very/long/url"}'

# 完整参数
curl -X POST http://localhost:8080/api/v1/links \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt-token>" \
  -d '{
    "url": "https://example.com/page",
    "custom_code": "my-link",
    "title": "营销活动页",
    "expire_at": "2026-12-31T23:59:59Z",
    "password": "secret",
    "rules": [
      {
        "priority": 1,
        "conditions": [
          {"field": "country", "op": "eq", "value": "CN"},
          {"field": "os", "op": "eq", "value": "ios"}
        ],
        "action": {"type": "redirect", "target": "https://apps.apple.com/app/id123456"}
      }
    ]
  }'
```

## 🏗️ 架构

```
Client → Nginx → Go API Server → PostgreSQL
                    ↓               ↑ (async)
                  Redis ←—— Click Log Worker
                    ↓
          Bloom Filter (防穿透)
          缓存 (短码→URL)
          限流 (令牌桶)
```

### 项目结构

```
url-shortener/
├── cmd/server/main.go          # 入口
├── internal/
│   ├── handler/                 # HTTP 处理器
│   │   ├── redirect.go         # 跳转 + 密码保护 + 级联查询
│   │   ├── link.go             # 链接 CRUD API
│   │   ├── auth.go             # 注册/登录
│   │   └── dashboard.go        # 数据看板 API
│   ├── service/                # 业务逻辑
│   │   ├── shortcode.go        # 短码生成引擎 (crypto/rand)
│   │   ├── link.go             # 链接服务
│   │   └── auth.go             # JWT 认证
│   ├── repository/             # 数据层 (PostgreSQL)
│   │   ├── link.go             # 链接 CRUD + Dashboard 聚合
│   │   ├── click_log.go        # 点击日志批量写入
│   │   ├── user.go             # 用户管理
│   │   └── apikey.go           # API Key 管理
│   ├── cache/                  # Redis 缓存
│   │   ├── redis.go            # 连接 + URL缓存 + 令牌桶Lua
│   │   └── bloom.go            # 布隆过滤器 (Bitmap)
│   ├── middleware/             # 中间件
│   │   ├── auth.go             # JWT 鉴权
│   │   ├── logger.go           # 请求日志 + Panic Recovery
│   │   ├── ratelimit.go        # 内存令牌桶 (降级)
│   │   └── ratelimit_redis.go  # Redis 令牌桶
│   ├── model/                  # 数据模型
│   └── util/                   # 工具
│       ├── base62.go           # Base62 编解码
│       ├── hash.go             # bcrypt + SHA-256 + API Key生成
│       ├── ua.go               # User-Agent 解析
│       ├── ip.go               # GeoIP 离线解析
│       └── rule_engine.go      # 智能分发规则引擎
├── web/                        # Vue 3 前端
│   └── src/
│       ├── views/              # 页面 (Dashboard/Links/Login)
│       ├── components/         # 图表组件 (ECharts)
│       ├── api/                # Axios 封装
│       ├── router/             # Vue Router
│       └── stores/             # Pinia 状态管理
├── migrations/                 # SQL 迁移
├── docker-compose.yml          # Docker 编排
├── Dockerfile                  # 多阶段构建
├── nginx.conf                  # Nginx 配置
└── Makefile
```

## 🔧 环境变量

参考 `.env.example`，本地开发时复制为 `.env`：

```bash
cp .env.example .env
# 编辑 .env 修改数据库密码、JWT密钥等
```

## 🐳 Docker 部署

```bash
# 本地
docker compose up -d

# 生产（腾讯云等）
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## 📚 课设答辩要点

| 层面 | 可讲内容 |
|------|----------|
| **应用层** | RESTful API、JWT/API Key 鉴权、301 vs 302 |
| **传输层** | HTTP Keep-Alive、连接池、Nginx 反向代理 |
| **网络层** | IP GeoIP 定位、DNS 解析链路 |
| **系统设计** | 缓存穿透/雪崩防护、异步日志管道、三层查询架构 |
| **数据结构** | Base62 编码、布隆过滤器、令牌桶算法 |
| **运维** | Docker 容器化、Docker Compose 编排、优雅关闭 |

## 📄 License

MIT
