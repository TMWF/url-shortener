package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseWriterWrapper оборачивает http.ResponseWriter для захвата статуса и размера ответа.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

func RequestLoggerMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Middleware is starting to work")
			start := time.Now()

			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Передаем запрос следующему обработчику
			next.ServeHTTP(wrapper, r)

			// Вычисляем время выполнения
			duration := time.Since(start)

			// Логируем сведения о запросе и ответе
			logger.Info("request completed",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Duration("duration", duration),
				zap.Int("status", wrapper.statusCode),
				zap.Int("bytes_written", wrapper.bytesWritten),
			)
		})
	}
}
