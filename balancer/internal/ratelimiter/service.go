package ratelimiter

import (
    "sync"
    "github.com/OlegrusWR/balancer_to_cloud/internal/logger"
	"net/http"
	"strings"
	"net"
)

type Service struct {
	config  Config
	storage *Storage
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
	logger  logger.Logger
}

func NewService(cfg Config, storage *Storage, logger logger.Logger) *Service {
	return &Service{
		config:  cfg,
		storage: storage,
		buckets: make(map[string]*TokenBucket),
		logger:  logger,
	}
}

func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := s.getClientID(r)
		s.logger.Printf("Проверка лимита запросов: %s", sanitizeForLog(clientID))
		
		client, err := s.storage.GetOrCreate(clientID, s.config.DefaultCapacity, s.config.DefaultRate)
		if err != nil {
			s.logger.Printf("Ошибка при получении клиента: %v", err)
			http.Error(w, "сервис недоступен", http.StatusServiceUnavailable)
			return
		}

		if !s.allow(client) {
			s.logger.Printf("Превышен лимит запросов для клиента: %s", clientID)
			http.Error(w, "Слишком много запросов", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Service) allow(client *Client) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[client.ID]
	if !exists {
		cap := client.Capacity
		rate := client.Rate
		
		if client.IsVIP {
			cap = s.config.VIPCapacity
			rate = s.config.VIPRate
			s.logger.Printf("Использование VIP лимита для: %s", client.ID)
		}

		bucket = NewBucket(cap, rate)
		s.buckets[client.ID] = bucket
	}

	return bucket.Allow()
}


func (s *Service) getClientID(r *http.Request) string {
    // Правильное извлечение IP
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        host = r.RemoteAddr // fallback
    }

    // Обработка IPv6 (удаляем квадратные скобки)
    host = strings.TrimPrefix(host, "[")
    host = strings.TrimSuffix(host, "]")

    // Проверка API-ключа
    if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
        return "api:" + apiKey
    }
    
    return "ip:" + host
}

func sanitizeForLog(s string) string {
    s = strings.ReplaceAll(s, "\n", "\\n")
    s = strings.ReplaceAll(s, "\r", "\\r")
    return s
}