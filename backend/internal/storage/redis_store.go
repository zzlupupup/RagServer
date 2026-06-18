package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
)

const (
	RedisIndexName = "idx:rag_chunks"
	RedisKeyPrefix = "rag:chunk:"
)

func OpenRedis(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:          addr,
		Password:      password,
		DB:            db,
		Protocol:      2,
		UnstableResp3: true,
	})
}

func EnsureVectorIndex(ctx context.Context, client *redis.Client, dimension int) error {
	if dimension <= 0 {
		return errors.New("embedding dimension must be positive")
	}
	_, err := client.FTInfo(ctx, RedisIndexName).Result()
	if err == nil {
		return nil
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unknown index") {
		return err
	}
	args := []any{
		"FT.CREATE", RedisIndexName,
		"ON", "HASH",
		"PREFIX", "1", RedisKeyPrefix,
		"SCHEMA",
		"kb_id", "TAG",
		"document_id", "TAG",
		"chunk_id", "TAG",
		"chunk_index", "NUMERIC",
		"filename", "TEXT",
		"content", "TEXT",
		"vector_content", "VECTOR", "FLAT", "6",
		"TYPE", "FLOAT32",
		"DIM", dimension,
		"DISTANCE_METRIC", "COSINE",
	}
	return client.Do(ctx, args...).Err()
}
