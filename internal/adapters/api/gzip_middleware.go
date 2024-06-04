package api

import (
	"compress/gzip"
	"net/http"
	"strings"
)

func checkContentEncoding(request *http.Request) bool {
	return strings.Contains(request.Header.Get("Content-Encoding"), "gzip")
}

func (api *API) gzipHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if !checkContentEncoding(request) {
			next.ServeHTTP(response, request)
			return
		}
		gz, err := gzip.NewReader(request.Body)
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		request.Body = gz
		next.ServeHTTP(response, request)

	})
}
