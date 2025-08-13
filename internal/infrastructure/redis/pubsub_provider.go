package redis

import (
	"context"
	"fmt"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/purushothdl/gochat-backend/internal/contracts"
	"github.com/redis/go-redis/v9"
)

// PubSubProvider implements the PubSub interface for Redis.
type PubSubProvider struct {
	rdb *redis.Client
}

func NewPubSubProvider(cfg *config.RedisConfig) (*PubSubProvider, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &PubSubProvider{rdb: rdb}, nil
}

func (p *PubSubProvider) Publish(ctx context.Context, channel string, message string) error {
	return p.rdb.Publish(ctx, channel, message).Err()
}

func (p *PubSubProvider) Subscribe(ctx context.Context, channels ...string) (<-chan *contracts.Message, error) {
	pubsub := p.rdb.Subscribe(ctx, channels...)

	// Wait for subscription confirmation.
	if _, err := pubsub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("failed to subscribe to channels: %w", err)
	}

	appChan := make(chan *contracts.Message)

	// Goroutine to bridge Redis channel to application channel.
	go func() {
		defer close(appChan)
		defer pubsub.Close()

		redisChan := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-redisChan:
				if !ok {
					return
				}
				appChan <- &contracts.Message{
					Channel: msg.Channel,
					Payload: msg.Payload,
				}
			}
		}
	}()

	return appChan, nil
}