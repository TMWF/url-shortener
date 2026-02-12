package repository

import (
	"context"
	"sync"

	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/util"
)

type memStorage struct {
	userID     int
	urlStorage map[string]string
	userUrls   map[int][]string
	lock       sync.RWMutex
}

func newMemStorage() *memStorage {
	return &memStorage{urlStorage: make(map[string]string), userUrls: make(map[int][]string)}
}

func (ms *memStorage) SaveURL(ctx context.Context, url string) (string, error) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	userID, ok := ctx.Value(util.UserID).(int)

	if !ok || userID < 1 {
		logger.GetLogger().Error("UserID unexpectedly not found in context")

		return "", ErrUserIDAbsent
	}

	id := util.RandomString(8, util.LatinCharSet)
	ms.urlStorage[id] = url
	ms.userUrls[userID] = append(ms.userUrls[userID], id)
	return id, nil
}

func (ms *memStorage) GetURL(ctx context.Context, id string) (string, bool) {
	ms.lock.RLock()
	defer ms.lock.RUnlock()
	url, found := ms.urlStorage[id]
	return url, found
}

func (ms *memStorage) SaveBatchURL(ctx context.Context, urlBatch []model.URLBatchRequestDto) ([]string, error) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	result := make([]string, 0, len(urlBatch))

	for _, urlModel := range urlBatch {
		id := util.RandomString(8, util.LatinCharSet)
		ms.urlStorage[id] = urlModel.OriginalURL
		result = append(result, id)
	}

	return result, nil
}

func (ms *memStorage) GetUsersURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error) {
	ms.lock.RLock()
	defer ms.lock.RUnlock()

	userID, ok := ctx.Value(util.UserID).(int)

	if !ok || userID < 1 {
		logger.GetLogger().Error("UserID unexpectedly not found in context")
		return nil, ErrUserIDAbsent
	}

	result := make([]model.GetUserURLsResponseModel, 0)

	urlIDs := ms.userUrls[userID]

	for _, urlID := range urlIDs {
		responseDto := model.GetUserURLsResponseModel{}
		responseDto.ShortURL = urlID
		responseDto.OriginalURL = ms.urlStorage[urlID]
		result = append(result, responseDto)
	}

	return result, nil
}

func (ms *memStorage) SaveUser(ctx context.Context) (int, error) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	ms.userID++
	ms.userUrls[ms.userID] = make([]string, 0)
	return ms.userID, nil
}
