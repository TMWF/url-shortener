package handler

import (
	"context"

	"github.com/TMWF/url-shortener/internal/model"
)

type RequestEventObserver interface {
	SaveEvent(context.Context, *model.AuditEvent) error
	GetID() string
}
