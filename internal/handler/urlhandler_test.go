package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TMWF/url-shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock для сервиса ---

type MockURLService struct {
	mock.Mock
}

func (m *MockURLService) ShortenURL(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *MockURLService) ShortenURLAPI(req model.ShortenURLRequest) (model.ShortenURLResponse, error) {
	args := m.Called(req)
	return args.Get(0).(model.ShortenURLResponse), args.Error(1)
}

func (m *MockURLService) GetOriginalURL(id string) (string, bool) {
	args := m.Called(id)
	return args.String(0), args.Bool(1)
}

// --- Тесты ---

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
				mockSvc.On("ShortenURL", tt.body).Return(tt.mockReturn, nil)
			} else if tt.mockError != nil {
				mockSvc.On("ShortenURL", tt.body).Return("", tt.mockError)
			}

			h := NewURLHandler(mockSvc)
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
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
	mockSvc := new(MockURLService)
	h := NewURLHandler(mockSvc)
	t.Run("Success JSON API", func(t *testing.T) {
		input := model.ShortenURLRequest{URL: "https://yandex.ru"}
		output := model.ShortenURLResponse{ShortenedURL: "http://localhost:8080/abc"}

		mockSvc.On("ShortenURLAPI", input).Return(output, nil)

		jsonBody, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(jsonBody))
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
	h := NewURLHandler(mockSvc)

	t.Run("Success Redirect", func(t *testing.T) {
		id := "xyz123"
		originalURL := "https://example.com"
		mockSvc.On("GetOriginalURL", id).Return(originalURL, true)

		req := httptest.NewRequest(http.MethodGet, "/"+id, nil)

		// Настройка chi context, чтобы chi.URLParam работал вне реального роутера
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
		mockSvc.On("GetOriginalURL", id).Return("", false)

		req := httptest.NewRequest(http.MethodGet, "/"+id, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		h.GetOriginalURL(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/TMWF/url-shortener/internal/config"
// 	"github.com/TMWF/url-shortener/internal/service"
// 	"github.com/go-chi/chi/v5"
// 	"github.com/go-resty/resty/v2"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// var shortenedURLHeaders = map[string]string{
// 	"Content-Type":   "text/plain",
// 	"Content-Length": "28",
// }

// var getOrigianlURLHeaders = map[string]string{
// 	"Location": "http://practicum.yandex.ru",
// }

// func TestShortenURL(t *testing.T) {
// 	cfg := config.Config{}
// 	cfg.BaseURL = "http://localhost:8080"
// 	var storage = MockStorage{mockID: "mockId", urlStorage: make(map[string]string, 1)}
// 	storage.urlStorage[storage.mockID] = "http://practicum.yandex.ru"
// 	urlService := service.NewURLService(&storage, &cfg)
// 	urlHandler := NewURLHandler(urlService)

// 	router := chi.NewRouter()
// 	router.Post(`/`, urlHandler.ShortenURL)
// 	srv := httptest.NewServer(router)
// 	defer srv.Close()

// 	type want struct {
// 		code     int
// 		response string
// 		headers  map[string]string
// 	}
// 	tests := []struct {
// 		name       string
// 		httpMethod string
// 		want       want
// 	}{
// 		{
// 			name:       "positive test",
// 			httpMethod: http.MethodPost,
// 			want: want{
// 				code:     http.StatusCreated,
// 				response: "http://localhost:8080/mockId",
// 				headers:  shortenedURLHeaders,
// 			},
// 		},
// 	}
// 	client := resty.New()
// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			resp, err := client.R().Execute(test.httpMethod, srv.URL)

// 			assert.Equal(t, test.want.code, resp.StatusCode())
// 			require.NoError(t, err)
// 			assert.Equal(t, test.want.response, string(resp.Body()))

// 			for key, value := range test.want.headers {
// 				assert.Equal(t, value, resp.Header().Get(key))
// 			}
// 		})
// 	}
// }

// func TestGetOriginalURL(t *testing.T) {
// 	cfg := config.Config{}
// 	cfg.ParseFlags()
// 	var storage = MockStorage{mockID: "mockId", urlStorage: make(map[string]string, 1)}
// 	storage.urlStorage[storage.mockID] = "http://practicum.yandex.ru"
// 	urlService := service.NewURLService(&storage, &cfg)
// 	urlHandler := NewURLHandler(urlService)

// 	router := chi.NewRouter()
// 	router.Get(`/{id}`, urlHandler.GetOriginalURL)
// 	srv := httptest.NewServer(router)
// 	defer srv.Close()

// 	type want struct {
// 		code    int
// 		headers map[string]string
// 	}

// 	tests := []struct {
// 		name       string
// 		httpMethod string
// 		want       want
// 	}{
// 		{
// 			name:       "positive _test",
// 			httpMethod: http.MethodGet,
// 			want: want{
// 				code:    http.StatusOK,
// 				headers: getOrigianlURLHeaders,
// 			},
// 		},
// 	}
// 	client := resty.New()
// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			resp, err := client.R().Execute(test.httpMethod, srv.URL+"/mockId")
// 			assert.Equal(t, test.want.code, resp.StatusCode())
// 			require.NoError(t, err)
// 		})
// 	}
// }

// type MockStorage struct {
// 	mockID     string
// 	urlStorage map[string]string
// }

// func (s *MockStorage) SaveURL(url string) (string, error) {
// 	return s.mockID, nil
// }

// func (s *MockStorage) GetURL(id string) (string, bool) {
// 	url, found := s.urlStorage[id]
// 	return url, found
// }
