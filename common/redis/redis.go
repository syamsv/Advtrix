package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/syamsv/go-template/config"
)

var client *redis.Client
var log *zap.Logger

func Init() {
	log = zap.L().Named("redis")

	client = redis.NewClient(&redis.Options{
		Addr:     config.REDIS_ADDRESS,
		Password: config.REDIS_PASSWORD,
		DB:       config.REDIS_DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal("failed to connect", zap.Error(err))
	}

	log.Info("connected", zap.String("address", config.REDIS_ADDRESS), zap.Int("db", config.REDIS_DB))
}

func Close() {
	if err := client.Close(); err != nil {
		log.Error("error closing", zap.Error(err))
	}
	log.Info("disconnected")
}

func Client() *redis.Client {
	return client
}

func Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return client.Set(ctx, key, value, ttl).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

func Delete(ctx context.Context, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

func Exists(ctx context.Context, key string) (bool, error) {
	n, err := client.Exists(ctx, key).Result()
	return n > 0, err
}

func SetJSON(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return client.Set(ctx, key, value, ttl).Err()
}

func GetJSON(ctx context.Context, key string) ([]byte, error) {
	return client.Get(ctx, key).Bytes()
}

func HSet(ctx context.Context, key string, values ...any) error {
	return client.HSet(ctx, key, values...).Err()
}

func HGet(ctx context.Context, key, field string) (string, error) {
	return client.HGet(ctx, key, field).Result()
}

func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return client.HGetAll(ctx, key).Result()
}

func HDel(ctx context.Context, key string, fields ...string) error {
	return client.HDel(ctx, key, fields...).Err()
}

func Expire(ctx context.Context, key string, ttl time.Duration) error {
	return client.Expire(ctx, key, ttl).Err()
}

func Keys(ctx context.Context, pattern string) ([]string, error) {
	return client.Keys(ctx, pattern).Result()
}
