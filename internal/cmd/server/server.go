package server

import (
	"log"
	"net/http"
	"os"
	ws "zalipuli/internal/games/ws_refactoring"

	"zalipuli/internal/api"
	"zalipuli/internal/games"
	"zalipuli/internal/storage"
	"zalipuli/internal/storage/inmemory"
	"zalipuli/internal/storage/redis"
	server "zalipuli/pkg/api"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(addr string) *http.Server {
	var levelRepo storage.LevelRepository
	var posRepo storage.PositionRepository

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr != "" {
		factory := func(gameName string) games.Level {
			switch gameName {
			case string(server.Watersort):
				return &ws.Level{}
			}

			return nil
		}
		var err error
		redisRepo, err := redis.New(redisAddr, factory)
		if err != nil {
			log.Fatalf("failed to connect to redis: %v", err)
		}

		levelRepo = redisRepo
		posRepo = redisRepo

		log.Printf("using redis storage at %s", redisAddr)
	} else {
		inmemoryRepo := inmemory.New()
		posRepo = inmemoryRepo
		levelRepo = inmemoryRepo
		log.Printf("using in-memory storage")
	}

	handler := api.NewApi(levelRepo, posRepo)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	server.HandlerFromMux(handler, r)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
