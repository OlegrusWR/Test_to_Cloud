package config

import (
	"fmt"
	"os"
	"time"
	"gopkg.in/yaml.v2"
)

type HealthCheckConfig struct {				// Структура для проверки сервера, на то жив он или нет
	IntervalStr string  `yaml:"interval"`
	TimeoutStr string  `yaml:"timeout"`
	Path string `yaml:"path"`
	Interval time.Duration
	Timeout time.Duration
}

type LogConfig struct {						// Структура для логгера, указываем уровень(в нашем случае это информауионны
	Level string `yaml:"level"`				// и фаил, куда все логи будут записываться
	File string `yaml:"file"`
}

type Config struct{							// Основная структура нашего конфига
	Listen_port string `yaml:"listen_port"`
	Algoritm string `yaml:"algoritm"`
	Backends []string `yaml:"backends"`
	HealthCheck HealthCheckConfig `yaml:"health_check"`
	Logging LogConfig `yaml:"logging"`
}

func LoadConf(path string) (*Config, error) {   // Читаем наш конфиг фаил и записываеи всё в структуру Config
	data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения конфигурационного файла: %w", err)
		}
	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err  != nil{
		return nil, fmt.Errorf("ошибка парсинга конфига: %w", err)
	}
	var parseErr error

	cfg.HealthCheck.Interval, parseErr = time.ParseDuration(cfg.HealthCheck.IntervalStr) // парсим строки в time.Duration
	if parseErr != nil {
		return nil, fmt.Errorf("недопустимый интервал проверки работоспособности: %w", parseErr)
	} 
	
	cfg.HealthCheck.Timeout, parseErr = time.ParseDuration(cfg.HealthCheck.TimeoutStr)
	if parseErr != nil {
		return nil, fmt.Errorf("недопустимый тайм-аут проверки работоспособности: %w", parseErr)
	} 
	if cfg.Listen_port == ""{
		return nil, fmt.Errorf("требуется порт для прослушивания")
	}
	if len(cfg.Backends) == 0{
		return nil, fmt.Errorf("необходим хотя бы один сервер")
	}
	return &cfg, nil
}

