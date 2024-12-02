// Package logger предоставляет инструменты для логирования HTTP-запросов.
// Включает middleware для логирования информации о запросах и ответах.
package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Sugar — глобальная переменная для использования sugared-логгера.
var Sugar zap.SugaredLogger

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int // Код статуса HTTP-ответа.
		size   int // Размер тела ответа в байтах.
	}

	// loggingResponseWriter реализует интерфейс http.ResponseWriter для захвата информации о HTTP-ответах.
	loggingResponseWriter struct {
		http.ResponseWriter               // Оригинальный http.ResponseWriter.
		responseData        *responseData // Структура для хранения данных о ответе.
	}
)

// Write перехватывает запись в тело ответа, чтобы захватить его размер.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

// WriteHeader перехватывает запись заголовка ответа, чтобы захватить код статуса.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// LoggingMiddleware добавляет логирование HTTP-запросов и ответов.
// Логируется URI, метод, статус ответа, длительность обработки и размер ответа.
func LoggingMiddleware(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(start)

		Sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)
	}
	return http.HandlerFunc(logFn)
}
