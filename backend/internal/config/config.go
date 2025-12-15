package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort      string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	JWTSecret       string
	StorageDir      string
	ThumbnailWidth  int
	ThumbnailHeight int
	MaxUploadSize   int64
	CORSOrigins     []string
}

func Load() Config {
	return Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		DBHost:          getEnv("DB_HOST", "127.0.0.1"),
		DBPort:          getEnv("DB_PORT", "3306"),
		DBUser:          getEnv("DB_USER", "root"),
		DBPassword:      getEnv("DB_PASSWORD", "13456301882dcx"),
		DBName:          getEnv("DB_NAME", "image_manager"),
		JWTSecret:       getEnv("JWT_SECRET", "3k136dd882bas21"),
		StorageDir:      getEnv("STORAGE_DIR", "./storage"),
		ThumbnailWidth:  getEnvAsInt("THUMBNAIL_WIDTH", 300),
		ThumbnailHeight: getEnvAsInt("THUMBNAIL_HEIGHT", 300),
		MaxUploadSize:   getEnvAsInt64("MAX_UPLOAD_SIZE", 10*1024*1024),
		CORSOrigins:     getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
		log.Printf("invalid value for %s, using fallback %d", key, fallback)
	}
	return fallback
}

func getEnvAsInt64(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
		log.Printf("invalid value for %s, using fallback %d", key, fallback)
	}
	return fallback
}

func getEnvAsSlice(key string, fallback []string) []string {
	if value, ok := os.LookupEnv(key); ok {
		parts := strings.Split(value, ",")
		var result []string
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return fallback
}
