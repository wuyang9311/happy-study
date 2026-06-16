package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	agent "github.com/wuyang9311/happy-study/internal/agent"
	"github.com/wuyang9311/happy-study/internal/agent/interviewer"
	"github.com/wuyang9311/happy-study/internal/agent/teacher"
	"github.com/wuyang9311/happy-study/internal/store"
)

// SessionManager 会话管理器（持久化 + Agent 解耦）
type SessionManager struct {
	mu          sync.RWMutex
	store       store.SessionStore
	counter     int
	interviewer *interviewer.Interviewer
	teacher     *teacher.Teacher
}

func NewSessionManager(s store.SessionStore, intv *interviewer.Interviewer, tchr *teacher.Teacher) *SessionManager {
	return &SessionManager{
		store:       s,
		counter:     s.Count(),
		interviewer: intv,
		teacher:     tchr,
	}
}

func (sm *SessionManager) CreateSession(topic, goal string, userID int64) (*store.SessionData, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.counter++
	sessionID := fmt.Sprintf("session_%d_%d", time.Now().Unix(), sm.counter)

	sd := &store.SessionData{
		ID:        sessionID,
		UserID:    userID,
		Topic:     topic,
		Goal:      goal,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := sm.store.Save(sd); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	return sd, nil
}

func (sm *SessionManager) SubmitAnswer(sessionID string, questionID int, content string) (*agent.Question, bool, error) {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return nil, false, err
	}

	sd.Answers = append(sd.Answers, agent.Answer{
		QuestionID: questionID,
		Content:    content,
	})

	sd.CurrentIndex++

	if err := sm.store.Save(sd); err != nil {
		return nil, false, fmt.Errorf("save after answer: %w", err)
	}

	if sd.CurrentIndex >= len(sd.Questions) {
		return nil, true, nil
	}

	return &sd.Questions[sd.CurrentIndex], false, nil
}

func (sm *SessionManager) SetQuestions(sessionID string, questions []agent.Question) error {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return err
	}

	sd.Questions = questions
	return sm.store.Save(sd)
}

func (sm *SessionManager) GenerateReport(ctx context.Context, sessionID string) (*agent.DiagnosisReport, error) {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return nil, err
	}

	if sd.Report != nil {
		return sd.Report, nil
	}

	req := &agent.TopicRequest{
		Topic: sd.Topic,
		Goal:  sd.Goal,
	}

	report, err := sm.interviewer.GenerateReport(ctx, req, sd.Answers)
	if err != nil {
		return nil, fmt.Errorf("generate report: %w", err)
	}

	sd.Report = report
	if err := sm.store.Save(sd); err != nil {
		return nil, fmt.Errorf("save report: %w", err)
	}

	return report, nil
}

func (sm *SessionManager) GenerateCurriculum(ctx context.Context, sessionID string) (*agent.Curriculum, error) {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return nil, err
	}

	if sd.Report == nil {
		return nil, fmt.Errorf("report not generated yet")
	}

	if sd.Curriculum != nil {
		return sd.Curriculum, nil
	}

	curriculum, err := sm.teacher.GenerateCurriculum(ctx, sd.Report)
	if err != nil {
		return nil, fmt.Errorf("generate curriculum: %w", err)
	}

	sd.Curriculum = curriculum
	if err := sm.store.Save(sd); err != nil {
		return nil, fmt.Errorf("save curriculum: %w", err)
	}

	return curriculum, nil
}

func (sm *SessionManager) GetReport(sessionID string) (*agent.DiagnosisReport, error) {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return nil, err
	}
	if sd.Report == nil {
		return nil, fmt.Errorf("report not ready yet")
	}
	return sd.Report, nil
}

func (sm *SessionManager) GetCurriculum(sessionID string) (*agent.Curriculum, error) {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return nil, err
	}
	if sd.Curriculum == nil {
		return nil, fmt.Errorf("curriculum not ready yet")
	}
	return sd.Curriculum, nil
}

func (sm *SessionManager) GetSessionInfo(sessionID string) map[string]interface{} {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return nil
	}

	done := sd.CurrentIndex >= len(sd.Questions)
	return map[string]interface{}{
		"session_id":       sd.ID,
		"topic":            sd.Topic,
		"goal":             sd.Goal,
		"current_index":    sd.CurrentIndex,
		"total_questions":  len(sd.Questions),
		"all_done":         done,
		"report_ready":     sd.Report != nil,
		"curriculum_ready": sd.Curriculum != nil,
		"created_at":       sd.CreatedAt,
	}
}

func (sm *SessionManager) GetTotalQuestions(sessionID string) int {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return 0
	}
	return len(sd.Questions)
}

func (sm *SessionManager) GetCurrentNumber(sessionID string) int {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return 0
	}
	return sd.CurrentIndex + 1
}

// ListUserCourses 列出某个用户的所有已生成课程方案的会话
func (sm *SessionManager) ListUserCourses(userID int64) []*store.SessionData {
	sessions := sm.store.ListByUser(userID)
	var courses []*store.SessionData
	for _, sd := range sessions {
		if sd.Curriculum != nil {
			courses = append(courses, sd)
		}
	}
	return courses
}

// GetSessionStoreData 直接获取 Store 中的 SessionData
func (sm *SessionManager) GetSessionStoreData(sessionID string) (*store.SessionData, error) {
	return sm.store.Get(sessionID)
}

// SaveLessonPlan 缓存生成的教案
func (sm *SessionManager) SaveLessonPlan(sessionID string, chapterIndex int, plan *agent.LessonPlan) error {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return err
	}
	if sd.LessonPlans == nil {
		sd.LessonPlans = make(map[int]*agent.LessonPlan)
	}
	sd.LessonPlans[chapterIndex] = plan
	return sm.store.Save(sd)
}

// SaveSectionOutlines 缓存某章的小节目录
func (sm *SessionManager) SaveSectionOutlines(sessionID string, chapterIndex int, sections []agent.Section) error {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return err
	}
	if sd.SectionOutlines == nil {
		sd.SectionOutlines = make(map[int][]agent.Section)
	}
	sd.SectionOutlines[chapterIndex] = sections
	return sm.store.Save(sd)
}

// GetSectionOutlines 获取某章的小节目录（如果已缓存）
func (sm *SessionManager) GetSectionOutlines(sessionID string, chapterIndex int) ([]agent.Section, error) {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return nil, err
	}
	if sd.SectionOutlines == nil {
		return nil, nil
	}
	sections, ok := sd.SectionOutlines[chapterIndex]
	if !ok || len(sections) == 0 {
		return nil, nil
	}
	return sections, nil
}

// SaveSectionContent 缓存某小节的学习内容
func (sm *SessionManager) SaveSectionContent(sessionID string, chapterIndex, sectionIndex int, content *agent.SectionContent) error {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%d_%d", chapterIndex, sectionIndex)
	if sd.SectionContents == nil {
		sd.SectionContents = make(map[string]*agent.SectionContent)
	}
	sd.SectionContents[key] = content
	return sm.store.Save(sd)
}

// GetSectionContent 获取某小节的学习内容（如果已缓存）
func (sm *SessionManager) GetSectionContent(sessionID string, chapterIndex, sectionIndex int) (*agent.SectionContent, error) {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return nil, err
	}
	if sd.SectionContents == nil {
		return nil, nil
	}
	key := fmt.Sprintf("%d_%d", chapterIndex, sectionIndex)
	content, ok := sd.SectionContents[key]
	if !ok {
		return nil, nil
	}
	return content, nil
}

// ====== 自适应诊断辅助方法 ======

// SaveSessionData 直接保存 SessionData
func (sm *SessionManager) SaveSessionData(sd *store.SessionData) error {
	return sm.store.Save(sd)
}

// AppendQuestion 向会话追加一道题
func (sm *SessionManager) AppendQuestion(sessionID string, q agent.Question) error {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return err
	}
	sd.Questions = append(sd.Questions, q)
	return sm.store.Save(sd)
}

// AppendAnswer 向会话追加一个答案（不管理索引）
func (sm *SessionManager) AppendAnswer(sessionID string, content string) error {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return err
	}
	nextID := len(sd.Questions) // 当前问题索引
	sd.Answers = append(sd.Answers, agent.Answer{
		QuestionID: nextID,
		Content:    content,
	})
	return sm.store.Save(sd)
}

// BuildConversation 从会话中构建对话文本（用于 LLM 上下文）
func (sm *SessionManager) BuildConversation(sessionID string) (string, error) {
	sd, err := sm.store.Get(sessionID)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for i, q := range sd.Questions {
		b.WriteString(fmt.Sprintf("面试官[%s]（%s）: %s\n", q.Category, q.Difficulty, q.Content))
		if i < len(sd.Answers) {
			b.WriteString(fmt.Sprintf("候选人: %s\n", sd.Answers[i].Content))
		} else {
			b.WriteString("候选人: （尚未回答）\n")
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}
