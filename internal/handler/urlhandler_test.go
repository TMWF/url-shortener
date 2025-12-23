package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TMWF/url-shortener/internal/config"
	"github.com/TMWF/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
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
	var storage = MockStorage{mockID: "mockId", urlStorage: make(map[string]string, 1)}
	storage.urlStorage[storage.mockID] = "http://practicum.yandex.ru"
	urlService := service.NewURLService(&storage)
	urlHandler := NewURLHandler(*urlService)

	router := chi.NewRouter()
	router.Post(`/`, urlHandler.ShortenURL)
	srv := httptest.NewServer(router)
	defer srv.Close()
	config.ParseFlags()

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
	}
	client := resty.New()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := client.R().Execute(test.httpMethod, srv.URL)

			assert.Equal(t, test.want.code, resp.StatusCode())
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resp.Body()))

			for key, value := range test.want.headers {
				assert.Equal(t, value, resp.Header().Get(key))
			}
			// request := httptest.NewRequest(
			// 	test.httpMethod,
			// 	"http://localhost:8080/",
			// 	strings.NewReader("http://practicum.yandex.ru"),
			// )
			// // создаём новый Recorder
			// w := httptest.NewRecorder()
			// urlHandler.ShortenURL(w, request)

			// res := w.Result()
			// // проверяем код ответа
			// assert.Equal(t, test.want.code, res.StatusCode)
			// // получаем и проверяем тело запроса
			// defer res.Body.Close()
			// resBody, err := io.ReadAll(res.Body)

			// require.NoError(t, err)

			// for key, value := range test.want.headers {
			// 	assert.Equal(t, value, res.Header.Get(key))
			// }
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	var storage = MockStorage{mockID: "mockId", urlStorage: make(map[string]string, 1)}
	storage.urlStorage[storage.mockID] = "http://practicum.yandex.ru"
	urlService := service.NewURLService(&storage)
	urlHandler := NewURLHandler(*urlService)

	router := chi.NewRouter()
	router.Get(`/{id}`, urlHandler.GetOriginalURL)
	srv := httptest.NewServer(router)
	defer srv.Close()

	type want struct {
		code    int
		headers map[string]string
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
				code:    http.StatusOK,
				headers: getOrigianlURLHeaders,
			},
		},
	}
	client := resty.New()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := client.R().Execute(test.httpMethod, srv.URL+"/mockId")
			assert.Equal(t, test.want.code, resp.StatusCode())
			require.NoError(t, err)
		})
	}
}

type MockStorage struct {
	mockID     string
	urlStorage map[string]string
}

func (s *MockStorage) SaveURL(url string) (string, error) {
	return s.mockID, nil
}

func (s *MockStorage) GetURL(id string) (string, bool) {
	url, found := s.urlStorage[id]
	return url, found
}
