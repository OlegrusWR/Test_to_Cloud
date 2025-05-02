package loadbalancer

import (

	"sync"
	"github.com/OlegrusWR/balancer_to_cloud/internal/backend"
	"sync/atomic"
)

// LeastConn - стратегия балансировки "Наименьшее количество соединений"
type LeastConn struct {
	mu sync.Mutex
}

// NewLeastConn создает новый экземпляр стратегии LeastConn
func NewLeastConn() *LeastConn {
	return &LeastConn{}
}

// SelectServer выбирает сервер с наименьшим количеством активных соединений
func (lc *LeastConn) SelectServer(servers []*backend.Server) (*backend.Server, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	var selected *backend.Server
	minConn := -1

	// Перебор всех серверов для поиска наименее нагруженного
	for _, server := range servers {
		if !server.IsAlive() {
			continue
		}

		// Получаем текущее количество соединений сервера
		connCount := server.GetConnCount()

		// Выбираем сервер если:
		// 1. Это первый проверяемый сервер (minConn == -1)
		// 2. Или у него меньше соединений, чем у текущего выбранного
		if minConn == -1 || connCount < minConn {
			minConn = connCount
			selected = server
		}
	}
	// Если ни один сервер не доступен, возвращаем ошибку
	if selected == nil {
		return nil, backend.ErrNoAliveServers
	}
	return selected, nil
}

type RoundRobin struct {
	counter uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (rr *RoundRobin) SelectServer(servers []*backend.Server) (*backend.Server, error) {
	start := atomic.AddUint64(&rr.counter, 1)
	size := uint64(len(servers))

	for i := uint64(0); i < size; i++ {
		idx := (start + i) % size
		server := servers[idx]
		
		if server.IsAlive() {
			return server, nil
		}
	}

	return nil, backend.ErrNoAliveServers
}