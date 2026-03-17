package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"viperai/internal/config"

	redis "github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Initialize() error {
	cfg := config.Get().Cache

	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		Protocol: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return Client.Ping(ctx).Err()
}

func SetCaptcha(email, captcha string) error {
	key := generateCaptchaKey(email)
	return Client.Set(context.Background(), key, captcha, 2*time.Minute).Err()
}

func VerifyCaptcha(email, input string) (bool, error) {
	key := generateCaptchaKey(email)

	stored, err := Client.Get(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	if stored == input {
		Client.Del(context.Background(), key)
		return true, nil
	}

	return false, nil
}

func generateCaptchaKey(email string) string {
	return fmt.Sprintf(config.DefaultCacheKeyConfig.CaptchaPrefix, email)
}

func GenerateIndexName(filename string) string {
	return fmt.Sprintf(config.DefaultCacheKeyConfig.IndexName, filename)
}

func GenerateIndexPrefix(filename string) string {
	return fmt.Sprintf(config.DefaultCacheKeyConfig.IndexNamePrefix, filename)
}

func CreateVectorIndex(ctx context.Context, filename string, dimension int) error {
	indexName := GenerateIndexName(filename)

	_, err := Client.Do(ctx, "FT.INFO", indexName).Result()
	if err == nil {
		return nil
	}

	prefix := GenerateIndexPrefix(filename)

	createArgs := []interface{}{
		"FT.CREATE", indexName,
		"ON", "HASH",
		"PREFIX", "1", prefix,
		"SCHEMA",
		"content", "TEXT",
		"metadata", "TEXT",
		"vector", "VECTOR", "FLAT",
		"6",
		"TYPE", "FLOAT32",
		"DIM", dimension,
		"DISTANCE_METRIC", "COSINE",
	}

	return Client.Do(ctx, createArgs...).Err()
}

func DeleteVectorIndex(ctx context.Context, filename string) error {
	indexName := GenerateIndexName(filename)
	return Client.Do(ctx, "FT.DROPINDEX", indexName).Err()
}
