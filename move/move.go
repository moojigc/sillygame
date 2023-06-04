package move

import (
	"time"

	"github.com/google/uuid"
)

type Move struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Event     string    `json:"event"`
	Timestamp int64     `json:"ts"`
	Coords    []float32 `json:"coords"`
	MouseDown bool      `json:"mouseDown"`
}

func (m *Move) SetUserID(id string) *Move {
	if m.UserID != "" {
		return m
	}
	m.UserID = id

	return m
}

func New(event string) *Move {
	return &Move{
		ID:        uuid.New().String(),
		Event:     event,
		Timestamp: time.Now().UnixMilli(),
	}
}
