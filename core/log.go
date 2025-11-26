package corepkg

import "os"

type Level int

const (
	LevelDebug Level = iota
	LevelFatal
	LevelError
	LevelWarning
	LevelInfo
)

// Logger is the interface that wraps basic logging methods.
type Logger interface {
	LogInfo(message ...string)
	LogInfof(format string, args ...any)
	LogWarning(err error)
	LogWarningf(format string, args ...any)
	LogError(err error, msg ...string) error
	LogErrorf(err error, format string, args ...any) error
	LogFatal(err error)
	LogFatalf(format string, args ...any)
}

var logger Logger = NewStandardLogger(LevelInfo)

func SetLogger(l Logger) {
	logger = l
}

func LogInf(message ...string) {
	logger.LogInfo(message...)
}

func LogInfo(message ...string) {
	logger.LogInfo(message...)
}

func LogInff(format string, args ...any) {
	logger.LogInfof(format, args...)
}

func LogInfof(format string, args ...any) {
	logger.LogInfof(format, args...)
}

func LogWarn(err error) {
	logger.LogWarning(err)
}

func LogWarning(err error) {
	logger.LogWarning(err)
}

func LogWarnf(format string, args ...any) {
	logger.LogWarningf(format, args...)
}
func LogWarningf(format string, args ...any) {
	logger.LogWarningf(format, args...)
}

func LogErr(err error, msg ...string) error {
	return logger.LogError(err, msg...)
}
func LogError(err error, msg ...string) error {
	return logger.LogError(err, msg...)
}

func LogErrf(err error, format string, args ...any) error {
	return logger.LogErrorf(err, format, args...)
}
func LogErrorf(err error, format string, args ...any) error {
	return logger.LogErrorf(err, format, args...)
}

func LogFatal(err error) {
	logger.LogFatal(err)
	os.Exit(1)
}

func LogFatalf(format string, args ...any) {
	logger.LogFatalf(format, args...)
	os.Exit(1)
}

type NillLogger struct {
}

func NewNillLogger() Logger {
	return &NillLogger{}
}

func (log *NillLogger) LogInfo(message ...string) {
}

func (log *NillLogger) LogInfof(format string, args ...any) {
}

func (log *NillLogger) LogWarning(err error) {
}

func (log *NillLogger) LogWarningf(format string, args ...any) {
}

func (log *NillLogger) LogError(err error, msg ...string) error {
	return err
}

func (log *NillLogger) LogErrorf(err error, format string, args ...any) error {
	return err
}

func (log *NillLogger) LogFatal(err error) {
}

func (log *NillLogger) LogFatalf(format string, args ...any) {
}
