package config

import (
	"fmt"
	"os"
	"time"
	"gopkg.in/yaml.v2"
)

type HelthCheckConfig struct {
	IntervalStr string  `yaml:"interval"`
	TimeoutStr string  `yaml:"timeout"`
	Path string `yaml:"path"`
	Interval time.Duration
	Timeout time.Duration
}

type LogConfig struct {
	Level string `yaml:"level"`
	File string `yaml:"file"`
}

type Config struct{
	Listen_port string `yaml:"listen_port"`
	Algoritm string `yaml:"algoritm"`
	Backends []string `yaml:"backends"`
	HelthCheck HelthCheckConfig `yaml:"helth_check"`
	Logging LogConfig `yaml:"logging"`
}

func LoadConf(path string) (*Config, error) {
	data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения конфигурационного файла: %w", err)
		}
	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err  != nil{
		return nil, fmt.Errorf("ошибка парсинга конфига: %w", err)
	}
	var parseErr error

	cfg.HelthCheck.Interval, parseErr = time.ParseDuration(cfg.HelthCheck.IntervalStr)
	if parseErr != nil {
		return nil, fmt.Errorf("недопустимый интервал проверки работоспособности: %w", parseErr)
	} 
	
	cfg.HelthCheck.Timeout, parseErr = time.ParseDuration(cfg.HelthCheck.TimeoutStr)
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

