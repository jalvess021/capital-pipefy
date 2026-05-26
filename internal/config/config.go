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

    port := os.Getenv("PORT")
    if port == "" {
        port = "8282"
    }

    cfg := &Config{
        Port:         port,
        DatabaseURL:  os.Getenv("DATABASE_URL"),
        RedisURL:     os.Getenv("REDIS_URL"),
        RabbitMQURL:  os.Getenv("RABBITMQ_URL"),
        PipefyPipeID: os.Getenv("PIPEFY_PIPE_ID"),
        PipefyAPIURL: os.Getenv("PIPEFY_API_URL"),
        PipefyToken:  os.Getenv("PIPEFY_TOKEN"),
    }

    if cfg.DatabaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }
    if cfg.RedisURL == "" {
        return nil, fmt.Errorf("REDIS_URL is required")
    }
    if cfg.RabbitMQURL == "" {
        return nil, fmt.Errorf("RABBITMQ_URL is required")
    }

    return cfg, nil
}