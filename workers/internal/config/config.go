package config

import (
	"fmt"
	"os"
)

type Config struct {
	RabbitURL        string
	S3Endpoint       string
	S3Region         string
	S3AccessKey      string
	S3SecretKey      string
	S3Bucket         string
	S3UsePathStyle   bool
	APIBaseURL       string
	APIToken         string
	LLMBaseURL       string
	LLMModel         string
	LLMAPIKey        string
	ChatHTTPAddr     string
	ChatServiceToken string
}

func Load() (Config, error) {
	cfg := Config{
		RabbitURL:      getenv("RABBITMQ_URL", "amqp://racecoach:racecoach@localhost:5672/"),
		S3Endpoint:     getenv("S3_ENDPOINT", "http://localhost:9000"),
		S3Region:       getenv("S3_REGION", "us-east-1"),
		S3AccessKey:    getenv("S3_ACCESS_KEY", "racecoach"),
		S3SecretKey:    getenv("S3_SECRET_KEY", "racecoachsecret"),
		S3Bucket:       getenv("S3_BUCKET", "racecoach-fits"),
		S3UsePathStyle: getenv("S3_USE_PATH_STYLE", "true") == "true",

		APIBaseURL: getenv("API_BASE_URL", "http://localhost:8000"),
		APIToken:   getenv("WORKER_API_TOKEN", ""),

		LLMBaseURL: getenv("LLM_BASE_URL", "http://localhost:11434/v1"),
		LLMModel:   getenv("LLM_MODEL", "llama3.1:8b"),
		LLMAPIKey:  getenv("LLM_API_KEY", "ollama"),

		ChatHTTPAddr:     getenv("CHAT_HTTP_ADDR", ":8081"),
		ChatServiceToken: getenv("CHAT_SERVICE_TOKEN", ""),
	}

	if cfg.S3Bucket == "" {
		return cfg, fmt.Errorf("S3_BUCKET is required")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
