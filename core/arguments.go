package corepkg

import "strings"

type Arguments struct {
	Args []string
}

func NewArguments(init int) *Arguments {
	init = max(init, 4)
	return &Arguments{Args: make([]string, 0, init)}
}

func NewArgumentsFromCmdline(cmdline string, removeEmptyEntries bool) *Arguments {
	var args []string
	for len(cmdline) > 0 {
		i := 0
		for i < len(cmdline) && cmdline[i] == ' ' {
			i++
		}
		cmdline = cmdline[i:] // Remove leading spaces
		if cmdline[0] == '"' {
			// Find the closing quote
			endQuote := strings.Index(cmdline[1:], "\"")
			if endQuote == -1 {
				// No closing quote found, return the original string
				args = append(args, cmdline)
				break
			}
			// Add the argument without the quotes
			if removeEmptyEntries && endQuote == 0 {
				// If we are removing empty entries, skip this argument
				cmdline = cmdline[endQuote+2:] // Move past the closing quote and space
				continue
			}
			args = append(args, cmdline[1:endQuote+1])
			cmdline = cmdline[endQuote+2:] // Move past the closing quote and space
		} else {
			// Find the next space
			nextSpace := strings.Index(cmdline, " ")
			if nextSpace == -1 {
				args = append(args, cmdline)
				break
			}
			// Add the argument before the space
			if removeEmptyEntries && nextSpace == 0 {
				cmdline = cmdline[nextSpace+1:] // Move past the space
				continue
			}
			args = append(args, cmdline[:nextSpace])
			cmdline = cmdline[nextSpace+1:] // Move past the space
		}
	}

	return &Arguments{Args: args}
}

func (a *Arguments) SmartSplit(s string) {
	strBegin := 0

	inQuotes := false
	escaped := false

	for i, c := range s {
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '"' {
			inQuotes = !inQuotes
			continue
		}
		if c == ' ' && !inQuotes {
			if strBegin < i {
				a.Args = append(a.Args, s[strBegin:i])
			}
			strBegin = i + 1
		}
	}
	if strBegin < len(s) {
		a.Args = append(a.Args, s[strBegin:])
	}
}

func (a *Arguments) Len() int {
	return len(a.Args)
}

func (a *Arguments) Clear() {
	a.Args = a.Args[:0]
}

func (a *Arguments) Truncate(len int) {
	if len < cap(a.Args) {
		a.Args = a.Args[:len]
	} else {
		a.Args = a.Args[:0]
	}
}

func (a *Arguments) Add(arg ...string) {
	a.Args = append(a.Args, arg...)
}

func (a *Arguments) AddWithPrefix(prefix string, args ...string) {
	for _, arg := range args {
		a.Args = append(a.Args, prefix+arg)
	}
}

func (a *Arguments) AddWithFunc(modFunc func(string) string, args ...string) {
	for _, arg := range args {
		a.Args = append(a.Args, modFunc(arg))
	}
}

func (a *Arguments) String() string {
	return strings.Join(a.Args, "\n")
}
