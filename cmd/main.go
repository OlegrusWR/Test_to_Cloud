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
)

func main () {
	
	cfg, err := config.LoadConf("config.yaml")
	if err != nil{
		log.Fatal("Ошибка загрузки конфига: %w", err)
	}

	
	logger.Init(cfg.Logging.File)
	lg := logger.Get()
	lg.Printf("Запуск балансировщика нагрузки с помощью конфигурации: %+v", cfg)
	defer lg.Println("работа балансировщика завершилась")

	if cfg.Algoritm != "least_Conn"{
		lg.Fatalf("Неподдерживаемый алгоритм балансировки: %s, поддерживается только алгоритм \"least_conn\"", cfg.Algoritm)
	}

	hc := &backend.HealthCheck{
		Interval: cfg.HealthCheck.Interval,
		Timeout: cfg.HealthCheck.Timeout,
		Path: cfg.HealthCheck.Path,
	}

	servers := []*backend.Server{}
	for _, url := range cfg.Backends{
		server, err := backend.NewServer(url, hc, lg)
		if err != nil{
			lg.Fatal("не удалось создать сервер")
		}
		servers = append(servers, server)
	}

	lb := loadbalancer.NewLoadBalancer(servers, loadbalancer.NewLeastConn(), lg)

	listenserver := http.Server{
		Addr: ":"+cfg.Listen_port,
		Handler: lb,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func () {
		lg.Printf("балансировщик начал работу на: %v", cfg.Listen_port)
		if err := listenserver.ListenAndServe(); err != nil {
			lg.Fatalf("ошибка прослушивания: %v", err)
		}
	}()

	<-done
	lg.Printf("Завершение работы")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := listenserver.Shutdown(ctx); err != nil {
		lg.Printf("SОшибка завершения: %v", err)
	}
}