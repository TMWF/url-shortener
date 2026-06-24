// Package service contains application configs
package service

import (
	"context"
	"errors"
	"time"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/TMWF/url-shortener/internal/util"
	"go.uber.org/zap"
)

// URLService описывает бизнес-логику работы с URL и пользователями.
//
// Интерфейс инкапсулирует операции сокращения ссылок, получения исходных URL,
// получения ссылок текущего пользователя, постановки задач на удаление URL,
// а также сохранения пользователя.
//
// Реализации URLService должны использовать переданный context.Context для
// контроля времени выполнения, отмены операций и передачи пользовательских
// данных между слоями приложения.
type URLService interface {
	// ShortenURL создаёт короткую ссылку для переданного URL.
	//
	// Возвращает сокращённый URL или ошибку, если создать ссылку не удалось.
	ShortenURL(ctx context.Context, url string) (string, error)

	// ShortenURLAPI создаёт короткую ссылку на основе API-запроса.
	//
	// Принимает модель запроса model.ShortenURLRequest и возвращает модель
	// ответа model.ShortenURLResponse с сокращённым URL.
	ShortenURLAPI(ctx context.Context, request *model.ShortenURLRequest) (*model.ShortenURLResponse, error)

	// ShortenURLBatch выполняет пакетное сокращение URL.
	//
	// Принимает список URL с корреляционными идентификаторами и возвращает
	// список результатов сокращения с теми же корреляционными идентификаторами.
	ShortenURLBatch(ctx context.Context, request []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error)

	// GetOriginalURL возвращает исходный URL по идентификатору короткой ссылки.
	//
	// Возвращает исходный URL или ошибку, если ссылка не найдена, удалена
	// или произошла внутренняя ошибка.
	GetOriginalURL(ctx context.Context, id string) (string, error)

	// GetUserURLs возвращает список URL, созданных текущим пользователем.
	//
	// Идентификатор пользователя ожидается в переданном контексте.
	GetUserURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error)

	// ScheduleUserURLsJob ставит задачу на асинхронное удаление URL пользователя.
	//
	// Принимает список идентификаторов коротких URL. Идентификатор пользователя
	// ожидается в переданном контексте.
	ScheduleUserURLsJob(ctx context.Context, urlIDs []string)

	// SaveUser сохраняет пользователя и возвращает его идентификатор.
	//
	// Используется, когда для текущего запроса необходимо создать или
	// зарегистрировать пользователя.
	SaveUser(ctx context.Context) (int, error)
}

type defaultURLService struct {
	storage            repository.Storage
	config             *config.Config
	userURLsDeleteJobs chan model.DeleteUserURLsJobModel
}

func NewURLService(storage repository.Storage, config *config.Config) *defaultURLService {
	service := &defaultURLService{
		storage:            storage,
		config:             config,
		userURLsDeleteJobs: make(chan model.DeleteUserURLsJobModel, 1024),
	}
	go service.urlDeletionWorker()
	return service
}

func (s *defaultURLService) ShortenURL(ctx context.Context, url string) (string, error) {
	context, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	id, err := s.storage.SaveURL(context, url)
	if err != nil && !errors.Is(err, repository.ErrConflict) {
		logger.GetLogger().Error("Error occurred while saving URL",
			zap.String("original error message", err.Error()),
			zap.String("url", url),
		)
		return "", err
	}
	return s.config.BaseURL + "/" + id, err
}

func (s *defaultURLService) ShortenURLAPI(ctx context.Context, request *model.ShortenURLRequest) (*model.ShortenURLResponse, error) {
	context, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	id, err := s.storage.SaveURL(context, request.URL)
	if err != nil && !errors.Is(err, repository.ErrConflict) {
		logger.GetLogger().Error("Error occured while getting shortened URL ",
			zap.String("original error message", err.Error()),
		)
		return nil, err
	}
	return &model.ShortenURLResponse{ShortenedURL: s.config.BaseURL + "/" + id}, err
}

func (s *defaultURLService) ShortenURLBatch(ctx context.Context, request []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error) {
	context, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ids, err := s.storage.SaveBatchURL(context, request)
	if err != nil {
		logger.GetLogger().Error("Error occured while getting shortened URL ",
			zap.String("original error message", err.Error()),
		)
		return nil, err
	}

	response := make([]model.URLBatchResponseDto, 0, len(request))
	for idx, id := range ids {
		shortURL := s.config.BaseURL + "/" + id
		responseModel := model.URLBatchResponseDto{CorrelationID: request[idx].CorrelationID, ShortURL: shortURL}
		response = append(response, responseModel)
	}

	return response, err
}

func (s *defaultURLService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return s.storage.GetURL(ctx, id)
}

func (s *defaultURLService) ScheduleUserURLsJob(ctx context.Context, urlIDs []string) {
	if len(urlIDs) == 0 {
		logger.GetLogger().Warn("Empty request body")
		return
	}

	userID, ok := ctx.Value(util.UserID).(int)

	if !ok || userID < 1 {
		logger.GetLogger().Error("UserID unexpectedly not found in context")

		return
	}

	jobModel := model.DeleteUserURLsJobModel{UserID: userID, URLIDs: urlIDs}

	s.userURLsDeleteJobs <- jobModel
}

func (s *defaultURLService) urlDeletionWorker() {
	dbStorage, ok := s.storage.(repository.DBStorage)

	if !ok {
		logger.GetLogger().Warn("Not db storage, not starting urlDeletionWorker")
		return
	}

	ticker := time.NewTicker(5 * time.Second)

	var jobs []model.DeleteUserURLsJobModel

	for {
		select {
		case job := <-s.userURLsDeleteJobs:
			jobs = append(jobs, job)
		case <-ticker.C:
			if len(jobs) == 0 {
				continue
			}
			err := dbStorage.DeleteUserURLs(jobs)
			if err != nil {
				logger.GetLogger().Debug("cannot delete messages", zap.Error(err))
				continue
			}
			jobs = nil
		}
	}
}

func (s *defaultURLService) GetUserURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error) {
	context, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := s.storage.GetUsersURLs(context)
	if err != nil {
		return nil, err
	}

	for idx := range response {
		response[idx].ShortURL = s.config.BaseURL + "/" + response[idx].ShortURL
	}
	return response, nil
}

func (s *defaultURLService) SaveUser(ctx context.Context) (int, error) {
	return s.storage.SaveUser(ctx)
}
