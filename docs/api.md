# ViperAI API 文档

## 概述

ViperAI 是一个智能对话平台，提供多种AI模型交互能力，包括普通对话、RAG增强检索、MCP工具调用、图像识别和语音合成等功能。

**基础URL**: `http://localhost:9090/api/v1`

**认证方式**: Bearer Token (JWT)

---

## 用户模块

### 1. 用户登录

**POST** `/user/login`

登录并获取访问令牌。

**请求体**:
```json
{
  "account": "string",
  "password": "string"
}
```

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**错误码**:
- `2003`: 用户不存在
- `2004`: 密码错误

---

### 2. 用户注册

**POST** `/user/register`

注册新用户账号。

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "password123",
  "captcha": "123456"
}
```

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**错误码**:
- `2002`: 用户已存在
- `2007`: 验证码无效

---

### 3. 发送验证码

**POST** `/user/captcha`

发送邮箱验证码。

**请求体**:
```json
{
  "email": "user@example.com"
}
```

**响应**:
```json
{
  "code": 1000,
  "message": "Success"
}
```

---

### 4. 获取用户信息

**GET** `/user/profile`

获取当前登录用户信息。

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "id": 1,
    "name": "username",
    "email": "user@example.com",
    "account": "12345678901"
  }
}
```

---

## 对话模块

### 1. 获取对话列表

**GET** `/chat/conversations`

获取用户的所有对话列表。

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "conversations": [
      {
        "id": "uuid-string",
        "title": "对话标题"
      }
    ]
  }
}
```

---

### 2. 创建新对话并发送消息

**POST** `/chat/send-new`

创建新对话并发送第一条消息。

**请求头**:
```
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "question": "你好，请介绍一下自己",
  "engineType": "1"
}
```

**engineType 说明**:
- `1`: 阿里百炼模型
- `2`: RAG增强检索
- `3`: MCP工具调用
- `4`: Ollama本地模型

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "conversationId": "uuid-string",
    "content": "你好！我是ViperAI助手..."
  }
}
```

---

### 3. 发送消息到现有对话

**POST** `/chat/send`

向现有对话发送消息。

**请求头**:
```
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "question": "请详细解释一下",
  "engineType": "1",
  "conversationId": "uuid-string"
}
```

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "content": "当然，让我为您详细解释..."
  }
}
```

---

### 4. 流式创建新对话

**POST** `/chat/stream-new`

创建新对话并以流式方式返回响应。

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json
Accept: text/event-stream
```

**请求体**:
```json
{
  "question": "写一篇关于AI的文章",
  "engineType": "1"
}
```

**响应** (Server-Sent Events):
```
data: {"conversationId": "uuid-string"}

data: 人工

data: 智能

data: （AI）

data: [DONE]
```

---

### 5. 流式发送消息

**POST** `/chat/stream`

向现有对话发送消息并以流式方式返回响应。

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json
Accept: text/event-stream
```

**请求体**:
```json
{
  "question": "继续写",
  "engineType": "1",
  "conversationId": "uuid-string"
}
```

---

### 6. 获取对话历史

**POST** `/chat/history`

获取指定对话的历史消息。

**请求头**:
```
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "conversationId": "uuid-string"
}
```

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "history": [
      {
        "isFromUser": true,
        "content": "你好"
      },
      {
        "isFromUser": false,
        "content": "你好！有什么可以帮助您的？"
      }
    ]
  }
}
```

---

## 文件模块

### 上传文件

**POST** `/file/upload`

上传文档文件用于RAG检索增强。

**请求头**:
```
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**请求体**:
```
file: <文件内容>
```

**支持的文件类型**: `.md`, `.txt`

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "filePath": "uploads/1/uuid.txt"
  }
}
```

---

## 图像识别模块

### 图像识别

**POST** `/image/recognize`

上传图片进行识别。

**请求头**:
```
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**请求体**:
```
image: <图片文件>
```

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "className": "golden retriever"
  }
}
```

---

## 语音合成模块

### 1. 创建语音任务

**POST** `/tts/create`

创建文本转语音任务。

**请求头**:
```
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "text": "你好，欢迎使用ViperAI"
}
```

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "taskId": "task-uuid-string"
  }
}
```

---

### 2. 查询语音任务

**GET** `/tts/query?taskId=<taskId>`

查询语音合成任务状态和结果。

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 1000,
  "message": "Success",
  "data": {
    "taskId": "task-uuid-string",
    "taskStatus": "Success",
    "taskResult": "https://audio-url.mp3"
  }
}
```

**taskStatus 可能的值**:
- `Pending`: 处理中
- `Running`: 正在合成
- `Success`: 合成成功
- `Failed`: 合成失败

---

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 1000 | 成功 |
| 2001 | 参数无效 |
| 2002 | 用户已存在 |
| 2003 | 用户不存在 |
| 2004 | 密码错误 |
| 2005 | Token无效 |
| 2006 | 未登录 |
| 2007 | 验证码无效 |
| 2008 | 记录不存在 |
| 3001 | 权限不足 |
| 4001 | 服务器错误 |
| 5001 | AI模型错误 |
| 6001 | TTS服务错误 |

---

## 通用说明

### 认证

除登录、注册和发送验证码接口外，其他接口均需要在请求头中携带JWT Token：

```
Authorization: Bearer <your_token>
```

### 请求格式

所有POST请求的请求体均为JSON格式，需要在请求头中指定：

```
Content-Type: application/json
```

### 响应格式

所有响应均为JSON格式，包含以下字段：

- `code`: 状态码，1000表示成功
- `message`: 状态消息
- `data`: 响应数据（可选）
