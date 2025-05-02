package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OlegrusWR/balancer_to_cloud/config"
	"github.com/OlegrusWR/balancer_to_cloud/internal/backend"
	"github.com/OlegrusWR/balancer_to_cloud/internal/loadbalancer"
	"github.com/OlegrusWR/balancer_to_cloud/internal/logger"
	"github.com/OlegrusWR/balancer_to_cloud/internal/ratelimiter"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConf("config.yaml")
	if err != nil {
		log.Fatalf("ошибка загрузки конфига: %v", err)
	}
	// Инициализация логгеров
	bLogger := logger.NewLogger("BALANCER", "../log/loadbalancer.log")
	dbLogger := logger.NewLogger("DATABASE", "../log/database.log")
	rlLogger := logger.NewLogger("RATELIMIT", "../log/ratelimit.log")

	defer func() {
		bLogger.Println("Завершение работы балансировщика")
		dbLogger.Println("Завершение работы базы данных")
		rlLogger.Println("Завершение работы rate limiter")
	}()

	
	rlStorage, err := ratelimiter.NewStorage(cfg.RateLimiter.DBPath, rlLogger)
	if err != nil {
		dbLogger.Fatal("Ошибка инициализации хранилища rate limiter")
	}

	rlConfig := ratelimiter.NewConfigFrom(cfg.RateLimiter)
	rateLimiter := ratelimiter.NewService(rlConfig, rlStorage, rlLogger)

	
	go func() {
		adminMux := http.NewServeMux()
		adminMux.Handle("/admin/vip", rateLimiter.AdminHandler())
		rlLogger.Printf("Админ-сервер запущен на :8888")
		if err := http.ListenAndServe(":9888", adminMux); err != nil {
			rlLogger.Fatalf("Ошибка админ-сервера: %v", err)
		}
	}()

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

	var algoritm loadbalancer.Balancer
	switch cfg.Algoritm {
	case "round_robin":
		algoritm = loadbalancer.NewRoundRobin()
	default:
		algoritm = loadbalancer.NewLeastConn()
	}

	// Инициализация балансировщика
	lb := loadbalancer.NewLoadBalancer(servers, algoritm, bLogger)

	// Настройка порта, если порт не казан, то будет выбрал по умолчанию 8080
	port := cfg.Listen_port
	if port == "" {
		port = "8080"
	}

	handler := rateLimiter.Middleware(lb)

	// Настройка HTTP-сервера
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
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