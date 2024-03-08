package apis

import (
	"encoding/json"
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

func NewShortenerAPI(baseURL string, generator util.Generator, repository repositories.Repository, shortIDSize uint) (*ShortenerAPI, error) {
	shortenerService, err := srv.NewShortenerService(generator, repository, shortIDSize)
	if err != nil {
		return nil, err
	}
	return &ShortenerAPI{
		Service: shortenerService,
		BaseURL: baseURL,
	}, nil
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

func (api *ShortenerAPI) JSONShorten(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var inputJSON models.InputJSON
	err = json.Unmarshal(body, &inputJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	serviceResult, err := api.Service.Add(models.NewServiceAddRecord(inputJSON.URL))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := &models.ResultJSON{
		ShortURL: fmt.Sprintf("%s/%s", api.BaseURL, serviceResult.ShortID),
	}
	body, err = json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)

}

func GetRoutesFunc(api *ShortenerAPI) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/{shortID:[A-Za-z]+}", api.Get)
		r.Post("/", api.Post)
		r.Post("/api/shorten", api.JSONShorten)
	}
}
