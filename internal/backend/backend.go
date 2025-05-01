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

//Server представляет бэкенд-сервер для балансировки нагрузки
type Server struct {                                    
	URL          string
	mu           sync.RWMutex
	connCount    int
	alive        bool
	healthCheck  *HealthCheck
	proxy        *httputil.ReverseProxy
	logger       *log.Logger
}

// HealthCheck содержит параметры проверки здоровья серверов
type HealthCheck struct {                               
	Interval time.Duration
	Timeout  time.Duration
	Path     string
	Client   *http.Client
}

var (
	ErrNoAliveServers = errors.New("нет живых серверов")
	ErrInvalidUrl     = errors.New("неверный URL")
)


// NewServer создаем новый экземпляр сервера с health-check и reverse proxy
func NewServer(serverURL string, hc *HealthCheck, logger *log.Logger) (*Server, error) {
	target, err := url.Parse(serverURL)
	if err != nil {
		return nil, ErrInvalidUrl
	}

	client := &http.Client{
		Timeout: hc.Timeout,
	}

	server := &Server{                // Создаем экземпляр сервера
		URL:  serverURL,
		alive: true,
		healthCheck: &HealthCheck{
			Interval: hc.Interval,
			Timeout:  hc.Timeout,
			Path:     hc.Path,
			Client:   client,
		},
		proxy:  httputil.NewSingleHostReverseProxy(target),  // Инициализация прокси
		logger: logger,
	}

	// обработчик ошибок прокси
	server.proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {  

		server.logger.Printf("ошибка прокси: %v", err)
		server.SetAlive(false)									// При ошибке помечаем сервер "мертвым"
		w.WriteHeader(http.StatusBadGateway)					// 502 Bad Gateway
	}

	go server.StartHealthCheck()
	return server, nil
}

// StartHealthCheck запускает периодическую проверку здоровья сервера
// Используем ticker для проверок с заданным интервалом
func (s *Server) StartHealthCheck() {
	ticker := time.NewTicker(s.healthCheck.Interval)
	defer ticker.Stop()

	// Бесконечный цикл проверок
	for range ticker.C {
		resp, err := s.healthCheck.Client.Get(s.URL + s.healthCheck.Path)
		alive := err == nil && resp.StatusCode == http.StatusOK
		if resp != nil {
			resp.Body.Close()
		}
		s.SetAlive(alive)
	}
}
// SetAlive обновляет статус сервера
func (s *Server) SetAlive(alive bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alive = alive
	s.logger.Printf("Сервер %s состояние: alive=%v", s.URL, alive)
}

//IsAlive возвращает текущий статус сервера
func (s *Server) IsAlive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alive
}

// ServeHTTP обрабатывает входящий HTTP-запрос
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Увеличение счетчика
	s.mu.Lock()
	s.connCount++
	s.mu.Unlock()
	
	// Уменишение счетчика по завершению
	defer func() {
		s.mu.Lock()
		s.connCount--
		s.mu.Unlock()
	}()

	// Проксирование запроса
	s.proxy.ServeHTTP(w, r)
}

// GetConnCount возвращает текущее количество активных соединений
func (s *Server) GetConnCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connCount
}