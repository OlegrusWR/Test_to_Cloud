package ratelimiter

import (
	"github.com/OlegrusWR/balancer_to_cloud/internal/logger"
	"sync"
	"os"
	"log"
)

var (

rlOnce sync.Once
rlInstance logger.Logger

)

func InitLogger(logFile string) logger.Logger {
    rlOnce.Do(func() {
        file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            logger.Get().Fatalf("ошибка открытия файла логов Rate Limiter: %v", err)
        }
        
        rlInstance = &logger.StdLogger{
            Logger: log.New(file, "RATELIMITER ", log.Ldate|log.Ltime|log.Lshortfile),
        }
    })
    return rlInstance
}

func GetLogger() logger.Logger {
    if rlInstance == nil {
        logger.Get().Fatal("логгер Rate Limiter не инициализирован")
    }
    return rlInstance
}

