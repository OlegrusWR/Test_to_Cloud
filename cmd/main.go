package main

import (
	"log"

	"github.com/OlegrusWR/balancer_to_cloud/config"
	"github.com/OlegrusWR/balancer_to_cloud/internal/backend"
	"github.com/OlegrusWR/balancer_to_cloud/internal/logger"
)

func main () {
	cfg, err := config.LoadConf("config.yaml")
	if err != nil{
		log.Fatal("Ошибка загрузки конфига: %w", err)
	}

	logger.Init(cfg.Logging.File)
	lg := logger.Get()
	lg.Printf("Запуск балансировщика нагрузки с помощью конфигурации: %+v", cfg)
	hc := backend.HealthCheck{
		Interval: cfg.HealthCheck.Interval,
		Timeout: cfg.HealthCheck.Timeout,
		Path: cfg.HealthCheck.Path,
	}

	
}