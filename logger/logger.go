package logger

import(
	"log"
	"os"
	"sync"
)

var (
	once sync.Once 
	instance *log.Logger
)

func Init(logFile string) *log.Logger {	
	once.Do(func() {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("ошибка открытия логгер файла", err)
		}
		instance = log.New(file, "BALANCER", log.Ldate|log.Ltime|log.Lshortfile)
	})
	return instance
}

func Get() *log.Logger{
	if instance == nil {
		log.Fatal("Логгер не инициализирован")
	}
	return instance
}