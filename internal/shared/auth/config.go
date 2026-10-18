package auth

import (
	"fmt"
	"time"
)

type Config struct {
	Secret []byte
	TTL    time.Duration
}

func (c Config) validate() error {
	if len(c.Secret) < 32 {
		return fmt.Errorf("config: secret must be at least 32 bytes")
	}

	if c.TTL <= 0 {
		return fmt.Errorf("config: ttl must be > 0")
	}

	return nil
}
