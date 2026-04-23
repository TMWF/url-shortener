package audit

import "github.com/TMWF/url-shortener/internal/model"

type RequestEventObserver interface {
	SaveEvent(*model.AuditEvent) error
	GetID() string
}
