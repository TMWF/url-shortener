package repository

import (
	"context"
	"sync"

	"github.com/TMWF/url-shortener/internal/util"
)

type memStorage struct {
	urlStorage map[string]string
	lock       sync.RWMutex
}

func newMemStorage() *memStorage {
	return &memStorage{urlStorage: make(map[string]string)}
}

func (ms *memStorage) SaveURL(ctx context.Context, url string) (string, error) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	id := util.RandomString(8, util.LatinCharSet)
	ms.urlStorage[id] = url
	return id, nil
}

func (ms *memStorage) GetURL(ctx context.Context, id string) (string, bool) {
	ms.lock.RLock()
	defer ms.lock.RUnlock()
	url, found := ms.urlStorage[id]
	return url, found
}
