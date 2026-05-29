package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	ServerPort string
	ServerHost string

	JWTSecret string

	DeepSeekAPIKey string
	DeepSeekAPIURL string

	FrontendURL string
	SkillsDir   string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioPublicURL string
	MinioUseSSL    bool

	EnableEmotionSummary  bool
	EmotionSummaryInterval int
}

var AppConfig *Config

func LoadConfig() *Config {
	godotenv.Load()

	AppConfig = &Config{
		DBHost:         getEnv("DB_HOST", "127.0.0.1"),
		DBPort:         getEnv("DB_PORT", "3306"),
		DBUser:         getEnv("DB_USER", "root"),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		DBName:         getEnv("DB_NAME", "rain_yi"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		ServerHost:     getEnv("SERVER_HOST", "0.0.0.0"),
		JWTSecret:      getEnv("JWT_SECRET", "rain-yi-secret"),
		DeepSeekAPIKey: getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekAPIURL: getEnv("DEEPSEEK_API_URL", "https://api.deepseek.com"),
		FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:5173"),
		SkillsDir:      getEnv("SKILLS_DIR", "../skills"),

		RedisHost:     getEnv("REDIS_HOST", ""),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		MinioEndpoint:  getEnv("MINIO_ENDPOINT", ""),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", ""),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", ""),
		MinioBucket:    getEnv("MINIO_BUCKET", "rain-yi"),
		MinioPublicURL: getEnv("MINIO_PUBLIC_URL", ""),
		MinioUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",

		EnableEmotionSummary:  getEnv("ENABLE_EMOTION_SUMMARY", "false") == "true",
		EmotionSummaryInterval: getEnvAsInt("EMOTION_SUMMARY_INTERVAL", 10),
	}

	return AppConfig
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
