package handler

import (
	"net/http"

	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/service"
	"go.uber.org/zap"
)

type pingHandler struct {
	pingService service.PingDBService
}

func NewPingHandler(pingService service.PingDBService) *pingHandler {
	return &pingHandler{pingService: pingService}
}

func (ph *pingHandler) PingDB(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Incorrect HTTP method, only GET methods allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := ph.pingService.PingDB(); err != nil {
		logger.GetLogger().Error(
			"Could not ping DB",
			zap.String("original error message", err.Error()),
		)

		http.Error(w, "Error occured while trying to ping database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
