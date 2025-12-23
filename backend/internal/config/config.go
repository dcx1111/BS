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
	// AI相关配置（使用智谱AI GLM-4 Vision，国内可用）
	AIApiKey        string  // 智谱AI API密钥，从 https://open.bigmodel.cn/ 获取
	AIApiURL        string  // 智谱AI API的URL
	AIModel         string  // 使用的AI模型名称，默认为glm-4v（支持图片分析）
	AIEnabled       bool    // 是否启用AI功能
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
		// AI配置，使用智谱AI GLM-4 Vision（国内可用）
		AIApiKey:        getEnv("AI_API_KEY", "990a23ed91bb4c18bff6feb63df0dea2.2y7qkV5jR2ceAg1f"),
		AIApiURL:        getEnv("AI_API_URL", "https://open.bigmodel.cn/api/paas/v4/chat/completions"),
		AIModel:         getEnv("AI_MODEL", "glm-4v"),
		AIEnabled:       getEnvAsBool("AI_ENABLED", true),  // 默认不启用，需要显式设置
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

func getEnvAsBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
		log.Printf("invalid value for %s, using fallback %v", key, fallback)
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
