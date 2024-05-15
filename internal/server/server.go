package server

import (
	"context"
	"net"
	"net/http"
)

func NewServer(ctx context.Context, addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:        addr,
		Handler:     handler,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
}
