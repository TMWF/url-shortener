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
	return s.config.BaseURL + "/" + id, err
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
	response, err := s.storage.SaveBatchURL(ctx, request)
	if err != nil {
		logger.GetLogger().Error("Error occured while getting shortened URL ",
			zap.String("original error message", err.Error()),
		)
		return nil, err
	}

	for idx, value := range response {
		value.ShortURL = s.config.BaseURL + "/" + value.ShortURL
		response[idx] = value
	}

	return response, err
}

func (s *defaultURLService) GetOriginalURL(ctx context.Context, id string) (string, bool) {
	url, found := s.storage.GetURL(ctx, id)
	return url, found
}
