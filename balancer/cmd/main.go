package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strings"

	"github.com/OlegrusWR/balancer_to_cloud/config"
	"github.com/OlegrusWR/balancer_to_cloud/internal/backend"
	"github.com/OlegrusWR/balancer_to_cloud/internal/loadbalancer"
	"github.com/OlegrusWR/balancer_to_cloud/internal/logger"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConf("config.yaml")
	if err != nil {
		log.Fatalf("ошибка загрузки конфига: %v", err)
	}
	// Инициализация логгеров
	// rlLogger := logger.InitRateLimitLogger(cfg.Logging.RateLimitFile)
	// dbLogger := logger.InitDataBaseLogger(cfg.Logging.DataBaseFile)
	bLogger := logger.InitBalancerLogger(cfg.Logging.LBalancerFile)
	defer bLogger.Println("завершение работы")

	// Проверка алгоритма балансировки (пока так, если успею, то будет проверка какой алгоритм выбраран в конфиге)
	if strings.ToLower(cfg.Algoritm) != "least_conn" {
		bLogger.Fatal("поддерживается только алгоритм 'least_conn'")
	}

	// Настройка health-check
	hc := &backend.HealthCheck{
		Interval: cfg.HealthCheck.Interval,
		Timeout:  cfg.HealthCheck.Timeout,
		Path:     cfg.HealthCheck.Path,
	}

	// Создание серверов
	servers := make([]*backend.Server, 0)
	for _, url := range cfg.Backends {
		server, err := backend.NewServer(url, hc, bLogger)
		if err != nil {
			bLogger.Fatalf("ошибка создания сервера: %v", err)
		}
		servers = append(servers, server)
	}

	// Настройка порта, если порт не казан, то будет выбрал по умолчанию 8080
	port := cfg.Listen_port
	if port == "" {
		port = "8080"
	}

	// Инициализация балансировщика
	lb := loadbalancer.NewLoadBalancer(servers, loadbalancer.NewLeastConn(), bLogger)

	// Настройка HTTP-сервера
	server := &http.Server{
		Addr:    ":" + port,
		Handler: lb,
	}

	// Настройка graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера в горутине
	go func() {
		bLogger.Printf("запуск сервера на :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			bLogger.Fatalf("ошибка сервера: %v", err)
		}
	}()

	//Ожидание сигнала завершения
	<-done
	bLogger.Println("завершение работы...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		bLogger.Printf("ошибка завершения: %v", err)
	}
}