// Package localmodel provides runtime discovery for locally hosted model
// runners. Phase 1 focuses on Ollama health, model listing, and metadata.
package localmodel

import "time"

type Status string

const (
	StatusUnavailable Status = "unavailable"
	StatusOffline     Status = "offline"
	StatusOnline      Status = "online"
	StatusError       Status = "error"
)

type RuntimeStatus struct {
	Installed bool
	Running   bool
	Status    Status
	Version   string
	Error     error
}

type Model struct {
	Runtime       string
	Name          string
	DisplayName   string
	Size          int64
	ModifiedAt    time.Time
	Family        string
	ParameterSize string
	Quantization  string
	ContextWindow int64
	MaxTokens     int64
	Capabilities  []string
	CodingCapable bool
}
