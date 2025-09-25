package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	appContext "github.com/alphabatem/common/context"
	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	appContext.DefaultService
	redis *redis.Client
}

const REDIS_SVC = "redis_svc"

func (svc RedisService) Id() string {
	return REDIS_SVC
}

func (svc *RedisService) Configure(ctx *appContext.Context) error {
	svc.initRedisClient()
	return svc.DefaultService.Configure(ctx)
}

func (svc *RedisService) Start() error {
	if svc.redis != nil {
		ctx := context.Background()
		_, err := svc.redis.Ping(ctx).Result()
		if err != nil {
			return fmt.Errorf("failed to connect to Redis: %w", err)
		}
	}
	return nil
}

func (svc *RedisService) initRedisClient() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	svc.redis = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
}

func (svc *RedisService) GetClient() *redis.Client {
	return svc.redis
}

func (svc *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if svc.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	var data []byte
	var err error

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	return svc.redis.Set(ctx, key, data, expiration).Err()
}

func (svc *RedisService) Get(ctx context.Context, key string) (string, error) {
	if svc.redis == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	result, err := svc.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return result, err
}

func (svc *RedisService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if svc.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	result, err := svc.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(result), dest)
}

func (svc *RedisService) Delete(ctx context.Context, keys ...string) error {
	if svc.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return svc.redis.Del(ctx, keys...).Err()
}

func (svc *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	if svc.redis == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	result, err := svc.redis.Exists(ctx, key).Result()
	return result > 0, err
}

func (svc *RedisService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if svc.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return svc.redis.Expire(ctx, key, expiration).Err()
}

func (svc *RedisService) TTL(ctx context.Context, key string) (time.Duration, error) {
	if svc.redis == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	return svc.redis.TTL(ctx, key).Result()
}

func (svc *RedisService) Increment(ctx context.Context, key string) (int64, error) {
	if svc.redis == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	return svc.redis.Incr(ctx, key).Result()
}

func (svc *RedisService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	if svc.redis == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	return svc.redis.IncrBy(ctx, key, value).Result()
}

func (svc *RedisService) Decrement(ctx context.Context, key string) (int64, error) {
	if svc.redis == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	return svc.redis.Decr(ctx, key).Result()
}

func (svc *RedisService) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	if svc.redis == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	return svc.redis.DecrBy(ctx, key, value).Result()
}

func (svc *RedisService) HSet(ctx context.Context, key string, values ...interface{}) error {
	if svc.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return svc.redis.HSet(ctx, key, values...).Err()
}

func (svc *RedisService) HGet(ctx context.Context, key, field string) (string, error) {
	if svc.redis == nil {
		return "", fmt.Errorf("redis client not initialized")
	}

	result, err := svc.redis.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil
	}
	return result, err
}

func (svc *RedisService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if svc.redis == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	return svc.redis.HGetAll(ctx, key).Result()
}

func (svc *RedisService) HDel(ctx context.Context, key string, fields ...string) error {
	if svc.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return svc.redis.HDel(ctx, key, fields...).Err()
}

func (svc *RedisService) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if svc.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return svc.redis.SAdd(ctx, key, members...).Err()
}

func (svc *RedisService) SMembers(ctx context.Context, key string) ([]string, error) {
	if svc.redis == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	return svc.redis.SMembers(ctx, key).Result()
}

func (svc *RedisService) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	if svc.redis == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	return svc.redis.SIsMember(ctx, key, member).Result()
}

func (svc *RedisService) SRem(ctx context.Context, key string, members ...interface{}) error {
	if svc.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return svc.redis.SRem(ctx, key, members...).Err()
}

func (svc *RedisService) FlushAll(ctx context.Context) error {
	if svc.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}

	return svc.redis.FlushAll(ctx).Err()
}

func (svc *RedisService) Keys(ctx context.Context, pattern string) ([]string, error) {
	if svc.redis == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	return svc.redis.Keys(ctx, pattern).Result()
}
