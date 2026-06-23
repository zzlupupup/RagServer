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
	EmbeddingProvider      string
	OpenAIBaseURL          string
	OpenAIAPIKey           string
	ArkBaseURL             string
	ArkAPIKey              string
	EmbeddingModel         string
	EmbeddingDimension     int
	IndexTimeoutSeconds    int
	JWTSecret              string
	JWTExpiresHours        int
	APIKeyEncryptionSecret string
	MCPUploadMaxMB         int
	MCPTmpDir              string
}

func Load() Config {
	return Config{
		ServerAddr:             getenv("SERVER_ADDR", ":8080"),
		MySQLDSN:               getenv("MYSQL_DSN", "rag:ragpass@tcp(mysql:3306)/ragserver?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:              getenv("REDIS_ADDR", "redis-stack:6379"),
		RedisPassword:          getenv("REDIS_PASSWORD", ""),
		RedisDB:                getenvInt("REDIS_DB", 0),
		FileStorageDir:         getenv("FILE_STORAGE_DIR", "../storage/files"),
		EmbeddingProvider:      getenv("EMBEDDING_PROVIDER", "ark"),
		OpenAIBaseURL:          getenv("OPENAI_BASE_URL", "https://api.openai.com"),
		OpenAIAPIKey:           getenv("OPENAI_API_KEY", ""),
		ArkBaseURL:             getenv("ARK_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"),
		ArkAPIKey:              getenv("ARK_API_KEY", ""),
		EmbeddingModel:         getenv("EMBEDDING_MODEL", ""),
		EmbeddingDimension:     getenvInt("EMBEDDING_DIMENSION", 0),
		IndexTimeoutSeconds:    getenvInt("INDEX_TIMEOUT_SECONDS", 300),
		JWTSecret:              getenv("JWT_SECRET", "dev-jwt-secret-change-me"),
		JWTExpiresHours:        getenvInt("JWT_EXPIRES_HOURS", 24),
		APIKeyEncryptionSecret: getenv("API_KEY_ENCRYPTION_SECRET", "dev-secret-change-me-32-byte-value"),
		MCPUploadMaxMB:         getenvInt("MCP_UPLOAD_MAX_MB", 20),
		MCPTmpDir:              getenv("MCP_TMP_DIR", "../storage/tmp/mcp"),
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
