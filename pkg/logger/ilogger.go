package logger

type ILogger interface {
	Fatal(message string, err error)
	Panic(message string, err error)
	Error(message string, err error)
	Warning(message string)
	Info(message string)
	Debug(message string)
	Sync()
}
