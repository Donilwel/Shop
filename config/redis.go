package config

import (
	"Shop/loging"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

var Rdb *redis.Client

func InitRedis() {

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	Rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := Rdb.Ping(ctx).Result(); err != nil {
		loging.Log.Error("Ошибка подключения к Redis")
		return
	}
	loging.Log.Info("Подключение к Redis успешно")
}
