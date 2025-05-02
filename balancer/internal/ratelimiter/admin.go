package ratelimiter

import (
	"net/http"
)

func (s *Service) AdminHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		clientID := r.FormValue("client_id")
		isVIP := r.FormValue("vip") == "true"

		s.logger.Printf("Обновление VIP-статуса для клиента: %s to %v", clientID, isVIP)

		if err := s.storage.SetVIP(clientID, isVIP); err != nil {
			s.logger.Printf("Не удалось обновить VIP-статус: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s.mu.Lock()
		delete(s.buckets, clientID)
		s.mu.Unlock()

		w.Write([]byte("Обновлен VIP-статус"))
		s.logger.Printf("VIP-статус успешно обновлен для: %s", clientID)
	}
}