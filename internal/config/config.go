package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type PipefyConfig struct {
	APIURL           string
	Token            string
	PipeID           string
	HTTPTimeout      time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
	CBThreshold      uint32
	CBOpenTimeout    time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RateLimitConfig struct {
	RedisURL string
	RPS      int
	Enabled  bool
}

type Config struct {
	Port      	string
	Database  	DatabaseConfig
	RateLimit 	RateLimitConfig
	RabbitMQURL string
	Pipefy    	PipefyConfig
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        os.Getenv("PORT"),
		RabbitMQURL: os.Getenv("RABBITMQ_URL"),
		Database: DatabaseConfig{
			URL:             os.Getenv("DATABASE_URL"),
			MaxOpenConns:    parseInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:    parseInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: parseDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		RateLimit: RateLimitConfig{
			RedisURL: os.Getenv("REDIS_URL"),
			RPS:      parseInt("RATE_LIMIT_RPS", 10),
			Enabled:  parseBool("RATE_LIMIT_ENABLED", true),
		},
		Pipefy: PipefyConfig{
			APIURL:        os.Getenv("PIPEFY_API_URL"),
			Token:         os.Getenv("PIPEFY_TOKEN"),
			PipeID:        os.Getenv("PIPEFY_PIPE_ID"),
			HTTPTimeout:   parseDuration("PIPEFY_HTTP_TIMEOUT", 10*time.Second),
			MaxRetries:    parseInt("PIPEFY_MAX_RETRIES", 3),
			RetryDelay:    parseDuration("PIPEFY_RETRY_DELAY", 500*time.Millisecond),
			CBThreshold:   uint32(parseInt("PIPEFY_CB_THRESHOLD", 5)),
			CBOpenTimeout: parseDuration("PIPEFY_CB_OPEN_TIMEOUT", 30*time.Second),
		},
	}

	if cfg.Port == "" {
		return nil, fmt.Errorf("PORT is required")
	}
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.Pipefy.APIURL == "" {
		return nil, fmt.Errorf("PIPEFY_API_URL is required")
	}
	if cfg.Pipefy.Token == "" {
		return nil, fmt.Errorf("PIPEFY_TOKEN is required")
	}
	if cfg.Pipefy.PipeID == "" {
		return nil, fmt.Errorf("PIPEFY_PIPE_ID is required")
	}

	return cfg, nil
}

func parseDuration(env string, fallback time.Duration) time.Duration {
	v := os.Getenv(env)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func parseBool(env string, fallback bool) bool {
	v := os.Getenv(env)
	if v == "" {
		return fallback
	}
	return v == "true" || v == "1"
}

func parseInt(env string, fallback int) int {
	v := os.Getenv(env)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
