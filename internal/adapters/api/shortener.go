package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type API struct {
	shortener ports.ShortenerService
	ping      ports.DBCheck
	logger    ports.Logger
	deleter   ports.DeleterService
	secretKey string
}

func NewAPI(
	shortener ports.ShortenerService,
	ping ports.DBCheck,
	logger ports.Logger,
	deleter ports.DeleterService,
	secretKey string,
) *API {
	return &API{
		shortener: shortener,
		ping:      ping,
		logger:    logger,
		deleter:   deleter,
		secretKey: secretKey,
	}
}

func (a *API) Routes() chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(a.loggingMiddleware)
	router.Use(a.gzipHandler)
	router.Use(middleware.Compress(5, "text/html", "application/json"))
	router.Use(a.cookieAuth)

	router.Get("/{shortID:[A-Za-z]+}", a.Get)
	router.Post("/", a.Post)
	router.Post("/api/shorten", a.JSONShorten)
	router.Get("/ping", a.Ping)
	router.Post("/api/shorten/batch", a.Batch)
	router.Get("/api/user/urls", a.GetAllUrls)
	router.Delete("/api/user/urls", a.DeleteUrls)

	return router
}

func (api *API) Post(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	if err != nil || len(url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body.Close()
	uid, ok := r.Context().Value(JWTKey("uid")).(string)
	if !ok {
		uid = ""
	}
	shortURL, err := api.shortener.Add(r.Context(), &domain.Record{
		UID: domain.UID(uid),
		URL: domain.URL(url),
	})
	if err != nil {
		if errors.Is(err, ports.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(string(shortURL)))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(string(shortURL)))
}

func (api *API) Get(w http.ResponseWriter, r *http.Request) {
	shortID := domain.ShortID(chi.URLParam(r, "shortID"))
	url, err := api.shortener.Get(r.Context(), shortID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			w.WriteHeader(http.StatusGone)
			return
		}
		api.logger.Errorln("get url by short id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", string(url))
	w.WriteHeader(http.StatusTemporaryRedirect)
}

type inputJSON struct {
	URL domain.URL `json:"url"`
}

type outJSON struct {
	ShortURL domain.ShortURL `json:"result"`
}

func (api *API) JSONShorten(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		api.logger.Errorln("haven't header application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		api.logger.Errorln("empty body", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body.Close()
	var inJSON inputJSON
	err = json.Unmarshal(body, &inJSON)
	if err != nil {
		api.logger.Error("couldn't unmarshal body", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uid, ok := r.Context().Value(JWTKey("uid")).(string)
	if !ok {
		uid = ""
	}
	shortURL, err := api.shortener.Add(r.Context(), &domain.Record{
		UID: domain.UID(uid),
		URL: inJSON.URL,
	})
	if err != nil {
		if errors.Is(err, ports.ErrAlreadyExists) {
			result := outJSON{
				ShortURL: shortURL,
			}
			body, err = json.Marshal(result)
			if err != nil {
				api.logger.Errorln("couldn't marshal:", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write(body)
			return
		}
		api.logger.Error("service error: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := outJSON{
		ShortURL: shortURL,
	}
	body, err = json.Marshal(result)
	if err != nil {
		api.logger.Errorln("couldn't marshal:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func (api *API) Ping(w http.ResponseWriter, r *http.Request) {
	err := api.ping.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (api *API) Batch(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body.Close()
	var batch []domain.BatchRecord
	err = json.Unmarshal(body, &batch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(batch) != 0 {
		batch, err = api.shortener.Batch(r.Context(), domain.UID(""), batch)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	body, err = json.Marshal(batch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)

}

func (api *API) GetAllUrls(response http.ResponseWriter, request *http.Request) {
	var uid string
	var ok bool
	if uid, ok = request.Context().Value(JWTKey("uid")).(string); !ok || uid == "" {
		api.logger.Error("not uid in context")
		response.WriteHeader(http.StatusUnauthorized)
		return
	}
	result, err := api.shortener.UserURLs(request.Context(), domain.UID(uid))
	if err != nil {
		api.logger.Errorln("service get user urls", "err", err)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(result) == 0 {
		api.logger.Infoln("not urls for user")
		response.WriteHeader(http.StatusNoContent)
		return
	}
	body, err := json.Marshal(result)
	if err != nil {
		api.logger.Error("couldn't marshal", "err", err)
		response.WriteHeader(http.StatusNoContent)
		return
	}
	response.Header().Add("Content-Type", "application/json")
	response.Write(body)
}

func (api *API) DeleteUrls(response http.ResponseWriter, request *http.Request) {
	var uid string
	var ok bool
	if uid, ok = request.Context().Value(JWTKey("uid")).(string); !ok || uid == "" {
		api.logger.Error("not uid in context")
		response.WriteHeader(http.StatusUnauthorized)
		return
	}
	contentType := request.Header.Get("Content-Type")
	if contentType != "application/json" {
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil || len(body) == 0 {
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	defer request.Body.Close()
	var shortIDs []string
	err = json.Unmarshal(body, &shortIDs)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	api.deleter.Process(request.Context(), uid, shortIDs)
	response.WriteHeader(http.StatusAccepted)
}
