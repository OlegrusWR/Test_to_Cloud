package backend

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Server struct{
	URL string
	mu sync.RWMutex
	connCount int
	alive bool
	healthCheck *HealthCheck
	proxy *httputil.ReverseProxy
}

type HealthCheck struct{
	Interval time.Duration
	Timeout time.Duration
	Path string
	Client *http.Client
}

var ErrNoAliveServers = errors.New("нет доступных серверов")

func NewServer(serverURL string, hc *HealthCheck) *Server {
	target, _ := url.Parse(serverURL)
	
	s := &Server{
		URL: serverURL,
		alive: true,
		healthCheck: &HealthCheck{
			Interval: hc.Interval,
			Timeout: hc.Timeout,
			Path: hc.Path,
			Client: &http.Client{
				Timeout: hc.Timeout,
			},
		},
		proxy: httputil.NewSingleHostReverseProxy(target),
	}

	s.proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("ошибка прокси для %s: %v", s.URL, err)
		s.SetAlive(false)
		w.WriteHeader(http.StatusBadGateway)
	}
	go s.StartHealthCheck()
	return s
}

func (s *Server) StartHealthCheck(){
	ticker := time.NewTicker(s.healthCheck.Interval)
	defer ticker.Stop()

	for range ticker.C {
		resp, err := s.healthCheck.Client.Get(s.URL + s.healthCheck.Path)
		alive := err == nil && resp.StatusCode == http.StatusOK
		if resp != nil {
			resp.Body.Close()
		}
		s.SetAlive(alive)
	}
}

func (s *Server) SetAlive(alive bool){
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.alive != alive {
		log.Printf("Сервер %s изменил состояние на %v", s.URL, alive)
	}
	s.alive = alive
}

func (s *Server) IsAlive() bool{
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alive
}
func (s *Server) Serve(w http.ResponseWriter, r *http.Request){
	s.mu.Lock()
	s.connCount++
	s.mu.Unlock()

	defer func ()  {
		s.mu.Lock()
		s.connCount--
		s.mu.Unlock()
	}()

	s.proxy.ServeHTTP(w, r)

}

func (s * Server) GetConnCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connCount
}