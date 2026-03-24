package server

import (
	"net/http"
	server "zalipuli/pkg/api"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"zalipuli/internal/api"
	"zalipuli/internal/storage"
)

func New(addr string) *http.Server {
	store := storage.New()
	handler := api.NewApi(store)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	server.HandlerFromMux(handler, r)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
