package utils

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

func GetOrSetCache[T any](ctx context.Context, rdb *redis.Client, db *gorm.DB, cacheKey string, query *gorm.DB, dest *[]T, ttl time.Duration) (bool, error) {
	cachedData, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(cachedData), dest); err == nil {
			logrus.Info("Данные достаны с помощью кэша: ", cacheKey)
			return true, nil
		}
	}

	if err := query.Find(dest).Error; err != nil {
		return false, err
	}

	if len(*dest) > 0 {
		jsonData, _ := json.Marshal(dest)
		_ = rdb.Set(ctx, cacheKey, jsonData, ttl).Err()
		logrus.Info("Данные достаны из базы данных и портированы в кэш на 5 минут: ", cacheKey)
	}

	return false, nil
}
