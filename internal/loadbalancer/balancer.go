package loadbalancer

import (
	"net/http"
	"sync"
	"log"
	"github.com/OlegrusWR/balancer_to_cloud/internal/backend"
)

type Balancer interface{
	SelectServer(servers []*backend.Server, r *http.Request) (*backend.Server, error)
}

type LoadBalancer struct{
	mu sync.RWMutex
	servers []*backend.Server
	strategy Balancer
	logger *log.Logger
}

func NewLoadBalancer(servers []*backend.Server, strategy Balancer, logger *log.Logger) *LoadBalancer{
	
	return &LoadBalancer{
		servers: servers,
		strategy: strategy,
		logger: logger,
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request){
	log.Printf("входящий запрос: %s %s", r.Method, r.URL.Path)

	lb.mu.RLock()
	server, err := lb.strategy.SelectServer(lb.servers, r)
	lb.mu.RUnlock()
	if err != nil{
		lb.logger.Printf("нет доступных серверов: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	go func() {
		server.ServeHTTP(w, r)
	}()
}