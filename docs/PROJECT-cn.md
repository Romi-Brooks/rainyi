# RainYi 项目文档

> [English](../PROJECT.md) | **中文**

## 数据库设计

### users（用户表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 PK | 主键自增 |
| username | string(100) | 用户昵称，默认「用户」 |
| email | string(200) UNIQUE | 邮箱（登录用） |
| avatar | string(500) | 用户头像 URL |
| password | string(200) | bcrypt 加密密码 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### conversations（会话表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 PK | 主键自增 |
| user_id | int64 FK | 关联 users.id |
| title | string(200) | 会话标题，默认「情感陪伴」 |
| ai_nickname | string(100) | AI 昵称，默认「RainYi」 |
| ai_avatar | string(500) | AI 头像 URL |
| persona_id | int64 FK nullable | 关联 personas.id |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| last_message | (虚拟字段) | 最近一条消息的 content/role/created_at，API 返回时不入库 |

### messages（消息表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 PK | 主键自增 |
| conversation_id | int64 FK | 关联 conversations.id |
| role | string(20) | user/assistant |
| content | text | 消息内容 |
| has_attachment | bool | 是否包含附件 |
| attachment_type | string(20) | 附件类型：image/file/voice |
| attachment_url | string(500) | 附件可访问 URL |
| created_at | datetime | 发送时间 |
| is_deleted | bool | 软删除标记 |

### personas（人格/角色表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 PK | 主键自增 |
| user_id | int64 FK | 关联 users.id，0=内置人格 |
| name | string(200) | 人格名称 |
| description | text | 人格描述 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### skill_nodes（技能节点表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 PK | 主键自增 |
| persona_id | int64 FK | 关联 personas.id |
| parent_id | int64 FK nullable | 父节点 ID，支持树形结构 |
| name | string(200) | 节点名称（来自 frontmatter name） |
| description | string(500) | 节点描述 |
| file_name | string(200) | 原始文件名 |
| content | text | 原始 Markdown 内容 |
| source | string(10) | 来源标记：db / fs（MinIO 备份） |
| storage_path | string(500) | MinIO 存储路径 |
| priority | int | 优先级排序 |
| created_at | datetime | 创建时间 |

### skill_kvs（技能键值对表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 PK | 主键自增 |
| skill_node_id | int64 FK | 关联 skill_nodes.id |
| key | string(200) | 规则标题（如"你的身份"、"核心规则"） |
| value | text | 规则内容 |
| sort_order | int | 排序序号 |
| created_at | datetime | 创建时间 |

### file_records（文件记录表）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 PK | 主键自增 |
| user_id | int64 FK | 上传用户 |
| file_type | string(30) | 文件类型：avatar / ai_avatar / skill_raw / message_image |
| reference_id | int64 | 关联业务 ID（如 skill_node.id） |
| reference_type | string(30) | 关联类型：user / skill_node / message |
| original_name | string(255) | 原始文件名 |
| storage_path | string(500) | MinIO 存储路径 |
| url | string(500) | 可访问 URL |
| size | int64 | 文件大小 |
| mime_type | string(100) | MIME 类型 |
| created_at | datetime | 创建时间 |
| deleted_at | datetime | 软删除时间 |

## API 接口

### 认证
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/auth/register | 用户注册 | 否 |
| POST | /api/auth/login | 用户登录，返回 JWT | 否 |
| POST | /api/auth/logout | 登出，将 token 加入 Redis 黑名单 | 否 |
| GET | /api/ws/chat | WebSocket 聊天 | Token(Query) |

### 会话
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /api/conversations | 获取会话列表（含 last_message） | 是 |
| POST | /api/conversations | 创建会话 | 是 |
| GET | /api/conversations/:id/messages | 获取历史消息（分页） | 是 |
| DELETE | /api/conversations/:id/messages | 清空会话消息 | 是 |
| PUT | /api/conversations/:id/config | 更新 AI 头像/昵称/标题 | 是 |

### 人格 & 技能
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /api/personas | 获取人格列表 | 是 |
| POST | /api/personas | 创建人格 | 是 |
| GET | /api/personas/:id | 获取人格详情（含 skill_nodes + kvs） | 是 |
| PUT | /api/personas/:id | 更新人格 | 是 |
| DELETE | /api/personas/:id | 删除人格 | 是 |
| POST | /api/personas/:id/files | 上传技能文件（支持多文件） | 是 |
| DELETE | /api/personas/:id/files/:fileId | 删除技能文件 | 是 |
| POST | /api/personas/load | 从 skills 目录批量加载 | 是 |
| GET | /api/conversations/:id/persona | 获取会话的当前人格 | 是 |
| PUT | /api/conversations/:id/persona | 设置会话人格 | 是 |

### 文件上传
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/upload | 通用文件上传（type=avatar/message_image/...） | 是 |
| POST | /api/upload/avatar | 快捷上传头像 | 是 |
| POST | /api/upload/image | 快捷上传图片 | 是 |
| GET | /api/upload/list?type= | 获取用户文件列表 | 是 |
| DELETE | /api/upload/:id | 删除文件（MinIO + 软删 FileRecord） | 是 |
| GET | /storage/* | 代理 MinIO 文件（同源加载，无跨域） | 否 |

### 用户
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | /api/user/profile | 获取用户信息 | 是 |
| PUT | /api/user/profile | 更新用户信息 | 是 |

## 技能系统

### 架构

```
skills/目录
  ├── SKILL-DEFAULT.md
  └── rain/
       └── Emotion-Companion.md
            │
            ▼ 解析后
persona
  └── skill_node（每个 .md 文件一个节点）
        ├── skill_kv: "你的身份" → "你是一个温柔的朋友..."
        ├── skill_kv: "核心规则" → "1. 语气口语化..."
        └── skill_kv: "情绪回应" → "..."
```

每个 `.md` 文件按 `# 标题` 拆分为多个 `skill_kv` 键值对，实现结构化规则存储，提高 System Prompt 的可查询性和可覆盖性。

### 技能文件格式

```markdown
---
name: emotional-companion
description: 温柔治愈的情感陪伴机器人
---

# 你的身份
你是一个温柔、耐心、治愈的情感陪伴助手...

# 核心规则
1. 语气口语化、简短、自然
2. 善于共情，先安抚情绪再展开对话
3. 不使用列表、代码块等生硬格式
```

### 多文件支持

一个人格可以绑定多个技能文件（skill_node），互不覆盖。上传时逐个解析、分别入库。删除只影响指定节点。

### PromptCache

- **第一层**：进程内 `map[int64]string`，毫秒级读取
- **第二层**：Redis `skill:prompt:{personaID}`，进程重启后不丢失
- **失效策略**：上传/删除技能文件时自动 `Invalidate(personaID)`
- **启动预热**：启动时遍历所有内置人格，预编译 System Prompt

## Redis 使用方案

| 用途 | Key 格式 | 结构 | TTL |
|------|----------|------|-----|
| 消息上下文 | `chat:context:{convID}` | List（最多 50 条） | 24h |
| System Prompt | `skill:prompt:{personaID}` | String | 无过期 |
| Token 黑名单 | `auth:blacklist:{jti}` | String | Token 剩余有效期 |
| 技能节点 | `skill:nodes:{personaID}` | Hash | 无过期 |

## MinIO 文件存储

### 路径规则

```
avatar/{userID}/{uuid}.{ext}           ← 用户头像
avatar/ai/{convID}/{uuid}.{ext}        ← AI 头像
skill_raw/{userID}/{uuid}.md           ← 技能文件原始 .md
message_image/{convID}/{uuid}.{ext}    ← 聊天消息图片
message_file/{convID}/{uuid}.{ext}     ← 聊天消息文件
```

### 代理机制

前端通过 `http://localhost:8080/storage/{bucket}/{path}` 加载文件，Go 后端从 MinIO 读取后返回，**浏览器同源请求，无跨域问题**。Vite 开发服务器通过 `/storage` 代理规则转发到后端。

## 安全措施

- **API Key 安全**：DeepSeek API Key 仅存储在后端环境变量，前端无任何 API Key 代码
- **XSS 防护**：前端使用 textContent/sanitizeHtml 防 XSS 注入，后端使用 html.EscapeString
- **路径穿越防护**：SKILL.md 解析使用 filepath.Clean + filepath.Abs 规范化路径，仅解析指定目录
- **SQL 注入防护**：使用 GORM 参数化查询
- **跨域限制**：CORS 仅允许指定前端域名访问
- **密码安全**：使用 bcrypt 加密存储用户密码
- **JWT 认证**：Token 过期时间 72 小时，支持 Redis 黑名单吊销
- **Token 吊销**：登出时将 jti 加入 Redis，AuthMiddleware 每次请求检查黑名单
- **文件限制**：头像上传限制 1MB，通用文件上传限制 10MB
- **MinIO 安全**：MinIO 存储桶设置为公开读，仅 Go 后端有写权限
