package logging

import (
	"fmt"
	"io"
	"log"
	"runtime"
)

// GetLongFileName gets filename that called the log function in long format.
// If it is unable to capture the calling filename, it returns an empty string.
func GetLongFileName() string {
	_, file, line, ok := runtime.Caller(2)
	longfile := ""
	if ok {
		longfile = fmt.Sprintf("%s:%d:", file, line)
	}

	return longfile
}

// Logger logs at debug, info, warn, and error levels.
type Logger interface {
	LogDebug(v ...any)
	LogInfo(v ...any)
	LogWarn(v ...any)
	LogError(v ...any)
}

// logger implements Logger.
type logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
}

// NewLogger returns a new Logger.
func NewLogger(stdout io.Writer, stderr io.Writer) *logger {
	return &logger{
		debugLogger: log.New(stdout, "[D] ", log.Ldate|log.Ltime|log.LUTC),
		infoLogger:  log.New(stdout, "[I] ", log.Ldate|log.Ltime|log.LUTC),
		warnLogger:  log.New(stdout, "[W] ", log.Ldate|log.Ltime|log.LUTC),
		errorLogger: log.New(stderr, "[E] ", log.Ldate|log.Ltime|log.LUTC),
	}
}

// LogDebug logs at debug level.
func (l *logger) LogDebug(v ...any) {
	longfile := GetLongFileName()
	l.debugLogger.Println(append([]any{longfile}, v...)...)
}

// LogInfo logs at info level.
func (l *logger) LogInfo(v ...any) {
	longfile := GetLongFileName()
	l.infoLogger.Println(append([]any{longfile}, v...)...)
}

// LogWarn logs at warn level.
func (l *logger) LogWarn(v ...any) {
	longfile := GetLongFileName()
	l.warnLogger.Println(append([]any{longfile}, v...)...)
}

// LogError logs at error level.
func (l *logger) LogError(v ...any) {
	longfile := GetLongFileName()
	l.errorLogger.Println(append([]any{longfile}, v...)...)
}
