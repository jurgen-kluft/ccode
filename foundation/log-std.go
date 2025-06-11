package foundation

import "fmt"

type StandardLogger struct {
	// LogLevel is the current logging level
	LogLevel Level
}

func NewStandardLogger(level Level) *StandardLogger {
	return &StandardLogger{
		LogLevel: level,
	}
}

func (log *StandardLogger) LogPrint(message string) {
	fmt.Print(message)
}

func (log *StandardLogger) LogPrintf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (log *StandardLogger) LogPrintln(message ...string) {
	for _, msg := range message {
		fmt.Println(msg)
	}
}

func (log *StandardLogger) LogInfo(message string) {
	if log.LogLevel <= LevelInfo {
		fmt.Print("INFO: ", message)
	}
}

func (log *StandardLogger) LogInfof(format string, args ...any) {
	if log.LogLevel <= LevelInfo {
		fmt.Printf("INFO: ")
		fmt.Printf(format, args...)
	}
}

func (log *StandardLogger) LogWarning(err error) {
	if log.LogLevel <= LevelWarning {
		fmt.Print("WARNING: ", err.Error())
	}
}

func (log *StandardLogger) LogWarningf(format string, args ...any) {
	if log.LogLevel <= LevelWarning {
		fmt.Print("WARNING: ")
		fmt.Printf(format, args...)
	}
}

func (log *StandardLogger) LogError(err error, msg ...string) error {
	if log.LogLevel <= LevelError {
		if len(msg) > 0 {
			fmt.Print("ERROR: ")
			for _, m := range msg {
				fmt.Print(m, " ")
			}
			fmt.Print(err.Error())
		} else {
			fmt.Print("ERROR: ", err.Error())
		}
	}
	return err
}

func (log *StandardLogger) LogErrorf(err error, format string, args ...any) error {
	if log.LogLevel <= LevelError {
		fmt.Print("ERROR: ")
		fmt.Printf(format, args...)
		if err != nil {
			fmt.Print(" - ", err.Error())
		}
	}
	return err
}

func (log *StandardLogger) LogFatal(err error) {
	panic(err)
}

func (log *StandardLogger) LogFatalf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}
