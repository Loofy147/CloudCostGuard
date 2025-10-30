package config

import (
    "fmt"
    "time"

    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Cache    CacheConfig
    Logging  LoggingConfig
}

type ServerConfig struct {
    Port            string        `envconfig:"PORT" default:"8080"`
    ReadTimeout     time.Duration `envconfig:"READ_TIMEOUT" default:"15s"`
    WriteTimeout    time.Duration `envconfig:"WRITE_TIMEOUT" default:"15s"`
    ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
}

type DatabaseConfig struct {
    Host            string        `envconfig:"DB_HOST" required:"true"`
    Port            int           `envconfig:"DB_PORT" default:"5432"`
    User            string        `envconfig:"DB_USER" required:"true"`
    Password        string        `envconfig:"DB_PASSWORD" required:"true"`
    Database        string        `envconfig:"DB_NAME" default:"pricing"`
    SSLMode         string        `envconfig:"DB_SSL_MODE" default:"disable"`
    MaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
    MaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
    ConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
}

type CacheConfig struct {
    RefreshInterval time.Duration `envconfig:"CACHE_REFRESH_INTERVAL" default:"6h"`
}

type LoggingConfig struct {
    Level  string `envconfig:"LOG_LEVEL" default:"info"`
    Format string `envconfig:"LOG_FORMAT" default:"json"`
}

func Load() (*Config, error) {
    var cfg Config
    if err := envconfig.Process("CCG", &cfg); err != nil {
        return nil, fmt.Errorf("failed to process env vars: %w", err)
    }
    return &cfg, nil
}
