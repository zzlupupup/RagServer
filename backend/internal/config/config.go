package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerAddr             string
	MySQLDSN               string
	RedisAddr              string
	RedisPassword          string
	RedisDB                int
	FileStorageDir         string
	OpenAIBaseURL          string
	OpenAIAPIKey           string
	EmbeddingModel         string
	EmbeddingDimension     int
	AdminToken             string
	APIKeyEncryptionSecret string
	MCPUploadMaxMB         int
}

func Load() Config {
	return Config{
		ServerAddr:             getenv("SERVER_ADDR", ":8080"),
		MySQLDSN:               getenv("MYSQL_DSN", "rag:ragpass@tcp(127.0.0.1:3306)/ragserver?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:              getenv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:          getenv("REDIS_PASSWORD", ""),
		RedisDB:                getenvInt("REDIS_DB", 0),
		FileStorageDir:         getenv("FILE_STORAGE_DIR", "../storage/files"),
		OpenAIBaseURL:          getenv("OPENAI_BASE_URL", "https://api.openai.com"),
		OpenAIAPIKey:           getenv("OPENAI_API_KEY", ""),
		EmbeddingModel:         getenv("EMBEDDING_MODEL", "text-embedding-3-small"),
		EmbeddingDimension:     getenvInt("EMBEDDING_DIMENSION", 1536),
		AdminToken:             getenv("ADMIN_TOKEN", "dev-admin-token"),
		APIKeyEncryptionSecret: getenv("API_KEY_ENCRYPTION_SECRET", "dev-secret-change-me-32-byte-value"),
		MCPUploadMaxMB:         getenvInt("MCP_UPLOAD_MAX_MB", 20),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
