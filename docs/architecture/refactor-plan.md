# Happy Study 后端重构计划

## 总览

当前 8 个源文件约 700 行，重构后预计 ~15 个源文件约 1200 行。
共 **10 项优化**，分 4 个阶段执行，每阶段完成后可独立编译验证。

---

## Phase 1 — 基础设施层

### P1-1: 共享工具函数 + 结构化日志

**目标：** 消除重复代码，建立统一的日志和错误处理机制

**改动：**
- `internal/common/clean.go` — 提取 `CleanJSON()`（现在两份重复）
- `internal/common/slog.go` — slog 初始化 + HTTP middleware（替代 fmt.Println）

**涉及文件：**
```
+ internal/common/clean.go
+ internal/common/slog.go
~ internal/agent/interviewer/agent.go  → 引用 common.CleanJSON
~ internal/agent/teacher/agent.go      → 引用 common.CleanJSON
```

---

## Phase 2 — LLM 层重构

### P2-1: Provider 抽象层

**目标：** 解耦 LLM 供应商，支持 DeepSeek/OpenAI/Moonshot 等切换

**改动：**
- `internal/llm/provider.go` — `Provider` 接口定义
- `internal/llm/deepseek.go` — DeepSeek 实现
- `internal/llm/openai.go` — OpenAI 兼容实现
- `internal/llm/config.go` — 配置加载，支持多供应商

**架构：**
```
Handler → Agent (依赖 Provider 接口，不依赖具体实现)
```

**涉及文件：**
```
+ internal/llm/provider.go
+ internal/llm/config.go
+ internal/llm/deepseek.go
+ internal/llm/openai.go
~ internal/agent/config.go           → 重构为 Provider 配置
~ internal/agent/interviewer/agent.go → 依赖 Provider 接口
~ internal/agent/teacher/agent.go     → 依赖 Provider 接口
```

### P2-2: LLM 调用重试 + 超时

**目标：** API 偶发失败时自动重试，避免 500 暴漏给前端

**改动：**
- `internal/llm/retry.go` — 带指数退避的重试包装器
- 所有 Agent LLM 调用 `Generate()` 包裹 retry

### P2-3: Prompt 模板化

**目标：** 把 prompt 从 Go 字符串字面量中提取出来

**改动：**
- `internal/agent/prompts/interviewer_breadth.yaml`
- `internal/agent/prompts/interviewer_deep.yaml`
- `internal/agent/prompts/interviewer_comprehensive.yaml`
- `internal/agent/prompts/report.yaml`
- `internal/agent/prompts/curriculum.yaml`
- `internal/agent/prompts/lesson.yaml`

**涉及文件：**
```
+ internal/agent/prompts/*.yaml
+ internal/agent/template.go  → Prompt 加载引擎
~ internal/agent/interviewer/agent.go → 引用模板
~ internal/agent/teacher/agent.go     → 引用模板
```

---

## Phase 3 — 架构解耦

### P3-1: Session 持久化（SQLite）

**目标：** 服务重启不丢数据

**改动：**
- `internal/store/store.go` — SQLite 接口
- `internal/store/session_store.go` — Session CRUD

**涉及文件：**
```
+ internal/store/store.go
+ internal/store/session_store.go
+ internal/store/schema.go  → 建表 SQL
~ internal/service/session.go → 底层从 map 改为 store
~ go.mod → 加 mattn/go-sqlite3 依赖
```

### P3-2: Session 解耦 Agent 引用

**目标：** Session 只存数据，不存 Agent 实例

**当前问题：**
```go
// Session 持有 Agent 引用 → Service 和 Agent 耦合
type Session struct {
    interviewer    *interviewer.Interviewer  // ❌
    teacher        *teacher.Teacher          // ❌
}
```

**改动方向：**
- Session 从 `service` 层移走 Agent 引用
- Agent 调用走 Handler → Service → callback 模式

### P3-3: Session TTL 自动清理

**目标：** 避免内存/磁盘无限增长

**改动：**
- `internal/store/cleanup.go` — 后台 goroutine 定期清理过期 session

---

## Phase 4 — 质量保障

### P4-1: 中间件栈

**目标：** 请求日志、panic 恢复、超时控制

**改动：**
- `internal/middleware/recovery.go`
- `internal/middleware/logger.go`

**涉及文件：**
```
+ internal/middleware/recovery.go
+ internal/middleware/logger.go
~ cmd/app/main.go → 注册中间件
```

### P4-2: 输入验证 & Prompt Injection 防护

**目标：** 防止恶意输入注入 prompt

**改动：**
- `internal/common/sanitize.go` — 用户输入转义清洗

### P4-3: 单元测试

**目标：** 核心 parsing + retry 逻辑有测试覆盖

**涉及文件：**
```
+ internal/llm/retry_test.go
+ internal/common/clean_test.go
```

---

## 执行顺序总表

| # | 任务 | 依赖 | 改动文件 | 编译验证 |
|---|------|------|---------|---------|
| 1 | P1-1 共享工具+日志 | 无 | +2 ~2 | ✅ go build |
| 2 | P2-1 Provider 抽象 | P1-1 | +4 ~4 | ✅ go build |
| 3 | P2-2 Retry + 超时 | P2-1 | +1 | ✅ go build |
| 4 | P2-3 Prompt 模板化 | P2-1 | +8 ~2 | ✅ go build |
| 5 | P3-1 Session 持久化 | P1-1 | +3 ~1 | ✅ go build + go test |
| 6 | P3-2 解耦 Agent 引用 | P2-1, P3-1 | ~2 | ✅ go build |
| 7 | P3-3 Session TTL | P3-1 | +1 | ✅ go build |
| 8 | P4-1 中间件栈 | P1-1 | +2 ~1 | ✅ go build |
| 9 | P4-2 输入验证 | 无 | +1 | ✅ go build |
| 10 | P4-3 单元测试 | P2-2 | +2 | ✅ go test |
