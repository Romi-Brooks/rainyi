# RainYi

> [English](../README.md) | **中文**

RainYi 是一款温柔治愈的情感陪伴聊天机器人，基于 LLM + Agent Skill 系统，支持连续对话、历史记忆、个性化设置及云原生存储。

## 功能清单

| 功能 | 说明 |
|------|------|
| 仿微信聊天界面 | 左侧会话列表 + 右侧聊天窗口，绿色/灰色气泡 |
| 连续对话 | 基于 WebSocket 流式响应，支持 AI "正在输入" 状态 |
| 历史上下文记忆 | 后端维护上下文队列（最大 20 条），持久化到 MySQL + Redis 缓存 |
| 多用户隔离 | JWT 认证（支持 Redis 黑名单吊销），用户数据完全隔离 |
| 多 Agent Skill 系统 | 自动解析 `/skills/*.md`，结构化存储为 SkillNode → SkillKV 树 |
| 技能热加载 | 通过 API 后台刷新技能，无需重启服务 |
| 自定义 AI 头像/昵称 | 每个会话独立支持上传头像或设置昵称，文件存储在 MinIO |
| 统一文件管理 | 通用文件上传接口，所有文件通过 FileRecord 表追踪 |
| 头像代理 | Go 后端代理 MinIO 图片到前端，消除跨域问题 |
| 清空聊天记录 | 设置页可清空当前会话所有消息，重置上下文 |
| 夜间模式 | 基于 Tailwind CSS 实现深色/浅色主题切换，全界面适配 |
| 情绪表情功能 | 输入框左侧「表情」入口按钮，打开表情选择弹窗 |
| 本地消息缓存 | 基于 IndexedDB 的消息缓存，切换会话零等待渲染 |

## 技术栈

### 前端
- **Vue 3** + **Vite 5** + **TypeScript**
- **Pinia** 状态管理
- **Tailwind CSS 3**
- **WebSocket** 实时通信
- **IndexedDB** 本地消息缓存
- 响应式设计（移动端 + 桌面端）

### 后端
- **Go 1.21+**
- **Gin** Web 框架
- **GORM** + **MySQL**
- **Redis**（上下文缓存、System Prompt 缓存、Token 黑名单）
- **MinIO**（文件存储：头像、附件、技能文件）
- **WebSocket**（gorilla/websocket）
- **JWT** 认证
- **DeepSeek API V4**

### AI 与技能
- DeepSeek API 流式响应
- SKILL.md 解析与热加载（标准 Markdown Frontmatter 格式）
- 结构化 SkillNode + SkillKV 模型（树形技能规则）
- 进程内 PromptCache + Redis 双级缓存
- 对话上下文管理（自动限制长度防溢出）

## 项目结构

```
rain-yi/
├── frontend/                  # 前端 Vue3 项目
│   ├── public/
│   │   └── vite.svg
│   ├── src/
│   │   ├── api/              # API 请求封装
│   │   │   └── index.ts
│   │   ├── assets/           # 样式
│   │   │   └── main.css
│   │   ├── components/       # 公共组件
│   │   │   ├── ChatBubble.vue
│   │   │   ├── EmojiPicker.vue
│   │   │   ├── LoadingState.vue
│   │   │   ├── TimeStamp.vue
│   │   │   └── VoiceButton.vue
│   │   ├── types/            # TypeScript 类型定义
│   │   │   └── api.ts
│   │   ├── views/            # 页面
│   │   │   ├── Login.vue
│   │   │   ├── MainChat.vue
│   │   │   └── Settings.vue
│   │   ├── store/            # Pinia 状态
│   │   │   ├── user.ts
│   │   │   ├── chat.ts
│   │   │   └── theme.ts
│   │   ├── router/           # 路由
│   │   │   └── index.ts
│   │   ├── utils/            # 工具函数
│   │   │   ├── index.ts
│   │   │   └── storage.ts    # IndexedDB 缓存层
│   │   ├── App.vue
│   │   ├── main.ts
│   │   └── env.d.ts
│   ├── .env.development          # 本地开发（不被 git 跟踪）
│   ├── .env.development.example
│   ├── .env.production           # 生产构建（不被 git 跟踪）
│   ├── .env.production.example
│   ├── .env.production.local     # APK 构建时的真实服务器地址（被 .gitignore 排除）
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   └── postcss.config.js
│
├── backend/                   # 后端 Go 项目
│   ├── cmd/
│   │   └── main.go           # 入口，组装所有依赖
│   ├── config/
│   │   ├── config.go         # 环境变量（DB、Redis、MinIO等）
│   │   ├── database.go       # MySQL 连接
│   │   └── redis.go          # Redis 客户端初始化
│   ├── controller/
│   │   ├── auth_controller.go
│   │   ├── chat_controller.go (WebSocket)
│   │   ├── conversation_controller.go
│   │   ├── persona_controller.go
│   │   ├── upload.go         # 文件上传（MinIO 后端）
│   │   └── user_controller.go
│   ├── service/
│   │   ├── ai_service.go     # DeepSeek API 调用
│   │   ├── context.go        # 上下文管理（Redis 缓存）
│   │   ├── storage.go        # FileStorage 接口 + MinIO 实现
│   │   └── websocket.go      # WebSocket Hub
│   ├── model/
│   │   └── models.go         # 全部数据模型
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── conversation_repo.go
│   │   ├── message_repo.go
│   │   ├── persona_repo.go   # Persona + SkillNode + SkillKV CRUD
│   │   └── file_repo.go      # FileRecord CRUD
│   ├── skill/
│   │   ├── loader.go         # MD 解析、SkillManager、System Prompt 拼接
│   │   └── prompt_cache.go   # 进程内 PromptCache + Redis 回写
│   ├── middleware/
│   │   └── auth.go           # JWT 中间件 + Redis 黑名单校验
│   ├── utils/
│   │   └── sanitize.go
│   ├── static/
│   │   └── default-avatar.svg
│   ├── .env
│   └── go.mod
│
├── skills/                    # 技能文件目录
│   ├── SKILL-DEFAULT.md       # 默认情感陪伴技能
│   └── rain/                  # 其他技能
│       └── Emotion-Companion.md
│
├── docs/
│   ├── README-cn.md          # 中文文档
│   └── PROJECT-cn.md         # 中文项目文档
│
├── TASKS.md                   # 开发任务清单
└── README.md
```

## 环境变量配置

### 后端 `.env`

```env
# 数据库
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password_here
DB_NAME=rain_yi

# 服务
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# JWT
JWT_SECRET=rain-yi-secret-key-change-in-production

# DeepSeek API（必须配置）
DEEPSEEK_API_KEY=your_deepseek_api_key_here
DEEPSEEK_API_URL=https://api.deepseek.com

# 前端地址（CORS）
FRONTEND_URL=http://localhost:5173

# 技能文件目录
SKILLS_DIR=../skills

# Redis（可选）
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# MinIO（可选）
MINIO_ENDPOINT=127.0.0.1:9000
MINIO_ACCESS_KEY=admin
MINIO_SECRET_KEY=your_secret_key
MINIO_BUCKET=rain-yi
MINIO_PUBLIC_URL=http://127.0.0.1:9000
MINIO_USE_SSL=false
```

## 启动步骤

### 环境要求
- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- pnpm 或 npm
- Redis（可选，不使用则降级为纯 MySQL）
- MinIO（可选，不使用则降级为纯 DB 存储）

### 1. 数据库准备

```bash
# 登录 MySQL 创建数据库
mysql -u root -p
CREATE DATABASE rain_yi CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
exit
```

### 2. 配置后端

```bash
cd backend

# 复制并修改配置
cp .env.example .env
# 编辑 .env 填入 MySQL 密码、DeepSeek API Key、Redis、MinIO 等

# 下载依赖
go mod tidy

# 启动后端
go run cmd/main.go
```

### 3. 配置前端

```bash
cd frontend

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev
```

### 4. 访问

打开浏览器访问 `http://localhost:5173`，注册账号后即可开始使用。

### 5. 构建 Android APK（Capacitor）

前端可以通过 Capacitor 打包为 Android APK。

#### 环境要求
- Android Studio（含 Android SDK）
- JDK 21+（推荐使用 Android Studio 自带的 JBR JDK）

#### 构建步骤

```bash
cd frontend

# 配置你的服务器地址
# 创建 frontend/.env.production.local，填入：
#   VITE_API_URL=http://your-server.com:8080/api
#   VITE_WS_URL=ws://your-server.com:8080/api/ws/chat
#
# 或者直接复制模板：
#   cp .env.production.example .env.production.local
#   然后编辑 .env.production.local 填入你的服务器地址

# （仅首次）添加 Android 平台
npx cap add android

# 构建 APK（编译前端 → 同步 Capacitor → 编译 APK）
pnpm cap:build
```

APK 生成位置：
```
frontend/android/app/build/outputs/apk/debug/app-debug.apk
```

> **注意：** `frontend/android/` 已被 `.gitignore` 排除，不会提交到 GitHub。

#### 构建说明
- APK 从**本地资源**加载前端界面（页面打包在 APK 内部）
- API 请求发向 `.env.production.local` 中配置的服务器地址
- 服务器地址在**编译时注入** JavaScript 代码
- 修改服务器地址后，编辑 `.env.production.local` 重新运行 `pnpm cap:build` 即可
- `.env.production.local` 已被 `.gitignore` 排除，不会误提交

### 内置技能

| 技能文件 | 名称 | 说明 |
|----------|------|------|
| SKILL-DEFAULT.md | emotional-companion | 默认情感陪伴，温和治愈风格 |

## 项目文档

详细技术文档包括数据库设计、API 端点、技能系统和安全措施，请参见：
- [项目文档（中文）](docs/PROJECT-cn.md)
- [Project Documentation (English)](PROJECT.md)

## 后续版本规划

| 功能 | 计划 |
|------|------|
| 选择性删除 | 支持选择性删除单条消息 |
| 关键事件记忆 | 清空记录后仍保留核心用户信息 |
| 表情配置 | 完整的表情选择器，支持自定义表情 |
| 语音输入/输出 | 接入 TTS 和语音识别 |
| 情绪摘要 | LLM 驱动的对话情绪分析 |
| 分布式 WebSocket | 基于 Redis Pub/Sub 的多节点支持 |
| Docker 部署 | 提供 docker-compose 一键部署 |

## License

MIT
