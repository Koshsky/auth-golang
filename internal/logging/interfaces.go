package logger

//go:generate mockery --name=Logger --output=./mocks --outpkg=mocks --filename=Logger.go
type Logger interface {
	Log(message string, params ...interface{})
}

type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

//go:generate mockery --name=StructuredLogger --output=./mocks --outpkg=mocks --filename=StructuredLogger.go
type StructuredLogger interface {
	Logger

	LogWithLevel(level LogLevel, message string, params ...interface{})
	Debug(message string, params ...interface{})
	Info(message string, params ...interface{})
	Warn(message string, params ...interface{})
	Error(message string, params ...interface{})
	Fatal(message string, params ...interface{})
}
