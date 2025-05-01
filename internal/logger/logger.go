package logger

import (
	"log"
	"os"
	"sync"
)

var (
	once     sync.Once     // Гарантирует однократную инициализацию
	instance *log.Logger   // Единственный экземпляр логгера
)

// Init инициализирует логгер с записью в файл
func Init(logFile string) *log.Logger {
	once.Do(func() {

		// Открываем файл логов с режимами:
		// O_APPEND - дописывать в конец
		// O_CREATE - создать если не существует
		// O_WRONLY - только запись
		// Права 0644: владелец (rw), группа (r), остальные (r)
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("ошибка открытия файла логов: %v", err)
		}
		// Создаем экземпляр логгера
		instance = log.New(file, "BALANCER ", log.Ldate|log.Ltime|log.Lshortfile)
	})
	return instance
}

// Get возвращает инициализированный экземпляр логгера
func Get() *log.Logger {
	if instance == nil {
		log.Fatal("логгер не инициализирован")
	}
	return instance
}