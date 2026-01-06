package service

import (
	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/repository"
)

type URLService interface {
	ShortenURL(url string) (string, error)
	ShortenURLAPI(request model.ShortenURLRequest) (model.ShortenURLResponse, error)
	GetOriginalURL(id string) (string, bool)
}

type defaultURLService struct {
	storage repository.Storage
	config  *config.Config
}

func NewURLService(storage repository.Storage, config *config.Config) *defaultURLService {
	return &defaultURLService{storage: storage, config: config}
}

func (s *defaultURLService) ShortenURL(url string) (string, error) {
	id, err := s.storage.SaveURL(url)
	return s.config.BaseURL + "/" + id, err
}

func (s *defaultURLService) ShortenURLAPI(request model.ShortenURLRequest) (model.ShortenURLResponse, error) {
	id, err := s.storage.SaveURL(request.URL)
	return model.ShortenURLResponse{ShortenedURL: s.config.BaseURL + "/" + id}, err
}

func (s *defaultURLService) GetOriginalURL(id string) (string, bool) {
	url, found := s.storage.GetURL(id)
	return url, found
}
