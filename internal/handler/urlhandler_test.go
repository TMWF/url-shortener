package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"testing"

	"github.com/TMWF/url-shortener/internal/middleware"
	"github.com/TMWF/url-shortener/internal/mocks"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/repository"
	"github.com/TMWF/url-shortener/internal/util"
	"github.com/go-chi/chi/v5"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockURLService struct {
	mock.Mock
}

// GetUserURLs implements [service.URLService].
func (m *MockURLService) GetUserURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error) {
	panic("unimplemented")
}

// SaveUser implements [service.URLService].
func (m *MockURLService) SaveUser(ctx context.Context) (int, error) {
	panic("unimplemented")
}

// ScheduleUserURLsJob implements [service.URLService].
func (m *MockURLService) ScheduleUserURLsJob(ctx context.Context, urlIDs []string) {
	m.Called(ctx, urlIDs)
}

// ShortenURLBatch implements [service.URLService].
func (m *MockURLService) ShortenURLBatch(ctx context.Context, request []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error) {
	args := m.Called(ctx, request)
	return args.Get(0).([]model.URLBatchResponseDto), args.Error(1)
}

func (m *MockURLService) ShortenURL(ctx context.Context, url string) (string, error) {
	args := m.Called(ctx, url)
	return args.String(0), args.Error(1)
}

func (m *MockURLService) ShortenURLAPI(ctx context.Context, req *model.ShortenURLRequest) (*model.ShortenURLResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.ShortenURLResponse), args.Error(1)
}

func (m *MockURLService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func gzipData(t *testing.T, data []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(data)
	assert.NoError(t, err)
	err = zw.Close()
	assert.NoError(t, err)
	return buf.Bytes()
}

func gunzipData(t *testing.T, data []byte) []byte {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	assert.NoError(t, err)
	res, err := io.ReadAll(zr)
	assert.NoError(t, err)
	return res
}

// --- Тесты ---

func TestGzipMiddlewareIntegration(t *testing.T) {
	controller := gomock.NewController(t)
	mockJwtBuilder := mocks.NewMockUserJWTBuilder(controller)
	mockSvc := new(MockURLService)
	h := NewURLHandler(mockSvc, mockJwtBuilder)

	gzipHandler := middleware.GzipMiddleware()(http.HandlerFunc(h.ShortenURLAPI))

	t.Run("should_compress_response", func(t *testing.T) {
		input := model.ShortenURLRequest{URL: "https://google.com"}
		output := &model.ShortenURLResponse{ShortenedURL: "http://localhost:8080/abc"}

		mockSvc.On("ShortenURLAPI", mock.Anything, &input).Return(output, nil).Once()
		context := context.WithValue(context.Background(), util.UserID, 1)
		body, _ := json.Marshal(input)
		req := httptest.NewRequestWithContext(context, http.MethodPost, "/api/shorten", bytes.NewReader(body))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		gzipHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
		assert.Empty(t, w.Header().Get("Content-Length"), "Content-Length should be deleted when gzipping")

		unzippedBody := gunzipData(t, w.Body.Bytes())
		var actualResp model.ShortenURLResponse
		json.Unmarshal(unzippedBody, &actualResp)
		assert.Equal(t, output.ShortenedURL, actualResp.ShortenedURL)
	})

	t.Run("should_decompress_request", func(t *testing.T) {
		input := model.ShortenURLRequest{URL: "https://yandex.ru"}
		output := model.ShortenURLResponse{ShortenedURL: "http://localhost:8080/def"}

		mockSvc.On("ShortenURLAPI", mock.Anything, &input).Return(&output, nil).Once()

		jsonBytes, _ := json.Marshal(input)
		compressedBody := gzipData(t, jsonBytes)

		context := context.WithValue(context.Background(), util.UserID, 1)
		req := httptest.NewRequestWithContext(context, http.MethodPost, "/api/shorten", bytes.NewReader(compressedBody))
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		gzipHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestShortenURL(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		body         string
		mockReturn   string
		mockError    error
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Success POST",
			method:       http.MethodPost,
			body:         "https://google.com",
			mockReturn:   "http://localhost:8080/short",
			mockError:    nil,
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/short",
		},
		{
			name:         "Wrong Method GET",
			method:       http.MethodGet,
			body:         "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "Service Error",
			method:       http.MethodPost,
			body:         "https://google.com",
			mockError:    errors.New("db error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockURLService)
			if tt.method == http.MethodPost && tt.mockError == nil && tt.name != "Service Error" {
				mockSvc.On("ShortenURL", mock.Anything, tt.body).Return(tt.mockReturn, nil)
			} else if tt.mockError != nil {
				mockSvc.On("ShortenURL", mock.Anything, tt.body).Return("", tt.mockError)
			}
			controller := gomock.NewController(t)
			jwtBuilder := mocks.NewMockUserJWTBuilder(controller)
			h := NewURLHandler(mockSvc, jwtBuilder)
			context := context.WithValue(context.Background(), util.UserID, 1)
			req := httptest.NewRequestWithContext(context, tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.ShortenURL(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
				assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestShortenURLAPI(t *testing.T) {
	controller := gomock.NewController(t)
	jwtBuilder := mocks.NewMockUserJWTBuilder(controller)
	mockSvc := new(MockURLService)
	h := NewURLHandler(mockSvc, jwtBuilder)
	t.Run("Success JSON API", func(t *testing.T) {
		input := model.ShortenURLRequest{URL: "https://yandex.ru"}
		output := &model.ShortenURLResponse{ShortenedURL: "http://localhost:8080/abc"}

		mockSvc.On("ShortenURLAPI", mock.Anything, &input).Return(output, nil)

		jsonBody, _ := json.Marshal(input)
		context := context.WithValue(context.Background(), util.UserID, 1)
		req := httptest.NewRequestWithContext(context, http.MethodPost, "/api/shorten", bytes.NewReader(jsonBody))
		w := httptest.NewRecorder()

		h.ShortenURLAPI(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var actualRes model.ShortenURLResponse
		json.NewDecoder(w.Body).Decode(&actualRes)
		assert.Equal(t, output.ShortenedURL, actualRes.ShortenedURL)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url": "no-closing-brace`))
		w := httptest.NewRecorder()

		h.ShortenURLAPI(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetOriginalURL(t *testing.T) {
	mockSvc := new(MockURLService)
	controller := gomock.NewController(t)
	jwtBuilder := mocks.NewMockUserJWTBuilder(controller)
	h := NewURLHandler(mockSvc, jwtBuilder)

	t.Run("Success Redirect", func(t *testing.T) {
		id := "xyz123"
		originalURL := "https://example.com"
		mockSvc.On("GetOriginalURL", mock.Anything, id).Return(originalURL, nil)

		req := httptest.NewRequest(http.MethodGet, "/"+id, nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.GetOriginalURL(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, originalURL, w.Header().Get("Location"))
	})

	t.Run("Not Found", func(t *testing.T) {
		id := "nonexistent"
		mockSvc.On("GetOriginalURL", mock.Anything, id).Return("", repository.ErrURLNotFound)

		req := httptest.NewRequest(http.MethodGet, "/"+id, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.GetOriginalURL(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Gone", func(t *testing.T) {
		id := "gone"
		mockSvc.On("GetOriginalURL", mock.Anything, id).Return("", repository.ErrURLDeleted)

		req := httptest.NewRequest(http.MethodGet, "/"+id, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.GetOriginalURL(w, req)

		assert.Equal(t, http.StatusGone, w.Code)
	})
}

func TestShortenURLBatch(t *testing.T) {
	tests := []struct {
		name                 string
		method               string
		requestBody          []model.URLBatchRequestDto
		mockServiceResponse  []model.URLBatchResponseDto
		mockServiceError     error
		expectedStatusCode   int
		expectedResponseBody string
		expectLogError       bool
	}{
		{
			name:   "Success Batch Shortening",
			method: http.MethodPost,
			requestBody: []model.URLBatchRequestDto{
				{CorrelationID: "1", OriginalURL: "http://example.com/long/url/1"},
				{CorrelationID: "2", OriginalURL: "http://example.com/long/url/2"},
			},
			mockServiceResponse: []model.URLBatchResponseDto{
				{CorrelationID: "1", ShortURL: "http://localhost:8080/abc"},
				{CorrelationID: "2", ShortURL: "http://localhost:8080/def"},
			},
			mockServiceError:     nil,
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `[{"correlation_id":"1","short_url":"http://localhost:8080/abc"},{"correlation_id":"2","short_url":"http://localhost:8080/def"}]`,
			expectLogError:       false,
		},
		{
			name:                 "Invalid Method GET",
			method:               http.MethodGet,
			requestBody:          nil,
			mockServiceResponse:  nil,
			mockServiceError:     nil,
			expectedStatusCode:   http.StatusMethodNotAllowed,
			expectedResponseBody: "Incorrect HTTP method, only POST methods allowed\n",
			expectLogError:       false,
		},
		{
			name:                 "Invalid JSON Body",
			method:               http.MethodPost,
			requestBody:          nil,
			mockServiceResponse:  nil,
			mockServiceError:     nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Error occured while decoding request body\n",
			expectLogError:       true,
		},
		// {
		// 	name:                 "Empty Batch Request",
		// 	method:               http.MethodPost,
		// 	requestBody:          []model.URLBatchRequestDto{},
		// 	mockServiceResponse:  nil, // Не будет вызвано
		// 	mockServiceError:     nil,
		// 	expectedStatusCode:   http.StatusBadRequest,
		// 	expectedResponseBody: "Request body cannot be empty\n",
		// 	expectLogError:       false,
		// },
		{
			name:   "Service Returns Error",
			method: http.MethodPost,
			requestBody: []model.URLBatchRequestDto{
				{CorrelationID: "1", OriginalURL: "http://example.com/error"},
			},
			mockServiceResponse:  nil,
			mockServiceError:     errors.New("database connection failed"),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Error occured while getting shortened url\n",
			expectLogError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockURLService)

			var reqBodyReader io.Reader
			if tt.requestBody != nil {
				bodyBytes, err := json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body for test setup: %v", err)
				}
				reqBodyReader = bytes.NewReader(bodyBytes)
			} else if tt.name == "Invalid JSON Body" {
				reqBodyReader = strings.NewReader(`{ "invalid": "json" `)
			} else {
				reqBodyReader = nil
			}

			req := httptest.NewRequest(tt.method, "/api/batch", reqBodyReader)
			if tt.method == http.MethodPost && tt.requestBody != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			if tt.method == http.MethodPost && tt.name != "Invalid JSON Body" && tt.name != "Empty Batch Request" {
				mockSvc.On("ShortenURLBatch", mock.Anything, tt.requestBody).Return(tt.mockServiceResponse, tt.mockServiceError)
			}

			controller := gomock.NewController(t)
			jwtBuilder := mocks.NewMockUserJWTBuilder(controller)
			w := httptest.NewRecorder()
			h := NewURLHandler(mockSvc, jwtBuilder)

			context := context.WithValue(context.Background(), util.UserID, 1)
			req = req.WithContext(context)
			h.ShortenURLBatch(w, req)

			// Проверки
			assert.Equal(t, tt.expectedStatusCode, w.Code, "Expected status code mismatch")
			assert.Equal(t, tt.expectedResponseBody, w.Body.String(), "Expected response body mismatch")

			if tt.method == http.MethodPost && tt.name != "Invalid JSON Body" && tt.name != "Empty Batch Request" && tt.mockServiceError == nil {
				mockSvc.AssertCalled(t, "ShortenURLBatch", mock.Anything, tt.requestBody)
			} else if tt.name == "Service Returns Error" {
				mockSvc.AssertCalled(t, "ShortenURLBatch", mock.Anything, tt.requestBody)
			} else {
				mockSvc.AssertNotCalled(t, "ShortenURLBatch", mock.Anything, mock.Anything)
			}
		})
	}
}
