package handler

import (
	"context"
	"encoding/json"
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

// ====== 自适应诊断（新） ======

type StartDiagnosisReq struct {
	Topic string `json:"topic"`
	Goal  string `json:"goal"`
}

type AdaptiveQuestionResp struct {
	ID         int    `json:"id"`
	Content    string `json:"content"`
	Category   string `json:"category"`
	Difficulty string `json:"difficulty"`
}

type StartDiagnosisResp struct {
	SessionID string              `json:"session_id"`
	Question  *AdaptiveQuestionResp `json:"question"`
}

// StartDiagnosis 开始自适应诊断 — 只创建会话，第一题由前端 SSE 流式获取
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

	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	// 只创建会话，不调 LLM
	session, err := h.sessionMgr.CreateSession(req.Topic, req.Goal, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "创建会话失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, StartDiagnosisResp{
		SessionID: session.ID,
	})
}

type AnswerReq struct {
	SessionID string `json:"session_id"`
	Answer    string `json:"answer"`
}

type AnswerResp struct {
	Done     bool                  `json:"done"`
	Question *AdaptiveQuestionResp `json:"question,omitempty"`
	Message  string                `json:"message,omitempty"`
}

// SubmitAnswer 提交答案并获取下一题（或完成诊断）
func (h *Handler) SubmitAnswer(ctx context.Context, c *app.RequestContext) {
	var req AnswerReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数无效"})
		return
	}

	// 获取会话信息
	sd, err := h.sessionMgr.GetSessionStoreData(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "会话不存在: " + err.Error()})
		return
	}

	// 保存回答
	if err := h.sessionMgr.AppendAnswer(req.SessionID, req.Answer); err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "保存回答失败: " + err.Error()})
		return
	}

	// 构建对话上下文
	conversation, err := h.sessionMgr.BuildConversation(req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "构建对话失败: " + err.Error()})
		return
	}

	questionsAsked := len(sd.Answers)

	// 调用 LLM 生成下一题或结束
	resp, err := h.intv.GenerateNextQuestion(ctx, sd.Topic, sd.Goal, conversation, questionsAsked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "生成下一题失败: " + err.Error()})
		return
	}

	if resp.Action == "done" {
		// 诊断完成，生成报告
		topicReq := &agent.TopicRequest{Topic: sd.Topic, Goal: sd.Goal}
		report, err := h.intv.GenerateReportFromConversation(ctx, topicReq, conversation)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.H{"error": "生成报告失败: " + err.Error()})
			return
		}

		// 保存报告到会话
		sd.Report = report
		if err := h.sessionMgr.SaveSessionData(sd); err != nil {
			c.JSON(http.StatusInternalServerError, utils.H{"error": "保存报告失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, AnswerResp{
			Done:    true,
			Message: resp.Summary,
		})
		return
	}

	// 继续问下一题
	nextID := len(sd.Questions) + 1
	newQ := agent.Question{
		ID:         nextID,
		Content:    resp.Question.Content,
		Category:   resp.Question.Category,
		Difficulty: resp.Question.Difficulty,
	}
	if err := h.sessionMgr.AppendQuestion(req.SessionID, newQ); err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "保存下一题失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, AnswerResp{
		Done: false,
		Question: &AdaptiveQuestionResp{
			ID:         nextID,
			Content:    resp.Question.Content,
			Category:   resp.Question.Category,
			Difficulty: resp.Question.Difficulty,
		},
	})
}

type StopDiagnosisReq struct {
	SessionID string `json:"session_id"`
}

// StopDiagnosis 停止诊断，基于已有回答生成报告
func (h *Handler) StopDiagnosis(ctx context.Context, c *app.RequestContext) {
	var req StopDiagnosisReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "参数无效"})
		return
	}

	sd, err := h.sessionMgr.GetSessionStoreData(req.SessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "会话不存在"})
		return
	}

	// 检查是否已有报告
	if sd.Report != nil {
		c.JSON(http.StatusOK, utils.H{
			"session_id": sd.ID,
			"message":    "诊断已停止",
		})
		return
	}

	// 基于已有对话生成报告
	conversation, err := h.sessionMgr.BuildConversation(req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "构建对话失败"})
		return
	}

	ctxReq := &agent.TopicRequest{Topic: sd.Topic, Goal: sd.Goal}
	report, err := h.intv.GenerateReportFromConversation(ctx, ctxReq, conversation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "生成报告失败: " + err.Error()})
		return
	}

	sd.Report = report
	if err := h.sessionMgr.SaveSessionData(sd); err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "保存报告失败"})
		return
	}

	c.JSON(http.StatusOK, utils.H{
		"session_id": sd.ID,
		"message":    "诊断已停止，报告已生成",
	})
}

// ====== 旧接口保留（原 report/curriculum 等） ======

// GetReport 获取诊断报告
func (h *Handler) GetReport(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("session_id")

	report, err := h.sessionMgr.GetReport(sessionID)
	if err != nil {
		// 如果还没报告，尝试先生成
		sd, _ := h.sessionMgr.GetSessionStoreData(sessionID)
		if sd != nil {
			conversation, _ := h.sessionMgr.BuildConversation(sessionID)
			if conversation != "" {
				req := &agent.TopicRequest{Topic: sd.Topic, Goal: sd.Goal}
				report, err = h.intv.GenerateReportFromConversation(ctx, req, conversation)
				if err == nil {
					sd.Report = report
					_ = h.sessionMgr.SaveSessionData(sd)
					c.JSON(http.StatusOK, utils.H{"report": report})
					return
				}
			}
		}
		c.JSON(http.StatusInternalServerError, utils.H{"error": "获取报告失败: " + err.Error()})
		return
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

	curriculum, err := h.sessionMgr.GenerateCurriculum(ctx, req.SessionID)
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
	lesson, err := h.tchr.GenerateLesson(ctx, report, &chapter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{"error": "生成教案失败: " + err.Error()})
		return
	}

	// 缓存教案
	h.sessionMgr.SaveLessonPlan(req.SessionID, req.ChapterIndex, lesson)

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

// StreamFirstQuestion SSE 流式返回诊断的第一道题
func (h *Handler) StreamFirstQuestion(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Param("session_id")

	sd, err := h.sessionMgr.GetSessionStoreData(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "会话不存在"})
		return
	}
	// 如果已经有题目了，直接 JSON 返回（不走 SSE）
	if len(sd.Questions) > 0 {
		q := sd.Questions[0]
		c.JSON(http.StatusOK, utils.H{"question": q})
		return
	}

	// 通知中间件不要缓冲这个响应
	c.Response.Header.Set("Content-Type", "text/event-stream")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")
	c.Response.Header.Set("X-Accel-Buffering", "no") // nginx 禁用缓冲

	// 先刷新一次 Header
	c.Flush()

	// 流式生成第一道题
	resp, err := h.intv.StreamFirstQuestion(ctx, sd.Topic, sd.Goal, func(token string) error {
		_, err := c.Write([]byte("data: " + token + "\n\n"))
		if err != nil {
			return err
		}
		c.Flush()
		return nil
	})
	if err != nil {
		c.Write([]byte("event: error\ndata: " + err.Error() + "\n\n"))
		c.Flush()
		return
	}

	if resp.Action != "ask" || resp.Question == nil {
		c.Write([]byte("event: error\ndata: unexpected response\n\n"))
		c.Flush()
		return
	}

	// 保存题目
	q := agent.Question{
		ID:         1,
		Content:    resp.Question.Content,
		Category:   resp.Question.Category,
		Difficulty: resp.Question.Difficulty,
	}
	if err := h.sessionMgr.SetQuestions(sessionID, []agent.Question{q}); err != nil {
		c.Write([]byte("event: error\ndata: save failed\n\n"))
		c.Flush()
		return
	}

	// 发送完成事件（包含完整题目 JSON）
	doneJSON, _ := json.Marshal(map[string]interface{}{
		"action":   "done",
		"question": q,
	})
	c.Write([]byte("data: " + string(doneJSON) + "\n\n"))
	c.Flush()
}

// ====== 旧接口保留的请求类型 ======

type GenerateCurriculumReq struct {
	SessionID string `json:"session_id"`
}

type GenerateLessonReq struct {
	SessionID    string `json:"session_id"`
	ChapterIndex int    `json:"chapter_index"`
}

// ====== 用户课程 ======

type UserCourseResp struct {
	SessionID     string  `json:"session_id"`
	Topic         string  `json:"topic"`
	Goal          string  `json:"goal"`
	Title         string  `json:"title"`
	TotalWeeks    int     `json:"total_weeks"`
	ChapterCount  int     `json:"chapter_count"`
	CreatedAt     string  `json:"created_at"`
	HasReport     bool    `json:"has_report"`
	OverallScore  float64 `json:"overall_score,omitempty"`
}

// ListUserCourses 列出当前用户的所有课程
func (h *Handler) ListUserCourses(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	courses := h.sessionMgr.ListUserCourses(uid)
	resp := make([]UserCourseResp, 0, len(courses))
	for _, sd := range courses {
		cr := UserCourseResp{
			SessionID:    sd.ID,
			Topic:        sd.Topic,
			Goal:         sd.Goal,
			Title:        sd.Curriculum.Title,
			TotalWeeks:   sd.Curriculum.TotalWeeks,
			ChapterCount: len(sd.Curriculum.Chapters),
			CreatedAt:    sd.CreatedAt,
			HasReport:    sd.Report != nil,
		}
		if sd.Report != nil {
			cr.OverallScore = sd.Report.OverallScore
		}
		resp = append(resp, cr)
	}

	c.JSON(http.StatusOK, utils.H{"courses": resp})
}

// GetUserCourseDetail 获取用户某门课程的详细信息（含课时）
func (h *Handler) GetUserCourseDetail(ctx context.Context, c *app.RequestContext) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	sessionID := c.Param("session_id")
	sd, err := h.sessionMgr.GetSessionStoreData(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "课程不存在"})
		return
	}
	if sd.UserID != uid {
		c.JSON(http.StatusForbidden, utils.H{"error": "无权访问该课程"})
		return
	}
	if sd.Curriculum == nil {
		c.JSON(http.StatusNotFound, utils.H{"error": "课程方案未生成"})
		return
	}

	c.JSON(http.StatusOK, utils.H{
		"session_id":   sd.ID,
		"topic":        sd.Topic,
		"goal":         sd.Goal,
		"created_at":   sd.CreatedAt,
		"report":       sd.Report,
		"curriculum":   sd.Curriculum,
		"lesson_plans": sd.LessonPlans,
	})
}
