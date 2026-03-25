package storage

import (
	"time"
	"zalipuli/internal/games"
)

const (
	LevelLifeTime   = time.Hour * 3
	CleanupInterval = time.Minute * 5
)

type Storage interface {
	Save(games.Level) error
	Get(string) (games.Level, error)
	Delete(string) error
	Close() error
}
