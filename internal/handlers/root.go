package handlers

import (
	"fmt"
	"io"
	"net/http"
)

func (s *Server) RootPost(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

}

func NewRootPost(server *Server) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		url, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		shortId := server.generator.RandString(10)
		err = server.repository.Add(shortId, string(url))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		result := fmt.Sprintf("http://%s:%d/%s", server.host, server.port, shortId)

		w.Write([]byte(result))
	})
}
