// Package logger пакет для логирования событий в приложениях.
// Содержит интерфейс Logger и его реализацию ZapLogger
package logger

type Logger interface {
	Fatal(message string, err error)
	Panic(message string, err error)
	Error(message string, err error)
	Warning(message string)
	Info(message string)
	Debug(message string)
	Sync()
}
