package teacher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	agent "github.com/wuyang9311/happy-study/internal/agent"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// Teacher 讲师 Agent
type Teacher struct {
	config *agent.LLMConfig
	llm    *openai.ChatModel
	ctx    context.Context
}

func NewTeacher(ctx context.Context, config *agent.LLMConfig) (*Teacher, error) {
	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   config.Model,
		APIKey:  config.APIKey,
		BaseURL: config.BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}
	return &Teacher{ctx: ctx, config: config, llm: llm}, nil
}

// GenerateCurriculum 根据诊断报告生成个性化课程方案（API 用，使用存储的 ctx）
func (t *Teacher) GenerateCurriculum(report *agent.DiagnosisReport) (*agent.Curriculum, error) {
	return t.generateCurriculum(t.ctx, report)
}

// GenerateLesson 生成某章节的详细教案（API 用，使用存储的 ctx）
func (t *Teacher) GenerateLesson(report *agent.DiagnosisReport, chapter *agent.CourseChapter) (*agent.LessonPlan, error) {
	return t.generateLesson(t.ctx, report, chapter)
}

func (t *Teacher) generateCurriculum(ctx context.Context, report *agent.DiagnosisReport) (*agent.Curriculum, error) {
	fmt.Println("\n📚 Teacher 正在根据诊断报告定制课程...")

	var scoresStr strings.Builder
	for _, s := range report.Scores {
		icon := ""
		switch s.Level {
		case "mastered":
			icon = "✅"
		case "familiar":
			icon = "👍"
		case "weak":
			icon = "⚠️"
		case "unknown":
			icon = "❌"
		}
		scoresStr.WriteString(fmt.Sprintf("%s %s: %.0f/100 (%s)\n", icon, s.Category, s.Score, s.Level))
	}

	prompt := fmt.Sprintf(`你是一个资深技术讲师，正在为一对一辅导准备课程。\n\n学习者背景：\n- 学习主题：%s\n- 目标职级：%s\n- 综合掌握度：%.0f/100\n- 评估周期：%d 周\n\n知识掌握情况：\n%s\n\n薄弱项：%s\n优势项：%s\n\n综合评语：%s\n\n请生成个性化课程方案，要求：\n1. 已掌握的知识点直接跳过或仅做快速回顾\n2. 薄弱项作为重点课程，分配更多课时\n3. 每章要有明确的主题、知识点清单和预估耗时\n4. 课程难度递进\n5. 最后安排综合实战和模拟面试\n\n以 JSON 格式返回：\n{\n  "topic": "%s",\n  "title": "课程总标题",\n  "goal": "课程目标描述",\n  "chapters": [\n    {\n      "title": "章节标题",\n      "description": "章节描述",\n      "duration": "预估学习时长",\n      "topics": ["知识点1", "知识点2"],\n      "difficulty": "beginner/intermediate/advanced"\n    }\n  ],\n  "total_weeks": 总周数\n}\n\n输出纯 JSON，不要 markdown 格式。`,
		report.Topic, report.TargetLevel, report.OverallScore, report.EstimatedWeeks,
		scoresStr.String(),
		strings.Join(report.Weaknesses, "、"),
		strings.Join(report.Strengths, "、"),
		report.Summary,
		report.Topic)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个经验丰富的技术讲师，擅长根据学生水平定制个性化课程。输出严格的 JSON 格式。"),
		schema.UserMessage(prompt),
	}

	resp, err := t.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate curriculum: %w", err)
	}

	return parseCurriculum(resp.Content)
}

func (t *Teacher) generateLesson(ctx context.Context, report *agent.DiagnosisReport, chapter *agent.CourseChapter) (*agent.LessonPlan, error) {
	fmt.Printf("\n📖 Teacher 正在编写「%s」章节的详细内容...\n", chapter.Title)

	prompt := fmt.Sprintf(`你是一个资深技术讲师，正在为学生编写详细教案。\n\n学生背景：\n- 主题：%s\n- 综合掌握度：%.0f/100\n- 薄弱项：%s\n\n章节信息：\n- 标题：%s\n- 描述：%s\n- 知识点：%s\n- 难度：%s\n\n请编写这一节的详细教学内容，要求：\n1. 用通俗易懂的语言讲解\n2. 包含具体的代码示例\n3. 指出常见的坑和面试考点\n4. 最后出一道练习题\n\n以 JSON 格式返回：\n{\n  "chapter_title": "章节标题",\n  "section_title": "本节标题",\n  "content": "详细教学内容",\n  "code_examples": ["代码示例1"],\n  "key_points": ["要点1"],\n  "practice_question": "练习题"\n}\n\n输出纯 JSON，不要 markdown 格式。`,
		report.Topic, report.OverallScore, strings.Join(report.Weaknesses, "、"),
		chapter.Title, chapter.Description, strings.Join(chapter.Topics, "、"), chapter.Difficulty)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个经验丰富的技术讲师，擅长把复杂概念讲清楚。输出严格的 JSON 格式。"),
		schema.UserMessage(prompt),
	}

	resp, err := t.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate lesson: %w", err)
	}

	return parseLessonPlan(resp.Content)
}

func parseCurriculum(content string) (*agent.Curriculum, error) {
	content = cleanJSON(content)
	var curriculum agent.Curriculum
	if err := json.Unmarshal([]byte(content), &curriculum); err != nil {
		return nil, fmt.Errorf("parse curriculum: %w", err)
	}
	return &curriculum, nil
}

func parseLessonPlan(content string) (*agent.LessonPlan, error) {
	content = cleanJSON(content)
	var plan agent.LessonPlan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("parse lesson plan: %w", err)
	}
	return &plan, nil
}

func cleanJSON(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	return content
}
