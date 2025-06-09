package ccode_utils

import "os"

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarning
	LevelError
)

var LogLevel Level = LevelInfo

func LogPrint(message string) {
}

func LogPrintf(format string, args ...any) {
}

func LogPrintln(message ...string) {
}

func LogInf(message string) {
}
func LogInfo(message string) {
}

func LogInff(format string, args ...any) {
}
func LogInfof(format string, args ...any) {
}

func LogWarn(err error) {
}
func LogWarning(err error) {
}

func LogWarnf(format string, args ...any) {
}
func LogWarningf(format string, args ...any) {
}

func LogErr(err error, msg ...string) error {
	return err
}
func LogError(err error, msg ...string) error {
	return err
}

func LogErrf(err error, format string, args ...any) error {
	return err
}
func LogErrorf(err error, format string, args ...any) error {
	return err
}

func LogFatal(err error) {
}

func LogFatalf(format string, args ...any) {

	os.Exit(1)
}
