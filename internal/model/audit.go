package model

// AuditEvent описывает событие аудита, возникающее при действиях пользователя
// с URL.
//
// Событие содержит время возникновения, тип выполненного действия, идентификатор
// пользователя и URL, связанный с действием. Используется для передачи информации
// наблюдателям аудита или внешним системам логирования.
type AuditEvent struct {
	// UnixTimeStamp содержит время возникновения события в формате Unix timestamp.
	UnixTimeStamp int64 `json:"ts"`

	// Action содержит тип действия, выполненного пользователем.
	Action Action `json:"action"`

	// UserID содержит идентификатор пользователя, выполнившего действие.
	// Поле опускается при JSON-сериализации, если значение пустое.
	UserID string `json:"user_id,omitempty"`

	// URL содержит URL, связанный с событием аудита.
	URL string `json:"url"`
}

type Action string

const (
	Shorten Action = "shorten"
	Follow  Action = "follow"
)
