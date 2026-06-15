package teacher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	agent "github.com/wuyang9311/happy-study/internal/agent"
	"github.com/wuyang9311/happy-study/internal/common"
	"github.com/wuyang9311/happy-study/internal/llm"

	"github.com/cloudwego/eino/schema"
)

// Teacher 讲师 Agent
type Teacher struct {
	provider llm.Provider
	prompts  *agent.PromptManager
}

func NewTeacher(provider llm.Provider, prompts *agent.PromptManager) *Teacher {
	return &Teacher{provider: provider, prompts: prompts}
}

// GenerateCurriculum 根据诊断报告生成个性化课程方案（API 用）
func (t *Teacher) GenerateCurriculum(ctx context.Context, report *agent.DiagnosisReport) (*agent.Curriculum, error) {
	return t.generateCurriculum(ctx, report)
}

// GenerateLesson 生成某章节的详细教案（API 用）
func (t *Teacher) GenerateLesson(ctx context.Context, report *agent.DiagnosisReport, chapter *agent.CourseChapter) (*agent.LessonPlan, error) {
	return t.generateLesson(ctx, report, chapter)
}

func (t *Teacher) generateCurriculum(ctx context.Context, report *agent.DiagnosisReport) (*agent.Curriculum, error) {
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

	prompt, err := t.prompts.Render("curriculum", map[string]interface{}{
		"Topic":          report.Topic,
		"TargetLevel":    report.TargetLevel,
		"OverallScore":   report.OverallScore,
		"EstimatedWeeks": report.EstimatedWeeks,
		"ScoreDetails":   scoresStr.String(),
		"Weaknesses":     strings.Join(report.Weaknesses, "、"),
		"Strengths":      strings.Join(report.Strengths, "、"),
		"Summary":        report.Summary,
	})
	if err != nil {
		return nil, fmt.Errorf("render curriculum prompt: %w", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个经验丰富的技术讲师，擅长根据学生水平定制个性化课程。输出严格的 JSON 格式。"),
		schema.UserMessage(prompt),
	}

	resp, err := t.provider.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate curriculum: %w", err)
	}

	return parseCurriculum(resp.Content)
}

func (t *Teacher) generateLesson(ctx context.Context, report *agent.DiagnosisReport, chapter *agent.CourseChapter) (*agent.LessonPlan, error) {
	prompt, err := t.prompts.Render("lesson", map[string]interface{}{
		"Topic":             report.Topic,
		"OverallScore":      report.OverallScore,
		"Weaknesses":        strings.Join(report.Weaknesses, "、"),
		"ChapterTitle":      chapter.Title,
		"ChapterDescription": chapter.Description,
		"ChapterTopics":     strings.Join(chapter.Topics, "、"),
		"ChapterDifficulty": chapter.Difficulty,
	})
	if err != nil {
		return nil, fmt.Errorf("render lesson prompt: %w", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个经验丰富的技术讲师，擅长把复杂概念讲清楚。输出严格的 JSON 格式。"),
		schema.UserMessage(prompt),
	}

	resp, err := t.provider.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm generate lesson: %w", err)
	}

	return parseLessonPlan(resp.Content)
}

func parseCurriculum(content string) (*agent.Curriculum, error) {
	content = common.CleanJSON(content)
	var curriculum agent.Curriculum
	if err := json.Unmarshal([]byte(content), &curriculum); err != nil {
		return nil, fmt.Errorf("parse curriculum: %w", err)
	}
	return &curriculum, nil
}

func parseLessonPlan(content string) (*agent.LessonPlan, error) {
	content = common.CleanJSON(content)
	var plan agent.LessonPlan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("parse lesson plan: %w", err)
	}
	return &plan, nil
}
