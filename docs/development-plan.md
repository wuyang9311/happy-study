# Happy Study — 开发计划

> 创建日期：2026-06-05
> 说明：按"从核心风险开始攻坚"的思路编排

---

## 总体原则

1. **从风险最高的地方开始** — 先验证 Eino 多 Agent 协作能跑通，再做周边
2. **每天都有可运行的东西** — 不埋头搭建基础设施半个月
3. **渐进引入基础设施** — 第一天纯内存，后面按需加数据库/缓存/消息队列
4. **先 CLI 验证，再做 Web** — 核心业务流程在终端跑通了，再写前端界面

---

## 第一阶段：Eino Agent 技术验证 ✅ 已完成

### 验证结果

```bash
# 完整流程
go run ./cmd/app --topic "Go 并发编程基础"

# 输出：
# ┌─────────────────────────────────┐
# │ Interviewer 三遍扫描（9道题）    │
# │ → 诊断报告（9个知识点评分）      │
# │ → Teacher 定制课程（12周方案）   │
# │ → 第一章详细教案                │
# └─────────────────────────────────┘
```

### 使用的技术
- `github.com/cloudwego/eino` — Eino 核心框架
- `github.com/cloudwego/eino-ext/components/model/openai` — ChatModel (DeepSeek)
- `github.com/cloudwego/eino/schema` — Message 类型
- DeepSeek API（OpenAI 兼容模式）

### 已验证的文件结构
```
internal/agent/
├── types.go                    # 共享类型定义
├── config.go                   # LLM 配置
├── interviewer/agent.go        # Interviewer Agent ✅
├── teacher/agent.go            # Teacher Agent ✅
└── workflow/                   # Eino Graph 编排（待构建）
```

---

## 第二阶段：最小闭环核心业务（待开始）

### 目标
跑通"先考后学"的完整闭环

### 待开发
- [ ] 引入 PostgreSQL，持久化诊断报告和课程
- [ ] 构建 Eino Graph 编排（compose.Graph）
- [ ] Mentor Agent（答疑/规划）
- [ ] Web 界面（Next.js）

---

## 开发原则

1. **每天都能跑** — 不出现"搭了两周架子还跑不起来"的情况
2. **先 CLI 后 Web** — 核心逻辑在终端验证通过，再做界面
3. **灵活替换模型** — Model Router 模式
4. **关注 Token 成本** — 每个 Agent 调用记录 token 消耗
5. **可观测性先行** — 从一开始就打好日志
