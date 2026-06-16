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

// ShortenURL обрабатывает HTTP-запрос на создание короткой ссылки.
//
// Метод принимает только POST-запросы. Тело запроса должно содержать исходный
// URL в текстовом виде. При необходимости метод получает или создаёт идентификатор
// пользователя в контексте запроса, после чего передаёт URL в сервис сокращения.
//
// В случае успешного создания новой короткой ссылки возвращает:
//   - HTTP 201 Created;
//   - Content-Type: text/plain;
//   - тело ответа с сокращённым URL.
//
// Если переданный URL уже был сохранён ранее, возвращает:
//   - HTTP 409 Conflict;
//   - Content-Type: text/plain;
//   - тело ответа с ранее созданным сокращённым URL.
//
// Возможные ошибки:
//   - HTTP 405 Method Not Allowed, если метод запроса не POST;
//   - HTTP 400 Bad Request, если не удалось прочитать тело запроса;
//   - HTTP 500 Internal Server Error, если произошла ошибка при сохранении
//     пользователя, создании короткой ссылки или установке JWT-cookie.
//
// Если в обработчике зарегистрированы наблюдатели аудита, после успешной
// обработки запроса асинхронно отправляется событие аудита с информацией
// о пользователе, действии и исходном URL.
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

// ShortenURLAPI обрабатывает HTTP API-запрос на создание короткой ссылки.
//
// Метод принимает только POST-запросы с телом в формате JSON. Размер тела запроса
// ограничивается 4096 байтами. Ожидается, что тело запроса соответствует структуре
// model.ShortenURLRequest и содержит исходный URL для сокращения.
//
// При необходимости ShortenURLAPI получает или создаёт идентификатор пользователя
// в контексте запроса, после чего передаёт данные в сервис сокращения URL.
//
// В случае успешного создания новой короткой ссылки метод возвращает:
//   - HTTP 201 Created;
//   - Content-Type: application/json;
//   - JSON-ответ со сведениями о сокращённой ссылке.
//
// Если переданный URL уже был сохранён ранее, метод возвращает:
//   - HTTP 409 Conflict;
//   - Content-Type: application/json;
//   - JSON-ответ с ранее созданной сокращённой ссылкой.
//
// Возможные ошибки:
//   - HTTP 405 Method Not Allowed, если метод запроса не POST;
//   - HTTP 400 Bad Request, если не удалось прочитать тело запроса;
//   - HTTP 400 Bad Request, если тело запроса содержит некорректный JSON;
//   - HTTP 500 Internal Server Error, если произошла ошибка при получении или
//     сохранении пользователя;
//   - HTTP 500 Internal Server Error, если произошла ошибка при создании
//     сокращённой ссылки;
//   - HTTP 500 Internal Server Error, если не удалось установить JWT-cookie
//     пользователя.
//
// Если в обработчике зарегистрированы наблюдатели аудита, после успешной
// обработки запроса асинхронно отправляется событие аудита с информацией
// о пользователе, действии Shorten и исходном URL.
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

	w.Header().Set("Content-Type", "application/json")

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

// GetOriginalURL обрабатывает HTTP-запрос на получение исходного URL
// по идентификатору короткой ссылки.
//
// Метод принимает только GET-запросы. Идентификатор короткой ссылки извлекается
// из URL-параметра "id". Если параметр отсутствует, метод возвращает ошибку.
//
// Для получения исходного URL создаётся контекст с таймаутом 5 секунд, после чего
// запрос передаётся в сервис URL.
//
// В случае успешного получения исходного URL метод возвращает:
//   - HTTP 307 Temporary Redirect;
//   - заголовок Location со значением исходного URL.
//
// Возможные ошибки:
//   - HTTP 405 Method Not Allowed, если метод запроса не GET;
//   - HTTP 400 Bad Request, если URL-параметр "id" отсутствует;
//   - HTTP 404 Not Found, если короткая ссылка не найдена;
//   - HTTP 410 Gone, если короткая ссылка была удалена;
//   - HTTP 500 Internal Server Error, если произошла ошибка при получении
//     исходного URL.
//
// Если в обработчике зарегистрированы наблюдатели аудита, после успешного
// получения исходного URL асинхронно отправляется событие аудита с информацией
// о пользователе, действии Follow и исходном URL.
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

// ShortenURLBatch обрабатывает HTTP API-запрос на пакетное создание коротких ссылок.
//
// Метод принимает только POST-запросы с телом в формате JSON. Ожидается, что тело
// запроса содержит массив объектов model.URLBatchRequestDto с исходными URL и
// корреляционными идентификаторами.
//
// При необходимости ShortenURLBatch получает или создаёт идентификатор пользователя
// в контексте запроса, после чего передаёт список URL в сервис пакетного сокращения.
//
// В случае успешного создания коротких ссылок метод возвращает:
//   - HTTP 201 Created;
//   - Content-Type: application/json;
//   - JSON-массив с результатами сокращения URL.
//
// Возможные ошибки:
//   - HTTP 405 Method Not Allowed, если метод запроса не POST;
//   - HTTP 400 Bad Request, если тело запроса не удалось декодировать;
//   - HTTP 500 Internal Server Error, если произошла ошибка при получении или
//     сохранении пользователя;
//   - HTTP 500 Internal Server Error, если произошла ошибка при пакетном создании
//     коротких ссылок;
//   - HTTP 500 Internal Server Error, если не удалось установить JWT-cookie
//     пользователя;
//   - HTTP 500 Internal Server Error, если не удалось записать тело ответа.
//
// Метод устанавливает JWT-cookie пользователя при необходимости.
func (h *urlHandler) ShortenURLBatch(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Incorrect HTTP method, only POST methods allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody = make([]model.URLBatchRequestDto, 0)
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&reqBody); err != nil {
		logger.GetLogger().Error("Error occured while decoding request body: " + err.Error())
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

	w.Header().Set("Content-Type", "application/json")

	if err = h.setUserJWTCookieIfNeeded(context, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.GetLogger().Error(err.Error())
		http.Error(w, "Error occured while writing response body", http.StatusInternalServerError)
		return
	}
}

// GetUserURLs обрабатывает HTTP-запрос на получение списка URL,
// сокращённых текущим пользователем.
//
// Метод принимает только GET-запросы. При необходимости GetUserURLs получает
// или создаёт идентификатор пользователя в контексте запроса, после чего
// запрашивает у сервиса список URL, связанных с этим пользователем.
//
// В случае успешного получения списка URL метод возвращает:
//   - HTTP 200 OK;
//   - Content-Type: application/json;
//   - JSON-массив URL пользователя.
//
// Если для пользователя не найдено ни одной ссылки, метод возвращает:
//   - HTTP 204 No Content.
//
// Возможные ошибки:
//   - HTTP 405 Method Not Allowed, если метод запроса не GET;
//   - HTTP 401 Unauthorized, если идентификатор пользователя отсутствует
//     или пользователь не авторизован;
//   - HTTP 500 Internal Server Error, если произошла ошибка при получении
//     или сохранении пользователя;
//   - HTTP 500 Internal Server Error, если произошла непредвиденная ошибка
//     при получении списка URL пользователя;
//   - HTTP 500 Internal Server Error, если не удалось установить JWT-cookie
//     пользователя;
//   - HTTP 500 Internal Server Error, если не удалось записать тело ответа.
//
// Метод устанавливает JWT-cookie пользователя при необходимости.
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.GetLogger().Error(err.Error())
		http.Error(w, "Error occured while writing response body", http.StatusInternalServerError)
		return
	}
}

// DeleteUserURLs обрабатывает HTTP-запрос на асинхронное удаление URL пользователя.
//
// Метод принимает только DELETE-запросы с телом в формате JSON. Ожидается, что
// тело запроса содержит массив строк с идентификаторами коротких URL, которые
// необходимо удалить.
//
// При необходимости DeleteUserURLs получает или создаёт идентификатор пользователя
// в контексте запроса, после чего передаёт список идентификаторов в сервис для
// постановки задачи удаления в очередь.
//
// Удаление выполняется асинхронно, поэтому при успешной постановке задачи метод
// возвращает:
//   - HTTP 202 Accepted.
//
// Возможные ошибки:
//   - HTTP 405 Method Not Allowed, если метод запроса не DELETE;
//   - HTTP 500 Internal Server Error, если произошла ошибка при получении или
//     сохранении пользователя;
//   - HTTP 400 Bad Request, если тело запроса не удалось декодировать;
//   - HTTP 500 Internal Server Error, если не удалось установить JWT-cookie
//     пользователя.
//
// Метод устанавливает JWT-cookie пользователя при необходимости.
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
