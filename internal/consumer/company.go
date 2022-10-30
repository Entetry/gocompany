// Package consumer provides consuming of messages of company
package consumer

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Company consuming company messages
type Company interface {
	Consume(ctx context.Context, callbackFunc func(id uuid.UUID, action, name string))
}

type redisCompany struct {
	redis  *redis.Client
	lastID string
}

// NewRedisCompanyConsumer creates redis company consumer object
func NewRedisCompanyConsumer(redisClient *redis.Client, startID string) Company {
	return &redisCompany{
		redis:  redisClient,
		lastID: startID}
}

// Consume get message from redis stream
func (c *redisCompany) Consume(ctx context.Context, callbackFunc func(id uuid.UUID, action, name string)) {
	for {
		args := &redis.XReadArgs{
			Streams: []string{"company", c.lastID},
		}
		r, err := c.redis.XRead(ctx, args).Result()
		if err != nil {
			log.Error(err)
		}

		for _, message := range r[0].Messages {
			id, action, name, decodeErr := decode(message)
			if decodeErr != nil {
				log.Error(err)
			}

			fmt.Printf("consumed message from redis: {%v, %s}\n", id, name)
			callbackFunc(id, action, name)
			c.lastID = message.ID
		}
	}
}

func decode(message redis.XMessage) (id uuid.UUID, action, name string, err error) {
	action, ok := message.Values["event"].(string)
	if !ok {
		return id, action, name, errors.New("cannot convert action to string")
	}
	idStr, ok := message.Values["id"].(string)
	if !ok {
		return id, action, name, errors.New("cannot convert id to string")
	}
	name, ok = message.Values["name"].(string)
	if !ok {
		return id, action, name, errors.New("cannot convert name to string")
	}

	id, err = uuid.Parse(idStr)
	if err != nil {
		return id, action, name, err
	}

	return id, action, name, nil
}
