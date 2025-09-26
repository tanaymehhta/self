package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string
	Env  string

	// Supabase
	SupabaseURL        string
	SupabaseAnonKey    string
	SupabaseServiceKey string
	DatabaseURL        string

	// JWT
	JWTSecret        string
	JWTRefreshSecret string

	// MinIO/Storage
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string

	// Redis
	RedisURL string

	// NATS
	NATSURL string

	// Qdrant
	QdrantURL string

	// File Upload
	MaxFileSize int64 // in bytes

	// AI Services
	OpenAIAPIKey string
	ClaudeAPIKey string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("ENVIRONMENT", "development"),

		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:    getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceKey: getEnv("SUPABASE_SERVICE_KEY", ""),
		DatabaseURL:        getEnv("DATABASE_URL", ""),

		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-this"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-change-this"),

		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinIOBucket:    getEnv("MINIO_BUCKET", "self-audio-files"),

		RedisURL: getEnv("REDIS_URL", "localhost:6379"),
		NATSURL:  getEnv("NATS_URL", "localhost:4222"),

		QdrantURL: getEnv("QDRANT_URL", "http://localhost:6333"),

		MaxFileSize: getEnvInt64("MAX_FILE_SIZE", 1024*1024*1024), // 1GB default

		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		ClaudeAPIKey: getEnv("CLAUDE_API_KEY", ""),
	}

	// Validate required config
	if config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}