package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/TMWF/url-shortener/internal/service"
	"github.com/TMWF/url-shortener/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type urlHandler struct {
	jwtHelper  util.UserJWTBuilder
	urlService service.URLService
}

func NewURLHandler(service service.URLService, jwtHelper util.UserJWTBuilder) *urlHandler {
	return &urlHandler{urlService: service, jwtHelper: jwtHelper}
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
	context, err := h.getContextWithUserIDIfNeeded(req.Context())
	if err != nil {
		logger.GetLogger().Error("error occured while trying to save user",
			zap.String("original error message", err.Error()),
		)
		http.Error(w, "Error occured while trying to save user", http.StatusInternalServerError)
		return
	}

	shortenedURL, err := h.urlService.ShortenURL(context, bodyString)
	var isConflictError = errors.Is(err, repository.ErrConflict)
	if err != nil && !isConflictError {
		logger.GetLogger().Error("Error occured while getting shortened url", zap.String("original error message", err.Error()))
		http.Error(w, "Error occured while getting shortened url", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(shortenedURL)))

	userID, err := h.requireUserIDFromContext(context)
	if err != nil {
		http.Error(w, "Unexpectedly not found user ID in context", http.StatusInternalServerError)
		return
	}

	jwtToken, err := h.jwtHelper.BuildJWTString(userID)
	if err != nil {
		logger.GetLogger().Error("error occured while getting jwtToken",
			zap.String("original error message", err.Error()),
		)
		http.Error(w, "Error occured while getting jwtToken", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{Name: util.USER_ID, Value: jwtToken}
	http.SetCookie(w, &cookie)

	if isConflictError {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	fmt.Fprint(w, shortenedURL)
}

func (h *urlHandler) ShortenURLAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Incorrect HTTP method, only POST methods allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody model.ShortenURLRequest
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&reqBody); err != nil {
		logger.GetLogger().Error("Error occured while decoding request body")
		http.Error(w, "Error occured while decoding request body", http.StatusBadRequest)
		return
	}

	context, err := h.getContextWithUserIDIfNeeded(req.Context())
	if err != nil {
		logger.GetLogger().Error("error occured while trying to save user",
			zap.String("original error message", err.Error()),
		)
		http.Error(w, "Error occured while trying to save user", http.StatusInternalServerError)
		return
	}

	response, err := h.urlService.ShortenURLAPI(context, &reqBody)
	var isConflictError = errors.Is(err, repository.ErrConflict)

	if err != nil && !isConflictError {
		logger.GetLogger().Error("Error occured while getting shortened url")
		http.Error(w, "Error occured while getting shortened url", http.StatusInternalServerError)
		return
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))

	userID, err := h.requireUserIDFromContext(context)
	if err != nil {
		http.Error(w, "Unexpectedly not found user ID in context", http.StatusInternalServerError)
		return
	}

	jwtToken, err := h.jwtHelper.BuildJWTString(userID)
	if err != nil {
		logger.GetLogger().Error("error occured while getting jwtToken",
			zap.String("original error message", err.Error()),
		)
		http.Error(w, "Error occured while getting jwtToken", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{Name: util.USER_ID, Value: jwtToken}
	http.SetCookie(w, &cookie)

	if isConflictError {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	w.Write(responseBody)
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

	context, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	originalURL, found := h.urlService.GetOriginalURL(context, shortID)
	if !found {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *urlHandler) ShortenURLBatch(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Incorrect HTTP method, only POST methods allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody = make([]model.URLBatchRequestDto, 0)
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&reqBody); err != nil {
		logger.GetLogger().Error("Error occured while decoding request body")
		http.Error(w, "Error occured while decoding request body", http.StatusBadRequest)
		return
	}

	context, err := h.getContextWithUserIDIfNeeded(req.Context())
	if err != nil {
		logger.GetLogger().Error("error occured while trying to save user",
			zap.String("original error message", err.Error()),
		)
		http.Error(w, "Error occured while trying to save user", http.StatusInternalServerError)
		return
	}

	response, err := h.urlService.ShortenURLBatch(context, reqBody)
	if err != nil {
		logger.GetLogger().Error("Error occured while getting shortened url")
		http.Error(w, "Error occured while getting shortened url", http.StatusInternalServerError)
		return
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))

	userID, err := h.requireUserIDFromContext(context)
	if err != nil {
		http.Error(w, "Unexpectedly not found user ID in context", http.StatusInternalServerError)
		return
	}

	jwtToken, err := h.jwtHelper.BuildJWTString(userID)
	if err != nil {
		logger.GetLogger().Error("error occured while getting jwtToken",
			zap.String("original error message", err.Error()),
		)
		http.Error(w, "Error occured while getting jwtToken", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{Name: util.USER_ID, Value: jwtToken}
	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusCreated)
	w.Write(responseBody)
}

func (h *urlHandler) GetUserURLs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Incorrect HTTP method, only GET methods allowed", http.StatusMethodNotAllowed)
		return
	}

	context, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	response, err := h.urlService.GetUserURLs(context)

	if errors.Is(err, repository.ErrUserIDAbsent) {
		logger.GetLogger().Error("User unathorized")
		http.Error(w, "User not authorized", http.StatusUnauthorized)
		return
	}

	if err != nil {
		logger.GetLogger().Error("Unexpected error occured while fetching user urls",
			zap.String("original error message", err.Error()),
		)
		http.Error(w, "Unexpected error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(response) == 0 {
		logger.GetLogger().Warn("No urls found for user")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		logger.GetLogger().Error("Error occured while marshaling json")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func (h *urlHandler) getContextWithUserIDIfNeeded(ctx context.Context) (context.Context, error) {
	_, ok := ctx.Value(util.USER_ID).(int)
	if !ok {
		logger.GetLogger().Info("UserID not found in context")
		userID, err := h.urlService.SaveUser(ctx)
		if err != nil {
			return nil, err
		}
		return context.WithValue(ctx, util.USER_ID, userID), nil
	}
	return ctx, nil
}

func (h *urlHandler) requireUserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(util.USER_ID).(int)
	if !ok {
		logger.GetLogger().Error("UserID unexpectedly not found in context")
		return -1, repository.ErrUserIDAbsent
	}

	return userID, nil
}
