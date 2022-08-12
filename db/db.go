package db

import (
	"context"
	"time"
)

type KeyValueDB interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val any, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Close(ctx context.Context) error
}
