<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Vue-3.2+-4FC08D?style=for-the-badge&logo=vue.js" alt="Vue Version">
  <img src="https://img.shields.io/badge/MySQL-8.0+-4479A1?style=for-the-badge&logo=mysql" alt="MySQL">
  <img src="https://img.shields.io/badge/Redis-7.0+-DC382D?style=for-the-badge&logo=redis" alt="Redis">
  <img src="https://img.shields.io/badge/RabbitMQ-3.12+-FF6600?style=for-the-badge&logo=rabbitmq" alt="RabbitMQ">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License">
</p>

<h1 align="center">🐍 ViperAI</h1>

<p align="center">
  <b>AI智能对话平台</b><br>
  <sub>基于领域驱动设计(DDD)架构的全栈AI应用解决方案</sub>
</p>

<p align="center">
  <a href="#-系统概述">系统概述</a> •
  <a href="#-核心功能">核心功能</a> •
  <a href="#-技术架构">技术架构</a> •
  <a href="#-快速开始">快速开始</a> •
  <a href="#-api文档">API文档</a> •
  <a href="#-部署指南">部署指南</a>
</p>

***

## 📖 系统概述

**ViperAI** 是一个功能完备的AI智能对话平台，采用前后端分离架构，集成了多种主流AI模型能力。系统支持普通对话、RAG检索增强、MCP工具调用、代码执行沙箱、图像识别、语音合成等多种AI能力，适用于智能客服、知识问答、代码助手等多种业务场景。

### 🎯 项目亮点

| 特性            | 描述                                     |
| ------------- | -------------------------------------- |
| 🏗️ **DDD架构** | 采用领域驱动设计，清晰的分层架构，高内聚低耦合                |
| 🤖 **多模型支持**  | 支持阿里百炼、OpenAI、Ollama等多种AI模型            |
| 📊 **RAG增强**  | 基于Redis Vector的检索增强生成，提升回答准确性          |
| 🔧 **MCP协议**  | 支持Model Context Protocol，可扩展工具调用       |
| 💻 **代码沙箱**   | 支持Python/JavaScript/Go/Java/C++多语言在线执行 |
| 🎨 **流式输出**   | 基于SSE的实时流式响应，用户体验流畅                    |
| 🔐 **JWT认证**  | 完整的用户认证授权体系                            |
| 📈 **高可用**    | 消息队列异步处理，支持水平扩展                        |

***

## 🚀 核心功能

### 1. 智能对话系统

```
┌─────────────────────────────────────────────────────────┐
│                    AI对话引擎架构                         │
├─────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐     │
│  │ OpenAI  │  │ 阿里百炼 │  │  Ollama │  │  RAG    │     │
│  │ Engine  │  │ Engine  │  │ Engine  │  │ Engine  │     │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘     │
│       │            │            │            │          │
│       └────────────┴─────┬──────┴────────────┘          │
│                          ▼                              │
│              ┌───────────────────────┐                  │
│              │   Assistant Manager   │                  │
│              │   (会话状态管理)       │                  │
│              └───────────────────────┘                  │
└─────────────────────────────────────────────────────────┘
```

- **多模型切换**: 支持在对话中动态切换AI模型
- **上下文管理**: 自动维护对话历史和上下文
- **流式响应**: SSE实时推送，打字机效果

### 2. RAG检索增强

```
文档上传 → 文本分块 → 向量嵌入 → Redis存储
                              ↓
用户提问 → 向量检索 → 上下文增强 → AI回答
```

- 支持Markdown/TXT文档上传
- 基于Redis Vector的向量索引
- 语义相似度检索

### 3. MCP工具调用

```json
{
  "tools": [
    {"name": "get_weather", "description": "获取城市天气"},
    {"name": "get_time", "description": "获取当前时间"},
    {"name": "calculate", "description": "数学计算"},
    {"name": "execute_code", "description": "代码执行"}
  ]
}
```

### 4. 代码执行沙箱

| 语言         | 编译器/解释器    | 超时时间 |
| ---------- | ---------- | ---- |
| Python     | python3    | 30s  |
| JavaScript | node       | 30s  |
| Go         | go build   | 30s  |
| Java       | javac/java | 30s  |
| C++        | g++        | 30s  |

### 5. 图像识别

- 基于ONNX Runtime
- MobileNetV2预训练模型
- 1000类ImageNet分类

### 6. 语音合成

- 百度TTS API集成
- 异步任务处理
- 轮询状态查询

***

## 🏛️ 技术架构

### 系统架构图

```
                                    ┌─────────────────┐
                                    │   Nginx/CDN     │
                                    │   (负载均衡)     │
                                    └────────┬────────┘
                                             │
                    ┌────────────────────────┼────────────────────────┐
                    │                        │                        │
                    ▼                        ▼                        ▼
           ┌───────────────┐       ┌───────────────┐       ┌───────────────┐
           │   Frontend    │       │   Backend     │       │  MCP Server   │
           │   (Vue 3)     │       │   (Go API)    │       │   (Go)        │
           │   Port: 80    │       │   Port: 9090  │       │   Port: 8081  │
           └───────────────┘       └───────┬───────┘       └───────────────┘
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    │                      │                      │
                    ▼                      ▼                      ▼
           ┌───────────────┐       ┌───────────────┐       ┌───────────────┐
           │    MySQL      │       │    Redis      │       │   RabbitMQ    │
           │   (主数据库)   │       │  (缓存/向量)   │       │   (消息队列)   │
           │   Port: 3306  │       │   Port: 6379  │       │   Port: 5672  │
           └───────────────┘       └───────────────┘       └───────────────┘
```

### 后端技术栈

| 层级        | 技术选型         | 说明                     |
| --------- | ------------ | ---------------------- |
| **Web框架** | Gin          | 高性能HTTP框架              |
| **ORM**   | GORM         | Go语言ORM库               |
| **AI框架**  | eino         | 字节跳动开源AI框架             |
| **认证**    | JWT          | golang-jwt/jwt         |
| **缓存**    | Redis        | go-redis v9            |
| **消息队列**  | RabbitMQ     | streadway/amqp         |
| **向量检索**  | Redis Vector | 原生向量索引支持               |
| **图像识别**  | ONNX Runtime | yalue/onnxruntime\_go  |
| **MCP协议** | mcp-go       | Model Context Protocol |

### 前端技术栈

| 技术           | 版本   | 说明                 |
| ------------ | ---- | ------------------ |
| Vue          | 3.2+ | 渐进式JavaScript框架    |
| Vue Router   | 4.x  | 官方路由管理器            |
| Element Plus | 2.x  | Vue 3组件库           |
| Axios        | 1.x  | HTTP客户端            |
| SSE          | -    | Server-Sent Events |

### 项目结构

```
ViperAI/
├── cmd/                          # 应用入口
│   ├── server/main.go            # API服务器
│   └── mcp-server/main.go        # MCP服务器
├── config/                       # 配置文件
│   └── settings.toml             # 主配置文件
├── docs/                         # 文档
│   ├── api.md                    # API文档
│   ├── database.md               # 数据库设计
│   └── deployment.md             # 部署指南
├── internal/                     # 内部模块
│   ├── config/                   # 配置管理
│   ├── domain/                   # 领域模型
│   │   ├── user.go               # 用户实体
│   │   ├── conversation.go       # 对话实体
│   │   └── message.go            # 消息实体
│   ├── engine/                   # AI引擎
│   │   ├── ai_engine.go          # 引擎抽象
│   │   ├── assistant.go          # 助手管理
│   │   ├── rag_engine.go         # RAG引擎
│   │   └── mcp_engine.go         # MCP引擎
│   ├── infrastructure/           # 基础设施
│   │   ├── cache/                # Redis缓存
│   │   ├── database/             # MySQL数据库
│   │   ├── mcp/                  # MCP服务器
│   │   └── queue/                # RabbitMQ队列
│   ├── pkg/                      # 公共工具
│   │   ├── auth/                 # JWT认证
│   │   ├── image/                # 图像处理
│   │   └── utils/                # 工具函数
│   ├── repository/               # 数据访问层
│   ├── service/                  # 业务逻辑层
│   └── transport/http/           # HTTP传输层
│       ├── handler/              # 控制器
│       ├── middleware/           # 中间件
│       ├── response/             # 响应封装
│       └── router/               # 路由配置
├── web/                          # 前端项目
│   ├── src/
│   │   ├── views/                # 页面组件
│   │   ├── router/               # 路由配置
│   │   └── utils/                # 工具函数
│   └── package.json
├── go.mod
└── README.md
```

***

## 🔧 快速开始

### 环境要求

- Go 1.24+
- Node.js 18+
- MySQL 8.0+
- Redis 7.0+
- RabbitMQ 3.12+

### 安装步骤

#### 1. 克隆项目

```bash
git clone https://github.com/your-username/viperai.git
cd viperai
```

#### 2. 配置环境变量

```bash
# Linux/macOS
export OPENAI_API_KEY="your-api-key"
export OPENAI_MODEL_NAME="qwen-turbo"
export OPENAI_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"

# Windows PowerShell
$env:OPENAI_API_KEY="your-api-key"
$env:OPENAI_MODEL_NAME="qwen-turbo"
$env:OPENAI_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"
```

#### 3. 修改配置文件

编辑 `config/settings.toml`，配置数据库、Redis、RabbitMQ等连接信息。

#### 4. 启动后端服务

```bash
# 安装依赖
go mod tidy

# 运行服务
go run cmd/server/main.go
```

#### 5. 启动MCP服务器（可选）

```bash
go run cmd/mcp-server/main.go
```

#### 6. 启动前端服务

```bash
cd web
npm install
npm run serve
```

#### 7. 访问应用

打开浏览器访问: <http://localhost:8080>

### Docker部署

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f backend
```

***

## 📚 API文档

### 基础信息

- **Base URL**: `http://localhost:9090/api/v1`
- **认证方式**: Bearer Token (JWT)

### 接口列表

#### 用户模块

| 方法   | 路径               | 描述     | 认证 |
| ---- | ---------------- | ------ | -- |
| POST | `/user/login`    | 用户登录   | ❌  |
| POST | `/user/register` | 用户注册   | ❌  |
| POST | `/user/captcha`  | 发送验证码  | ❌  |
| GET  | `/user/profile`  | 获取用户信息 | ✅  |

#### 对话模块

| 方法   | 路径                    | 描述         | 认证 |
| ---- | --------------------- | ---------- | -- |
| GET  | `/chat/conversations` | 获取对话列表     | ✅  |
| POST | `/chat/send-new`      | 创建新对话并发送消息 | ✅  |
| POST | `/chat/send`          | 发送消息到现有对话  | ✅  |
| POST | `/chat/stream-new`    | 流式创建新对话    | ✅  |
| POST | `/chat/stream`        | 流式发送消息     | ✅  |
| POST | `/chat/history`       | 获取对话历史     | ✅  |

#### 代码执行模块

| 方法   | 路径                | 描述      | 认证 |
| ---- | ----------------- | ------- | -- |
| POST | `/code/execute`   | 执行代码    | ✅  |
| GET  | `/code/languages` | 获取支持的语言 | ✅  |
| POST | `/code/analyze`   | 分析代码    | ✅  |
| POST | `/code/format`    | 格式化代码   | ✅  |
| POST | `/code/test`      | 运行测试用例  | ✅  |

#### 文件模块

| 方法   | 路径             | 描述      | 认证 |
| ---- | -------------- | ------- | -- |
| POST | `/file/upload` | 上传RAG文档 | ✅  |

#### 图像模块

| 方法   | 路径                 | 描述   | 认证 |
| ---- | ------------------ | ---- | -- |
| POST | `/image/recognize` | 图像识别 | ✅  |

#### TTS模块

| 方法   | 路径            | 描述     | 认证 |
| ---- | ------------- | ------ | -- |
| POST | `/tts/create` | 创建语音任务 | ✅  |
| GET  | `/tts/query`  | 查询任务状态 | ✅  |

### 请求示例

```bash
# 登录
curl -X POST http://localhost:9090/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"account": "testuser", "password": "password123"}'

# 发送消息
curl -X POST http://localhost:9090/api/v1/chat/send-new \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"question": "你好", "engineType": "1"}'

# 执行代码
curl -X POST http://localhost:9090/api/v1/code/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"language": "python", "code": "print(\"Hello World\")"}'
```

详细API文档请查看 [docs/api.md](docs/api.md)

***

## 🗄️ 数据库设计

### ER图

```
┌─────────────────┐       ┌─────────────────┐
│     users       │       │  conversations  │
├─────────────────┤       ├─────────────────┤
│ id (PK)         │───┐   │ id (PK)         │
│ name            │   │   │ user_id (FK)    │───┐
│ email           │   │   │ title           │   │
│ account (UK)    │   │   │ created_at      │   │
│ password        │   │   │ updated_at      │   │
│ created_at      │   │   │ deleted_at      │   │
│ updated_at      │   │   └─────────────────┘   │
│ deleted_at      │   │           │             │
└─────────────────┘   │           │ 1:N         │
                      │           ▼             │
                      │   ┌─────────────────┐   │
                      │   │  chat_messages  │   │
                      │   ├─────────────────┤   │
                      │   │ id (PK)         │   │
                      │   │ conversation_id │◄──┘
                      │   │ user_id         │
                      │   │ content         │
                      │   │ is_from_user    │
                      │   │ created_at      │
                      │   └─────────────────┘
                      │
                      └──────────────────────► (关联)
```

详细数据库设计请查看 [docs/database.md](docs/database.md)

***

## 🚢 部署指南

### 生产环境部署

#### 1. 编译后端

```bash
# Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o viperai ./cmd/server

# Windows
go build -o viperai.exe ./cmd/server
```

#### 2. 构建前端

```bash
cd web
npm run build
```

#### 3. Nginx配置

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        root /var/www/viperai/web/dist;
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://127.0.0.1:9090;
        proxy_set_header Host $host;
        proxy_buffering off;
        proxy_cache off;
    }
}
```

#### 4. Systemd服务

```ini
[Unit]
Description=ViperAI Backend Service
After=network.target mysql.service redis.service

[Service]
Type=simple
User=viperai
WorkingDirectory=/opt/viperai
ExecStart=/opt/viperai/viperai
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

详细部署指南请查看 [docs/deployment.md](docs/deployment.md)

***

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./... -v

# 运行特定包测试
go test ./internal/service/... -v

# 测试覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 测试覆盖

| 模块         | 覆盖率  |
| ---------- | ---- |
| service    | 85%+ |
| repository | 90%+ |
| pkg/auth   | 95%+ |
| pkg/utils  | 90%+ |
| engine     | 80%+ |

***

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 贡献流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 编写单元测试
- 更新相关文档

***

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

***

***

<p align="center">
  <sub>⭐ 如果这个项目对你有帮助，请给一个 Star 支持一下！</sub>
</p>
