package audit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/TMWF/url-shortener/internal/model"
)

var ExternalServerError = errors.New("external server error")

type RemoteAuditEventHandler struct {
	auditServerURL string
	client         *http.Client
}

func NewRemoteAuditEventHandler(auditServerURL string) *RemoteAuditEventHandler {
	return &RemoteAuditEventHandler{
		auditServerURL: auditServerURL,
		client:         &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *RemoteAuditEventHandler) GetID() string {
	return "RemoteAuditEventHandler"
}

func (h *RemoteAuditEventHandler) SaveEvent(event *model.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("audit marshal error: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, h.auditServerURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send audit to remote server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return ExternalServerError
	}

	return nil
}
