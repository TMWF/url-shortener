package service

import (
	"context"
	"errors"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/repository"
	"go.uber.org/zap"
)

type URLService interface {
	ShortenURL(ctx context.Context, url string) (string, error)
	ShortenURLAPI(ctx context.Context, request *model.ShortenURLRequest) (*model.ShortenURLResponse, error)
	ShortenURLBatch(ctx context.Context, request []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error)
	GetOriginalURL(ctx context.Context, id string) (string, bool)
}

type defaultURLService struct {
	storage repository.Storage
	config  *config.Config
}

func NewURLService(storage repository.Storage, config *config.Config) *defaultURLService {
	return &defaultURLService{storage: storage, config: config}
}

func (s *defaultURLService) ShortenURL(ctx context.Context, url string) (string, error) {
	id, err := s.storage.SaveURL(ctx, url)
	if err != nil && !errors.Is(err, repository.ErrConflict) {
		logger.GetLogger().Error("Error occurred while saving URL",
			zap.String("original error message", err.Error()),
			zap.String("url", url),
		)
		return "", err
	}
	return s.config.BaseURL + "/" + id, nil
}

func (s *defaultURLService) ShortenURLAPI(ctx context.Context, request *model.ShortenURLRequest) (*model.ShortenURLResponse, error) {
	id, err := s.storage.SaveURL(ctx, request.URL)
	if err != nil && !errors.Is(err, repository.ErrConflict) {
		logger.GetLogger().Error("Error occured while getting shortened URL ",
			zap.String("original error message", err.Error()),
		)
		return nil, err
	}
	return &model.ShortenURLResponse{ShortenedURL: s.config.BaseURL + "/" + id}, err
}

func (s *defaultURLService) ShortenURLBatch(ctx context.Context, request []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error) {
	ids, err := s.storage.SaveBatchURL(ctx, request)
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

func (s *defaultURLService) GetOriginalURL(ctx context.Context, id string) (string, bool) {
	url, found := s.storage.GetURL(ctx, id)
	return url, found
}
