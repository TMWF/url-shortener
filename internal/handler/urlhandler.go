package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/TMWF/url-shortener/internal/service"
)

type URLHandler struct {
	urlService service.URLService
}

func NewURLHandler(service service.URLService) *URLHandler {
	return &URLHandler{urlService: service}
}

func (h *URLHandler) ShortenURL(w http.ResponseWriter, req *http.Request) {
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

	shortenedUrl, err := h.urlService.ShortenURL(bodyString)
	if err != nil {
		http.Error(w, "Error occured while etting shortened url", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(shortenedUrl)))
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, shortenedUrl)
}

func (h *URLHandler) GetOriginalURL(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Incorrect HTTP method, only GET methods allowed", http.StatusMethodNotAllowed)
		return
	}

	shortID := strings.TrimPrefix(req.URL.Path, "/")
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
