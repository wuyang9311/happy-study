package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	agent "github.com/wuyang9311/happy-study/internal/agent"
)

// SessionData 持久化的会话数据（不含 Agent 引用）
type SessionData struct {
	ID           string                     `json:"id"`
	UserID       int64                      `json:"user_id"`
	Topic        string                     `json:"topic"`
	Goal         string                     `json:"goal"`
	Questions    []agent.Question           `json:"questions"`
	Answers      []agent.Answer             `json:"answers"`
	CurrentIndex int                        `json:"current_index"`
	Report       *agent.DiagnosisReport     `json:"report,omitempty"`
	Curriculum   *agent.Curriculum          `json:"curriculum,omitempty"`
	LessonPlans  map[int]*agent.LessonPlan  `json:"lesson_plans,omitempty"`
	CreatedAt    string                     `json:"created_at"`
}

// Store 持久化存储
type Store struct {
	mu       sync.RWMutex
	dataDir  string
	sessions map[string]*SessionData
}

// NewStore 创建或加载持久化存储
func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	s := &Store{
		dataDir:  dataDir,
		sessions: make(map[string]*SessionData),
	}

	// 从磁盘加载已有会话
	if err := s.loadAll(); err != nil {
		return nil, fmt.Errorf("load sessions: %w", err)
	}

	return s, nil
}

// filePath 获取某个会话的文件路径
func (s *Store) filePath(id string) string {
	return filepath.Join(s.dataDir, id+".json")
}

// loadAll 从磁盘加载所有会话
func (s *Store) loadAll() error {
	entries, err := os.ReadDir(s.dataDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read data dir: %w", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.dataDir, entry.Name()))
		if err != nil {
			continue // 跳过损坏的文件
		}

		var sd SessionData
		if err := json.Unmarshal(data, &sd); err != nil {
			continue
		}
		s.sessions[sd.ID] = &sd
	}

	return nil
}

// Save 保存会话
func (s *Store) Save(sd *SessionData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sd.ID] = sd

	data, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	if err := os.WriteFile(s.filePath(sd.ID), data, 0644); err != nil {
		return fmt.Errorf("write session file: %w", err)
	}

	return nil
}

// Get 获取会话
func (s *Store) Get(id string) (*SessionData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sd, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return sd, nil
}

// Delete 删除会话
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
	os.Remove(s.filePath(id)) // 忽略错误
	return nil
}

// ListAll 列出所有会话 ID
func (s *Store) ListAll() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.sessions))
	for id := range s.sessions {
		ids = append(ids, id)
	}
	return ids
}

// Count 会话数量
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// ListByUser 列出某个用户的所有会话
func (s *Store) ListByUser(userID int64) []*SessionData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*SessionData
	for _, sd := range s.sessions {
		if sd.UserID == userID {
			result = append(result, sd)
		}
	}
	return result
}
