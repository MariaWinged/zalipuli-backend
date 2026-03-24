package inmemory

import (
	"errors"
	"sync"
	"zalipuli/internal/games"
)

type Storage struct {
	levels sync.Map
}

func New() *Storage {
	return &Storage{
		levels: sync.Map{},
	}
}

func (s *Storage) Save(level games.Level) error {
	s.levels.Store(level.Id(), level)
	return nil
}

func (s *Storage) Get(id string) (games.Level, error) {
	level, ok := s.levels.Load(id)
	if !ok {
		return nil, errors.New("level not found")
	}
	return level.(games.Level), nil
}

func (s *Storage) Delete(id string) error {
	_, ok := s.levels.LoadAndDelete(id)
	if !ok {
		return errors.New("level not found")
	}
	return nil
}
