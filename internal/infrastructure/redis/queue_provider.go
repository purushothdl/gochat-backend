package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/redis/go-redis/v9"
)

// QueueProvider implements the Queue interface using Redis Lists.
type QueueProvider struct {
	rdb *redis.Client
}

func NewQueueProvider(cfg *config.RedisConfig) (*QueueProvider, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &QueueProvider{rdb: rdb}, nil
}

func (p *QueueProvider) Enqueue(ctx context.Context, queueName string, job interface{}) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	if err := p.rdb.LPush(ctx, queueName, jobData).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}
	return nil
}

func (p *QueueProvider) Dequeue(ctx context.Context, queueName string, result interface{}) error {
	// Use BRPop for blocking dequeue with a timeout of 0 (wait forever).
	jobData, err := p.rdb.BRPop(ctx, 0, queueName).Result()
	if err != nil {
		// This could be redis.Nil if context is cancelled.
		return fmt.Errorf("failed to dequeue job: %w", err)
	}

	// BRPop returns a slice: [queueName, value]
	if len(jobData) < 2 {
		return fmt.Errorf("invalid job data received from queue")
	}

	if err := json.Unmarshal([]byte(jobData[1]), result); err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}
	return nil
}

func (p *QueueProvider) Close() error {
	return p.rdb.Close()
}