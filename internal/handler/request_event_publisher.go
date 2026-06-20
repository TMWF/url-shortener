package handler

import (
	"github.com/TMWF/url-shortener/internal/model"
)

type RequestEventPublisher interface {
	RegisterObserver(RequestEventObserver)
	DeregisterObserver(string)
	Notify(*model.AuditEvent)
}
