package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Piccadilly98/incidents_service/internal/config"
	"github.com/Piccadilly98/incidents_service/internal/models/entities"
	"github.com/redis/go-redis/v9"
)

const (
	ActiveIncidentPrefix = "incident:active:"
	WebhookInQueuePrefix = "incident:in_queue"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(cfg *config.Config, ctx context.Context) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return &RedisCache{client: client}, nil
}

func (rc *RedisCache) SetActiveIncident(ctx context.Context, data *entities.ReadIncident) error {
	key := ActiveIncidentPrefix + data.Id

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return rc.client.Set(ctx, key, b, -1).Err()
}

func (rc *RedisCache) GetActiveIncident(ctx context.Context, id string) (*entities.ReadIncident, error) {
	key := ActiveIncidentPrefix + id

	data, err := rc.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis GET failed: %w", err)
	}

	var inc *entities.ReadIncident
	if err := json.Unmarshal(data, &inc); err != nil {
		return nil, fmt.Errorf("unmarshal incident failed: %w", err)
	}

	return inc, nil
}

func (rc *RedisCache) DeleteActiveIncident(ctx context.Context, id string) error {
	key := ActiveIncidentPrefix + id

	err := rc.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis DEL failed: %w", err)
	}

	return nil
}
