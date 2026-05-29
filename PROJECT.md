# RainYi Project Documentation

> **English** | [中文](docs/PROJECT-cn.md)

## Database Design

### users
| Field | Type | Description |
|-------|------|-------------|
| id | int64 PK | Auto-increment primary key |
| username | string(100) | User nickname, default "User" |
| email | string(200) UNIQUE | Email (used for login) |
| avatar | string(500) | User avatar URL |
| password | string(200) | Bcrypt-hashed password |
| created_at | datetime | Creation timestamp |
| updated_at | datetime | Last update timestamp |

### conversations
| Field | Type | Description |
|-------|------|-------------|
| id | int64 PK | Auto-increment primary key |
| user_id | int64 FK | References users.id |
| title | string(200) | Conversation title, default "Emotional Companion" |
| ai_nickname | string(100) | AI nickname, default "RainYi" |
| ai_avatar | string(500) | AI avatar URL |
| persona_id | int64 FK nullable | References personas.id |
| created_at | datetime | Creation timestamp |
| updated_at | datetime | Last update timestamp |
| last_message | (virtual) | Last message content/role/created_at, API-only field |

### messages
| Field | Type | Description |
|-------|------|-------------|
| id | int64 PK | Auto-increment primary key |
| conversation_id | int64 FK | References conversations.id |
| role | string(20) | user / assistant |
| content | text | Message content |
| has_attachment | bool | Whether has attachment |
| attachment_type | string(20) | image / file / voice |
| attachment_url | string(500) | Public attachment URL |
| created_at | datetime | Sent timestamp |
| is_deleted | bool | Soft delete flag |

### personas
| Field | Type | Description |
|-------|------|-------------|
| id | int64 PK | Auto-increment primary key |
| user_id | int64 FK | References users.id, 0=built-in |
| name | string(200) | Persona name |
| description | text | Persona description |
| created_at | datetime | Creation timestamp |
| updated_at | datetime | Last update timestamp |

### skill_nodes
| Field | Type | Description |
|-------|------|-------------|
| id | int64 PK | Auto-increment primary key |
| persona_id | int64 FK | References personas.id |
| parent_id | int64 FK nullable | Parent node for tree structure |
| name | string(200) | Node name (from frontmatter) |
| description | string(500) | Node description |
| file_name | string(200) | Original filename |
| content | text | Raw Markdown content |
| source | string(10) | Source: db / fs (MinIO-backed) |
| storage_path | string(500) | MinIO storage path |
| priority | int | Priority order |
| created_at | datetime | Creation timestamp |

### skill_kvs
| Field | Type | Description |
|-------|------|-------------|
| id | int64 PK | Auto-increment primary key |
| skill_node_id | int64 FK | References skill_nodes.id |
| key | string(200) | Rule title (e.g. "Your Role", "Core Rules") |
| value | text | Rule content |
| sort_order | int | Sort order |
| created_at | datetime | Creation timestamp |

### file_records
| Field | Type | Description |
|-------|------|-------------|
| id | int64 PK | Auto-increment primary key |
| user_id | int64 FK | Uploader |
| file_type | string(30) | avatar / ai_avatar / skill_raw / message_image |
| reference_id | int64 | Business reference ID |
| reference_type | string(30) | user / skill_node / message |
| original_name | string(255) | Original filename |
| storage_path | string(500) | MinIO storage path |
| url | string(500) | Public access URL |
| size | int64 | File size in bytes |
| mime_type | string(100) | MIME type |
| created_at | datetime | Creation timestamp |
| deleted_at | datetime | Soft delete timestamp |

## API Endpoints

### Auth
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/auth/register | Register | No |
| POST | /api/auth/login | Login, returns JWT | No |
| POST | /api/auth/logout | Logout, adds token to Redis blacklist | No |

### Chat
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/ws/chat | WebSocket chat | Token(Query) |

### Conversations
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/conversations | List conversations (with last_message) | Yes |
| POST | /api/conversations | Create conversation | Yes |
| GET | /api/conversations/:id/messages | Get messages (paginated) | Yes |
| DELETE | /api/conversations/:id/messages | Clear messages | Yes |
| PUT | /api/conversations/:id/config | Update AI avatar/nickname/title | Yes |

### Personas & Skills
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/personas | List personas | Yes |
| POST | /api/personas | Create persona | Yes |
| GET | /api/personas/:id | Get persona details (with skill_nodes + kvs) | Yes |
| PUT | /api/personas/:id | Update persona | Yes |
| DELETE | /api/personas/:id | Delete persona | Yes |
| POST | /api/personas/:id/files | Upload skill files (multi-file support) | Yes |
| DELETE | /api/personas/:id/files/:fileId | Delete skill file | Yes |
| POST | /api/personas/load | Batch load from skills directory | Yes |
| GET | /api/conversations/:id/persona | Get conversation persona | Yes |
| PUT | /api/conversations/:id/persona | Set conversation persona | Yes |

### File Upload
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/upload | Generic upload (type=avatar/message_image/...) | Yes |
| POST | /api/upload/avatar | Quick avatar upload | Yes |
| POST | /api/upload/image | Quick image upload | Yes |
| GET | /api/upload/list?type= | List user files | Yes |
| DELETE | /api/upload/:id | Delete file (MinIO + soft delete record) | Yes |
| GET | /storage/* | Proxy MinIO files (same-origin, no CORS) | No |

### User
| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/user/profile | Get user profile | Yes |
| PUT | /api/user/profile | Update user profile | Yes |

## Skill System

### Architecture

```
skills/ directory
  ├── SKILL-DEFAULT.md
  └── rain/
       └── Emotion-Companion.md
            │
            ▼ After parsing
persona
  └── skill_node (one per .md file)
        ├── skill_kv: "Your Role" → "You are a gentle friend..."
        ├── skill_kv: "Core Rules" → "1. Speak naturally..."
        └── skill_kv: "Emotional Response" → "..."
```

Each `.md` file is split into multiple `skill_kv` entries by `# Heading` sections, enabling structured rule storage and queryable system prompts.

### Skill File Format

```markdown
---
name: emotional-companion
description: A gentle, comforting emotional companion
---

# Your Role
You are a warm, patient, healing companion...

# Core Rules
1. Speak in a natural, conversational tone
2. Empathize first, then expand the conversation
3. Avoid lists, code blocks, or rigid formatting
```

### Multi-File Support

Each persona can have multiple skill files (skill_nodes) that coexist without overwriting. Uploads are parsed and stored individually. Deletion only affects the specified node.

### PromptCache

- **Layer 1**: In-memory `map[int64]string`, millisecond access
- **Layer 2**: Redis `skill:prompt:{personaID}`, survives process restart
- **Invalidation**: Auto-triggered on skill upload/delete
- **Startup Warmup**: Pre-compiles all built-in personas on boot

## Redis Usage

| Purpose | Key Pattern | Type | TTL |
|---------|-------------|------|-----|
| Message Context | `chat:context:{convID}` | List (max 50) | 24h |
| System Prompt | `skill:prompt:{personaID}` | String | None |
| Token Blacklist | `auth:blacklist:{jti}` | String | Remaining token TTL |
| Skill Nodes | `skill:nodes:{personaID}` | Hash | None |

## MinIO File Storage

### Path Convention

```
avatar/{userID}/{uuid}.{ext}           ← User avatar
skill_raw/{userID}/{uuid}.md           ← Raw skill file
message_image/{convID}/{uuid}.{ext}    ← Chat image attachment
message_file/{convID}/{uuid}.{ext}     ← Chat file attachment
```

### Proxy Mechanism

Frontend loads files via `http://localhost:8080/storage/{bucket}/{path}`. The Go backend fetches from MinIO and streams the response, **eliminating CORS issues** (same-origin requests). Vite dev server proxies `/storage` to the backend.

## Security Measures

- **API Key Security**: DeepSeek API Key stored only in backend env vars
- **XSS Prevention**: Frontend uses textContent/sanitizeHtml; backend uses html.EscapeString
- **Path Traversal**: SKILL.md parsing uses filepath.Clean + filepath.Abs for path normalization
- **SQL Injection**: GORM parameterized queries
- **CORS Restriction**: Only allow specified frontend domain
- **Password Security**: Bcrypt-hashed passwords
- **Token Revocation**: JWT blacklist via Redis, checked on every request
- **File Limits**: Avatar upload max 1MB, general upload max 10MB
- **MinIO Access**: Read access is public, write access restricted to Go backend only

---

> 查看中文版本： [项目文档（中文）](docs/PROJECT-cn.md)
