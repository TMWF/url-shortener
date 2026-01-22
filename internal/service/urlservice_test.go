package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Мок Storage ---
type mockStorage struct {
	mock.Mock
}

// SaveURL implements repository.Storage.
func (m *mockStorage) SaveURL(ctx context.Context, url string) (string, error) {
	args := m.Called(ctx, url)
	return args.String(0), args.Error(1)
}

// SaveBatchURL implements repository.Storage.
func (m *mockStorage) SaveBatchURL(ctx context.Context, request []model.URLBatchRequestDto) ([]string, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// GetURL implements repository.Storage.
func (m *mockStorage) GetURL(ctx context.Context, id string) (string, bool) {
	args := m.Called(ctx, id)
	return args.String(0), args.Bool(1)
}

// --- Тесты ---

func TestDefaultURLService_ShortenURL(t *testing.T) {
	// Создаем моки
	mockRepo := new(mockStorage)
	cfg := &config.Config{BaseURL: "http://short.url"}

	// Настраиваем мок репозитория
	shortID := "test_id"
	originalURL := "http://original.com"
	mockRepo.On("SaveURL", mock.Anything, originalURL).Return(shortID, nil)

	urlService := NewURLService(mockRepo, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	shortened, err := urlService.ShortenURL(ctx, originalURL)

	assert.NoError(t, err)
	assert.Equal(t, cfg.BaseURL+"/"+shortID, shortened)
	mockRepo.AssertCalled(t, "SaveURL", mock.Anything, originalURL)
}

func TestDefaultURLService_ShortenURL_StorageError(t *testing.T) {
	mockRepo := new(mockStorage)
	cfg := &config.Config{BaseURL: "http://short.url"}
	originalURL := "http://original.com"
	dbErr := errors.New("storage failed")

	// Настраиваем мок репозитория на возврат ошибки
	mockRepo.On("SaveURL", mock.Anything, originalURL).Return("", dbErr)

	urlService := NewURLService(mockRepo, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	shortened, err := urlService.ShortenURL(ctx, originalURL)

	assert.Error(t, err)
	assert.Equal(t, dbErr, err) // Ожидаем, что сервис вернет ту же ошибку
	assert.Empty(t, shortened)
	mockRepo.AssertCalled(t, "SaveURL", mock.Anything, originalURL)
}

func TestDefaultURLService_ShortenURLAPI(t *testing.T) {

	tests := []struct {
		name             string
		originalURL      string
		shortID          string
		mockStorageError error
		expectedResponse *model.ShortenURLResponse
		expectedError    error
		expectLogError   bool
		isConflictError  bool
	}{
		{
			name:             "Success API Shortening",
			originalURL:      "http://original.com/api",
			shortID:          "api_id_success",
			mockStorageError: nil,
			expectedResponse: &model.ShortenURLResponse{ShortenedURL: "http://short.url/api_id_success"},
			expectedError:    nil,
			expectLogError:   false,
		},
		{
			name:             "API Shortening with Storage Error",
			originalURL:      "http://original.com/api_error",
			shortID:          "",
			mockStorageError: dbErr,
			expectedResponse: nil,
			expectedError:    dbErr,
			expectLogError:   true,
		},
		{
			name:             "API Shortening with Conflict Error",
			originalURL:      "http://original.com/api_conflict",
			shortID:          "existing_api_id",
			mockStorageError: repository.ErrConflict,
			expectedResponse: &model.ShortenURLResponse{ShortenedURL: "http://short.url/existing_api_id"},
			expectedError:    repository.ErrConflict,
			expectLogError:   false,
			isConflictError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockStorage)
			cfg := &config.Config{BaseURL: "http://short.url"}
			urlService := NewURLService(mockRepo, cfg)

			request := &model.ShortenURLRequest{URL: tt.originalURL}

			if tt.isConflictError {
				mockRepo.On("SaveURL", mock.Anything, tt.originalURL).Return(tt.shortID, tt.mockStorageError)
			} else {
				mockRepo.On("SaveURL", mock.Anything, tt.originalURL).Return(tt.shortID, tt.mockStorageError)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			response, err := urlService.ShortenURLAPI(ctx, request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResponse, response)
			mockRepo.AssertCalled(t, "SaveURL", mock.Anything, tt.originalURL)
		})
	}
}

func TestDefaultURLService_ShortenURLBatch(t *testing.T) {

	tests := []struct {
		name             string
		requestBody      []model.URLBatchRequestDto
		mockStorageIDs   []string
		mockStorageError error
		expectedResponse []model.URLBatchResponseDto
		expectedError    error
		expectLogError   bool
	}{
		{
			name: "Success Batch Shortening",
			requestBody: []model.URLBatchRequestDto{
				{CorrelationID: "1", OriginalURL: "http://example.com/1"},
				{CorrelationID: "2", OriginalURL: "http://example.com/2"},
			},
			mockStorageIDs:   []string{"batch_id_1", "batch_id_2"},
			mockStorageError: nil,
			expectedResponse: []model.URLBatchResponseDto{
				{CorrelationID: "1", ShortURL: "http://short.url/batch_id_1"},
				{CorrelationID: "2", ShortURL: "http://short.url/batch_id_2"},
			},
			expectedError:  nil,
			expectLogError: false,
		},
		{
			name: "Batch Shortening with Storage Error",
			requestBody: []model.URLBatchRequestDto{
				{CorrelationID: "1", OriginalURL: "http://example.com/error_batch"},
			},
			mockStorageIDs:   nil,
			mockStorageError: dbErr,
			expectedResponse: nil,
			expectedError:    dbErr,
			expectLogError:   true,
		},
		// {
		// 	name:             "Empty Batch Request",
		// 	requestBody:      []model.URLBatchRequestDto{},
		// 	mockStorageIDs:   []string{},
		// 	mockStorageError: nil,
		// 	expectedResponse: []model.URLBatchResponseDto{},
		// 	expectedError:    nil,
		// 	expectLogError:   false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockStorage)
			cfg := &config.Config{BaseURL: "http://short.url"}
			urlService := NewURLService(mockRepo, cfg)

			mockRepo.On("SaveBatchURL", mock.Anything, tt.requestBody).Return(tt.mockStorageIDs, tt.mockStorageError)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			response, err := urlService.ShortenURLBatch(ctx, tt.requestBody)

			// Проверки
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResponse, response)
			mockRepo.AssertCalled(t, "SaveBatchURL", mock.Anything, tt.requestBody)
		})
	}
}

func TestDefaultURLService_GetOriginalURL(t *testing.T) {
	tests := []struct {
		name            string
		shortID         string
		mockOriginalURL string
		mockFoundStatus bool
		expectedURL     string
		expectedFound   bool
	}{
		{
			name:            "URL Found",
			shortID:         "get_id_found",
			mockOriginalURL: "http://get.original.com/found",
			mockFoundStatus: true,
			expectedURL:     "http://get.original.com/found",
			expectedFound:   true,
		},
		{
			name:            "URL Not Found",
			shortID:         "get_id_not_found",
			mockOriginalURL: "",
			mockFoundStatus: false,
			expectedURL:     "",
			expectedFound:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockStorage)
			cfg := &config.Config{BaseURL: "http://short.url"}
			urlService := NewURLService(mockRepo, cfg)

			mockRepo.On("GetURL", mock.Anything, tt.shortID).Return(tt.mockOriginalURL, tt.mockFoundStatus)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			foundURL, found := urlService.GetOriginalURL(ctx, tt.shortID)

			assert.Equal(t, tt.expectedFound, found, "Found status mismatch")
			assert.Equal(t, tt.expectedURL, foundURL, "Original URL mismatch")
			mockRepo.AssertCalled(t, "GetURL", mock.Anything, tt.shortID)
		})
	}
}

// --- Вспомогательные переменные (если нужны) ---
var dbErr = errors.New("database error")
