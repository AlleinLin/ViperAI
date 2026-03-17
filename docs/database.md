# ViperAI 数据库设计文档

## 概述

ViperAI 使用 MySQL 作为主数据库，采用 GORM 作为 ORM 框架。数据库设计遵循第三范式，支持软删除和时间戳记录。

---

## 数据库表结构

### 1. 用户表 (users)

存储用户基本信息。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 用户ID |
| name | VARCHAR(50) | | 用户昵称 |
| email | VARCHAR(100) | INDEX | 邮箱地址 |
| account | VARCHAR(50) | UNIQUE INDEX | 账号（自动生成） |
| password | VARCHAR(255) | NOT NULL | 密码（MD5加密） |
| created_at | DATETIME | | 创建时间 |
| updated_at | DATETIME | | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

**建表SQL**:
```sql
CREATE TABLE `users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) DEFAULT NULL,
  `email` varchar(100) DEFAULT NULL,
  `account` varchar(50) DEFAULT NULL,
  `password` varchar(255) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_account` (`account`),
  KEY `idx_users_email` (`email`),
  KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

### 2. 对话表 (conversations)

存储用户的对话会话信息。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | VARCHAR(36) | PRIMARY KEY | 对话ID (UUID) |
| user_id | BIGINT | INDEX, NOT NULL | 所属用户ID |
| title | VARCHAR(100) | | 对话标题 |
| created_at | DATETIME | | 创建时间 |
| updated_at | DATETIME | | 更新时间 |
| deleted_at | DATETIME | INDEX | 软删除时间 |

**建表SQL**:
```sql
CREATE TABLE `conversations` (
  `id` varchar(36) NOT NULL,
  `user_id` bigint NOT NULL,
  `title` varchar(100) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_conversations_user_id` (`user_id`),
  KEY `idx_conversations_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_conversations_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

### 3. 消息表 (chat_messages)

存储对话中的消息记录。

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | BIGINT | PRIMARY KEY, AUTO_INCREMENT | 消息ID |
| conversation_id | VARCHAR(36) | INDEX, NOT NULL | 所属对话ID |
| user_id | BIGINT | | 用户ID |
| content | TEXT | | 消息内容 |
| is_from_user | TINYINT(1) | NOT NULL | 是否来自用户 |
| created_at | DATETIME | | 创建时间 |

**建表SQL**:
```sql
CREATE TABLE `chat_messages` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `conversation_id` varchar(36) NOT NULL,
  `user_id` bigint DEFAULT NULL,
  `content` text,
  `is_from_user` tinyint(1) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_chat_messages_conversation_id` (`conversation_id`),
  CONSTRAINT `fk_messages_conversation` FOREIGN KEY (`conversation_id`) REFERENCES `conversations` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

## ER图

```
┌─────────────────────┐
│       users         │
├─────────────────────┤
│ PK  id              │───┐
│     name            │   │
│     email           │   │
│     account         │   │
│     password        │   │
│     created_at      │   │
│     updated_at      │   │
│     deleted_at      │   │
└─────────────────────┘   │
                          │ 1:N
                          │
┌─────────────────────┐   │
│   conversations     │   │
├─────────────────────┤   │
│ PK  id              │◄──┘
│ FK  user_id         │───┐
│     title           │   │
│     created_at      │   │
│     updated_at      │   │
│     deleted_at      │   │
└─────────────────────┘   │
                          │ 1:N
                          │
┌─────────────────────┐   │
│   chat_messages     │   │
├─────────────────────┤   │
│ PK  id              │   │
│ FK  conversation_id │◄──┘
│     user_id         │
│     content         │
│     is_from_user    │
│     created_at      │
└─────────────────────┘
```

---

## 关系说明

### users ↔ conversations (一对多)

- 一个用户可以创建多个对话
- 每个对话属于一个用户
- 外键约束：`conversations.user_id` → `users.id`
- 删除策略：级联删除

### conversations ↔ chat_messages (一对多)

- 一个对话包含多条消息
- 每条消息属于一个对话
- 外键约束：`chat_messages.conversation_id` → `conversations.id`
- 删除策略：级联删除

---

## 索引设计

### 主键索引

| 表名 | 索引名 | 字段 |
|------|--------|------|
| users | PRIMARY | id |
| conversations | PRIMARY | id |
| chat_messages | PRIMARY | id |

### 唯一索引

| 表名 | 索引名 | 字段 |
|------|--------|------|
| users | idx_users_account | account |

### 普通索引

| 表名 | 索引名 | 字段 | 用途 |
|------|--------|------|------|
| users | idx_users_email | email | 邮箱查询 |
| users | idx_users_deleted_at | deleted_at | 软删除过滤 |
| conversations | idx_conversations_user_id | user_id | 用户对话查询 |
| conversations | idx_conversations_deleted_at | deleted_at | 软删除过滤 |
| chat_messages | idx_chat_messages_conversation_id | conversation_id | 对话消息查询 |

---

## Redis 数据结构设计

### 验证码存储

```
Key: captcha:{email}
Type: String
Value: 6位数字验证码
TTL: 2分钟
```

### 向量索引

```
Index: knowledge:{filename}:idx
Type: FT (Full-Text + Vector)
Fields:
  - content: TEXT
  - metadata: TEXT
  - vector: VECTOR (FLOAT32, DIM=1024, COSINE)
```

### 文档存储

```
Key: knowledge:{filename}:{doc_id}
Type: Hash
Fields:
  - content: 文档内容
  - metadata: 元数据
  - vector: 向量数据
```

---

## 数据迁移

项目使用 GORM 的 AutoMigrate 功能自动创建和更新表结构：

```go
func autoMigrate() error {
    return DB.AutoMigrate(
        &domain.User{},
        &domain.Conversation{},
        &domain.ChatMessage{},
    )
}
```

---

## 性能优化建议

1. **分页查询**: 对消息列表使用分页，避免一次性加载大量数据
2. **索引优化**: 根据查询模式添加复合索引
3. **连接池配置**: 设置合理的连接池参数
4. **读写分离**: 高并发场景下考虑读写分离
5. **缓存策略**: 热点数据使用 Redis 缓存

---

## 备份策略

1. **全量备份**: 每日凌晨执行全量备份
2. **增量备份**: 每小时执行 binlog 增量备份
3. **异地备份**: 备份文件同步到异地存储
4. **恢复测试**: 定期进行备份恢复测试
