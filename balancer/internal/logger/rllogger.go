package logger

import (
	"sync"
	"os"
	"log"
)

var (

rlOnce sync.Once
rlInstance Logger

)

func InitRLLogger(logFile string) Logger {
    rlOnce.Do(func() {
        file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            RLGet().Fatalf("ошибка открытия файла логов Rate Limiter: %v", err)
        }
        
        rlInstance = &StdLogger{
            Logger: log.New(file, "RATELIMITER ", log.Ldate|log.Ltime|log.Lshortfile),
        }
    })
    return rlInstance
}

func RLGet() Logger {
	if instance == nil {
		log.Fatal("логгер не инициализирован")
	}
	return instance
}

func InitRateLimitLogger(logfile string) Logger {
	logger := InitRLLogger("ratelimit.log")
	logger.Println("Rate Limiter логгер инициализирован")
	return logger
}

