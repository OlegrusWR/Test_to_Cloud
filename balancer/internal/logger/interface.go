package logger

import "log"

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