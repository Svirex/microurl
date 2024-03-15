package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Svirex/microurl/internal/pkg/logging"
	appmiddleware "github.com/Svirex/microurl/internal/pkg/middleware"
	"github.com/Svirex/microurl/internal/pkg/models"
	"github.com/Svirex/microurl/internal/pkg/services"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type ShortenerAPI struct {
	shortenerService services.Shortener
	BaseURL          string
	pingService      services.DBCheck
}

func NewShortenerAPI(service services.Shortener, dbCheckService services.DBCheck, baseURL string) *ShortenerAPI {
	return &ShortenerAPI{
		shortenerService: service,
		BaseURL:          baseURL,
		pingService:      dbCheckService,
	}
}

func (api *ShortenerAPI) Post(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil || len(url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	serviceResult, err := api.shortenerService.Add(r.Context(), models.NewServiceAddRecord(string(url)))
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
	serviceResult, err := api.shortenerService.Get(r.Context(), models.NewServiceGetRecord(shortID))
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
	serviceResult, err := api.shortenerService.Add(r.Context(), models.NewServiceAddRecord(inputJSON.URL))
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

func (api *ShortenerAPI) Ping(w http.ResponseWriter, r *http.Request) {
	err := api.pingService.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (api *ShortenerAPI) Routes(logger logging.Logger) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(appmiddleware.NewLoggingMiddleware(logger))
	router.Use(appmiddleware.GzipHandler)
	router.Use(middleware.Compress(5, "text/html", "application/json"))

	router.Get("/{shortID:[A-Za-z]+}", api.Get)
	router.Post("/", api.Post)
	router.Post("/api/shorten", api.JSONShorten)
	router.Get("/ping", api.Ping)

	return router
}
