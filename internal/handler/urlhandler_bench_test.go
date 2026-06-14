package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TMWF/url-shortener/internal/handler"
	"github.com/TMWF/url-shortener/internal/model"
	"github.com/TMWF/url-shortener/internal/util"
)

type mockURLService struct {
	ShortenURLFunc          func(ctx context.Context, url string) (string, error)
	ShortenURLAPIFunc       func(ctx context.Context, request *model.ShortenURLRequest) (*model.ShortenURLResponse, error)
	ShortenURLBatchFunc     func(ctx context.Context, request []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error)
	GetOriginalURLFunc      func(ctx context.Context, id string) (string, error)
	GetUserURLsFunc         func(ctx context.Context) ([]model.GetUserURLsResponseModel, error)
	ScheduleUserURLsJobFunc func(ctx context.Context, urlIDs []string)
	SaveUserFunc            func(ctx context.Context) (int, error)
}

func (m *mockURLService) ShortenURL(ctx context.Context, url string) (string, error) {
	if m.ShortenURLFunc == nil {
		panic("mockURLService: ShortenURLFunc is not defined")
	}
	return m.ShortenURLFunc(ctx, url)
}

func (m *mockURLService) ShortenURLAPI(ctx context.Context, request *model.ShortenURLRequest) (*model.ShortenURLResponse, error) {
	if m.ShortenURLAPIFunc == nil {
		panic("mockURLService: ShortenURLAPIFunc is not defined")
	}
	return m.ShortenURLAPIFunc(ctx, request)
}

func (m *mockURLService) ShortenURLBatch(ctx context.Context, request []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error) {
	if m.ShortenURLBatchFunc == nil {
		panic("mockURLService: ShortenURLBatchFunc is not defined")
	}
	return m.ShortenURLBatchFunc(ctx, request)
}

func (m *mockURLService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	if m.GetOriginalURLFunc == nil {
		panic("mockURLService: GetOriginalURLFunc is not defined")
	}
	return m.GetOriginalURLFunc(ctx, id)
}

func (m *mockURLService) GetUserURLs(ctx context.Context) ([]model.GetUserURLsResponseModel, error) {
	if m.GetUserURLsFunc == nil {
		panic("mockURLService: GetUserURLsFunc is not defined")
	}
	return m.GetUserURLsFunc(ctx)
}

func (m *mockURLService) ScheduleUserURLsJob(ctx context.Context, urlIDs []string) {
	if m.ScheduleUserURLsJobFunc == nil {
		panic("mockURLService: ScheduleUserURLsJobFunc is not defined")
	}
	m.ScheduleUserURLsJobFunc(ctx, urlIDs)
}

func (m *mockURLService) SaveUser(ctx context.Context) (int, error) {
	if m.SaveUserFunc == nil {
		panic("mockURLService: SaveUserFunc is not defined")
	}
	return m.SaveUserFunc(ctx)
}

type mockUserJWTBuilder struct {
	BuildJWTStringFunc func(userID int) (string, error)
}

// Гарантируем на этапе компиляции, что mockUserJWTBuilder соответствует интерфейсу.
var _ util.UserJWTBuilder = (*mockUserJWTBuilder)(nil)

func (m *mockUserJWTBuilder) BuildJWTString(userID int) (string, error) {
	if m.BuildJWTStringFunc == nil {
		panic("mockUserJWTBuilder: BuildJWTStringFunc is not defined")
	}
	return m.BuildJWTStringFunc(userID)
}

func BenchmarkShortenURL(b *testing.B) {
	mockJwtBuilder := &mockUserJWTBuilder{
		BuildJWTStringFunc: func(userID int) (string, error) {
			return "test.jwt.token", nil
		},
	}
	mockSvc := &mockURLService{
		ShortenURLFunc: func(ctx context.Context, longURL string) (string, error) {
			// Имитируем успешный быстрый ответ сервиса
			return "http://localhost:8080/shrtn", nil
		},
		SaveUserFunc: func(ctx context.Context) (int, error) {
			return 123456, nil
		},
	}

	h := handler.NewURLHandler(mockSvc, mockJwtBuilder)

	longURL := "https://very-long-and-complicated-url-to-shorten-and-test-performance.com/path/to/resource?query=1"

	b.ReportAllocs()
	b.ResetTimer()

	// 2. Основной цикл бенчмарка
	for i := 0; i < b.N; i++ {
		// Создаем новое тело запроса для каждой итерации
		body := strings.NewReader(longURL)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		w := httptest.NewRecorder()

		h.ShortenURL(w, req)

		// Базовая валидация, чтобы убедиться, что хендлер вообще работает
		resp := w.Result()
		if resp.StatusCode != http.StatusCreated {
			b.Fatalf("expected status 201, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// Параллельный бенчмарк для проверки работы под высокой конкурентной нагрузкой (concurrency)
func BenchmarkShortenURL_Parallel(b *testing.B) {
	mockJwtBuilder := &mockUserJWTBuilder{
		BuildJWTStringFunc: func(userID int) (string, error) {
			return "test.jwt.token", nil
		},
	}
	mockSvc := &mockURLService{
		ShortenURLFunc: func(ctx context.Context, longURL string) (string, error) {
			// Имитируем успешный быстрый ответ сервиса
			return "http://localhost:8080/shrtn", nil
		},
		SaveUserFunc: func(ctx context.Context) (int, error) {
			return 123456, nil
		},
	}

	h := handler.NewURLHandler(mockSvc, mockJwtBuilder)

	longURL := "https://example.com/long"

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			body := strings.NewReader(longURL)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			w := httptest.NewRecorder()

			h.ShortenURL(w, req)
			w.Result().Body.Close()
		}
	})
}

func BenchmarkShortenURLAPI(b *testing.B) {
	// Инициализация моков
	mockSvc := &mockURLService{
		ShortenURLAPIFunc: func(ctx context.Context, request *model.ShortenURLRequest) (*model.ShortenURLResponse, error) {
			return &model.ShortenURLResponse{
				ShortenedURL: "http://localhost:8080/shrtn",
			}, nil
		},
		SaveUserFunc: func(ctx context.Context) (int, error) {
			return 123456, nil
		},
	}

	mockJwtBuilder := &mockUserJWTBuilder{
		BuildJWTStringFunc: func(userID int) (string, error) {
			return "test.jwt.token", nil
		},
	}

	h := handler.NewURLHandler(mockSvc, mockJwtBuilder)

	const jsonPayload = `{"url":"https://very-long-and-complicated-url-to-shorten-and-test-performance.com/path/to/resource?query=1"}`

	b.ReportAllocs() // Включаем сбор метрик памяти
	b.ResetTimer()   // Сбрасываем время, затраченное на инициализацию хендлеров и моков

	for i := 0; i < b.N; i++ {
		// strings.NewReader работает быстрее, чем bytes.NewBuffer
		body := strings.NewReader(jsonPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ShortenURLAPI(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusCreated {
			b.Fatalf("expected status 201, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// Параллельный бенчмарк для эмуляции конкурентной нагрузки (Concurrency)
func BenchmarkShortenURLAPI_Parallel(b *testing.B) {
	mockSvc := &mockURLService{
		ShortenURLAPIFunc: func(ctx context.Context, request *model.ShortenURLRequest) (*model.ShortenURLResponse, error) {
			return &model.ShortenURLResponse{
				ShortenedURL: "http://localhost:8080/shrtn",
			}, nil
		},
		SaveUserFunc: func(ctx context.Context) (int, error) {
			return 123456, nil
		},
	}

	mockJwtBuilder := &mockUserJWTBuilder{
		BuildJWTStringFunc: func(userID int) (string, error) {
			return "test.jwt.token", nil
		},
	}

	h := handler.NewURLHandler(mockSvc, mockJwtBuilder)

	const jsonPayload = `{"url":"https://example.com/long"}`

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			body := strings.NewReader(jsonPayload)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.ShortenURLAPI(w, req)
			w.Result().Body.Close()
		}
	})
}

func BenchmarkShortenURLBatch(b *testing.B) {
	mockSvc := &mockURLService{
		ShortenURLBatchFunc: func(ctx context.Context, request []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error) {
			resp := make([]model.URLBatchResponseDto, len(request))
			for i, req := range request {
				resp[i] = model.URLBatchResponseDto{
					CorrelationID: req.CorrelationID,
					ShortURL:      "http://localhost:8080/" + req.CorrelationID,
				}
			}
			return resp, nil
		},
		SaveUserFunc: func(ctx context.Context) (int, error) {
			return 123456, nil
		},
	}

	mockJwtBuilder := &mockUserJWTBuilder{
		BuildJWTStringFunc: func(userID int) (string, error) {
			return "test.jwt.token", nil
		},
	}

	h := handler.NewURLHandler(mockSvc, mockJwtBuilder)

	const batchPayload = `[
		{"correlation_id": "req-1", "original_url": "https://yandex.ru/maps"},
		{"correlation_id": "req-2", "original_url": "https://google.com/search?q=golang"},
		{"correlation_id": "req-3", "original_url": "https://github.com/golang/go"},
		{"correlation_id": "req-4", "original_url": "https://random1.com/golang/go"},
		{"correlation_id": "req-5", "original_url": "https://random2.com/golang/go"},
		{"correlation_id": "req-6", "original_url": "https://random3.com/golang/go"}
	]`

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		body := strings.NewReader(batchPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h.ShortenURLBatch(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusCreated {
			b.Fatalf("expected status 201, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// Параллельный бенчмарк (нагрузка в несколько горутин)
func BenchmarkShortenURLBatch_Parallel(b *testing.B) {
	mockSvc := &mockURLService{
		ShortenURLBatchFunc: func(ctx context.Context, request []model.URLBatchRequestDto) ([]model.URLBatchResponseDto, error) {
			resp := make([]model.URLBatchResponseDto, len(request))
			for i, req := range request {
				resp[i] = model.URLBatchResponseDto{
					CorrelationID: req.CorrelationID,
					ShortURL:      "http://localhost:8080/shrt",
				}
			}
			return resp, nil
		},
		SaveUserFunc: func(ctx context.Context) (int, error) {
			return 123456, nil
		},
	}

	mockJwtBuilder := &mockUserJWTBuilder{
		BuildJWTStringFunc: func(userID int) (string, error) {
			return "test.jwt.token", nil
		},
	}

	h := handler.NewURLHandler(mockSvc, mockJwtBuilder)

	const batchPayload = `[
		{"correlation_id": "1", "original_url": "https://yandex.ru"},
		{"correlation_id": "2", "original_url": "https://google.com"}
	]`

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			body := strings.NewReader(batchPayload)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.ShortenURLBatch(w, req)
			w.Result().Body.Close()
		}
	})
}

func BenchmarkGetUserURLs(b *testing.B) {
	mockSvc := &mockURLService{
		GetUserURLsFunc: func(ctx context.Context) ([]model.GetUserURLsResponseModel, error) {
			return []model.GetUserURLsResponseModel{
				{
					ShortURL:    "http://localhost:8080/sh1",
					OriginalURL: "https://yandex.ru",
				},
				{
					ShortURL:    "http://localhost:8080/sh2",
					OriginalURL: "https://google.com",
				},
			}, nil
		},
		SaveUserFunc: func(ctx context.Context) (int, error) {
			return 123456, nil
		},
	}

	mockJwtBuilder := &mockUserJWTBuilder{
		BuildJWTStringFunc: func(userID int) (string, error) {
			return "test.jwt.token", nil
		},
	}

	h := handler.NewURLHandler(mockSvc, mockJwtBuilder)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		w := httptest.NewRecorder()

		h.GetUserURLs(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

// Параллельный бенчмарк для проверки конкурентного чтения списка ссылок
func BenchmarkGetUserURLs_Parallel(b *testing.B) {
	mockSvc := &mockURLService{
		GetUserURLsFunc: func(ctx context.Context) ([]model.GetUserURLsResponseModel, error) {
			return []model.GetUserURLsResponseModel{
				{
					ShortURL:    "http://localhost:8080/shrt",
					OriginalURL: "https://example.com",
				},
			}, nil
		},
		SaveUserFunc: func(ctx context.Context) (int, error) {
			return 123456, nil
		},
	}

	mockJwtBuilder := &mockUserJWTBuilder{
		BuildJWTStringFunc: func(userID int) (string, error) {
			return "test.jwt.token", nil
		},
	}

	h := handler.NewURLHandler(mockSvc, mockJwtBuilder)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			w := httptest.NewRecorder()

			h.GetUserURLs(w, req)
			w.Result().Body.Close()
		}
	})
}
