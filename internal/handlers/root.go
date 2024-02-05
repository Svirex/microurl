package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Svirex/microurl/internal/pkg/context"
)

func Post(w http.ResponseWriter, r *http.Request, appCtx *context.AppContext) {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortID := appCtx.Generator.RandString(8)
	err = appCtx.Repository.Add(shortID, string(url))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := fmt.Sprintf("http://%s:%d/%s", appCtx.Config.Host, appCtx.Config.Port, shortID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result))
}

func Get(w http.ResponseWriter, r *http.Request, appCtx *context.AppContext) {
	splitted := strings.Split(r.RequestURI, "/")
	if len(splitted) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortID := splitted[1]
	originURL, err := appCtx.Repository.Get(shortID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", *originURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func NewMainHandler(appCtx *context.AppContext) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			Post(w, r, appCtx)
		} else if r.Method == http.MethodGet {
			Get(w, r, appCtx)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	})
}
