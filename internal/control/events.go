package control

import (
	"time"

	"github.com/jrafaelca/nodara/internal/storage"
)

type AgentEvent struct {
	Type       string                `json:"type"`
	OccurredAt time.Time             `json:"occurred_at"`
	Agent      *storage.AgentStatus  `json:"agent,omitempty"`
	Agents     []storage.AgentStatus `json:"agents,omitempty"`
	Message    string                `json:"message,omitempty"`
}
