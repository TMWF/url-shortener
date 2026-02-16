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

type URLService interface {
	ShortenURL(ctx context.Context, url string) (string, error)
	ShortenURLAPI(ctx context.Context, request *model.ShortenURLRequest) (*model.ShortenURLResponse, error)
	ShortenURLBatch(ctx context.Context, request []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error)
	GetOriginalURL(ctx context.Context, id string) (string, error)
	GetUserURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error)
	ScheduleUserURLsJob(ctx context.Context, urlIDs []string)
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

	// context, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()

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
				logger.GetLogger().Debug("cannot save messages", zap.Error(err))
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
