// Package logger 
package logger

import (
    "io"
    "log"
    "os"
)

// Init настраивает вывод логов в консоль и файл
func Init(logFile string) {
    f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal("Cannot open log file:", err)
    }
    multi := io.MultiWriter(os.Stdout, f)
    log.SetOutput(multi)
}