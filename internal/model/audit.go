package model

type AuditEvent struct {
	UnixTimeStamp int64  `json:"ts"`
	Action        Action `json:"action"`
	UserID        string `json:"user_id,omitempty"`
	URL           string `json:"url"`
}

type Action string

const (
	Shorten Action = "shorten"
	Follow  Action = "follow"
)
