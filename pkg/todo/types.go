package todo

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/google/uuid"
)

type TaskStatus int

const (
	Pending TaskStatus = iota
	InProgress
	Paused
	Completed
)

const (
	StatusPending    = "⏳ Pending"
	StatusInProgress = "▶️ In Progress"
	StatusPaused     = "⏸️ Paused"
	StatusCompleted  = "✅ Completed"
)

func (s TaskStatus) String() string {
	switch s {
	case Pending:
		return StatusPending
	case InProgress:
		return StatusInProgress
	case Paused:
		return StatusPaused
	case Completed:
		return StatusCompleted
	default:
		return "Unknown"
	}
}

type Task struct {
	ID            uuid.UUID     `json:"id"`
	Description   string        `json:"description"`
	Status        TaskStatus    `json:"status"`
	TimeSpent     time.Duration `json:"time_spent"`
	LastStartedAt time.Time     `json:"last_started_at"`
	CreatedAt     time.Time     `json:"created_at"`
}

type KeyMap struct {
	Add, Delete, Toggle, Complete, Up, Down, Quit, Enter, Esc, ScrollUp, ScrollDown, ToggleLineNumbers, ToggleCalendar key.Binding
}
