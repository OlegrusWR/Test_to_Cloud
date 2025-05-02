package ratelimiter

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"os"
	"path/filepath"
	_ "modernc.org/sqlite"
	"github.com/OlegrusWR/balancer_to_cloud/internal/logger"

)

type Storage struct {
	db     *sql.DB
	logger logger.Logger
}

func NewStorage(path string, logger logger.Logger) (*Storage, error) {
	 // Создаём директорию, если её нет
	 if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return nil, fmt.Errorf("не удалось создать директорию БД: %w", err)
    }

    // Открываем БД с правильными параметрами
    db, err := sql.Open("sqlite", "file:"+path+"?cache=shared")
    if err != nil {
        return nil, fmt.Errorf("не удалось открыть БД: %w", err)
    }

    // Проверяем соединение
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("не удалось связаться с БД: %w", err)
    }

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS clients (
			id TEXT PRIMARY KEY,
			ip TEXT,
			api_key TEXT,
			capacity INTEGER,
			rate TEXT,
			is_vip BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		logger.Printf("Failed to create table: %v", err)
		return nil, fmt.Errorf("не удалось создать таблицу: %w", err)
	}

	logger.Println("БД успешно инициализирована")
	return &Storage{db: db, logger: logger}, nil
}

func (s *Storage) GetOrCreate(clientID string, defaultCap int, defaultRate time.Duration) (*Client, error) {
	
	if defaultRate <= 0 {
        defaultRate = time.Second 
    }

	client, err := s.get(clientID)
	if err != nil {
		s.logger.Printf("Ошибка при получении клиента %s: %v", clientID, err)
		return nil, err
	}

	if client == nil {
		s.logger.Printf("создание нового клиента: %s", clientID)
		return s.create(clientID, defaultCap, defaultRate)
	}

	return client, nil
}

func (s *Storage) get(clientID string) (*Client, error) {
	var c Client
	var rateStr string

	row := s.db.QueryRow("SELECT id, capacity, rate, is_vip FROM clients WHERE id = ?", clientID)
	if err := row.Scan(&c.ID, &c.Capacity, &rateStr, &c.IsVIP); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	rate, err := time.ParseDuration(rateStr)
	if err != nil {
		return nil, err
	}

	c.Rate = rate
	return &c, nil
}

func (s *Storage) create(clientID string, cap int, rate time.Duration) (*Client, error) {
	parts := strings.SplitN(clientID, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("неверная форма Id клиента")
	}

	clientType, id := parts[0], parts[1]
	query := "INSERT INTO clients (id, %s, capacity, rate) VALUES (?, ?, ?, ?)"
	
	if clientType == "ip" {
		query = fmt.Sprintf(query, "ip")
	} else {
		query = fmt.Sprintf(query, "api_key")
	}

	if _, err := s.db.Exec(query, clientID, id, cap, rate.String()); err != nil {
		return nil, err
	}

	return &Client{
		ID:       clientID,
		Capacity: cap,
		Rate:     rate,
		IsVIP:    false,
	}, nil
}

func (s *Storage) SetVIP(clientID string, isVIP bool) error {
	_, err := s.db.Exec(
		"UPDATE clients SET is_vip = ? WHERE id = ?",
		isVIP,
		clientID,
	)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}