package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func Post(w http.ResponseWriter, r *http.Request, server *Server) {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortId := server.generator.RandString(8)
	err = server.repository.Add(shortId, string(url))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := fmt.Sprintf("http://%s:%d/%s", server.host, server.port, shortId)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result))
}

func Get(w http.ResponseWriter, r *http.Request, server *Server) {
	splitted := strings.Split(r.RequestURI, "/")
	if len(splitted) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortId := splitted[1]
	origin_url, err := server.repository.Get(shortId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(*origin_url))
}

func NewMainHandler(server *Server) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			Post(w, r, server)
		} else if r.Method == http.MethodGet {
			Get(w, r, server)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	})
}
