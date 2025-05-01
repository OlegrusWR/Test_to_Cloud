package loadbalancer

import (
	"net/http"
	"sync"
	"log"
	"github.com/OlegrusWR/balancer_to_cloud/internal/backend"
)
// Balancer определяет интерфейс для стратегий балансировки (Если успею, то сделаю еще одну стратегию)
type Balancer interface {
	SelectServer(servers []*backend.Server, r *http.Request) (*backend.Server, error)
}

// LoadBalancer - основной тип, реализующий балансировку HTTP-запросов
// Содержит:
// - mu: RWMutex для потокобезопасного доступа к списку серверов
// - servers: список доступных бэкенд-серверов
// - strategy: алгоритм выбора сервера (реализует интерфейс Balancer)
// - logger: логгер для записи событий
type LoadBalancer struct {
	mu      sync.RWMutex
	servers []*backend.Server
	strategy Balancer
	logger  *log.Logger
}

// NewLoadBalancer создает новый экземпляр балансировщика
// Параметры:
// - servers: список бэкенд-серверов для балансировки
// - strategy: стратегия выбора сервера
// - logger: логгер для записи событий
// Возвращает:
// - *LoadBalancer: готовый к работе экземпляр балансировщика
func NewLoadBalancer(servers []*backend.Server, strategy Balancer, logger *log.Logger) *LoadBalancer {
	return &LoadBalancer{
		servers:  servers,
		strategy: strategy,
		logger:   logger,
	}
}

// ServeHTTP обрабатывает HTTP-запросы, реализуя интерфейс http.Handler
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    lb.logger.Printf("Входящий запрос: %s %s", r.Method, r.URL.Path)
    
	// Безопасное чтение списка серверов
    lb.mu.RLock()
    server, err := lb.strategy.SelectServer(lb.servers, r)
    lb.mu.RUnlock()

	// Обработка ошибок выбора сервера
    if err != nil {
        lb.logger.Printf("Ошибка выбора сервера: %v", err)
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }

    lb.logger.Printf("Перенаправление на сервер: %s", server.URL)
    server.ServeHTTP(w, r)
}