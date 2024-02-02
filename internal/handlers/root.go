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
	shortID := server.generator.RandString(8)
	err = server.repository.Add(shortID, string(url))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := fmt.Sprintf("http://%s:%d/%s", server.host, server.port, shortID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result))
}

func Get(w http.ResponseWriter, r *http.Request, server *Server) {
	splitted := strings.Split(r.RequestURI, "/")
	if len(splitted) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortID := splitted[1]
	originURL, err := server.repository.Get(shortID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", *originURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
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
