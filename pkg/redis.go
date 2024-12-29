package KVStore

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Redis struct {
	client *redis.Client
}

func NewRedis(addr string, password string, db int) KVStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	return Redis{client: rdb}
}

func (r Redis) Set(key string, value interface{}) error {
	err := r.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r Redis) Get(key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r Redis) Delete(key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}
func (r Redis) LPush(key string, values ...interface{}) error {
	err := r.client.LPush(ctx, key, values...).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r Redis) RPush(key string, values ...interface{}) error {
	err := r.client.RPush(ctx, key, values...).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r Redis) LPop(key string) (string, error) {
	val, err := r.client.LPop(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r Redis) RPop(key string) (string, error) {
	val, err := r.client.RPop(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r Redis) LLen(key string) (int64, error) {
	val, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}
func (r Redis) LIndex(key string, index int64) (string, error) {
	val, err := r.client.LIndex(ctx, key, index).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r Redis) LRange(key string, start, stop int64) ([]string, error) {
	val, err := r.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (r Redis) INCR(key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (r Redis) DECR(key string) (int64, error) {
	val, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}
