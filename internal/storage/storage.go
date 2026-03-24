package storage

import (
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

func (s *Storage) Save(level games.Level) {
	s.levels.Store(level.Id(), level)
}

func (s *Storage) Get(id string) (games.Level, bool) {
	level, ok := s.levels.Load(id)
	if !ok {
		return nil, false
	}
	return level.(games.Level), true
}

func (s *Storage) Delete(id string) bool {
	_, ok := s.levels.LoadAndDelete(id)
	return ok
}
