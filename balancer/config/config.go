package config

import (
	"fmt"
	"os"
	"time"
	"gopkg.in/yaml.v2"
)

type HealthCheckConfig struct {                  // Структура для проверки сервера, на то жив он или нет
	Interval time.Duration `yaml:"interval"` 
	Timeout  time.Duration `yaml:"timeout"`
	Path     string        `yaml:"path"`
}

type LogConfig struct {                          // Структура для логгера, указываем уровень(инфо)
	Level string `yaml:"level"`  				 // и фаил, куда все логи будут записываться
	LBalancerFile  string `yaml:"lbfile"` 
	DataBaseFile string `yaml:"dbfile"`
	RateLimitFile string `yaml:"rlfile"`
}
type RateLimiterConfig struct {
    Enabled         bool            `yaml:"enabled"`
    DefaultCapacity int             `yaml:"default_capacity"`
    DefaultRate     time.Duration   `yaml:"default_rate"`
    VIPCapacity     int             `yaml:"vip_capacity"`
    VIPRate         time.Duration   `yaml:"vip_rate"`
    DBPath          string          `yaml:"db_path"`
}

type Config struct {
    Listen_port  string           `yaml:"listen_port"`
    Algoritm     string           `yaml:"algoritm"`
    Backends     []string         `yaml:"backends"`
    HealthCheck  HealthCheckConfig `yaml:"health_check"`
    Logging      LogConfig        `yaml:"logging"`
    RateLimiter  RateLimiterConfig `yaml:"rate_limiter"`
}

func LoadConf(path string) (*Config, error) {            // Читаем наш конфиг фаил и записываем всё в структуру Config
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфига: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга: %w", err)
	}

	if cfg.Listen_port == "" {
		return nil, fmt.Errorf("порт не указан")
	}
	if len(cfg.Backends) == 0 {
		return nil, fmt.Errorf("нет бэкендов")
	}

	return &cfg, nil
}