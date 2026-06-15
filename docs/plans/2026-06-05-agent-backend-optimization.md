# Happy Study Agent 后端优化实施计划

> **Goal:** 参照 Hermes Agent 的工程标准，对后端代码进行模块化、可扩展、可观测的重构

**架构原则：**
- **Provider 抽象** — 不硬编码 LLM 供应商，通过接口切换
- **关注点分离** — Handler → Service → Agent（不混层）
- **可观测** — 结构化日志代替 `fmt.Println`
- **可恢复** — Session 持久化，不丢数据
- **可测试** — 核心逻辑有单元测试覆盖

**Tech Stack:** Go 1.26, Hertz, Eino, SQLite (mattn/go-sqlite3), slog

---

## Phase 1：基础设施加固

### Task 1: 抽取共享工具函数

**Objective:** 消除 interviewer/teacher 中重复的 `cleanJSON`，提到公共包

**Files:**
- Create: `internal/agent/util.go`

**Code:**

```go
// internal/agent/util.go
package agent

import "strings"

// CleanJSON 清理 LLM 返回内容中的 markdown 代码块标记，提取纯 JSON
func CleanJSON(content string) string {
    content = strings.TrimSpace(content)
    content = strings.TrimPrefix(content, "```json")
    content = strings.TrimPrefix(content, "```")
    content = strings.TrimSuffix(content, "```")
    content = strings.TrimSpace(content)
    return content
}
```

**然后在 interviewer/agent.go 和 teacher/agent.go 中：**
- 删除各自文件中的 `cleanJSON` 函数
- 将 `cleanJSON(content)` 改为 `agent.CleanJSON(content)`

---

### Task 2: 结构化日志替换 fmt.Println

**Objective:** 用 `slog` 替代所有业务代码中的 `fmt.Println`

**Files:**
- Modify: `cmd/app/main.go`
- Modify: `internal/agent/interviewer/agent.go`
- Modify: `internal/agent/teacher/agent.go`

**Changes:**

`cmd/app/main.go` — 初始化 slog：

```go
// 在 main 函数开头
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)
```

`internal/agent/interviewer/agent.go` — 删掉所有 `fmt.Println`，改为：

```go
import "log/slog"

// 在 generateReport 开头
slog.Info("generating diagnosis report",
    "topic", req.Topic,
    "goal", req.Goal,
    "answers_count", len(answers),
)
```

`internal/agent/teacher/agent.go` — 同理：

```go
// 在 generateCurriculum 开头
slog.Info("generating curriculum",
    "topic", report.Topic,
    "overall_score", report.OverallScore,
)
```

**注意：** 删除 interviewer/agent.go:39-45 的 CLI 边框输出（那是 CLI 交互版本的残留，Web 服务不需要）

---

### Task 3: 添加中间件栈

**Objective:** 添加 logging / recovery / request-timeout 中间件

**Files:**
- Modify: `cmd/app/main.go`

**Code — 新增 3 个 middleware 函数：**

```go
// recovery middleware
func recoveryMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        defer func() {
            if r := recover(); r != nil {
                slog.Error("panic recovered", "panic", r)
                c.JSON(500, map[string]string{"error": "internal server error"})
            }
        }()
        c.Next(ctx)
    }
}

// logging middleware
func loggingMiddleware() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        start := time.Now()
        c.Next(ctx)
        slog.Info("request",
            "method", string(c.Method()),
            "path", string(c.Path()),
            "status", c.Response.StatusCode(),
            "duration", time.Since(start).String(),
        )
    }
}
```

**在 h.Use() 中注册：**

```go
h := hertz.New(server.WithMaxKeepBodySize(10 * 1024 * 1024))
h.Use(recoveryMiddleware())
h.Use(loggingMiddleware())
h.Use(cors.Default())
```

---

## Phase 2：Provider 抽象层（核心重构）

### Task 4: 定义 LLM Provider 接口

**Objective:** 抽象 LLM 供应商接口，不再直接依赖 specific model

**Files:**
- Create: `internal/agent/provider.go`
- Modify: `internal/agent/config.go`

**Code:**

```go
// internal/agent/provider.go
package agent

import (
    "context"
    "github.com/cloudwego/eino/schema"
)

// LLMProvider LLM 供应商接口
type LLMProvider interface {
    // Generate 发送消息并获取回复
    Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
    // ModelName 返回当前使用的模型名称
    ModelName() string
}

// ProviderType 供应商类型
type ProviderType string

const (
    ProviderDeepSeek ProviderType = "deepseek"
    ProviderOpenAI   ProviderType = "openai"
    ProviderMoonshot ProviderType = "moonshot"
    ProviderCustom   ProviderType = "custom"
)

// ProviderConfig 供应商配置
type ProviderConfig struct {
    Type    ProviderType `json:"type"`
    APIKey  string       `json:"api_key"`
    BaseURL string       `json:"base_url"`
    Model   string       `json:"model"`
}
```

**修改 `internal/agent/config.go`：**

```go
package agent

import (
    "os"
    "log/slog"
)

// DefaultProviderConfig 从环境变量读取 LLM 配置
func DefaultProviderConfig() *ProviderConfig {
    providerType := ProviderType(os.Getenv("LLM_PROVIDER"))
    if providerType == "" {
        providerType = ProviderDeepSeek
    }

    apiKey := os.Getenv("LLM_API_KEY")
    if apiKey == "" {
        apiKey = os.Getenv("DEEPSEEK_API_KEY")
    }

    baseURL := os.Getenv("LLM_BASE_URL")
    if baseURL == "" {
        baseURL = os.Getenv("DEEPSEEK_BASE_URL")
    }
    if baseURL == "" {
        switch providerType {
        case ProviderDeepSeek:
            baseURL = "https://api.deepseek.com/v1"
        case ProviderOpenAI:
            baseURL = "https://api.openai.com/v1"
        default:
            baseURL = "https://api.deepseek.com/v1"
        }
    }

    model := os.Getenv("LLM_MODEL")
    if model == "" {
        model = os.Getenv("DEEPSEEK_MODEL")
    }
    if model == "" {
        switch providerType {
        case ProviderDeepSeek:
            model = "deepseek-chat"
        case ProviderOpenAI:
            model = "gpt-4o-mini"
        default:
            model = "deepseek-chat"
        }
    }

    slog.Info("llm provider config",
        "provider", providerType,
        "model", model,
        "base_url", baseURL,
    )

    return &ProviderConfig{
        Type:    providerType,
        APIKey:  apiKey,
        BaseURL: baseURL,
        Model:   model,
    }
}

// Deprecated: 保留向后兼容，新代码用 DefaultProviderConfig
func DefaultLLMConfig() *LLMConfig {
    cfg := DefaultProviderConfig()
    return &LLMConfig{
        APIKey:  cfg.APIKey,
        BaseURL: cfg.BaseURL,
        Model:   cfg.Model,
    }
}
```

---

### Task 5: 实现 OpenAI-compatible Provider

**Objective:** 基于 Eino 的 openai.ChatModel 实现 LLMProvider 接口，覆盖 DeepSeek / OpenAI / Moonshot 等

**Files:**
- Create: `internal/agent/provider_openai.go`

**Code:**

```go
// internal/agent/provider_openai.go
package agent

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/cloudwego/eino-ext/components/model/openai"
    "github.com/cloudwego/eino/schema"
)

// OpenAIProvider 基于 OpenAI 兼容协议的 LLM 供应商实现
type OpenAIProvider struct {
    model  string
    llm    *openai.ChatModel
}

// NewProvider 根据配置创建 LLM 供应商
func NewProvider(ctx context.Context, cfg *ProviderConfig) (LLMProvider, error) {
    llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
        Model:   cfg.Model,
        APIKey:  cfg.APIKey,
        BaseURL: cfg.BaseURL,
    })
    if err != nil {
        return nil, fmt.Errorf("create provider %s: %w", cfg.Type, err)
    }

    slog.Info("provider initialized",
        "type", cfg.Type,
        "model", cfg.Model,
    )

    return &OpenAIProvider{
        model: cfg.Model,
        llm:   llm,
    }, nil
}

func (p *OpenAIProvider) ModelName() string {
    return p.model
}

func (p *OpenAIProvider) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
    // 带超时的 context
    ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
    defer cancel()

    resp, err := p.llm.Generate(ctx, messages)
    if err != nil {
        return nil, fmt.Errorf("provider generate: %w", err)
    }
    return resp, nil
}
```

---

### Task 6: 为 LLM call 添加重试 + backoff

**Objective:** 对 Provider.Generate 包装重试逻辑（指数退避）

**Files:**
- Create: `internal/agent/retry.go`

**Code:**

```go
// internal/agent/retry.go
package agent

import (
    "context"
    "fmt"
    "log/slog"
    "time"
)

// RetryConfig 重试配置
type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
    return &RetryConfig{
        MaxAttempts: 3,
        BaseDelay:   1 * time.Second,
        MaxDelay:    10 * time.Second,
    }
}

// GenerateWithRetry 带重试的 LLM 调用
func GenerateWithRetry(ctx context.Context, provider LLMProvider, messages []*schema.Message, rc *RetryConfig) (*schema.Message, error) {
    if rc == nil {
        rc = DefaultRetryConfig()
    }

    var lastErr error
    for attempt := 1; attempt <= rc.MaxAttempts; attempt++ {
        resp, err := provider.Generate(ctx, messages)
        if err == nil {
            return resp, nil
        }

        lastErr = err
        if attempt < rc.MaxAttempts {
            delay := rc.BaseDelay * (1 << uint(attempt-1))
            if delay > rc.MaxDelay {
                delay = rc.MaxDelay
            }

            slog.Warn("llm call failed, retrying",
                "attempt", attempt,
                "max_attempts", rc.MaxAttempts,
                "delay", delay.String(),
                "error", err,
            )

            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(delay):
            }
        }
    }

    return nil, fmt.Errorf("llm call failed after %d attempts: %w", rc.MaxAttempts, lastErr)
}
```

---

## Phase 3：Agent 层重构

### Task 7: 重构 Interviewer Agent — 使用 Provider 接口

**Objective:** Interviewer 不再直接持有 `openai.ChatModel`，改为依赖 `LLMProvider` 接口

**Files:**
- Modify: `internal/agent/interviewer/agent.go`

**Key changes:**

```go
type Interviewer struct {
    provider       agent.LLMProvider
    retryConfig    *agent.RetryConfig
}

func NewInterviewer(provider agent.LLMProvider) *Interviewer {
    return &Interviewer{
        provider:    provider,
        retryConfig: agent.DefaultRetryConfig(),
    }
}
```

- 删除 `config` 和 `llm` 字段
- 删除 `ctx` 字段（context 由参数传入，不存储）
- 所有 `iv.llm.Generate(ctx, messages)` 改为 `agent.GenerateWithRetry(ctx, iv.provider, messages, iv.retryConfig)`
- ModelName() 从 `iv.provider.ModelName()` 获取
- 删除 `fmt.Println` 输出，改为 `slog`
- 删除本地 `cleanJSON` 函数，使用 `agent.CleanJSON`
- `GenerateAllQuestions` 和 `GenerateReport` 不再使用存储的 `iv.ctx`

---

### Task 8: 重构 Teacher Agent — 使用 Provider 接口

**Objective:** 与 Interviewer 同步重构

**Files:**
- Modify: `internal/agent/teacher/agent.go`

**Same pattern as Task 7:**
- 删除 `config`、`llm`、`ctx` 字段
- 依赖 `agent.LLMProvider` 接口
- 使用 `agent.GenerateWithRetry` + `agent.CleanJSON`
- 删除 `fmt.Println`

---

## Phase 4：Service 层解耦

### Task 9: Session 数据模型 — 剥离 Agent 引用

**Objective:** Session 只存业务数据，不存 Agent 实例

**Files:**
- Modify: `internal/service/session.go`

**Changes：**

```go
// Session 业务会话 — 只存数据，不存 agent 引用
type Session struct {
    ID        string    `json:"id"`
    Topic     string    `json:"topic"`
    Goal      string    `json:"goal"`
    Answers   []agent.Answer         `json:"answers"`
    Questions []agent.Question       `json:"questions"`
    Report    *agent.DiagnosisReport `json:"report,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**SessionManager 改动：**

```go
type SessionManager struct {
    mu       sync.RWMutex
    sessions map[string]*Session
    nextID   atomic.Int64
}

func NewSessionManager() *SessionManager {
    return &SessionManager{
        sessions: make(map[string]*Session),
    }
}

func (sm *SessionManager) Create(topic, goal string) *Session {
    id := fmt.Sprintf("sess_%s_%d", time.Now().Format("20060102150405"), sm.nextID.Add(1))
    now := time.Now()
    sess := &Session{
        ID:        id,
        Topic:     topic,
        Goal:      goal,
        Answers:   make([]agent.Answer, 0),
        Questions: make([]agent.Question, 0),
        CreatedAt: now,
        UpdatedAt: now,
    }
    sm.mu.Lock()
    sm.sessions[id] = sess
    sm.mu.Unlock()
    return sess
}
```

- 删除 `interviewer` 和 `teacher` 字段
- Agent 实例由 Handler 层创建和管理，不存储在 Session 中
- `Service` 结构体可持有 agent 实例，但 Session 不持有

---

### Task 10: 重构 Handler — 关注点分离

**Objective:** Handler 负责 HTTP 请求/响应处理，调用 Service/Agent，不做业务逻辑

**Files:**
- Modify: `internal/handler/handler.go`
- Optionally create: `internal/service/diagnosis.go`

**Handler 结构：**

```go
type Handler struct {
    interviewer *interviewer.Interviewer
    teacher     *teacher.Teacher
    sessions    *service.SessionManager
}

func NewHandler(
    interviewer *interviewer.Interviewer,
    teacher *teacher.Teacher,
    sessions *service.SessionManager,
) *Handler {
    return &Handler{
        interviewer: interviewer,
        teacher:     teacher,
        sessions:    sessions,
    }
}
```

**Handler 方法只做：**
1. 解析请求参数
2. 调用 agent（通过依赖的 agent 实例）
3. 管理 Session 数据存取
4. 返回 HTTP 响应

**不再做的事情：**
- ❌ 不在 Handler 里创建 LLM 连接
- ❌ 不在 Handler 里管理 context 生命周期（由 main 统一创建）

---

## Phase 5：持久化 & 健壮性

### Task 11: SQLite Session 持久化

**Objective:** Session 数据持久化到 SQLite，重启不丢失。使用 `modernc.org/sqlite`（纯 Go 实现，无需 CGO）

```bash
go get modernc.org/sqlite
```

**Files:**
- Create: `internal/service/store.go`

**Code:**

```go
// internal/service/store.go
package service

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log/slog"
    "os"
    "path/filepath"
    "time"

    _ "modernc.org/sqlite"
)

const DefaultDBPath = "data/happy-study.db"

// Store SQLite 持久化存储
type Store struct {
    db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
    if dbPath == "" {
        dbPath = DefaultDBPath
    }
    if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
        return nil, fmt.Errorf("create db dir: %w", err)
    }

    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("ping db: %w", err)
    }

    store := &Store{db: db}
    if err := store.migrate(); err != nil {
        return nil, fmt.Errorf("migrate: %w", err)
    }

    slog.Info("store initialized", "path", dbPath)
    return store, nil
}

func (s *Store) migrate() error {
    _, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS sessions (
            id         TEXT PRIMARY KEY,
            topic      TEXT NOT NULL,
            goal       TEXT NOT NULL DEFAULT '',
            answers    TEXT NOT NULL DEFAULT '[]',
            questions  TEXT NOT NULL DEFAULT '[]',
            report     TEXT,
            created_at TEXT NOT NULL,
            updated_at TEXT NOT NULL
        )
    `)
    return err
}

func (s *Store) SaveSession(sess *Session) error {
    answersJSON, _ := json.Marshal(sess.Answers)
    questionsJSON, _ := json.Marshal(sess.Questions)
    var reportJSON *string
    if sess.Report != nil {
        b, _ := json.Marshal(sess.Report)
        s := string(b)
        reportJSON = &s
    }

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    _, err := s.db.ExecContext(ctx, `
        INSERT INTO sessions (id, topic, goal, answers, questions, report, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET
            topic=excluded.topic, goal=excluded.goal,
            answers=excluded.answers, questions=excluded.questions,
            report=excluded.report, updated_at=excluded.updated_at`,
        sess.ID, sess.Topic, sess.Goal,
        answersJSON, questionsJSON, reportJSON,
        sess.CreatedAt.Format(time.RFC3339),
        sess.UpdatedAt.Format(time.RFC3339),
    )
    return err
}

// 再加上 LoadSession / ListSessions / DeleteSession 等方法
// ...
```

**SessionManager 改为同时写入 SQLite：**

```go
func (sm *SessionManager) SetStore(store *Store) {
    sm.store = store
}

func (sm *SessionManager) Create(topic, goal string) *Session {
    sess := // ...创建session...
    if sm.store != nil {
        if err := sm.store.SaveSession(sess); err != nil {
            slog.Error("failed to save session", "id", sess.ID, "error", err)
        }
    }
    return sess
}
```

---

### Task 12: Session 过期自动清理

**Objective:** 定期清理超过 24 小时未使用的 Session

**Files:**
- Modify: `internal/service/session.go`

**Code:**

```go
// StartCleanup 启动后台 goroutine 定期清理过期 session
func (sm *SessionManager) StartCleanup(ctx context.Context, interval time.Duration, ttl time.Duration) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                sm.cleanupExpired(ttl)
            }
        }
    }()
}

func (sm *SessionManager) cleanupExpired(ttl time.Duration) {
    now := time.Now()
    threshold := now.Add(-ttl)

    sm.mu.Lock()
    defer sm.mu.Unlock()

    var expired []string
    for id, sess := range sm.sessions {
        if sess.UpdatedAt.Before(threshold) {
            expired = append(expired, id)
        }
    }
    for _, id := range expired {
        delete(sm.sessions, id)
    }
    if len(expired) > 0 {
        slog.Info("cleaned up expired sessions", "count", len(expired))
    }
}
```

---

## Phase 6：Prompt 模板化

### Task 13: 将 Prompt 抽离为独立模板文件

**Objective:** 不再在 Go 代码中硬编码 prompt，改为读取外部 `.prompt` 文件

**Files:**
- Create: `internal/agent/prompts/interviewer_breadth.prompt`
- Create: `internal/agent/prompts/interviewer_deep.prompt`
- Create: `internal/agent/prompts/interviewer_comprehensive.prompt`
- Create: `internal/agent/prompts/report_generation.prompt`
- Create: `internal/agent/prompts/curriculum_generation.prompt`
- Create: `internal/agent/prompts/lesson_generation.prompt`
- Create: `internal/agent/prompts.go`

**示例 prompt 文件：**

```yaml
# internal/agent/prompts/report_generation.prompt
system: >
  你是一个严谨的面试评估专家，根据问答记录生成诊断报告。
  输出严格的 JSON 格式，不要 markdown 代码块。

user: >
  学习主题：{{.Topic}}
  目标职级：{{.Goal}}

  问答记录：
  {{.QA}}

  请分析候选人的知识掌握情况，以 JSON 格式返回：

  {
    "topic": "{{.Topic}}",
    "overall_score": 整体掌握度(0-100),
    "scores": [
      {
        "category": "知识点名称",
        "score": 掌握度(0-100),
        "level": "mastered|familiar|weak|unknown",
        "feedback": "具体评价"
      }
    ],
    "weaknesses": ["薄弱点"],
    "strengths": ["优势"],
    "summary": "综合评语",
    "target_level": "推荐目标职级",
    "estimated_weeks": 推荐学习周期(周)
  }
```

**prompt loader：**

```go
// internal/agent/prompts.go
package agent

import (
    "embed"
    "fmt"
    "strings"
    "text/template"
)

//go:embed prompts/*.prompt
var promptFS embed.FS

// PromptData 所有 prompt 共用模板变量
type PromptData struct {
    Topic    string
    Goal     string
    QA       string
    Score    float64
    // ... 按需扩展
}

// LoadPrompt 加载并渲染 prompt 模板
func LoadPrompt(name string, data any) (system, user string, err error) {
    content, err := promptFS.ReadFile(fmt.Sprintf("prompts/%s.prompt", name))
    if err != nil {
        return "", "", fmt.Errorf("load prompt %s: %w", name, err)
    }

    tmpl, err := template.New(name).Parse(string(content))
    if err != nil {
        return "", "", fmt.Errorf("parse prompt %s: %w", name, err)
    }

    var buf strings.Builder
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", "", fmt.Errorf("execute prompt %s: %w", name, err)
   }

    // 解析 system / user 部分
    lines := strings.SplitN(buf.String(), "\nuser: ", 2)
    // lines[0] = system: ...
    // lines[1] = user: ...
    system = strings.TrimPrefix(lines[0], "system: ")
    system = strings.TrimSpace(system)
    if len(lines) > 1 {
        user = strings.TrimSpace(lines[1])
    }
    return system, user, nil
}
```

---

## Phase 7：测试覆盖

### Task 14: 为工具函数加测试

**Files:**
- Create: `internal/agent/util_test.go`

```go
package agent

import "testing"

func TestCleanJSON(t *testing.T) {
    tests := []struct{
        input    string
        expected string
    }{
        {`{"key": "value"}`, `{"key": "value"}`},
        {"```json\n{\"key\": \"value\"}\n```", `{"key": "value"}`},
        {"```\n{\"key\": \"value\"}\n```", `{"key": "value"}`},
        {"  {\"key\": \"value\"}  ", `{"key": "value"}`},
    }
    for _, tt := range tests {
        got := CleanJSON(tt.input)
        if got != tt.expected {
            t.Errorf("CleanJSON(%q) = %q, want %q", tt.input, got, tt.expected)
        }
    }
}
```

### Task 15: 为 Provider 重试逻辑加测试

**Files:**
- Create: `internal/agent/retry_test.go`

测试 retry 在 transient error 时重试，在 persistent error 时放弃。

---

## Phase 8：Prompt Injection 防护

### Task 16: 用户输入清洗

**Objective:** 对用户输入的 topic/goal 做转义，防止 prompt injection

**Files:**
- Modify: `internal/agent/prompts.go`（加 escape 函数）

```go
// EscapePromptInput 转义用户输入中的特殊字符
func EscapePromptInput(s string) string {
    s = strings.ReplaceAll(s, `"`, `\"`)
    s = strings.ReplaceAll(s, "\n", " ")
    s = strings.ReplaceAll(s, "\r", " ")
    return s
}
```

---

## 执行顺序总览

```
Phase 1 — 基础设施
  Task 1  →  抽取 CleanJSON 工具函数
  Task 2  →  slog 结构化日志
  Task 3  →  middleware 栈

Phase 2 — Provider 抽象层
  Task 4  →  定义 LLMProvider 接口
  Task 5  →  OpenAI-compatible 实现
  Task 6  →  重试 + backoff

Phase 3 — Agent 层
  Task 7  →  重构 Interviewer
  Task 8  →  重构 Teacher

Phase 4 — Service 层
  Task 9  →  Session 剥离 agent 引用
  Task 10 →  Handler 重构

Phase 5 — 持久化
  Task 11 →  SQLite Session 存储
  Task 12 →  Session 过期清理

Phase 6 — Prompt 模板化
  Task 13 → .prompt 文件 + loader

Phase 7 — 测试
  Task 14 → 工具函数测试
  Task 15 → 重试逻辑测试

Phase 8 — 安全
  Task 16 → Prompt injection 防护
```

---

## 验证方式

每次修改后：
1. `go build ./...` — 确认编译通过
2. `go vet ./...` — 确认无静态问题
3. 启动后端：`go run ./cmd/app`
4. 测试 API：`curl http://localhost:8080/api/health`
5. 提交：`git commit -m "refactor: xxx"`
