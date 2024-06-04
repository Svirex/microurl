package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/Svirex/microurl/internal/core/domain"
	"github.com/Svirex/microurl/internal/core/ports"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// API - структура апи.
type API struct {
	shortener ports.ShortenerService
	ping      ports.DBCheck
	logger    ports.Logger
	deleter   ports.DeleterService
	secretKey string
}

// NewAPI - создание нового апи.
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

// NewServer - создание нового сервера.
func NewServer(ctx context.Context, addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:        addr,
		Handler:     handler,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
}

// Routes - создание роутов.
func (api *API) Routes() chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(api.loggingMiddleware)
	router.Use(api.gzipHandler)
	router.Use(middleware.Compress(5, "text/html", "application/json"))
	router.Use(api.cookieAuth)

	router.Get("/{shortID:[A-Za-z]+}", api.GetUrl)
	router.Post("/", api.PostAddUrl)
	router.Get("/ping", api.GetPingDB)
	router.Route("/api", func(router chi.Router) {
		router.Post("/shorten", api.JSONShorten)
		router.Post("/shorten/batch", api.PostAddBatch)
		router.Get("/user/urls", api.GetAllUrls)
		router.Delete("/user/urls", api.DeleteUrls)
	})

	return router
}

// PostAddUrl - обработка запроса на добавление записи.
func (api *API) PostAddUrl(w http.ResponseWriter, r *http.Request) {
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

// GetUrl - обработка запроса на получение урла.
func (api *API) GetUrl(w http.ResponseWriter, r *http.Request) {
	shortID := domain.ShortID(chi.URLParam(r, "shortID"))
	url, err := api.shortener.Get(r.Context(), shortID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			w.WriteHeader(http.StatusGone)
			return
		}
		api.logger.Errorln("get url by short id: ", err)
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

// JSONShorten - добавление записи из json запроса.
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
			api.marshalAndSendJSON(result, http.StatusConflict, w)
			return
		}
		api.logger.Error("service error: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := outJSON{
		ShortURL: shortURL,
	}
	api.marshalAndSendJSON(result, http.StatusCreated, w)
}

// GetPingDB - проверка работоспособности подключения к БД.
func (api *API) GetPingDB(w http.ResponseWriter, r *http.Request) {
	err := api.ping.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// PostAddBatch - добавление нескольких записей.
func (api *API) PostAddBatch(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		api.logger.Errorf("api, batch, Content-Type not json: %s", contentType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		api.logger.Errorf("api, batch, read body: %w", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body.Close()
	var batch []domain.BatchRecord
	err = json.Unmarshal(body, &batch)
	if err != nil {
		api.logger.Errorf("api, batch, unmarshal: %w", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(batch) != 0 {
		batch, err = api.shortener.Batch(r.Context(), domain.UID(""), batch)
		if err != nil {
			api.logger.Errorf("api, batch, service error: %w, result: %v", err, batch)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	api.marshalAndSendJSON(batch, http.StatusCreated, w)
}

// GetAllUrls - получение всех записей для пользователя.
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
	api.marshalAndSendJSON(result, http.StatusOK, response)
}

// DeleteUrls - обработка запроса для удаления записей
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
		api.logger.Errorf("api, delete, Content-Type not json: %s", contentType)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil || len(body) == 0 {
		api.logger.Errorf("api, delete, read bode: %w", err)
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

func (api *API) marshalAndSendJSON(data any, okStatus int, w http.ResponseWriter) {
	body, err := json.Marshal(data)
	if err != nil {
		api.logger.Errorf("api, marshal and send json: %w", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(okStatus)
	w.Write(body)
}
