package agent

// TopicRequest 用户学习目标
type TopicRequest struct {
	Topic     string // 学习主题，如 "Go 并发编程"
	TechStack string // 技术栈，如 "Go", "Java"
	Goal      string // 学习目标，如 "面试 P6"
}

// Question 面试题
type Question struct {
	ID         int    `json:"id"`
	Content    string `json:"content"`
	Category   string `json:"category"`  // 所属知识点类别
	Difficulty string `json:"difficulty"` // easy/medium/hard
	Round      string `json:"round"`     // 广度扫描/深度追问/综合题
}

// Answer 用户回答
type Answer struct {
	QuestionID int    `json:"question_id"`
	Content    string `json:"content"`
}

// KnowledgeScore 知识点掌握度评分
type KnowledgeScore struct {
	Category string  `json:"category"` // 知识点名称
	Score    float64 `json:"score"`    // 0-100
	Level    string  `json:"level"`    // mastered/familiar/weak/unknown
	Feedback string  `json:"feedback"` // 评价
}

// DiagnosisReport 诊断报告
type DiagnosisReport struct {
	Topic         string           `json:"topic"`
	OverallScore  float64          `json:"overall_score"`
	Scores        []KnowledgeScore `json:"scores"`
	Weaknesses    []string         `json:"weaknesses"`
	Strengths     []string         `json:"strengths"`
	Summary       string           `json:"summary"`
	TargetLevel   string           `json:"target_level"`
	EstimatedWeeks int             `json:"estimated_weeks"`
}

// CourseChapter 课程章节
type CourseChapter struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Duration    string   `json:"duration"`
	Topics      []string `json:"topics"`
	Difficulty  string   `json:"difficulty"`
}

// Curriculum 课程方案
type Curriculum struct {
	Topic      string          `json:"topic"`
	Title      string          `json:"title"`
	Goal       string          `json:"goal"`
	Chapters   []CourseChapter `json:"chapters"`
	TotalWeeks int             `json:"total_weeks"`
}

// LessonPlan 单节课的详细教案
type LessonPlan struct {
	ChapterTitle     string   `json:"chapter_title"`
	SectionTitle     string   `json:"section_title"`
	Content          string   `json:"content"`
	CodeExamples     []string `json:"code_examples"`
	KeyPoints        []string `json:"key_points"`
	PracticeQuestion string   `json:"practice_question"`
}
