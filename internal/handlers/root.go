package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Svirex/microurl/internal/pkg/context"
	"github.com/go-chi/chi/v5"
)

func Post(appCtx *context.AppContext) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url, err := io.ReadAll(r.Body)
		if err != nil || len(url) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		shortID := appCtx.Generator.RandString(8)
		err = appCtx.Repository.Add(shortID, string(url))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		result := fmt.Sprintf("http://%s/%s", appCtx.Config.BaseURL, shortID)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(result))
	})
}

func Get(appCtx *context.AppContext) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortID := chi.URLParam(r, "shortID")
		originURL, err := appCtx.Repository.Get(shortID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", *originURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
}
