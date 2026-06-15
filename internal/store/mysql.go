package store

import (
	"encoding/json"
	"fmt"
	"time"

	agent "github.com/wuyang9311/happy-study/internal/agent"
	"gorm.io/gorm"
)

// SessionStore 会话存储接口
type SessionStore interface {
	Save(sd *SessionData) error
	Get(id string) (*SessionData, error)
	Delete(id string) error
	ListByUser(userID int64) []*SessionData
	Count() int
}

// SessionModel GORM 模型（对应 MySQL sessions 表）
type SessionModel struct {
	ID           string    `gorm:"primaryKey;type:varchar(100)"`
	UserID       int64     `gorm:"index;not null;default:0"`
	Topic        string    `gorm:"type:text"`
	Goal         string    `gorm:"type:text"`
	Data         string    `gorm:"type:longtext"` // JSON: Questions, Answers, Report, Curriculum, LessonPlans
	CurrentIndex int       `gorm:"default:0"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (SessionModel) TableName() string {
	return "sessions"
}

// MySQLStore MySQL 版会话存储
type MySQLStore struct {
	db *gorm.DB
}

// NewMySQLStore 创建 MySQL 会话存储，自动建表
func NewMySQLStore(db *gorm.DB) (*MySQLStore, error) {
	if err := db.AutoMigrate(&SessionModel{}); err != nil {
		return nil, fmt.Errorf("迁移 sessions 表失败: %w", err)
	}
	return &MySQLStore{db: db}, nil
}

// modelToSession 将数据库模型转换为 SessionData
func modelToSession(m *SessionModel) (*SessionData, error) {
	sd := &SessionData{
		ID:           m.ID,
		UserID:       m.UserID,
		Topic:        m.Topic,
		Goal:         m.Goal,
		CurrentIndex: m.CurrentIndex,
		CreatedAt:    m.CreatedAt.Format(time.RFC3339),
	}

	// 解析 JSON data
	if m.Data != "" {
		type sessionJSON struct {
			Questions   []agent.Question            `json:"questions"`
			Answers     []agent.Answer              `json:"answers"`
			Report      *agent.DiagnosisReport      `json:"report,omitempty"`
			Curriculum  *agent.Curriculum           `json:"curriculum,omitempty"`
			LessonPlans map[int]*agent.LessonPlan   `json:"lesson_plans,omitempty"`
		}
		var sj sessionJSON
		if err := json.Unmarshal([]byte(m.Data), &sj); err != nil {
			return nil, fmt.Errorf("解析 session data 失败: %w", err)
		}
		sd.Questions = sj.Questions
		sd.Answers = sj.Answers
		sd.Report = sj.Report
		sd.Curriculum = sj.Curriculum
		sd.LessonPlans = sj.LessonPlans
	}

	return sd, nil
}

// sessionToModel 将 SessionData 转换为数据库模型
func sessionToModel(sd *SessionData) (*SessionModel, error) {
	m := &SessionModel{
		ID:           sd.ID,
		UserID:       sd.UserID,
		Topic:        sd.Topic,
		Goal:         sd.Goal,
		CurrentIndex: sd.CurrentIndex,
	}

	// 序列化复杂字段为 JSON
	type sessionJSON struct {
		Questions   []agent.Question            `json:"questions"`
		Answers     []agent.Answer              `json:"answers"`
		Report      *agent.DiagnosisReport      `json:"report,omitempty"`
		Curriculum  *agent.Curriculum           `json:"curriculum,omitempty"`
		LessonPlans map[int]*agent.LessonPlan   `json:"lesson_plans,omitempty"`
	}
	sj := sessionJSON{
		Questions:   sd.Questions,
		Answers:     sd.Answers,
		Report:      sd.Report,
		Curriculum:  sd.Curriculum,
		LessonPlans: sd.LessonPlans,
	}
	data, err := json.Marshal(sj)
	if err != nil {
		return nil, fmt.Errorf("序列化 session data 失败: %w", err)
	}
	m.Data = string(data)

	// 解析创建时间
	if sd.CreatedAt != "" {
		t, err := time.Parse(time.RFC3339, sd.CreatedAt)
		if err == nil {
			m.CreatedAt = t
		}
	}

	return m, nil
}

// Save 保存会话
func (ms *MySQLStore) Save(sd *SessionData) error {
	m, err := sessionToModel(sd)
	if err != nil {
		return err
	}
	return ms.db.Save(m).Error
}

// Get 获取会话
func (ms *MySQLStore) Get(id string) (*SessionData, error) {
	var m SessionModel
	err := ms.db.Where("id = ?", id).First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, err
	}
	return modelToSession(&m)
}

// Delete 删除会话
func (ms *MySQLStore) Delete(id string) error {
	return ms.db.Delete(&SessionModel{}, "id = ?", id).Error
}

// ListByUser 列出某个用户的所有会话
func (ms *MySQLStore) ListByUser(userID int64) []*SessionData {
	var models []SessionModel
	ms.db.Where("user_id = ?", userID).Order("created_at desc").Find(&models)

	result := make([]*SessionData, 0, len(models))
	for i := range models {
		sd, err := modelToSession(&models[i])
		if err != nil {
			continue
		}
		result = append(result, sd)
	}
	return result
}

// Count 会话数量
func (ms *MySQLStore) Count() int {
	var count int64
	ms.db.Model(&SessionModel{}).Count(&count)
	return int(count)
}
