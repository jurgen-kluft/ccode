package corepkg

import "fmt"

type StandardLogger struct {
	LogLevel Level
}

func NewStandardLogger(level Level) *StandardLogger {
	return &StandardLogger{
		LogLevel: level,
	}
}

func (log *StandardLogger) LogInfo(message ...string) {
	if LevelInfo <= log.LogLevel {
		if len(message) > 0 {
			fmt.Print("INFO: ")
			for _, m := range message {
				fmt.Print(m)
			}
		}
		fmt.Println()
	}
}

func (log *StandardLogger) LogInfof(format string, args ...any) {
	if LevelInfo <= log.LogLevel {
		fmt.Printf("INFO: ")
		fmt.Printf(format, args...)
		fmt.Println()
	}
}

func (log *StandardLogger) LogWarning(err error) {
	if LevelWarning <= log.LogLevel {
		fmt.Print("WARNING: ", err.Error())
		fmt.Println()
	}
}

func (log *StandardLogger) LogWarningf(format string, args ...any) {
	if LevelWarning <= log.LogLevel {
		fmt.Print("WARNING: ")
		fmt.Printf(format, args...)
		fmt.Println()
	}
}

func (log *StandardLogger) LogError(err error, msg ...string) error {
	if LevelError <= log.LogLevel {
		if len(msg) > 0 {
			fmt.Print("ERROR: ")
			for _, m := range msg {
				fmt.Print(m, " ")
			}
			fmt.Print(err.Error())
		} else {
			fmt.Print("ERROR: ", err.Error())
		}
		fmt.Println()
	}
	return err
}

func (log *StandardLogger) LogErrorf(err error, format string, args ...any) error {
	if LevelError <= log.LogLevel {
		fmt.Print("ERROR: ")
		fmt.Printf(format, args...)
		if err != nil {
			fmt.Print(" - ", err.Error())
		}
		fmt.Println()
	}
	return err
}

func (log *StandardLogger) LogFatal(err error) {
	panic(err)
}

func (log *StandardLogger) LogFatalf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}
