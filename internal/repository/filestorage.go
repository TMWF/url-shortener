package repository

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/logger"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/util"
	"go.uber.org/zap"
)

type fileStorage struct {
	userID     int
	userUrls   map[int][]string
	urlStorage map[string]model.URLModel
	lock       sync.RWMutex
	cfg        *config.Config
}

func NewFileStorage(config *config.Config) *fileStorage {
	fileStorage := fileStorage{cfg: config}
	urlStorage := make(map[string]model.URLModel)
	file, err := os.OpenFile(fileStorage.cfg.URLStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.GetLogger().Error("Error occured while opening file ",
			zap.String("originalErrorMessage", err.Error()),
		)
		fileStorage.urlStorage = urlStorage
		return &fileStorage
	}

	if err := json.NewDecoder(file).Decode(&urlStorage); err != nil {
		logger.GetLogger().Error("Error occured while decoding urls from file ",
			zap.String("originalErrorMessage", err.Error()),
		)

		fileStorage.urlStorage = urlStorage
		return &fileStorage
	}

	fileStorage.urlStorage = urlStorage
	return &fileStorage
}

func (fs *fileStorage) GetURL(ctx context.Context, id string) (string, bool) {
	fs.lock.RLock()
	defer fs.lock.RUnlock()
	urlModel, found := fs.urlStorage[id]
	return urlModel.OriginalURL, found
}

func (fs *fileStorage) SaveURL(ctx context.Context, url string) (string, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	id := util.RandomString(8, util.LatinCharSet)
	urlModel := model.URLModel{ShortURL: id, OriginalURL: url}
	file, err := os.OpenFile(fs.cfg.URLStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.GetLogger().Error("Error occured while opening file ",
			zap.String("originalErrorMessage", err.Error()),
		)
		return "", err
	}
	defer file.Close()
	fs.urlStorage[id] = urlModel

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(fs.urlStorage); err != nil {
		logger.GetLogger().Error("Error occured while encoding JSON ",
			zap.String("originalErrorMessage", err.Error()),
		)
		return "", err
	}
	return id, nil
}

func (fs *fileStorage) SaveBatchURL(ctx context.Context, urlBatch []model.URLBatchRequestDto) ([]string, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	file, err := os.OpenFile(fs.cfg.URLStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.GetLogger().Error("Error occured while opening file ",
			zap.String("originalErrorMessage", err.Error()),
		)
		return nil, err
	}
	defer file.Close()

	result := make([]string, 0, len(urlBatch))

	for _, urlBatchModel := range urlBatch {
		id := util.RandomString(8, util.LatinCharSet)
		urlModel := model.URLModel{ShortURL: id, OriginalURL: urlBatchModel.OriginalURL}
		fs.urlStorage[id] = urlModel
		result = append(result, id)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(fs.urlStorage); err != nil {
		logger.GetLogger().Error("Error occured while encoding JSON ",
			zap.String("originalErrorMessage", err.Error()),
		)
		return nil, err
	}
	return result, nil
}

func (fs *fileStorage) GetUsersURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error) {
	fs.lock.RLock()
	defer fs.lock.RUnlock()

	userID, ok := ctx.Value(util.UserID).(int)

	if !ok || userID < 1 {
		logger.GetLogger().Error("UserID unexpectedly not found in context")
		return nil, ErrUserIDAbsent
	}

	result := make([]model.GetUserURLsResponseModel, 0)

	urlIDs := fs.userUrls[userID]

	for _, urlID := range urlIDs {
		responseDto := model.GetUserURLsResponseModel{}
		responseDto.ShortURL = urlID
		responseDto.OriginalURL = fs.urlStorage[urlID].OriginalURL
		result = append(result, responseDto)
	}

	return result, nil
}

func (fs *fileStorage) SaveUser(ctx context.Context) (int, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	fs.userID++
	fs.userUrls[fs.userID] = make([]string, 0)
	return fs.userID, nil
}
