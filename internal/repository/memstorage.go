package repository

import (
	"github.com/TMWF/url-shortener/internal/util"
)

type memStorage struct {
	urlStorage map[string]string
}

func NewMemStorage() *memStorage {
	return &memStorage{urlStorage: make(map[string]string)}
}

func (ms *memStorage) SaveURL(url string) (string, error) {
	id := util.RandomString(8, util.LatinCharSet)
	ms.urlStorage[id] = url
	return id, nil
}

func (ms *memStorage) GetURL(id string) (string, bool) {
	url, found := ms.urlStorage[id]
	return url, found
}
