package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/TMWF/url-shortener/internal/model"
)

type localAuditEventhandler struct {
	filePath string
	mu       sync.Mutex
}

func NewLocalAuditEventHandler(filePath string) *localAuditEventhandler {
	return &localAuditEventhandler{filePath: filePath}
}

func (h *localAuditEventhandler) GetID() string {
	return "LocalAuditEventhandler"
}

func (h *localAuditEventhandler) SaveEvent(event *model.AuditEvent) error {
	file, err := os.OpenFile(h.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open audit file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(*event)
	if err != nil {
		return fmt.Errorf("audit marshal error: %w", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("audit write error: %w", err)
	}

	return nil
}
