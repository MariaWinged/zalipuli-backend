package inmemory

import (
	"errors"
	"sync"
	"time"
	"zalipuli/internal/games"
	"zalipuli/internal/storage"
)

type Storage struct {
	levels   sync.Map
	stopChan chan struct{}
}

func New() *Storage {
	st := &Storage{
		levels:   sync.Map{},
		stopChan: make(chan struct{}),
	}
	st.StartCleanup()
	return st
}

func (s *Storage) Save(level games.Level) error {
	s.levels.Store(level.Id(), expireLevel{
		level:    level,
		expireAt: time.Now().Add(storage.LevelLifeTime),
	})

	return nil
}

func (s *Storage) Get(id string) (games.Level, error) {
	el, ok := s.levels.Load(id)
	if !ok {
		return nil, errors.New("level not found")
	}

	expLvl := el.(expireLevel)
	if time.Now().After(expLvl.expireAt) {
		s.levels.Delete(id)
		return nil, errors.New("level expired")
	}

	return expLvl.level, nil
}

func (s *Storage) Delete(id string) error {
	_, ok := s.levels.LoadAndDelete(id)
	if !ok {
		return errors.New("level not found")
	}
	return nil
}

func (s *Storage) cleanup() {
	now := time.Now()
	s.levels.Range(func(key, value interface{}) bool {
		el := value.(expireLevel)
		if now.After(el.expireAt) {
			s.levels.Delete(key)
		}
		return true
	})
}

func (s *Storage) StartCleanup() {
	go func() {
		ticker := time.NewTicker(storage.CleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanup()
			case <-s.stopChan:
				return
			}
		}
	}()
}

func (s *Storage) Close() error {
	close(s.stopChan)
	return nil
}

type expireLevel struct {
	level    games.Level
	expireAt time.Time
}
