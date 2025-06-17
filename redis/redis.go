package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"gin-web-api/config"

	"github.com/go-redis/redis/v8"
)

var Client *redis.Client
var ctx = context.Background()

func InitRedis(cfg *config.Config) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Redis 连接失败:", err)
	}

	Client = rdb
	log.Println("Redis 连接成功")
}

func GetClient() *redis.Client {
	return Client
}

func Set(key string, value interface{}, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

func Get(key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

func Del(key string) error {
	return Client.Del(ctx, key).Err()
}

func Exists(key string) (bool, error) {
	count, err := Client.Exists(ctx, key).Result()
	return count > 0, err
} 