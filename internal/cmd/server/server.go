package server

import (
	"log"
	"net/http"
	"os"

	"zalipuli/internal/api"
	"zalipuli/internal/games"
	"zalipuli/internal/games/watersort"
	"zalipuli/internal/storage"
	"zalipuli/internal/storage/inmemory"
	"zalipuli/internal/storage/redis"
	server "zalipuli/pkg/api"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(addr string) *http.Server {
	var store storage.LevelRepository

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr != "" {
		factory := func(st storage.LevelRepository, gameName string) games.Level {
			switch gameName {
			case string(server.Watersort):
				return watersort.EmptyWaterSortLevel(st)
			}

			return nil
		}
		var err error
		store, err = redis.New(redisAddr, factory)
		if err != nil {
			log.Fatalf("failed to connect to redis: %v", err)
		}

		log.Printf("using redis storage at %s", redisAddr)
	} else {
		store = inmemory.New()
		log.Printf("using in-memory storage")
	}

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
