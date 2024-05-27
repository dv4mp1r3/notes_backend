package main

import (
	"fmt"
	"time"
)

type Session struct {
	ID        string
	Data      map[string]interface{}
	ExpiresAt time.Time
	UserId    int
}

type SessionStore interface {
	CreateSession() (*Session, error)
	GetSession(id string) (*Session, error)
	DeleteSession(id string) error
}

func NewSession(userId int) *Session {
	return &Session{
		ID:        generateSessionID(),
		Data:      make(map[string]interface{}),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserId:    userId,
	}
}

func generateSessionID() string {
	return "session-" + fmt.Sprintf("%d", time.Now().UnixNano())
}
