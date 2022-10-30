// Package producer provides producing of messages of requested company
package producer

import (
	"context"

	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
)

// Company producer company interface
type Company interface {
	Produce(ctx context.Context, id uuid.UUID, event, name string) error
}

type redisCompany struct {
	redis *redis.Client
}

// NewRedisCompanyProducer creates new producer to price stream
func NewRedisCompanyProducer(redisClient *redis.Client) Company {
	return &redisCompany{
		redis: redisClient,
	}
}

// Produce Push new company record into redis stream
func (r *redisCompany) Produce(ctx context.Context, id uuid.UUID, event, name string) error {
	args := &redis.XAddArgs{
		Stream: "company",
		Values: map[string]interface{}{
			"id":    id.String(),
			"event": event,
			"name":  name,
		},
	}
	return r.redis.XAdd(ctx, args).Err()
}
