package database

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DSN                   string
	MaxConns              int32
	MinConns              int32
	MaxConnLifetime       time.Duration
	MaxConnLifetimeJitter time.Duration
	MaxConnIdleTime       time.Duration
	HealthCheckPeriod     time.Duration
	ConnectTimeout        time.Duration
}

func LoadConfigFromEnv() (Config, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is not set")
	}

	cfg := Config{
		DSN:                   dsn,
		MaxConns:              envInt32("DB_MAX_CONNS", 10),
		MinConns:              envInt32("DB_MIN_CONNS", 2),
		MaxConnLifetime:       envDuration("DB_MAX_CONN_LIFETIME", time.Hour),
		MaxConnLifetimeJitter: envDuration("DB_MAX_CONN_LIFETIME_JITTER", 5*time.Minute),
		MaxConnIdleTime:       envDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
		HealthCheckPeriod:     envDuration("DB_HEALTH_CHECK_PERIOD", time.Minute),
		ConnectTimeout:        envDuration("DB_CONNECT_TIMEOUT", 5*time.Second),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.MaxConns <= 0 {
		return fmt.Errorf("DB_MAX_CONNS must be > 0")
	}

	if c.MinConns < 0 || c.MinConns > c.MaxConns {
		return fmt.Errorf("DB_MIN_CONNS must be in [0, DB_MAX_CONNS]")
	}

	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("DB_CONNECT_TIMEOUT must be > 0")
	}

	return nil
}

func envInt32(key string, def int32) int32 {
	value := os.Getenv(key)
	if value == "" {
		return def
	}

	number, err := strconv.Atoi(value)
	if err != nil {
		return def
	}

	return int32(number)
}

func envDuration(key string, def time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return def
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return def
	}

	return duration
}
