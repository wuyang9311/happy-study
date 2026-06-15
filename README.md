# Happy Study 🎉

> **先考后学，只学不会的** — AI 驱动的自适应学习引擎

Happy Study 是一个 AI 驱动的自适应学习平台。它通过 AI 面试官诊断你的知识水平，精确识别知识盲区，然后生成千人千面的个性化课程方案。

## 核心模式

传统的学习平台从第一课开始教，但这意味着大量时间花在已经掌握的内容上。

**Happy Study 的流程：**

```
诊断面试 → 精准备课 → 循环学习 → 验收面试
```

| 对比项 | 传统学习平台 | Happy Study |
|--------|------------|-------------|
| 起点 | 从第一课开始学 | 先诊断，从你不会的地方开始 |
| 课程 | 所有人一样 | 千人千面，动态定制 |
| 效率 | 60% 时间在学已会的 | 100% 时间在攻克盲区 |

## Architecture

```
┌─────────────┐     HTTP/JSON      ┌──────────────┐
│  Frontend    │ ◄──────────────► │   Backend     │
│  Next.js 16  │    API + SSE      │  Go / Hertz   │
│  React 19    │                   │  Eino LLM     │
└─────────────┘                    └──────┬───────┘
                                          │
                                    ┌─────▼──────┐
                                    │   MySQL    │
                                    │  Sessions  │
                                    │  + Users   │
                                    └────────────┘
```

### Backend (Go)

- **框架:** Hertz (CloudWeGo)
- **LLM 编排:** Eino (CloudWeGo)
- **LLM 供应商:** DeepSeek (流式 + 非流式)
- **存储:** MySQL (用户 + 会话数据)
- **认证:** JWT (HS256, 7天有效期)
- **当前 Agent:**
  - 🎙️ **Interviewer** — 自适应诊断面试官
  - 🧑‍🏫 **Teacher** — 课程方案生成 + 教案生成

### Frontend (Next.js)

- **框架:** Next.js 16 (App Router)
- **UI:** React 19 + Tailwind CSS 4 + Radix UI
- **图标:** Lucide React
- **图表:** Recharts

## Features

### ✅ 已实现

- **用户系统** — 注册/登录/个人信息 (JWT)
- **自适应诊断** — AI 面试官逐题诊断，根据回答动态调整难度和方向
  - 一次一道题，流式输出（SSE）
  - 正确 → 深挖细问，不会 → 换方向
  - 可随时停止，基于已有回答生成报告
- **诊断报告** — 多维度评分、强弱项分析
- **课程方案** — AI 根据诊断结果生成个性化课程大纲
- **教案生成** — AI 为每个章节生成详细教案
- **课程管理** — 用户课程列表/详情

### 🚧 规划中

- 🧭 **Mentor** (导师) — 路线规划、实时答疑
- 🔍 **Reviewer** (评审官) — AI 代码审查
- 🏗️ **Project Designer** — 实战项目设计
- 学习进度追踪与艾宾浩斯复习提醒

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+
- MySQL 8.0+

### 1. 环境变量

```bash
cp .env.example .env
# 编辑 .env，填入你的 DEEPSEEK_API_KEY
```

### 2. 配置数据库

确保 MySQL 已运行，创建数据库：

```sql
CREATE DATABASE IF NOT EXISTS happy_study CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

默认连接 `root:qwer1234@tcp(172.30.48.1:3306)/happy_study`，可在 `.env` 中通过 `MYSQL_DSN` 自定义。

### 3. 启动后端

```bash
# 安装依赖
go mod tidy

# 编译运行
go build -o app ./cmd/app && ./app

# 或使用 Makefile
make run
```

后端默认运行在 `http://localhost:8080`。

### 4. 启动前端

```bash
cd web
npm install

# 开发模式
npm run dev

# 生产模式
npm run build && npm start
```

前端默认运行在 `http://localhost:3000`。

### 5. 打开浏览器

访问 `http://localhost:3000` → 注册账号 → 输入想学的主题 → 开始诊断！

## Project Structure

```
happy-study/
├── cmd/
│   └── app/
│       └── main.go              # 应用入口，路由注册
├── internal/
│   ├── agent/                   # AI Agent 定义
│   │   ├── interviewer/         # 面试官 Agent
│   │   │   └── agent.go         # 自适应诊断逻辑
│   │   ├── teacher/             # 讲师 Agent
│   │   │   └── agent.go         # 课程/教案生成逻辑
│   │   └── prompts.go           # Prompt 模板管理
│   ├── auth/                    # 用户认证
│   │   ├── handler.go           # 注册/登录 API
│   │   ├── middleware.go        # JWT 中间件
│   │   └── store.go             # 用户模型 + CRUD
│   ├── handler/
│   │   └── handler.go           # HTTP handlers (诊断、报告、课程)
│   ├── llm/                     # LLM 供应商抽象
│   │   ├── provider.go          # Provider 接口
│   │   ├── config.go            # LLM 配置
│   │   ├── deepseek.go          # DeepSeek 实现 (eino)
│   │   └── retry.go             # 重试包装器
│   ├── middleware/
│   │   └── middleware.go        # 通用中间件 (Recovery, Logger)
│   ├── service/
│   │   └── session.go           # 会话管理业务逻辑
│   └── store/
│       └── mysql.go             # MySQL 会话存储
├── pkg/
│   └── version/                 # 版本信息
├── web/                         # 前端 (Next.js App Router)
│   └── src/
│       ├── app/
│       │   ├── page.tsx                 # 首页
│       │   ├── login/page.tsx           # 登录
│       │   ├── register/page.tsx        # 注册
│       │   ├── interview/[sessionId]/   # 自适应诊断
│       │   ├── report/[sessionId]/      # 诊断报告
│       │   ├── curriculum/[sessionId]/  # 课程方案
│       │   └── profile/courses/         # 用户课程
│       ├── components/                  # UI 组件
│       └── lib/api.ts                   # API 客户端
├── docs/                        # 设计文档
│   ├── design/                  # 产品设计
│   └── architecture/            # 技术架构
├── Makefile
├── go.mod / go.sum
└── README.md
```

## API Overview

### 认证 (无需 Token)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/login` | 登录 |

### 诊断 (需 Token)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/diagnosis/start` | 创建诊断会话 |
| GET | `/api/diagnosis/question/:session_id` | **流式(SSE)** 获取第一道题 |
| POST | `/api/diagnosis/answer` | 提交答案，获取下一题 |
| POST | `/api/diagnosis/stop` | 停止诊断，生成报告 |
| GET | `/api/diagnosis/report/:session_id` | 获取诊断报告 |

### 课程 (需 Token)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/curriculum/generate` | 生成课程方案 |
| GET | `/api/curriculum/:session_id` | 获取课程方案 |
| POST | `/api/curriculum/lesson` | 生成章节教案 |

### 用户 (需 Token)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/auth/profile` | 个人信息 |
| GET | `/api/user/courses` | 课程列表 |
| GET | `/api/user/courses/:session_id` | 课程详情 |

## Tech Stack

| Category | Technology |
|----------|-----------|
| Backend Framework | [Hertz](https://github.com/cloudwego/hertz) (CloudWeGo) |
| LLM Framework | [Eino](https://github.com/cloudwego/eino) (CloudWeGo) |
| LLM Provider | DeepSeek (V3 / R1) |
| Database | MySQL 8.0 |
| Authentication | JWT (HS256) |
| Frontend | Next.js 16 + React 19 |
| UI | Tailwind CSS 4 + Radix UI |
| Language | Go + TypeScript |

## Design Philosophy

- **自适应诊断** — AI 面试官一次一道题，回答正确→深挖，回答不会→换方向
- **千人千面** — 每个用户的课程方案基于诊断结果动态生成
- **流式优先** — 所有 LLM 交互使用 SSE 流式，减少首字等待延迟
- **三层循环** — 大循环(课程级) → 中循环(章节级) → 小循环(知识点级)

## License

MIT
