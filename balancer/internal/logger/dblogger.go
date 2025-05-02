package logger

import (
	"log"
	"os"
	"sync"
)

var (
	dbOnce     sync.Once
	dbInstance Logger
)


func InitDBLogger(logFile string) Logger {
    rlOnce.Do(func() {
        file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            DBGet().Fatalf("ошибка открытия файла логов DB: %v", err)
        }
        
        rlInstance = &StdLogger{
            Logger: log.New(file, "DATABASE ", log.Ldate|log.Ltime|log.Lshortfile),
        }
    })
    return rlInstance
}

func DBGet() Logger {
	if instance == nil {
		log.Fatal("логгер не инициализирован")
	}
	return instance
}

func InitDataBaseLogger(logfile string) Logger {
	logger := InitDBLogger("db.log")
	logger.Println("Database logger initialized")
	return logger
}