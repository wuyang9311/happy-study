package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"

	agent "github.com/wuyang9311/happy-study/internal/agent"
	"github.com/wuyang9311/happy-study/internal/agent/interviewer"
	"github.com/wuyang9311/happy-study/internal/agent/teacher"
	"github.com/wuyang9311/happy-study/internal/service"
)

// Handler 所有 API 处理函数
type Handler struct {
	sessionMgr *service.SessionManager
	intv       *interviewer.Interviewer
	tchr       *teacher.Teacher
}

func NewHandler(sm *service.SessionManager, intv *interviewer.Interviewer, tchr *teacher.Teacher) *Handler {
	return &Handler{sessionMgr: sm, intv: intv, tchr: tchr}
}

type StartDiagnosisReq struct {
	Topic string `json:"topic"`
	Goal  string `json:"goal"`
}

type StartDiagnosisResp struct {
	SessionID      string          `json:"session_id"`
	Question       *agent.Question `json:"question"`
	TotalQuestions int             `json:"total_questions"`
	QuestionNumber int             `json:"question_number"`
}

type AnswerReq struct {
	SessionID  string `json:"session_id"`
	QuestionID int    `json:"question_id"`
	Answer     string `json:"answer"`
}

type AnswerResp struct {
	Done           bool             `json:"done"`
	Question       *agent.Question  `json:"question,omitempty"`
	QuestionNumber int              `json:"question_number,omitempty"`
	Message        string           `json:"message,omitempty"`
}

type GenerateCurriculumReq struct {
	SessionID string `json:"session_id"`
}

type GenerateLessonReq struct {
	SessionID    string `json:"session_id"`
	ChapterIndex int    `json:"chapter_index"`
}

// StartDiagnosis 开始诊断面试
func (h *Handler) StartDiagnosis(ctx context.Context, c *app.RequestContext) {
	var req StartDiagnosisReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数无效"})
		return
	}
	if req.Topic == "" {
		req.Topic = "Go 并发编程"
	}
	if req.Goal == "" {
		req.Goal = "面试 P6"
	}

	session, err := h.sessionMgr.CreateSession(h.intv, h.tchr, req.Topic, req.Goal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "创建会话失败: " + err.Error()})
		return
	}

	questions, err := h.intv.GenerateAllQuestions(req.Topic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "生成题目失败: " + err.Error()})
		return
	}
	session.Questions = questions

	resp := StartDiagnosisResp{
		SessionID:      session.ID,
		TotalQuestions: len(questions),
		QuestionNumber: 1,
	}
	if len(questions) > 0 {
		resp.Question = &questions[0]
	}

	c.JSON(http.StatusOK, resp)
}

// SubmitAnswer 提交答案
func (h *Handler) SubmitAnswer(ctx context.Context, c *app.RequestContext) {
	var req AnswerReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数无效"})
		return
	}

	nextQuestion, done, err := h.sessionMgr.SubmitAnswer(req.SessionID, req.QuestionID, req.Answer)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	resp := AnswerResp{Done: done}

	if done {
		resp.Message = "诊断完成，正在生成报告..."
		info := h.sessionMgr.GetSessionInfo(req.SessionID)
		if total, ok := info["total_questions"].(int); ok {
			resp.QuestionNumber = total
		}
	} else {
		resp.Question = nextQuestion
		resp.QuestionNumber = h.sessionMgr.GetCurrentNumber(req.SessionID)
	}

	c.JSON(http.StatusOK, resp)
}

// GetReport 获取诊断报告
func (h *Handler) GetReport(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("session_id")

	report, err := h.sessionMgr.GetReport(sessionID)
	if err != nil {
		report, err = h.sessionMgr.GenerateReport(sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.H{"error": "生成报告失败: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, utils.H{"report": report})
}

// GenerateCurriculum 生成课程方案
func (h *Handler) GenerateCurriculum(ctx context.Context, c *app.RequestContext) {
	var req GenerateCurriculumReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数无效"})
		return
	}

	curriculum, err := h.sessionMgr.GenerateCurriculum(req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "生成课程失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.H{"curriculum": curriculum})
}

// GetCurriculum 获取课程方案
func (h *Handler) GetCurriculum(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("session_id")

	curriculum, err := h.sessionMgr.GetCurriculum(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.H{"curriculum": curriculum})
}

// GenerateLesson 生成章节教案
func (h *Handler) GenerateLesson(ctx context.Context, c *app.RequestContext) {
	var req GenerateLessonReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数无效"})
		return
	}

	report, err := h.sessionMgr.GetReport(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "报告未就绪: " + err.Error()})
		return
	}

	curriculum, err := h.sessionMgr.GetCurriculum(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "课程方案未就绪: " + err.Error()})
		return
	}

	if req.ChapterIndex < 0 || req.ChapterIndex >= len(curriculum.Chapters) {
		c.JSON(http.StatusBadRequest, utils.H{"error": "章节索引无效"})
		return
	}

	chapter := curriculum.Chapters[req.ChapterIndex]
	lesson, err := h.tchr.GenerateLesson(report, &chapter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "生成教案失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.H{"lesson_plan": lesson})
}

// GetSessionInfo 获取会话信息
func (h *Handler) GetSessionInfo(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("session_id")
	info := h.sessionMgr.GetSessionInfo(sessionID)
	if info == nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "会话不存在"})
		return
	}
	c.JSON(http.StatusOK, info)
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, utils.H{
		"status":  "ok",
		"service": "Happy Study API",
	})
}
