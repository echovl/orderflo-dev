package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/layerhub-io/api/db"
)

type Config struct {
	Addr         string
	Username     string
	Password     string
	DB           int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type RedisDB struct {
	c *redis.Client
}

var _ db.KeyValueDB = (*RedisDB)(nil)

func New(conf *Config) (db.KeyValueDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Username:     conf.Username,
		Password:     conf.Password,
		DB:           conf.DB,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		IdleTimeout:  conf.IdleTimeout,
	})

	err := client.Ping(context.TODO()).Err()
	if err != nil {
		return nil, err
	}

	return &RedisDB{client}, nil
}

func (r *RedisDB) Get(ctx context.Context, key string) ([]byte, error) {
	cmd := r.c.Get(ctx, key)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Bytes()
}

func (r *RedisDB) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	cmd := r.c.Set(ctx, key, val, expiration)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (r *RedisDB) Del(ctx context.Context, keys ...string) error {
	return r.c.Del(ctx, keys...).Err()
}

func (r *RedisDB) Close(ctx context.Context) error {
	return r.c.Close()
}
