package config

import (
    "fmt"
    "os"
)

type Config struct {
    Port         string
    DatabaseURL  string
    RedisURL     string
    RabbitMQURL  string
    PipefyPipeID string
    PipefyAPIURL string
    PipefyToken  string
}

func Load() (*Config, error) {

    cfg := &Config{
        Port:         os.Getenv("PORT"),
        DatabaseURL:  os.Getenv("DATABASE_URL"),
        RedisURL:     os.Getenv("REDIS_URL"),
        RabbitMQURL:  os.Getenv("RABBITMQ_URL"),
        PipefyPipeID: os.Getenv("PIPEFY_PIPE_ID"),
        PipefyAPIURL: os.Getenv("PIPEFY_API_URL"),
        PipefyToken:  os.Getenv("PIPEFY_TOKEN"),
    }

    if cfg.Port == "" {
        return nil, fmt.Errorf("PORT is required")
    }
    if cfg.DatabaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }
    if cfg.PipefyAPIURL == "" {
        return nil, fmt.Errorf("PIPEFY_API_URL is required")
    }
    if cfg.PipefyToken == "" {
        return nil, fmt.Errorf("PIPEFY_TOKEN is required")
    }
    if cfg.PipefyPipeID == "" {
        return nil, fmt.Errorf("PIPEFY_PIPE_ID is required")
    }

    return cfg, nil
}