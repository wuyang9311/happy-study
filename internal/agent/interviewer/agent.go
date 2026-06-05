package interviewer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	agent "github.com/wuyang9311/happy-study/internal/agent"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// Interviewer 面试官 Agent
type Interviewer struct {
	config *agent.LLMConfig
	llm    *openai.ChatModel
	ctx    context.Context
}

// NewInterviewer 创建面试官
func NewInterviewer(ctx context.Context, config *agent.LLMConfig) (*Interviewer, error) {
	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   config.Model,
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}
	return &Interviewer{ctx: ctx, config: config, llm: llm}, nil
}

// ConductDiagnosis 执行诊断面试（三遍扫描）— CLI 交互版本
func (iv *Interviewer) ConductDiagnosis(ctx context.Context, req *agent.TopicRequest) (*agent.DiagnosisReport, error) {
	fmt.Println("\n╔══════════════════════════════════════════════════════╗")
	fmt.Println("║     🎙️  Interviewer 诊断面试                        ║")
	fmt.Println("╠══════════════════════════════════════════════════════╣")
	fmt.Printf("║  主题: %s\n", req.Topic)
	fmt.Printf("║  目标: %s\n", req.Goal)
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// 第一遍：广度扫描
	fmt.Println("📋 第一轮：广度扫描 — 快速了解你的知识覆盖面")
	fmt.Println(strings.Repeat("─", 50))
	breadthQuestions, err := iv.generateQuestions(ctx, req.Topic, "广度扫描", 5)
	if err != nil {
		return nil, fmt.Errorf("generate breadth questions: %w", err)
	}

	var allAnswers []agent.Answer
	for _, q := range breadthQuestions {
		fmt.Printf("\n❓ [%s] %s\n", q.Difficulty, q.Content)
		fmt.Print("> ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		allAnswers = append(allAnswers, agent.Answer{
			QuestionID: q.ID,
			Content:    answer,
		})
	}

	// 第二遍：深度追问
	fmt.Println("\n\n📋 第二轮：深度追问 — 深入挖掘关键知识点")
	fmt.Println(strings.Repeat("─", 50))
	deepQuestions, err := iv.generateQuestions(ctx, req.Topic, "深度追问", 3)
	if err != nil {
		return nil, fmt.Errorf("generate deep questions: %w", err)
	}

	for _, q := range deepQuestions {
		fmt.Printf("\n🔍 [%s] %s\n", q.Difficulty, q.Content)
		fmt.Print("> ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		allAnswers = append(allAnswers, agent.Answer{
			QuestionID: q.ID,
			Content:    answer,
		})
	}

	// 第三遍：综合题
	fmt.Println("\n\n📋 第三轮：综合实战 — 考察综合运用能力")
	fmt.Println(strings.Repeat("─", 50))
	comprehensiveQuestions, err := iv.generateQuestions(ctx, req.Topic, "综合题", 1)
	if err != nil {
		return nil, fmt.Errorf("generate comprehensive questions: %w", err)
	}

	for _, q := range comprehensiveQuestions {
		fmt.Printf("\n🏗️  [%s] %s\n", q.Difficulty, q.Content)
		fmt.Print("> ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		allAnswers = append(allAnswers, agent.Answer{
			QuestionID: q.ID,
			Content:    answer,
		})
	}

	// 生成诊断报告
	fmt.Println("\n\n📋 正在生成诊断报告...")
	report, err := iv.generateReport(ctx, req, allAnswers)
	if err != nil {
		return nil, fmt.Errorf("generate report: %w", err)
	}

	return report, nil
}

// GenerateAllQuestions 生成所有面试题（API 用，非交互版）
func (iv *Interviewer) GenerateAllQuestions(topic string) ([]agent.Question, error) {
	var allQuestions []agent.Question

	rounds := []struct {
		name  string
		count int
	}{
		{"广度扫描", 5},
		{"深度追问", 3},
		{"综合题", 1},
	}

	for _, r := range rounds {
		questions, err := iv.generateQuestions(iv.ctx, topic, r.name, r.count)
		if err != nil {
			return nil, fmt.Errorf("generate %s questions: %w", r.name, err)
		}
		allQuestions = append(allQuestions, questions...)
	}

	return allQuestions, nil
}

// GenerateReport 生成诊断报告（API 用，非交互版）
func (iv *Interviewer) GenerateReport(req *agent.TopicRequest, answers []agent.Answer) (*agent.DiagnosisReport, error) {
	return iv.generateReport(iv.ctx, req, answers)
}

// generateQuestions 用 LLM 生成面试题
func (iv *Interviewer) generateQuestions(ctx context.Context, topic, round string, count int) ([]agent.Question, error) {
	prompt := fmt.Sprintf(`你是一个资深技术面试官，正在面试候选人。\n\n主题：%s\n面试轮次：%s\n需要出题数量：%d\n\n请生成面试题，要求：\n1. 题目要覆盖主题的核心知识点\n2. 难度递进（从基础到进阶）\n3. 考察理解深度而非死记硬背\n4. 如果是 "%s" 轮次，出综合应用题\n\n请以 JSON 数组格式返回，每个元素包含：\n- id: 编号\n- content: 题目内容\n- category: 所属知识点分类\n- difficulty: easy/medium/hard\n- round: 当前轮次名称\n\n只返回 JSON，不要其他文字。`, topic, round, count, round)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个资深技术面试官，擅长考察候选人的技术深度和广度。输出严格的 JSON 格式。"),
		schema.UserMessage(prompt),
	}

	resp, err := iv.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate: %w", err)
	}

	return parseQuestions(resp.Content)
}

func parseQuestions(content string) ([]agent.Question, error) {
	content = cleanJSON(content)

	var questions []agent.Question
	if err := json.Unmarshal([]byte(content), &questions); err != nil {
		var wrapper struct {
			Questions []agent.Question `json:"questions"`
		}
		if err2 := json.Unmarshal([]byte(content), &wrapper); err2 != nil {
			return nil, fmt.Errorf("parse json: %w", err)
		}
		questions = wrapper.Questions
	}

	for i := range questions {
		questions[i].ID = i + 1
	}
	return questions, nil
}

func (iv *Interviewer) generateReport(ctx context.Context, req *agent.TopicRequest, answers []agent.Answer) (*agent.DiagnosisReport, error) {
	var qaBuilder strings.Builder
	for i, a := range answers {
		qaBuilder.WriteString(fmt.Sprintf("Q%d: (题目内容见上文)\n", i+1))
		qaBuilder.WriteString(fmt.Sprintf("A%d: %s\n", i+1, a.Content))
	}

	prompt := fmt.Sprintf(`你是一个资深技术面试官，基于以下面试问答记录，生成一份详细的诊断报告。\n\n学习主题：%s\n目标职级：%s\n\n问答记录：\n%s\n\n请分析候选人的知识掌握情况，以 JSON 格式返回：\n\n{\n  "topic": "%s",\n  "overall_score": 整体掌握度(0-100),\n  "scores": [\n    {\n      "category": "知识点名称",\n      "score": 掌握度(0-100),\n      "level": "mastered|familiar|weak|unknown",\n      "feedback": "具体评价"\n    }\n  ],\n  "weaknesses": ["薄弱点1", "薄弱点2"],\n  "strengths": ["优势1", "优势2"],\n  "summary": "综合评语",\n  "target_level": "推荐目标职级",\n  "estimated_weeks": 推荐学习周期(周)\n}\n\n要求：\n1. 知识点按主题的实际情况细分\n2. 评分要客观\n3. 评语要具体有针对性\n4. 输出纯 JSON，不要 markdown 格式`,
		req.Topic, req.Goal, qaBuilder.String(), req.Topic)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个严谨的面试评估专家，输出严格的 JSON 格式。"),
		schema.UserMessage(prompt),
	}

	resp, err := iv.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate report: %w", err)
	}

	return parseReport(resp.Content)
}

func parseReport(content string) (*agent.DiagnosisReport, error) {
	content = cleanJSON(content)
	var report agent.DiagnosisReport
	if err := json.Unmarshal([]byte(content), &report); err != nil {
		return nil, fmt.Errorf("parse report: %w", err)
	}
	return &report, nil
}

func cleanJSON(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	return content
}
