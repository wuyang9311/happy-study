package store

import (
	"log"
	"time"
)

// StartCleanup 启动后台会话清理 goroutine
// interval: 检查间隔
// maxAge:   会话最大存活时间
func (s *Store) StartCleanup(interval, maxAge time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// 首次运行前先清理一次
		s.cleanupOnce(maxAge)

		for range ticker.C {
			s.cleanupOnce(maxAge)
		}
	}()

	log.Printf("🧹 会话清理已启动（间隔 %v，保留 %v）", interval, maxAge)
}

func (s *Store) cleanupOnce(maxAge time.Duration) {
	now := time.Now()
	var expiredIDs []string

	s.mu.RLock()
	for id, sd := range s.sessions {
		created, err := time.Parse(time.RFC3339, sd.CreatedAt)
		if err != nil {
			continue
		}
		if now.Sub(created) > maxAge {
			expiredIDs = append(expiredIDs, id)
		}
	}
	s.mu.RUnlock()

	for _, id := range expiredIDs {
		if err := s.Delete(id); err != nil {
			log.Printf("清理过期会话失败 %s: %v", id, err)
		}
	}
}
