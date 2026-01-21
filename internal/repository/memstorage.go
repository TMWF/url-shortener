package repository

import (
	"context"
	"sync"

	"github.com/TMWF/url-shortener/internal/model"
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

func (ms *memStorage) SaveBatchURL(ctx context.Context, urlBatch []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	result := make([]model.URLBatchResponseDto, len(urlBatch))

	for _, urlModel := range urlBatch {
		id := util.RandomString(8, util.LatinCharSet)
		ms.urlStorage[id] = urlModel.OriginalURL
		responseModel := model.URLBatchResponseDto{CorrelationID: urlModel.CorrelationID, ShortURL: id}
		result = append(result, responseModel)
	}

	return result, nil
}
