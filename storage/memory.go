package storage

import (
	"sync"
)

type MemoryStorage struct {
	mu      sync.RWMutex
	history map[int64][]string
	limit   int
}

func NewMemoryStorage(limit int) *MemoryStorage {
	return &MemoryStorage{
		history: make(map[int64][]string),
		limit:   limit,
	}
}

func (s *MemoryStorage) GetHistory(chatID int64) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	history := s.history[chatID]
	copied := make([]string, len(history))
	copy(copied, history)
	return copied
}

func (s *MemoryStorage) AddMessage(chatID int64, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	history := s.history[chatID]
	history = append(history, msg)
	
	if len(history) > s.limit {
		history = history[len(history)-s.limit:]
	}
	
	s.history[chatID] = history
}
