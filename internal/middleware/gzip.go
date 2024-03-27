package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// func checkAcceptEncoding(request *http.Request) bool {
// 	for _, v := range request.Header.Values("Accept-Encoding") {
// 		splitted := strings.Split(v, ",")
// 		for i := range splitted {
// 			if strings.Contains(splitted[i], "gzip") {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

func checkContentEncoding(request *http.Request) bool {
	return strings.Contains(request.Header.Get("Content-Encoding"), "gzip")
}

// type gzipWriter struct {
// 	http.ResponseWriter
// 	Writer io.Writer
// }

// func (w gzipWriter) Write(b []byte) (int, error) {
// 	return w.Writer.Write(b)
// }

func GzipHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if checkContentEncoding(request) {
			gz, err := gzip.NewReader(request.Body)
			if err != nil {
				response.WriteHeader(http.StatusBadRequest)
				return
			}
			request.Body = gz
		}
		next.ServeHTTP(response, request)

	})
}
