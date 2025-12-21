package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TMWF/url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var shortenedURLHeaders = map[string]string{
	"Content-Type":   "text/plain",
	"Content-Length": "28",
}

var getOrigianlURLHeaders = map[string]string{
	"Location": "http://practicum.yandex.ru",
}

func TestShortenURL(t *testing.T) {
	var storage = MockStorage{mockId: "mockId", urlStorage: make(map[string]string, 1)}
	storage.urlStorage[storage.mockId] = "http://practicum.yandex.ru"
	type want struct {
		code     int
		response string
		headers  map[string]string
	}
	tests := []struct {
		name       string
		httpMethod string
		want       want
	}{
		{
			name:       "positive test",
			httpMethod: http.MethodPost,
			want: want{
				code:     http.StatusCreated,
				response: "http://localhost:8080/mockId",
				headers:  shortenedURLHeaders,
			},
		},
		{
			name:       "negative test - wrong http method",
			httpMethod: http.MethodGet,
			want: want{
				code:     http.StatusMethodNotAllowed,
				response: "Incorrect HTTP method, only POST methods allowed\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				test.httpMethod,
				"http://localhost:8080/",
				strings.NewReader("http://practicum.yandex.ru"),
			)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			urlService := service.NewURLService(&storage)
			urlHandler := NewURLHandler(*urlService)
			urlHandler.ShortenURL(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			for key, value := range test.want.headers {
				assert.Equal(t, value, res.Header.Get(key))
			}
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	var storage = MockStorage{mockId: "mockId", urlStorage: make(map[string]string, 1)}
	storage.urlStorage[storage.mockId] = "http://practicum.yandex.ru"
	type want struct {
		code     int
		response string
		headers  map[string]string
	}
	tests := []struct {
		name       string
		httpMethod string
		want       want
	}{
		{
			name:       "positive _test",
			httpMethod: http.MethodGet,
			want: want{
				code:    http.StatusTemporaryRedirect,
				headers: getOrigianlURLHeaders,
			},
		},
		{
			name:       "negative test - wrong http method",
			httpMethod: http.MethodPost,
			want: want{
				code:     http.StatusMethodNotAllowed,
				response: "Incorrect HTTP method, only GET methods allowed\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				test.httpMethod,
				"http://localhost:8080/mockId",
				nil,
			)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			urlService := service.NewURLService(&storage)
			urlHandler := NewURLHandler(*urlService)
			urlHandler.GetOriginalURL(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			for key, value := range test.want.headers {
				assert.Equal(t, value, res.Header.Get(key))
			}
		})
	}
}

type MockStorage struct {
	mockId     string
	urlStorage map[string]string
}

func (s *MockStorage) SaveURL(url string) (string, error) {
	return s.mockId, nil
}

func (s *MockStorage) GetURL(id string) (string, bool) {
	url, found := s.urlStorage[id]
	return url, found
}
