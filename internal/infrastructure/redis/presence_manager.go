package redis

import (
	"context"
	"fmt"

	"github.com/purushothdl/gochat-backend/internal/config"
	"github.com/redis/go-redis/v9"
)

const presenceKeyPrefix = "presence:room:"

// PresenceManager implements the PresenceManager contract using Redis Sets.
type PresenceManager struct {
	rdb *redis.Client
}

func NewPresenceManager(cfg *config.RedisConfig) (*PresenceManager, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis for presence: %w", err)
	}

	return &PresenceManager{rdb: rdb}, nil
}

// keyForRoom generates the specific Redis key for a given room.
func (p *PresenceManager) keyForRoom(roomID string) string {
	return fmt.Sprintf("%s%s", presenceKeyPrefix, roomID)
}

func (p *PresenceManager) AddToRoom(ctx context.Context, roomID string, userID string) error {
	key := p.keyForRoom(roomID)
	return p.rdb.SAdd(ctx, key, userID).Err()
}

func (p *PresenceManager) RemoveFromRoom(ctx context.Context, roomID string, userID string) error {
	key := p.keyForRoom(roomID)
	return p.rdb.SRem(ctx, key, userID).Err()
}

func (p *PresenceManager) GetOnlineUserIDs(ctx context.Context, roomID string) ([]string, error) {
	key := p.keyForRoom(roomID)
	return p.rdb.SMembers(ctx, key).Result()
}