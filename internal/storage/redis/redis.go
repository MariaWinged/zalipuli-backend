package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"zalipuli/internal/games"
	"zalipuli/internal/storage"

	"github.com/redis/go-redis/v9"
)

type LevelFactory func(string) games.Level

type Storage struct {
	client  *redis.Client
	factory LevelFactory
	ctx     context.Context
}

func New(addr string, factory LevelFactory) (*Storage, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Storage{
		client:  client,
		factory: factory,
		ctx:     ctx,
	}, nil
}

func (s *Storage) SavePosition(gameName string, hash string, position any) error {
	data, err := json.Marshal(position)
	if err != nil {
		return fmt.Errorf("failed to marshal position: %w", err)
	}

	key := fmt.Sprintf("%s:%s", gameName, hash)

	err = s.client.Set(s.ctx, key, data, storage.PositionLifeTime).Err()
	if err != nil {
		return fmt.Errorf("failed to save position to redis: %w", err)
	}

	return nil
}

func (s *Storage) GetPosition(gameName string, hash string, position any) error {
	key := fmt.Sprintf("%s:%s", gameName, hash)

	data, err := s.client.Get(s.ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return storage.ErrNotFound
		}

		return fmt.Errorf("failed to get position from redis: %w", err)
	}

	err = json.Unmarshal(data, position)

	return err
}

func (s *Storage) DeletePosition(gameName string, hash string) error {
	key := fmt.Sprintf("%s:%s", gameName, hash)
	deleted, err := s.client.Del(s.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete position from redis: %w", err)
	}

	if deleted == 0 {
		return storage.ErrNotFound
	}

	return nil
}

type levelRecord struct {
	GameName string          `json:"game_name"`
	Data     json.RawMessage `json:"data"`
}

func (s *Storage) SaveLevel(level games.Level) error {
	data, err := json.Marshal(level)
	if err != nil {
		return fmt.Errorf("failed to marshal level: %w", err)
	}

	record := levelRecord{
		GameName: string(level.GameName()),
		Data:     data,
	}

	recordData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal level record: %w", err)
	}

	err = s.client.Set(s.ctx, level.Id(), recordData, storage.LevelLifeTime).Err()
	if err != nil {
		return fmt.Errorf("failed to save level to redis: %w", err)
	}

	return nil
}

func (s *Storage) GetLevel(id string) (games.Level, error) {
	recordData, err := s.client.Get(s.ctx, id).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get level from redis: %w", err)
	}

	var record levelRecord
	if err := json.Unmarshal(recordData, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal level record: %w", err)
	}

	level := s.factory(record.GameName)
	if level == nil {
		return nil, errors.New("failed to create level from factory")
	}

	err = json.Unmarshal(record.Data, level)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal level: %w", err)
	}

	return level, nil
}

func (s *Storage) DeleteLevel(id string) error {
	deleted, err := s.client.Del(s.ctx, id).Result()
	if err != nil {
		return fmt.Errorf("failed to delete level from redis: %w", err)
	}

	if deleted == 0 {
		return storage.ErrNotFound
	}

	return nil
}
