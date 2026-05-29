# RainYi

> **English** | [中文](docs/README-cn.md)

A gentle, comforting AI companion chatbot powered by LLM + Agent Skill system. Supports continuous conversation, history memory, personalized settings, and cloud-native storage.

## Features

| Feature | Description |
|---------|-------------|
| WeChat-like UI | Sidebar conversation list + chat window, green/gray message bubbles |
| Continuous Chat | WebSocket streaming responses with AI "typing" indicator |
| Context Memory | Backend maintains context queue (max 20 messages), persisted to MySQL + Redis |
| Multi-user Isolation | JWT authentication (with Redis blacklist support), complete user data isolation |
| Agent Skill System | Auto-parses `/skills/*.md` files, structured as SkillNode → SkillKV tree |
| Hot-reload Skills | Refresh skills via API without restarting the server |
| Custom AI Avatar/Nickname | Upload avatar or set nickname per conversation, stored in MinIO |
| File Upload & Management | Unified file upload endpoint, records tracked via FileRecord table |
| Avatar Proxy | Go backend proxies MinIO images to frontend, eliminating CORS issues |
| Clear Chat History | Delete all messages in a conversation, reset context |
| Dark Mode | Full UI dark/light theme toggle via Tailwind CSS |
| Emoji Picker | Emoji selection popup in the chat input area |
| Local Message Cache | IndexedDB-based message storage for instant rendering |

## Tech Stack

### Frontend
- **Vue 3** + **Vite 5** + **TypeScript**
- **Pinia** state management
- **Tailwind CSS 3**
- **WebSocket** real-time communication
- **IndexedDB** local message caching
- Responsive design (mobile + desktop)

### Backend
- **Go 1.21+**
- **Gin** Web framework
- **GORM** + **MySQL**
- **Redis** (context cache, system prompt cache, token blacklist)
- **MinIO** (file storage: avatars, attachments, skill files)
- **WebSocket** (gorilla/websocket)
- **JWT** authentication
- **DeepSeek API V4**

### AI & Skills
- DeepSeek API streaming responses
- SKILL.md parsing with hot-reload (standard Markdown Frontmatter format)
- Structured SkillNode + SkillKV model (tree-based skill rules)
- In-memory PromptCache with Redis fallback
- Conversation context management (auto-truncation to prevent overflow)

## Project Structure

```
rain-yi/
├── frontend/                  # Vue3 frontend
│   ├── public/
│   │   └── vite.svg
│   ├── src/
│   │   ├── api/              # API request layer
│   │   │   └── index.ts
│   │   ├── assets/           # Styles
│   │   │   └── main.css
│   │   ├── components/       # Shared components
│   │   │   ├── ChatBubble.vue
│   │   │   ├── EmojiPicker.vue
│   │   │   ├── LoadingState.vue
│   │   │   ├── TimeStamp.vue
│   │   │   └── VoiceButton.vue
│   │   ├── types/            # TypeScript definitions
│   │   │   └── api.ts
│   │   ├── views/            # Pages
│   │   │   ├── Login.vue
│   │   │   ├── MainChat.vue
│   │   │   └── Settings.vue
│   │   ├── store/            # Pinia stores
│   │   │   ├── user.ts
│   │   │   ├── chat.ts
│   │   │   └── theme.ts
│   │   ├── router/           # Routes
│   │   │   └── index.ts
│   │   ├── utils/            # Utilities
│   │   │   ├── index.ts
│   │   │   └── storage.ts    # IndexedDB cache layer
│   │   ├── App.vue
│   │   ├── main.ts
│   │   └── env.d.ts
│   ├── .env.development
│   ├── .env.production
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   └── postcss.config.js
│
├── backend/                   # Go backend
│   ├── cmd/
│   │   └── main.go           # Entry point, wires all dependencies
│   ├── config/
│   │   ├── config.go         # Environment variables (DB, Redis, MinIO, etc.)
│   │   ├── database.go       # MySQL connection
│   │   └── redis.go          # Redis client initialization
│   ├── controller/
│   │   ├── auth_controller.go
│   │   ├── chat_controller.go (WebSocket)
│   │   ├── conversation_controller.go
│   │   ├── persona_controller.go
│   │   ├── upload.go         # File upload endpoint (MinIO-backed)
│   │   └── user_controller.go
│   ├── service/
│   │   ├── ai_service.go     # DeepSeek API calls
│   │   ├── context.go        # Context management (Redis-backed)
│   │   ├── storage.go        # FileStorage interface + MinIO implementation
│   │   └── websocket.go      # WebSocket Hub
│   ├── model/
│   │   └── models.go         # All data models (User, Conversation, Message, Persona, SkillNode, SkillKV, FileRecord)
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── conversation_repo.go
│   │   ├── message_repo.go
│   │   ├── persona_repo.go   # Persona + SkillNode + SkillKV CRUD
│   │   └── file_repo.go      # FileRecord CRUD
│   ├── skill/
│   │   ├── loader.go         # MD parsing, SkillManager, SystemPrompt assembly
│   │   └── prompt_cache.go   # In-memory prompt cache with Redis fallback
│   ├── middleware/
│   │   └── auth.go           # JWT middleware with Redis blacklist check
│   ├── utils/
│   │   └── sanitize.go
│   ├── static/
│   │   └── default-avatar.svg
│   ├── .env
│   └── go.mod
│
├── skills/                    # Skill files directory
│   ├── SKILL-DEFAULT.md       # Default emotional companion skill
│   └── rain/
│       └── Emotion-Companion.md
│
├── docs/
│   ├── README-cn.md          # Chinese README
│   └── PROJECT-cn.md         # Chinese project documentation
│
├── TASKS.md                   # Development task list & roadmap
└── README.md
```

## Environment Variables

### Backend `.env`

```env
# Database
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password_here
DB_NAME=rain_yi

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# JWT
JWT_SECRET=rain-yi-secret-key-change-in-production

# DeepSeek API (required)
DEEPSEEK_API_KEY=your_deepseek_api_key_here
DEEPSEEK_API_URL=https://api.deepseek.com

# Frontend URL (CORS)
FRONTEND_URL=http://localhost:5173

# Skills directory (relative to backend/ or absolute)
SKILLS_DIR=../skills

# Redis (optional)
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# MinIO (optional)
MINIO_ENDPOINT=127.0.0.1:9000
MINIO_ACCESS_KEY=admin
MINIO_SECRET_KEY=your_secret_key
MINIO_BUCKET=rain-yi
MINIO_USE_SSL=false
```

## Getting Started

### Prerequisites
- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- pnpm (or npm)
- Redis (optional, fallback to MySQL without it)
- MinIO (optional, fallback to DB-only without it)

### 1. Database Setup

```bash
# Login to MySQL and create database
mysql -u root -p
CREATE DATABASE rain_yi CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
exit
```

### 2. Configure Backend

```bash
cd backend

# Copy and edit configuration
cp .env.example .env
# Edit .env to set MySQL password, DeepSeek API Key, Redis, MinIO

# Download dependencies
go mod tidy

# Start backend
go run cmd/main.go
```

### 3. Configure Frontend

```bash
cd frontend

# Install dependencies
pnpm install

# Start dev server
pnpm dev
```

### 4. Access

Open your browser at `http://localhost:5173`, register an account, and start chatting.

### Built-in Skills

| Skill File | Name | Description |
|------------|------|-------------|
| SKILL-DEFAULT.md | emotional-companion | Default emotional companion, gentle and comforting |

## Project Documentation

For detailed technical documentation including database design, API endpoints, skill system, and security measures, see:
- [Project Documentation (English)](PROJECT.md)
- [项目文档（中文）](docs/PROJECT-cn.md)

## Roadmap

| Feature | Plan |
|---------|------|
| Selectable message deletion | Support deleting individual messages |
| Key event memory | Retain core user info after clearing history |
| Full emoji system | Complete emoji picker with custom emojis |
| Voice input/output | TTS and speech recognition integration |
| Emotion summary | LLM-driven conversation mood analysis |
| Distributed WebSocket | Redis Pub/Sub for multi-node support |
| Docker deployment | One-click docker-compose deployment |

## License

MIT
