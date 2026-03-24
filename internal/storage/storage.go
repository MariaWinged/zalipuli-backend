package storage

import "zalipuli/internal/games"

type Storage interface {
	Save(games.Level) error
	Get(string) (games.Level, error)
	Delete(string) error
}
