package server

import (
	"log"
	"net/http"
	"os"
	"zalipuli/internal/games"
	"zalipuli/internal/games/watersort"
	"zalipuli/internal/storage"
	"zalipuli/internal/storage/inmemory"
	"zalipuli/internal/storage/redis"
	server "zalipuli/pkg/api"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"zalipuli/internal/api"
)

func New(addr string) *http.Server {
	var store storage.Storage

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr != "" {
		redisStore, err := redis.New(redisAddr)
		if err != nil {
			log.Fatalf("failed to connect to redis: %v", err)
		}
		factory := func(gameName string) games.Level {
			switch gameName {
			case string(server.Watersort):
				return watersort.EmptyWaterSortLevel(redisStore)
			}

			return nil
		}
		redisStore.SetFactory(factory)

		store = redisStore
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
