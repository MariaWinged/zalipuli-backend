package storage

import (
	"errors"
	"time"
	"zalipuli/internal/games"
)

const (
	LevelLifeTime    = time.Hour * 3
	PositionLifeTime = time.Hour * 24 * 7
	CleanupInterval  = time.Minute * 5
)

var ErrNotFound = errors.New("not found")

type LevelRepository interface {
	SaveLevel(games.Level) error
	GetLevel(string) (games.Level, error)
	DeleteLevel(string) error
}

type PositionRepository interface {
	SavePosition(gameName string, hash string, position any) error
	GetPosition(gameName string, hash string, position any) error
	DeletePosition(gameName string, hash string) error
}
