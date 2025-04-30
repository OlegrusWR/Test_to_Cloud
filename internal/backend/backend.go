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
	logger *log.Logger
}

type HealthCheck struct{
	Interval time.Duration
	Timeout time.Duration
	Path string
	Client *http.Client
}

var ErrNoAliveServers = errors.New("нет доступных серверов")
var ErrInvalidUrl = errors.New("неверный URL-адрес сервера")

func NewServer(serverURL string, hc *HealthCheck, logger *log.Logger) (*Server, error) {
	target, err := url.Parse(serverURL)
		if err != nil {
			return nil, ErrInvalidUrl
		}
	
		client := &http.Client{
			Timeout: hc.Timeout,
			Transport: &http.Transport{
				MaxIdleConns: 100,
				IdleConnTimeout: 90,
				DisableCompression: true,
			},
		}

	s := &Server{
		URL: serverURL,
		alive: true,
		healthCheck: &HealthCheck{
			Interval: hc.Interval,
			Timeout: hc.Timeout,
			Path: hc.Path,
			Client: client,
		},
		proxy: httputil.NewSingleHostReverseProxy(target),
	}

	s.proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		s.logger.Printf("ошибка прокси для %s: %v", s.URL, err)
		s.SetAlive(false)
		w.WriteHeader(http.StatusBadGateway)
	}
	go s.StartHealthCheck()
	return s, nil
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
		s.logger.Printf("Сервер %s изменил состояние на %v", s.URL, alive)
	}
	s.alive = alive
}

func (s *Server) IsAlive() bool{
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alive
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request){
	s.mu.Lock()
	s.connCount++
	s.mu.Unlock()

	defer func ()  {
		s.mu.Lock()
		s.connCount--
		s.mu.Unlock()
	}()
	s.logger.Printf("проксирующий запрос на %s%s", s.URL, r.URL.Path)
	s.proxy.ServeHTTP(w, r)

}

func (s * Server) GetConnCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connCount
}