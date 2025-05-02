package logger

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface{
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatal(v ...interface{})
}

type StdLogger struct{
	*log.Logger
}

func (s *StdLogger) Printf(format string, v ...interface{}){
	s.Logger.Printf(format, v ...)
}

func (s *StdLogger) Println(v ...interface{}){
	s.Logger.Println(v ...)
}

func (s *StdLogger) Fatalf(format string, v ...interface{}){
	s.Logger.Fatalf(format, v ...)
}

func (s *StdLogger) Fatal(v ...interface{}){
	s.Logger.Fatal(v ...)
}

type stdLogger struct {
	*log.Logger
}

var (
	loggers = make(map[string]Logger)
	mu      sync.Mutex
)

func NewLogger(name, filename string) Logger {
	mu.Lock()
	defer mu.Unlock()

	if lg, exists := loggers[name]; exists {
		return lg
	}

	// Создаем директорию для логов если ее нет
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		log.Fatalf("Не удалось создать директорию для логов: %v", err)
	}

	// Настройка ротации логов
	logWriter := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    10,    // Максимальный размер в мегабайтах
		MaxBackups: 3,     // Максимальное количество старых логов
		MaxAge:     30,    // Максимальное количество дней хранения
		Compress:   true,  // Сжатие старых логов
	}

	lg := &stdLogger{
		Logger: log.New(logWriter, name+" ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	loggers[name] = lg
	return lg
}