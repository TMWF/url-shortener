package repository

import (
	"sync"

	"github.com/TMWF/url-shortener/internal/util"
)

type memStorage struct {
	urlStorage map[string]string
	lock       sync.RWMutex
}

func NewMemStorage() *memStorage {
	return &memStorage{urlStorage: make(map[string]string)}
}

func (fs *memStorage) SaveURL(url string) (string, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	id := util.RandomString(8, util.LatinCharSet)
	fs.urlStorage[id] = url
	return id, nil
}

func (ms *memStorage) GetURL(id string) (string, bool) {
	ms.lock.RLock()
	defer ms.lock.RUnlock()
	url, found := ms.urlStorage[id]
	return url, found
}
