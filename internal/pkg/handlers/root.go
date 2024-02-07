package handlers

import (
	"github.com/go-chi/chi/v5"

	"github.com/Svirex/microurl/internal/handlers"
	"github.com/Svirex/microurl/internal/pkg/context"
)

func GetRoutesFunc(appCtx *context.AppContext) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/{shortID:[A-Za-z]+}", handlers.Get(appCtx))
		r.Post("/", handlers.Post(appCtx))
	}
}
