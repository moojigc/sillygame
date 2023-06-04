package subs

import (
	"github.com/google/uuid"
)

type Subscriber struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Msgs      chan []byte `json:"-"`
	CloseSlow func()      `json:"-"`
}

func New(name string, bufSize int, closeSlow func()) *Subscriber {
	return &Subscriber{
		ID:        uuid.New().String(),
		Msgs:      make(chan []byte, bufSize),
		Name:      name,
		CloseSlow: closeSlow,
	}
}
