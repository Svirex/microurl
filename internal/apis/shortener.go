package apis

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/repositories"
	"github.com/Svirex/microurl/internal/pkg/services"
	"github.com/Svirex/microurl/internal/pkg/util"
	srv "github.com/Svirex/microurl/internal/services"
	"github.com/go-chi/chi/v5"
)

type ShortenerAPI struct {
	Service services.Shortener
	BaseURL string
}

func NewShortenerAPI(generator util.Generator, repository repositories.Repository, baseURL string, shortIDSize uint) *ShortenerAPI {
	return &ShortenerAPI{
		Service: srv.NewShortenerService(generator, repository, shortIDSize),
		BaseURL: baseURL,
	}
}

func (api *ShortenerAPI) Post(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil || len(url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	serviceResult, err := api.Service.Add(models.NewServiceAddRecord(string(url)))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := fmt.Sprintf("%s/%s", api.BaseURL, serviceResult.ShortID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result))
}

func (api *ShortenerAPI) Get(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "shortID")
	serviceResult, err := api.Service.Get(models.NewServiceGetRecord(shortID))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", serviceResult.URL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func GetRoutesFunc(api *ShortenerAPI) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/{shortID:[A-Za-z]+}", api.Get)
		r.Post("/", api.Post)
	}
}
