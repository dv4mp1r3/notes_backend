package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	notFoundErrorMessage = "session not found"
)

type RAMSessionStore struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

func NewRAMSessionStore() *RAMSessionStore {
	return &RAMSessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *RAMSessionStore) CreateSession() (*Session, error) {
	session := NewSession()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.sessions[session.ID] = session
	return session, nil
}

func (s *RAMSessionStore) GetSession(id string) (*Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	session, ok := s.sessions[id]
	if !ok {
		if !time.Now().Before(session.ExpiresAt) {
			s.DeleteSession(id)
		}
		return nil, fmt.Errorf(notFoundErrorMessage)
	}

	return session, nil
}

func (s *RAMSessionStore) DeleteSession(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.sessions, id)
	return nil
}
