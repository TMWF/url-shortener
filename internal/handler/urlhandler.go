package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/TMWF/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

type urlHandler struct {
	urlService service.URLService
}

func NewURLHandler(service service.URLService) *urlHandler {
	return &urlHandler{urlService: service}
}

func (h *urlHandler) ShortenURL(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Incorrect HTTP method, only POST methods allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	bodyString := string(bodyBytes)

	shortenedURL, err := h.urlService.ShortenURL(bodyString)
	if err != nil {
		http.Error(w, "Error occured while etting shortened url", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(shortenedURL)))
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortenedURL)
}

func (h *urlHandler) GetOriginalURL(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Incorrect HTTP method, only GET methods allowed", http.StatusMethodNotAllowed)
		return
	}

	shortID := chi.URLParam(req, "id")
	if shortID == "" {
		http.Error(w, "Short URL ID is missing", http.StatusBadRequest)
		return
	}

	originalURL, found := h.urlService.GetOriginalURL(shortID)
	if !found {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
