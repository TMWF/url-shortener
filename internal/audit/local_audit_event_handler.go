package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/TMWF/url-shortener/internal/model"
)

type localAuditEventhandler struct {
	file *os.File
	mu   sync.Mutex
}

func NewLocalAuditEventHandler(auditFile *os.File) *localAuditEventhandler {
	return &localAuditEventhandler{file: auditFile}
}

func (h *localAuditEventhandler) GetID() string {
	return "LocalAuditEventhandler"
}

func (h *localAuditEventhandler) SaveEvent(event *model.AuditEvent) error {
	data, err := json.Marshal(*event)
	if err != nil {
		return fmt.Errorf("audit marshal error: %w", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, err := h.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("audit write error: %w", err)
	}

	return nil
}
