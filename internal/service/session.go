package service

import (
	"fmt"
	"sync"
	"time"

	agent "github.com/wuyang9311/happy-study/internal/agent"
	"github.com/wuyang9311/happy-study/internal/agent/interviewer"
	"github.com/wuyang9311/happy-study/internal/agent/teacher"
)

// Session 一次学习会话（纯内存）
type Session struct {
	ID            string
	Topic         string
	Goal          string
	Questions     []agent.Question
	Answers       []agent.Answer
	CurrentIndex  int
	Report        *agent.DiagnosisReport
	Curriculum    *agent.Curriculum
	CreatedAt     time.Time

	interviewer    *interviewer.Interviewer
	teacher        *teacher.Teacher
}

// SessionManager 内存会话管理器
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	counter  int
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) CreateSession(intv *interviewer.Interviewer, tchr *teacher.Teacher, topic, goal string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.counter++
	sessionID := fmt.Sprintf("session_%d_%d", time.Now().Unix(), sm.counter)

	session := &Session{
		ID:          sessionID,
		Topic:       topic,
		Goal:        goal,
		CreatedAt:   time.Now(),
		interviewer: intv,
		teacher:     tchr,
	}

	sm.sessions[sessionID] = session
	return session, nil
}

func (sm *SessionManager) GetCurrentQuestion(sessionID string) (*agent.Question, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if session.CurrentIndex >= len(session.Questions) {
		return nil, nil
	}

	return &session.Questions[session.CurrentIndex], nil
}

func (sm *SessionManager) SubmitAnswer(sessionID string, questionID int, content string) (*agent.Question, bool, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, false, fmt.Errorf("session not found: %s", sessionID)
	}

	session.Answers = append(session.Answers, agent.Answer{
		QuestionID: questionID,
		Content:    content,
	})

	session.CurrentIndex++

	if session.CurrentIndex >= len(session.Questions) {
		return nil, true, nil
	}

	return &session.Questions[session.CurrentIndex], false, nil
}

func (sm *SessionManager) GenerateReport(sessionID string) (*agent.DiagnosisReport, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Report != nil {
		return session.Report, nil
	}

	req := &agent.TopicRequest{
		Topic: session.Topic,
		Goal:  session.Goal,
	}

	report, err := session.interviewer.GenerateReport(req, session.Answers)
	if err != nil {
		return nil, fmt.Errorf("generate report: %w", err)
	}
	session.Report = report
	return report, nil
}

func (sm *SessionManager) GenerateCurriculum(sessionID string) (*agent.Curriculum, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Report == nil {
		return nil, fmt.Errorf("report not generated yet")
	}

	if session.Curriculum != nil {
		return session.Curriculum, nil
	}

	curriculum, err := session.teacher.GenerateCurriculum(session.Report)
	if err != nil {
		return nil, fmt.Errorf("generate curriculum: %w", err)
	}
	session.Curriculum = curriculum
	return curriculum, nil
}

func (sm *SessionManager) GetReport(sessionID string) (*agent.DiagnosisReport, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if session.Report == nil {
		return nil, fmt.Errorf("report not ready yet")
	}
	return session.Report, nil
}

func (sm *SessionManager) GetCurriculum(sessionID string) (*agent.Curriculum, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if session.Curriculum == nil {
		return nil, fmt.Errorf("curriculum not ready yet")
	}
	return session.Curriculum, nil
}

func (sm *SessionManager) GetSessionInfo(sessionID string) map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil
	}

	done := session.CurrentIndex >= len(session.Questions)
	return map[string]interface{}{
		"session_id":       session.ID,
		"topic":            session.Topic,
		"goal":             session.Goal,
		"current_index":    session.CurrentIndex,
		"total_questions":  len(session.Questions),
		"all_done":         done,
		"report_ready":     session.Report != nil,
		"curriculum_ready": session.Curriculum != nil,
		"created_at":       session.CreatedAt,
	}
}

func (sm *SessionManager) GetTotalQuestions(sessionID string) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return 0
	}
	return len(session.Questions)
}

func (sm *SessionManager) GetCurrentNumber(sessionID string) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return 0
	}
	return session.CurrentIndex + 1
}
