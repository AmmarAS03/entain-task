// Package repository contains data storage methods
package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type redisRepo struct {
	client *redis.Client
}

// NewRedisRepository creates a new instance of a Repository using Redis as the persistance layer
func NewRedisRepository(ctx context.Context, address string, password string) (Repository, error) {
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0, // default DB
	})

	// Check connection
	rslt := &redisRepo{client: rdb}
	rslt.HealthCheck(ctx)
	if !rslt.HealthCheck(ctx) {
		return nil, fmt.Errorf("failed_to_init_redis")
	}

	return &redisRepo{client: rdb}, nil
}

func (c *redisRepo) HealthCheck(ctx context.Context) bool {
	if err := c.client.Ping(ctx).Err(); err != nil {
		logrus.Errorf("could not connect to Redis: %v", err)
		return false
	}

	return true
}

func (c *redisRepo) UpdateEvent(ctx context.Context, event *model.Event) error {
	data, mErr := json.Marshal(event)
	if mErr != nil {
		logrus.Errorf("could not marshall event %v", mErr)
		return mErr
	}
	rslt := c.client.Set(ctx, event.ID, data, 0)
	if rslt.Err() != nil {
		logrus.Errorf("could not update event %v", rslt.Err())
		return rslt.Err()
	}

	return nil
}

func (c *redisRepo) GetEventByID(ctx context.Context, id string) (*model.Event, error) {
	rslt, err := c.client.Get(ctx, id).Result()
	if err == redis.Nil {
		logrus.Infof("Event not found")
		return nil, nil
	} else if err != nil {
		logrus.Errorf("could not get event %v", err)
		return nil, err
	}

	event := &model.Event{}
	if err := json.Unmarshal([]byte(rslt), event); err != nil {
		logrus.Errorf("failed to unmarshal event %v", err)
		return nil, err
	}

	return event, nil
}

func (c *redisRepo) DeleteEventByID(ctx context.Context, id string) error {
	_, err := c.client.Del(ctx, id).Result()
	if err != nil {
		logrus.Errorf("could not delete event %v", err)
	}

	return err
}
