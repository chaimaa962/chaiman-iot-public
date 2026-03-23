package utils

import (
    "log"
    "os"
)

type Logger struct {
    info  *log.Logger
    error *log.Logger
}

var logger *Logger

func InitLogger() {
    logger = &Logger{
        info:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
        error: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
    }
}

func GetLogger() *Logger {
    if logger == nil {
        InitLogger()
    }
    return logger
}

func (l *Logger) Info(msg string, keys ...interface{}) {
    l.info.Println(msg, keys)
}

func (l *Logger) Error(msg string, keys ...interface{}) {
    l.error.Println(msg, keys)
}
func (l *Logger) Warn(msg string, keys ...interface{}) {
    l.info.Println("WARN: "+msg, keys)  // Ou utilise error selon ton besoin
}
