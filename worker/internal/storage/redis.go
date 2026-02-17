package storage

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

// Redis wraps Redis client
type Redis struct {
    client *redis.Client
}

// NewRedis creates a new Redis client
func NewRedis(ctx context.Context, addr, password string, db int) (*Redis, error) {
    client := redis.NewClient(&redis.Options{
        Addr:         addr,
        Password:     password,
        DB:           db,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolSize:     10,
        MinIdleConns: 5,
    })

    // Ping to verify connection
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("ping redis: %w", err)
    }

    return &Redis{client: client}, nil
}

// Close closes the Redis connection
func (r *Redis) Close() error {
    if err := r.client.Close(); err != nil {
        return fmt.Errorf("closing redis: %w", err)
    }
    return nil
}

// Client returns the underlying Redis client
func (r *Redis) Client() *redis.Client {
    return r.client
}

// AcquireLock acquires a distributed lock with expiry
func (r *Redis) AcquireLock(ctx context.Context, key string, expiry time.Duration) (bool, error) {
    result, err := r.client.SetNX(ctx, key, "locked", expiry).Result()
    if err != nil {
        return false, fmt.Errorf("acquiring lock: %w", err)
    }
    return result, nil
}

// ReleaseLock releases a distributed lock
func (r *Redis) ReleaseLock(ctx context.Context, key string) error {
    if err := r.client.Del(ctx, key).Err(); err != nil {
        return fmt.Errorf("releasing lock: %w", err)
    }
    return nil
}

// Publish publishes a message to a channel
func (r *Redis) Publish(ctx context.Context, channel string, message string) error {
    if err := r.client.Publish(ctx, channel, message).Err(); err != nil {
        return fmt.Errorf("publishing message: %w", err)
    }
    return nil
}
