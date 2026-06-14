package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	json "github.com/goccy/go-json"

	"github.com/TMWF/url-shortener/internal/audit"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/TMWF/url-shortener/internal/service"
	"github.com/TMWF/url-shortener/internal/util"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type urlHandler struct {
	jwtHelper           util.UserJWTBuilder
	urlService          service.URLService
	auditEventObservers map[string]audit.RequestEventObserver
}

func (h *urlHandler) RegisterObserver(observer audit.RequestEventObserver) {
	if h.auditEventObservers == nil {
		h.auditEventObservers = make(map[string]audit.RequestEventObserver)
	}

	h.auditEventObservers[observer.GetID()] = observer
}

func (h *urlHandler) DeregisterObserver(observerID string) {
	delete(h.auditEventObservers, observerID)
}

func (h *urlHandler) Notify(event *model.AuditEvent) {
	for _, observer := range h.auditEventObservers {
		go func() {
			err := observer.SaveEvent(event)
			if err != nil {
				logger.GetLogger().Error(
					"Error occured while handling audit event",
					zap.Error(err),
				)
			}
		}()
	}
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

	if err = h.setUserJWTCookieIfNeeded(context, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if isConflictError {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	fmt.Fprint(w, shortenedURL)

	if len(h.auditEventObservers) == 0 {
		return
	}

	userID, err := requireUserIDFromContext(context)
	if err != nil {
		logger.GetLogger().Error("Error occured while trying to get user id from context")
		return
	}

	auditEvent := &model.AuditEvent{
		UnixTimeStamp: time.Now().Unix(),
		Action:        model.Shorten,
		UserID:        strconv.Itoa(userID),
		URL:           bodyString,
	}

	go h.Notify(auditEvent)
}

func (h *urlHandler) ShortenURLAPI(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Incorrect HTTP method, only POST methods allowed", http.StatusMethodNotAllowed)
		return
	}

	limitReader := io.LimitReader(req.Body, 4096)
	bodyBytes, err := io.ReadAll(limitReader)
	if err != nil {
		http.Error(w, "Ошибка чтения запроса", http.StatusBadRequest)
		return
	}

	var reqBody model.ShortenURLRequest
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
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
		logger.GetLogger().Error("Error occured while getting shortened url", zap.Error(err))
		http.Error(w, "Error occured while getting shortened url", http.StatusInternalServerError)
		return
	}

	// if err != nil {
	// 	logger.GetLogger().Error("Error occured while getting shortened url", zap.Error(err))
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	w.Header().Set("Content-Type", "application/json")
	// w.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))

	if err = h.setUserJWTCookieIfNeeded(context, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if isConflictError {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	json.NewEncoder(w).Encode(response)

	if len(h.auditEventObservers) == 0 {
		return
	}

	userID, err := requireUserIDFromContext(context)
	if err != nil {
		logger.GetLogger().Error("Error occured while trying to get user id from context")
		return
	}

	auditEvent := &model.AuditEvent{
		UnixTimeStamp: time.Now().Unix(),
		Action:        model.Shorten,
		UserID:        strconv.Itoa(userID),
		URL:           reqBody.URL,
	}

	go h.Notify(auditEvent)
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

	originalURL, err := h.urlService.GetOriginalURL(context, shortID)

	if errors.Is(err, repository.ErrURLNotFound) {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	if errors.Is(err, repository.ErrURLDeleted) {
		http.Error(w, err.Error(), http.StatusGone)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)

	if len(h.auditEventObservers) == 0 {
		return
	}

	userID, err := requireUserIDFromContext(context)
	if err != nil {
		logger.GetLogger().Error("Error occured while trying to get user id from context")
		return
	}

	auditEvent := &model.AuditEvent{
		UnixTimeStamp: time.Now().Unix(),
		Action:        model.Follow,
		UserID:        strconv.Itoa(userID),
		URL:           originalURL,
	}

	go h.Notify(auditEvent)
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

	if err = h.setUserJWTCookieIfNeeded(context, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(responseBody)
}

func (h *urlHandler) GetUserURLs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Incorrect HTTP method, only GET methods allowed", http.StatusMethodNotAllowed)
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

	if err = h.setUserJWTCookieIfNeeded(context, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}

func (h *urlHandler) DeleteUserURLs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		http.Error(w, "Incorrect HTTP method, only DELETE methods allowed", http.StatusMethodNotAllowed)
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

	var reqBody = make([]string, 0)
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&reqBody); err != nil {
		logger.GetLogger().Error("Error occured while decoding request body")
		http.Error(w, "Error occured while decoding request body", http.StatusBadRequest)
		return
	}

	h.urlService.ScheduleUserURLsJob(context, reqBody)

	if err = h.setUserJWTCookieIfNeeded(context, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *urlHandler) getContextWithUserIDIfNeeded(ctx context.Context) (context.Context, error) {
	_, ok := ctx.Value(util.UserID).(int)
	if !ok {
		userID, err := h.urlService.SaveUser(ctx)
		if err != nil {
			return nil, err
		}
		contextWithNeedToSetCookieProperty := context.WithValue(ctx, util.NeedToSetUserJWTCookie, true)
		return context.WithValue(contextWithNeedToSetCookieProperty, util.UserID, userID), nil
	}
	return ctx, nil
}

func requireUserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(util.UserID).(int)
	if !ok {
		logger.GetLogger().Error("UserID unexpectedly not found in context")
		return -1, repository.ErrUserIDAbsent
	}

	return userID, nil
}

func (h *urlHandler) setUserJWTCookieIfNeeded(ctx context.Context, w http.ResponseWriter) error {
	if _, ok := ctx.Value(util.NeedToSetUserJWTCookie).(bool); !ok {
		logger.GetLogger().Debug("No need to set JWT Cookie")
		return nil
	}

	userID, err := requireUserIDFromContext(ctx)
	if err != nil {
		return errors.New("unexpectedly not found user ID in context")
	}

	jwtToken, err := h.jwtHelper.BuildJWTString(userID)
	if err != nil {
		logger.GetLogger().Error("error occured while getting jwtToken",
			zap.String("original error message", err.Error()),
		)
		return errors.New("error occured while getting jwtToken")
	}

	cookie := http.Cookie{Name: string(util.UserID), Value: jwtToken}
	http.SetCookie(w, &cookie)
	logger.GetLogger().Debug("Successfully set jwt cookie")

	return nil
}
