package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"

	agent "github.com/wuyang9311/happy-study/internal/agent"
	"github.com/wuyang9311/happy-study/internal/agent/interviewer"
	"github.com/wuyang9311/happy-study/internal/agent/teacher"
)

func main() {
	// 加载 .env 文件（不存在则忽略）
	_ = godotenv.Load()

	// CLI 参数
	topic := flag.String("topic", "", "学习主题，如 \"Go 并发编程\"")
	goal := flag.String("goal", "面试 P6", "学习目标")
	diagnosisOnly := flag.Bool("diagnosis-only", false, "只做诊断面试，不生成课程")
	flag.Parse()

	// 如果没有指定主题，交互式输入
	if *topic == "" {
		fmt.Println("🎉 Happy Study - AI 学习与面试准备平台")
		fmt.Println(strings.Repeat("=", 50))
		fmt.Println("\n📚 你想学什么？")
		fmt.Print("> ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		*topic = strings.TrimSpace(input)

		if *topic == "" {
			log.Fatal("请指定学习主题")
		}
	}

	ctx := context.Background()

	// 读取 API Key（先读环境变量，再读 .env）
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 DEEPSEEK_API_KEY 环境变量（或在 .env 文件中配置）")
	}

	llmConfig := &agent.LLMConfig{
		APIKey:  apiKey,
		BaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1"),
		Model:   getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
	}

	req := &agent.TopicRequest{
		Topic:     *topic,
		TechStack: guessTechStack(*topic),
		Goal:      *goal,
	}

	// 第一阶段：诊断面试
	interviewerAgent, err := interviewer.NewInterviewer(ctx, llmConfig)
	if err != nil {
		log.Fatalf("创建面试官失败: %v", err)
	}

	report, err := interviewerAgent.ConductDiagnosis(ctx, req)
	if err != nil {
		log.Fatalf("诊断面试失败: %v", err)
	}

	// 输出诊断报告
	printDiagnosisReport(report)

	if *diagnosisOnly {
		return
	}

	// 第二阶段：生成课程方案
	teacherAgent, err := teacher.NewTeacher(ctx, llmConfig)
	if err != nil {
		log.Fatalf("创建讲师失败: %v", err)
	}

	curriculum, err := teacherAgent.GenerateCurriculum(ctx, report)
	if err != nil {
		log.Fatalf("生成课程方案失败: %v", err)
	}

	// 输出课程方案
	printCurriculum(curriculum)

	// 第三阶段：生成第一章详细教案
	if len(curriculum.Chapters) > 0 {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("📖 以下为第一章的详细教案（预览）")
		fmt.Println(strings.Repeat("=", 60))

		lesson, err := teacherAgent.GenerateLesson(ctx, report, &curriculum.Chapters[0])
		if err != nil {
			log.Fatalf("生成教案失败: %v", err)
		}

		printLessonPlan(lesson)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ 第一阶段验证完成！")
	fmt.Println("   核心流程：用户输入 → Interviewer诊断 → Teacher备课 → 课程输出")
	fmt.Println("   已使用 Eino：ChatModel (openai) + schema.Message")
	fmt.Println(strings.Repeat("=", 60))
}

// printDiagnosisReport 输出诊断报告
func printDiagnosisReport(r *agent.DiagnosisReport) {
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("  📋 诊断报告")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("  主题：%s\n", r.Topic)
	fmt.Printf("  综合掌握度：%.0f/100\n", r.OverallScore)
	fmt.Printf("  目标水平：%s\n", r.TargetLevel)
	fmt.Printf("  建议周期：%d 周\n", r.EstimatedWeeks)
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("  知识热力图：")
	for _, s := range r.Scores {
		barLen := int(s.Score / 5)
		if barLen < 0 {
			barLen = 0
		}
		if barLen > 20 {
			barLen = 20
		}
		bar := strings.Repeat("█", barLen)
		space := strings.Repeat("░", 20-barLen)
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
		fmt.Printf("  %s %-12s %s%s %3.0f%%\n", icon, s.Category, bar, space, s.Score)
	}
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  💪 优势：%s\n", strings.Join(r.Strengths, "、"))
	fmt.Printf("  🚨 薄弱：%s\n", strings.Join(r.Weaknesses, "、"))
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  💬 评语：%s\n", r.Summary)
	fmt.Println(strings.Repeat("═", 60))
}

// printCurriculum 输出课程方案
func printCurriculum(c *agent.Curriculum) {
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("  📚 个性化课程方案")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("  %s\n", c.Title)
	fmt.Printf("  目标：%s\n", c.Goal)
	fmt.Printf("  总周期：%d 周\n", c.TotalWeeks)
	fmt.Println(strings.Repeat("─", 60))

	for i, ch := range c.Chapters {
		difficultyIcon := ""
		switch ch.Difficulty {
		case "beginner":
			difficultyIcon = "🌱"
		case "intermediate":
			difficultyIcon = "🌿"
		case "advanced":
			difficultyIcon = "🌳"
		}
		fmt.Printf("  %d. %s %s\n", i+1, difficultyIcon, ch.Title)
		fmt.Printf("     难度：%s | 时长：%s\n", ch.Difficulty, ch.Duration)
		fmt.Printf("     📝 %s\n", ch.Description)
		fmt.Printf("     📌 %s\n", strings.Join(ch.Topics, "、"))
		fmt.Println()
	}
	fmt.Println(strings.Repeat("═", 60))
}

// printLessonPlan 输出教案
func printLessonPlan(l *agent.LessonPlan) {
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Printf("  📖 %s — %s\n", l.ChapterTitle, l.SectionTitle)
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()
	fmt.Println(l.Content)
	fmt.Println()

	if len(l.CodeExamples) > 0 {
		fmt.Println("  💻 代码示例：")
		for _, code := range l.CodeExamples {
			fmt.Println(strings.Repeat("─", 40))
			fmt.Println(code)
		}
	}

	if len(l.KeyPoints) > 0 {
		fmt.Println("\n  🔑 核心要点：")
		for _, p := range l.KeyPoints {
			fmt.Printf("    • %s\n", p)
		}
	}

	if l.PracticeQuestion != "" {
		fmt.Println("\n  ✏️  练习题：")
		fmt.Printf("    %s\n", l.PracticeQuestion)
	}
	fmt.Println(strings.Repeat("─", 60))

	// 保存到文件
	savePath := fmt.Sprintf("lesson_%s.json", l.ChapterTitle)
	data, _ := json.MarshalIndent(l, "", "  ")
	os.WriteFile(savePath, data, 0644)
	fmt.Printf("\n  💾 教案已保存到 %s\n", savePath)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func guessTechStack(topic string) string {
	topicLower := strings.ToLower(topic)
	if strings.Contains(topicLower, "go") || strings.Contains(topicLower, "golang") {
		return "Go"
	}
	if strings.Contains(topicLower, "java") {
		return "Java"
	}
	if strings.Contains(topicLower, "python") {
		return "Python"
	}
	if strings.Contains(topicLower, "前端") || strings.Contains(topicLower, "react") || strings.Contains(topicLower, "vue") {
		return "Frontend"
	}
	return "General"
}
