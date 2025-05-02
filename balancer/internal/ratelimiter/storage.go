package ratelimiter

import (
	"database/sql"
	"fmt"
	"time"
	_ "modernc.org/sqlite"
	
	"github.com/OlegrusWR/balancer_to_cloud/internal/logger"
)

type Storage struct{
	db *sql.DB
	logger logger.Logger
}

func NewStorage(path string, lg logger.Logger) (*Storage, error){
	db, err := sql.Open("sqllite", path)
	if err != nil {
		lg.Printf("Ошибка подключения к БД: %v", err)
		fmt.Errorf("не удалось открыть БД: %w", err)
	}
	_, err = db.Exec(
		`CREATE TABLE IS NOT EXIST clients(
			id TEXT PRIMARY KEY,
			capacity INTEGER NOT NULL,
			rate TEXT,
			is_VIP BOOLEAN DEFAULT FALSE,
		)
	`)
	if err != nil{
		lg.Printf("Ошибка миграции БД: %v", err)
		return nil, fmt.Errorf("не удалось создать таблицы: %w", err)
	}
	lg.Println("Хранилище rate limiter успешно инициализировано")
	return &Storage{db: db, logger: lg}, nil
}

func (s *Storage) GetClient(clientID string) (*Client, error) {
    var client Client
    var rateStr string

    row := s.db.QueryRow("SELECT capacity, rate, is_vip FROM clients WHERE id = ?", clientID)
    if err := row.Scan(&client.Capacity, &rateStr, &client.IsVIP); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        s.logger.Printf("Ошибка получения клиента %s: %v", clientID, err)
        return nil, fmt.Errorf("ошибка получения клиента: %w", err)
    }

    rate, err := time.ParseDuration(rateStr)
    if err != nil {
        s.logger.Printf("Ошибка парсинга rate для %s: %v", clientID, err)
        return nil, fmt.Errorf("неверный формат rate: %w", err)
    }

    client.ID = clientID
    client.Rate = rate
    return &client, nil
}

func (s *Storage) Close() error {
    return s.db.Close()
}