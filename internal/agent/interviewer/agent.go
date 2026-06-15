package interviewer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	agent "github.com/wuyang9311/happy-study/internal/agent"
	"github.com/wuyang9311/happy-study/internal/common"
	"github.com/wuyang9311/happy-study/internal/llm"

	"github.com/cloudwego/eino/schema"
)

// Interviewer 面试官 Agent
type Interviewer struct {
	provider llm.Provider
	prompts  *agent.PromptManager
}

// NewInterviewer 创建面试官
func NewInterviewer(provider llm.Provider, prompts *agent.PromptManager) *Interviewer {
	return &Interviewer{provider: provider, prompts: prompts}
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

// ====== 自适应诊断（新） ======

// AdaptiveQuestion LLM 返回的下一题
type AdaptiveQuestion struct {
	Content    string `json:"content"`
	Category   string `json:"category"`
	Difficulty string `json:"difficulty"`
}

// AdaptiveResponse LLM 返回的完整响应
type AdaptiveResponse struct {
	Action        string           `json:"action"` // "ask" or "done"
	Question      *AdaptiveQuestion `json:"question,omitempty"`
	Summary       string           `json:"summary,omitempty"`
	Confidence    string           `json:"confidence,omitempty"`
	QuestionsAsked int             `json:"questions_asked,omitempty"`
}

// GenerateFirstQuestion 生成诊断的第一道题
func (iv *Interviewer) GenerateFirstQuestion(ctx context.Context, topic, goal string) (*AdaptiveQuestion, error) {
	resp, err := iv.generateAdaptiveQuestion(ctx, topic, goal, "", 0)
	if err != nil {
		return nil, err
	}
	if resp.Action != "ask" || resp.Question == nil {
		return nil, fmt.Errorf("unexpected response: LLM returned done instead of question")
	}
	return resp.Question, nil
}

// GenerateNextQuestion 根据对话历史生成下一题（或标记完成）
func (iv *Interviewer) GenerateNextQuestion(ctx context.Context, topic, goal string, conversation string, questionsAsked int) (*AdaptiveResponse, error) {
	resp, err := iv.generateAdaptiveQuestion(ctx, topic, goal, conversation, questionsAsked)
	if err != nil {
		return nil, err
	}
	// generateAdaptiveQuestion already parsed the response
	return resp, nil
}

// generateAdaptiveQuestion 调用 LLM 生成自适应问题
func (iv *Interviewer) generateAdaptiveQuestion(ctx context.Context, topic, goal string, conversation string, questionsAsked int) (*AdaptiveResponse, error) {
	if conversation == "" {
		conversation = "（尚未提问，请生成第一道题）"
	}

	prompt, err := iv.prompts.Render("adaptive_interview", map[string]interface{}{
		"Topic":        topic,
		"Goal":         goal,
		"Conversation": conversation,
		"QuestionsAsked": questionsAsked,
	})
	if err != nil {
		return nil, fmt.Errorf("render adaptive prompt: %w", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个资深技术面试官。输出严格的 JSON 格式，不要 markdown。"),
		schema.UserMessage(prompt),
	}

	resp, err := iv.provider.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate: %w", err)
	}

	content := common.CleanJSON(resp.Content)

	var result AdaptiveResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("parse adaptive response: %w", err)
	}

	return &result, nil
}

// GenerateReportFromConversation 从对话历史生成诊断报告
func (iv *Interviewer) GenerateReportFromConversation(ctx context.Context, req *agent.TopicRequest, conversation string) (*agent.DiagnosisReport, error) {
	prompt, err := iv.prompts.Render("report", map[string]interface{}{
		"Topic":  req.Topic,
		"Goal":   req.Goal,
		"QAText": conversation,
	})
	if err != nil {
		return nil, fmt.Errorf("render report prompt: %w", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个严谨的面试评估专家，输出严格的 JSON 格式。"),
		schema.UserMessage(prompt),
	}

	resp, err := iv.provider.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate report: %w", err)
	}

	return parseReport(resp.Content)
}

// StreamFirstQuestion 流式生成第一道题，逐 token 发送
func (iv *Interviewer) StreamFirstQuestion(ctx context.Context, topic, goal string, cb func(token string) error) (*AdaptiveResponse, error) {
	return iv.streamAdaptiveQuestion(ctx, topic, goal, "", 0, cb)
}

// StreamNextQuestion 流式生成下一题，逐 token 发送
func (iv *Interviewer) StreamNextQuestion(ctx context.Context, topic, goal string, conversation string, questionsAsked int, cb func(token string) error) (*AdaptiveResponse, error) {
	return iv.streamAdaptiveQuestion(ctx, topic, goal, conversation, questionsAsked, cb)
}

// streamAdaptiveQuestion 流式调用 LLM，逐 token 回调，返回完整解析结果
func (iv *Interviewer) streamAdaptiveQuestion(ctx context.Context, topic, goal string, conversation string, questionsAsked int, cb func(token string) error) (*AdaptiveResponse, error) {
	if conversation == "" {
		conversation = "（尚未提问，请生成第一道题）"
	}

	prompt, err := iv.prompts.Render("adaptive_interview", map[string]interface{}{
		"Topic":         topic,
		"Goal":          goal,
		"Conversation":  conversation,
		"QuestionsAsked": questionsAsked,
	})
	if err != nil {
		return nil, fmt.Errorf("render adaptive prompt: %w", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个资深技术面试官。输出严格的 JSON 格式，不要 markdown。"),
		schema.UserMessage(prompt),
	}

	stream, err := iv.provider.GenerateStream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm stream: %w", err)
	}
	defer stream.Close()

	var fullContent strings.Builder
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("stream recv: %w", err)
		}
		token := msg.Content
		fullContent.WriteString(token)
		if cb != nil {
			if err := cb(token); err != nil {
				return nil, err
			}
		}
	}

	content := common.CleanJSON(fullContent.String())
	var result AdaptiveResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("parse adaptive response: %w", err)
	}

	return &result, nil
}

// GenerateAllQuestions 生成所有面试题（API 用，非交互版）
func (iv *Interviewer) GenerateAllQuestions(ctx context.Context, topic string) ([]agent.Question, error) {
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
		questions, err := iv.generateQuestions(ctx, topic, r.name, r.count)
		if err != nil {
			return nil, fmt.Errorf("generate %s questions: %w", r.name, err)
		}
		allQuestions = append(allQuestions, questions...)
	}

	return allQuestions, nil
}

// GenerateReport 生成诊断报告（API 用，非交互版）
func (iv *Interviewer) GenerateReport(ctx context.Context, req *agent.TopicRequest, answers []agent.Answer) (*agent.DiagnosisReport, error) {
	return iv.generateReport(ctx, req, answers)
}

// generateQuestions 用 LLM 生成面试题
func (iv *Interviewer) generateQuestions(ctx context.Context, topic, round string, count int) ([]agent.Question, error) {
	prompt, err := iv.prompts.Render("interviewer_questions", map[string]interface{}{
		"Topic": topic,
		"Round": round,
		"Count": count,
	})
	if err != nil {
		return nil, fmt.Errorf("render question prompt: %w", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个资深技术面试官，擅长考察候选人的技术深度和广度。输出严格的 JSON 格式。"),
		schema.UserMessage(prompt),
	}

	resp, err := iv.provider.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate: %w", err)
	}

	return parseQuestions(resp.Content)
}

func parseQuestions(content string) ([]agent.Question, error) {
	content = common.CleanJSON(content)

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

	prompt, err := iv.prompts.Render("report", map[string]interface{}{
		"Topic":  req.Topic,
		"Goal":   req.Goal,
		"QAText": qaBuilder.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("render report prompt: %w", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个严谨的面试评估专家，输出严格的 JSON 格式。"),
		schema.UserMessage(prompt),
	}

	resp, err := iv.provider.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate report: %w", err)
	}

	return parseReport(resp.Content)
}

func parseReport(content string) (*agent.DiagnosisReport, error) {
	content = common.CleanJSON(content)
	var report agent.DiagnosisReport
	if err := json.Unmarshal([]byte(content), &report); err != nil {
		return nil, fmt.Errorf("parse report: %w", err)
	}
	return &report, nil
}
