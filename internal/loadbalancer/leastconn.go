package loadbalancer

import (
	"net/http"
	"sync"

	"github.com/OlegrusWR/balancer_to_cloud/internal/backend"
)

type LeastConn struct{
	mu sync.Mutex
}

func NewLeastConn() *LeastConn {
	return &LeastConn{}
}

func (lc *LeastConn) SelectServer (servers []*backend.Server, r *http.Request) (*backend.Server, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	var selected *backend.Server
	minConnect := -1

	for _, server := range servers{
		if !server.IsAlive(){
			continue
		}
		currentCountConn := server.GetConnCount()

		if minConnect == -1 || currentCountConn < minConnect{
			minConnect = currentCountConn
			selected = server
		}
	}
	if selected == nil {
		return nil, backend.ErrNoAliveServers
	}
	return selected, nil
}