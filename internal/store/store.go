package store

import (
	"sync"
	"time"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]Entry
}

type Entry struct {
	Value  string
	Expiry time.Time
}

func New() *Store {
	return &Store{
		data: make(map[string]Entry),
	}
}

func (s *Store) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.data[key]
	if !ok {
		return false
	}
	entry.Expiry = time.Now().Add(ttl)
	s.data[key] = entry
	return true
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = Entry{Value: value}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.data[key]
	if !ok {
		return "", false
	}

	if !entry.Expiry.IsZero() && time.Now().After(entry.Expiry) {
		return "", false
	}
	return entry.Value, true
}

func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.data[key]
	delete(s.data, key)
	return ok
}

func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.data[key]
	if !ok {
		return false
	}
	if !entry.Expiry.IsZero() && time.Now().After(entry.Expiry) {
		return false
	}
	return true
}

func (s *Store) Ttl(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.data[key]
	if !ok {
		return -2
	}
	if entry.Expiry.IsZero() {
		return -1
	}

	if time.Now().After(entry.Expiry) {
		return -2
	}
	return int(time.Until(entry.Expiry).Seconds())
}
