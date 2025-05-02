package ratelimiter

import (
	"time"
"github.com/OlegrusWR/balancer_to_cloud/config"
)
type Config struct {
    DefaultCapacity int           `yaml:"default_capacity"`
    DefaultRate     time.Duration `yaml:"default_rate"`
    VIPCapacity     int           `yaml:"vip_capacity"`
    VIPRate         time.Duration `yaml:"vip_rate"`
    DBPath          string        `yaml:"db_path"`
}

type Client struct {
    ID       string
    Capacity int
    Rate     time.Duration
    IsVIP    bool
}

func NewConfigFrom(cfg config.RateLimiterConfig) Config {
    return Config{
        DefaultCapacity: cfg.DefaultCapacity,
        DefaultRate:     cfg.DefaultRate,
        VIPCapacity:     cfg.VIPCapacity,
        VIPRate:         cfg.VIPRate,
        DBPath:          cfg.DBPath,
    }
}