package config

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var RedisCtx = context.Background()

func InitRedis(cfg *Config) *redis.Client {
	if cfg.RedisHost == "" {
		log.Println("Redis 未配置，跳过初始化")
		return nil
	}

	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)

	options := &redis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	RDB = redis.NewClient(options)

	_, err := RDB.Ping(RedisCtx).Result()
	if err != nil {
		log.Printf("警告: Redis 连接失败: %v", err)
		return nil
	}

	log.Printf("Redis 已连接: %s (DB %d)", addr, cfg.RedisDB)
	return RDB
}
