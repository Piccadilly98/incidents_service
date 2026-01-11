package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/redis/go-redis/v9"
)

const (
	KeyQueue = "webhook:queue"
)

type RedisQueue struct {
	client         *redis.Client
	durationForPop time.Duration
}

func NewRedisQueue(client *redis.Client, ctx context.Context, durationForPop int64) (*RedisQueue, error) {
	if durationForPop <= 0 {
		return nil, fmt.Errorf("duration cannot be <= 0")
	}
	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	return &RedisQueue{
		client:         client,
		durationForPop: time.Duration(durationForPop) * time.Second,
	}, nil
}

func (rq *RedisQueue) AddToQueue(read *dto.WebhookTask, ctx context.Context) error {
	b, err := json.Marshal(read)
	if err != nil {
		return err
	}

	return rq.client.RPush(ctx, KeyQueue, b).Err()
}

func (rq *RedisQueue) PopFromQueue(ctx context.Context) (*dto.WebhookTask, bool, error) {
	result, err := rq.client.BRPop(ctx, rq.durationForPop, KeyQueue).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("brpop failed: %s", err.Error())
	}

	if len(result) != 2 {
		return nil, false, fmt.Errorf("failed to pop task from queue, args < 2")
	}
	var task *dto.WebhookTask
	err = json.Unmarshal([]byte(result[1]), &task)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal queue, err: %s", err.Error())
	}

	return task, true, nil
}

func (rq *RedisQueue) PushTask(task *dto.WebhookTask, ctx context.Context) error {
	b, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return rq.client.RPush(ctx, KeyQueue, b).Err()
}

func (rq *RedisQueue) PingWithCtx(ctx context.Context) error {
	return rq.client.Ping(ctx).Err()
}

func (rq *RedisQueue) Name() string {
	return "RedisQueue"
}
