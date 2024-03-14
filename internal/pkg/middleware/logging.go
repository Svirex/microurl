package middleware

import (
	"net/http"
	"time"

	"github.com/Svirex/microurl/internal/pkg/logging"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (writer *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := writer.ResponseWriter.Write(data)
	writer.responseData.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

func NewLoggingMiddleware(logger logging.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(writer http.ResponseWriter, request *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}

			loggingWriter := loggingResponseWriter{
				ResponseWriter: writer,
				responseData:   responseData,
			}

			next.ServeHTTP(&loggingWriter, request)

			duration := time.Since(start)

			if logger != nil {
				logger.Info(
					"uri", request.RequestURI,
					"method", request.Method,
					"status", responseData.status, // получаем перехваченный код статуса ответа
					"duration", duration,
					"size", responseData.size, // получаем перехваченный размер ответа
				)
			}
		}
		return http.HandlerFunc(fn)
	}

}
