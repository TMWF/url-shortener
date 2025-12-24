package service

import (
	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/repository"
)

type URLService interface {
	ShortenURL(url string) (string, error)
	GetOriginalURL(id string) (string, bool)
}

type defaultURLService struct {
	storage repository.Storage
}

func NewURLService(storage repository.Storage) *defaultURLService {
	return &defaultURLService{storage: storage}
}

func (s *defaultURLService) ShortenURL(url string) (string, error) {
	id, err := s.storage.SaveURL(url)
	return config.BaseURL + "/" + id, err
}

func (s *defaultURLService) GetOriginalURL(id string) (string, bool) {
	url, found := s.storage.GetURL(id)
	return url, found
}
