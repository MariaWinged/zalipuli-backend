package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"zalipuli/internal/games"

	"github.com/redis/go-redis/v9"
)

type LevelFactory func(gameName string) games.Level

type Storage struct {
	client  *redis.Client
	factory LevelFactory
	ctx     context.Context
}

type levelRecord struct {
	GameName string          `json:"game_name"`
	Data     json.RawMessage `json:"data"`
}

func New(addr string) (*Storage, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Storage{
		client: client,
		ctx:    ctx,
	}, nil
}

func (s *Storage) SetFactory(factory LevelFactory) {
	s.factory = factory
}

func (s *Storage) Save(level games.Level) error {
	data, err := level.ToJson()
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

	err = s.client.Set(s.ctx, level.Id(), recordData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to save level to redis: %w", err)
	}

	return nil
}

func (s *Storage) Get(id string) (games.Level, error) {
	recordData, err := s.client.Get(s.ctx, id).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("level not found")
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

	if err := level.FromJson(record.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal level data: %w", err)
	}

	return level, nil
}

func (s *Storage) Delete(id string) error {
	deleted, err := s.client.Del(s.ctx, id).Result()
	if err != nil {
		return fmt.Errorf("failed to delete level from redis: %w", err)
	}

	if deleted == 0 {
		return errors.New("level not found")
	}

	return nil
}
