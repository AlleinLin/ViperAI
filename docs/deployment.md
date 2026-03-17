# ViperAI 部署指南

## 环境要求

### 后端环境

- Go 1.24+
- MySQL 8.0+
- Redis 7.0+
- RabbitMQ 3.12+

### 前端环境

- Node.js 18+
- npm 9+

---

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/your-org/viperai.git
cd viperai
```

### 2. 配置环境变量

创建 `.env` 文件或设置环境变量：

```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_MODEL_NAME="qwen-turbo"
export OPENAI_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"
```

### 3. 修改配置文件

编辑 `config/settings.toml`：

```toml
[app]
name = "ViperAI"
host = "0.0.0.0"
port = 9090

[database]
host = "127.0.0.1"
port = 3306
user = "root"
password = "your_password"
database = "viperai"
charset = "utf8mb4"

[cache]
host = "127.0.0.1"
port = 6379
password = ""
db = 0

[queue]
host = "localhost"
port = 5672
user = "root"
password = "your_password"
vhost = "/"

[auth]
secret = "your-secret-key-change-in-production"
issuer = "viperai"
subject = "auth"
duration = 8760

[mail]
address = "your_email@qq.com"
auth_code = "your_auth_code"
smtp_server = "smtp.qq.com"
smtp_port = 587

[ai_model]
provider = "openai"
embedding_model = "text-embedding-v4"
chat_model = "qwen-turbo"
base_url = "https://dashscope.aliyuncs.com/compatible-mode/v1"
dimension = 1024
doc_directory = "./docs"

[voice]
api_key = "baidu_api_key"
secret_key = "baidu_secret_key"
```

---

## Docker 部署

### 1. 构建镜像

创建 `Dockerfile`：

```dockerfile
# 后端 Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o viperai ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/viperai .
COPY --from=builder /app/config ./config

EXPOSE 9090
CMD ["./viperai"]
```

### 2. Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root123
      MYSQL_DATABASE: viperai
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    command: --default-authentication-plugin=mysql_native_password

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  rabbitmq:
    image: rabbitmq:3.12-management-alpine
    environment:
      RABBITMQ_DEFAULT_USER: root
      RABBITMQ_DEFAULT_PASS: root123
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

  backend:
    build: .
    ports:
      - "9090:9090"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - OPENAI_MODEL_NAME=${OPENAI_MODEL_NAME}
      - OPENAI_BASE_URL=${OPENAI_BASE_URL}
    depends_on:
      - mysql
      - redis
      - rabbitmq
    volumes:
      - ./config:/app/config
      - uploads_data:/app/uploads

  frontend:
    build:
      context: ./web
      dockerfile: Dockerfile
    ports:
      - "80:80"
    depends_on:
      - backend

volumes:
  mysql_data:
  redis_data:
  rabbitmq_data:
  uploads_data:
```

### 3. 启动服务

```bash
docker-compose up -d
```

---

## Kubernetes 部署

### 1. 创建命名空间

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: viperai
```

### 2. 创建配置映射

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: viperai-config
  namespace: viperai
data:
  settings.toml: |
    [app]
    name = "ViperAI"
    host = "0.0.0.0"
    port = 9090
    # ... 其他配置
```

### 3. 创建密钥

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: viperai-secrets
  namespace: viperai
type: Opaque
stringData:
  OPENAI_API_KEY: "your-api-key"
  DB_PASSWORD: "your-db-password"
  REDIS_PASSWORD: ""
  RABBITMQ_PASSWORD: "your-rabbitmq-password"
```

### 4. 部署后端

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: viperai-backend
  namespace: viperai
spec:
  replicas: 3
  selector:
    matchLabels:
      app: viperai-backend
  template:
    metadata:
      labels:
        app: viperai-backend
    spec:
      containers:
      - name: backend
        image: viperai/backend:latest
        ports:
        - containerPort: 9090
        envFrom:
        - secretRef:
            name: viperai-secrets
        volumeMounts:
        - name: config
          mountPath: /app/config
      volumes:
      - name: config
        configMap:
          name: viperai-config
---
apiVersion: v1
kind: Service
metadata:
  name: viperai-backend
  namespace: viperai
spec:
  selector:
    app: viperai-backend
  ports:
  - port: 9090
    targetPort: 9090
```

### 5. 部署前端

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: viperai-frontend
  namespace: viperai
spec:
  replicas: 2
  selector:
    matchLabels:
      app: viperai-frontend
  template:
    metadata:
      labels:
        app: viperai-frontend
    spec:
      containers:
      - name: frontend
        image: viperai/frontend:latest
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: viperai-frontend
  namespace: viperai
spec:
  type: LoadBalancer
  selector:
    app: viperai-frontend
  ports:
  - port: 80
    targetPort: 80
```

---

## 手动部署

### 1. 安装依赖服务

#### MySQL

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install mysql-server
sudo mysql_secure_installation

# 创建数据库
mysql -u root -p
CREATE DATABASE viperai CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

#### Redis

```bash
# Ubuntu/Debian
sudo apt install redis-server
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

#### RabbitMQ

```bash
# Ubuntu/Debian
sudo apt install rabbitmq-server
sudo systemctl enable rabbitmq-server
sudo systemctl start rabbitmq-server

# 启用管理界面
sudo rabbitmq-plugins enable rabbitmq_management
```

### 2. 编译后端

```bash
cd viperai
go mod download
go build -o viperai ./cmd/server
```

### 3. 构建前端

```bash
cd web
npm install
npm run build
```

### 4. 配置 Nginx

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
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # SSE 支持
        proxy_buffering off;
        proxy_cache off;
    }
}
```

### 5. 启动服务

```bash
# 使用 systemd 管理后端服务
sudo cat > /etc/systemd/system/viperai.service << EOF
[Unit]
Description=ViperAI Backend Service
After=network.target mysql.service redis.service rabbitmq-server.service

[Service]
Type=simple
User=viperai
WorkingDirectory=/opt/viperai
ExecStart=/opt/viperai/viperai
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable viperai
sudo systemctl start viperai
```

---

## 监控与日志

### Prometheus 监控

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'viperai'
    static_configs:
      - targets: ['localhost:9090']
```

### 日志配置

建议使用 logrotate 管理日志文件：

```
/var/log/viperai/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0644 viperai viperai
}
```

---

## 安全建议

1. **HTTPS**: 生产环境必须使用 HTTPS
2. **密钥管理**: 使用专业的密钥管理服务
3. **数据库安全**: 限制数据库访问IP，使用强密码
4. **Redis安全**: 启用密码认证
5. **RabbitMQ安全**: 修改默认端口和密码
6. **定期备份**: 设置自动备份策略
7. **安全更新**: 定期更新系统和依赖包

---

## 故障排查

### 常见问题

1. **数据库连接失败**
   - 检查 MySQL 服务状态
   - 验证连接参数
   - 检查防火墙设置

2. **Redis 连接失败**
   - 检查 Redis 服务状态
   - 验证密码配置

3. **RabbitMQ 连接失败**
   - 检查 RabbitMQ 服务状态
   - 验证用户权限

4. **AI 模型调用失败**
   - 检查 API Key 是否有效
   - 验证网络连接
   - 检查模型名称

### 日志查看

```bash
# 查看服务日志
sudo journalctl -u viperai -f

# 查看 Nginx 日志
tail -f /var/log/nginx/error.log
```

---

## 性能调优

### 后端优化

1. 调整 GOMAXPROCS
2. 优化数据库连接池
3. 启用 Redis 缓存
4. 使用消息队列异步处理

### 数据库优化

1. 配置合适的缓冲池大小
2. 优化查询语句
3. 添加必要的索引

### 前端优化

1. 启用 Gzip 压缩
2. 配置浏览器缓存
3. 使用 CDN 加速
