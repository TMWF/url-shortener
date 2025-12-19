package service

import (
	"github.com/TMWF/url-shortener/internal/repository"
)

const baseURL = "http://localhost:8080/"

type URLService struct {
	storage repository.Storage
}

func NewURLService(storage repository.Storage) *URLService {
	return &URLService{storage: storage}
}

func (s *URLService) ShortenURL(url string) (string, error) {
	id, err := s.storage.SaveURL(url)
	return baseURL + id, err
}

func (s *URLService) GetOriginalURL(id string) (string, bool) {
	url, found := s.storage.GetURL(id)
	return url, found
}
