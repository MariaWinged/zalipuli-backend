package inmemory

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"zalipuli/internal/games"
	"zalipuli/internal/storage"
)

type Storage struct {
	levels    sync.Map
	positions sync.Map
	stopChan  chan struct{}
}

func New() *Storage {
	st := &Storage{
		levels:    sync.Map{},
		positions: sync.Map{},
		stopChan:  make(chan struct{}),
	}
	st.StartCleanup()

	return st
}

func (s *Storage) SavePosition(gameName string, hash string, position any) error {
	key := fmt.Sprintf("%s:%s", gameName, hash)
	s.positions.Store(key, position)

	return nil
}

func (s *Storage) GetPosition(gameName string, hash string, position any) error {
	key := fmt.Sprintf("%s:%s", gameName, hash)
	val, ok := s.positions.Load(key)
	if !ok {
		return storage.ErrNotFound
	}

	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal position: %w", err)
	}

	err = json.Unmarshal(data, position)
	if err != nil {
		return fmt.Errorf("failed to unmarshal position: %w", err)
	}

	return nil
}

func (s *Storage) DeletePosition(gameName string, hash string) error {
	key := fmt.Sprintf("%s:%s", gameName, hash)
	_, ok := s.positions.LoadAndDelete(key)
	if !ok {
		return storage.ErrNotFound
	}

	return nil
}

func (s *Storage) SaveLevel(level games.Level) error {
	s.levels.Store(level.Id(), expireLevel{
		level:    level,
		expireAt: time.Now().Add(storage.LevelLifeTime),
	})

	return nil
}

func (s *Storage) GetLevel(id string) (games.Level, error) {
	el, ok := s.levels.Load(id)
	if !ok {
		return nil, storage.ErrNotFound
	}

	expLvl := el.(expireLevel)
	if time.Now().After(expLvl.expireAt) {
		s.levels.Delete(id)
		return nil, storage.ErrNotFound
	}

	return expLvl.level, nil
}

func (s *Storage) DeleteLevel(id string) error {
	_, ok := s.levels.LoadAndDelete(id)
	if !ok {
		return storage.ErrNotFound
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
