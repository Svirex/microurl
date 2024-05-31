package api

import (
	"net/http"
	"time"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write - запись тела ответа
func (w *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.responseData.size += size
	return size, err
}

// WriteHeader - запись заголовка ответа
func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

func (api *API) loggingMiddleware(next http.Handler) http.Handler {
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

		if api.logger != nil {
			api.logger.Infoln(
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
