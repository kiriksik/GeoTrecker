package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/kiriksik/GeoTrecker/config"
	"github.com/redis/go-redis/v9"
)

var (
	NilError = redis.Nil
	Ctx      = context.Background()
	RBD      *redis.Client
)

func InitRedis() {
	cfg := config.Cfg

	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	RBD = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	_, err := RBD.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Cant connect to redis: %v", err)
	}

	log.Println("Connected to Redis")
}
