package vars

import (
	"fmt"
	"strings"
)

// Replacer is providing functionality to replace a variable in a large body of text
type Replacer interface {
	ReplaceInLine(variable string, replacement string, body string) string // Replaces occurences of @variable with @replace and thus will remove the variable from @body
	ReplaceInLines(variable string, replacement string, lines []string)    // Replaces occurences of @variable with @replace and thus will remove the variable from @body
	InsertInLine(variable string, insertment string, line string) string   //Inserts @variable at places where @variable occurs without removing the variable in @body
	InsertInLines(variable string, insertment string, lines []string)      //Inserts @variable at places where @variable occurs without removing the variable in @body
}

type basicReplacer struct {
}

// NewReplacer will return an instance of a Replacer object
func NewReplacer() Replacer {
	return &basicReplacer{}
}

func (v *basicReplacer) ReplaceInLine(variable string, replacement string, body string) string {
	for true {
		n := strings.Count(body, variable)
		if n > 0 {
			body = strings.Replace(body, variable, replacement, n)
		} else {
			break
		}
	}
	return body
}

func (v *basicReplacer) ReplaceInLines(variable string, replacement string, lines []string) {
	for i, line := range lines {
		lines[i] = v.ReplaceInLine(variable, replacement, line)
	}
}

func (v *basicReplacer) InsertInLine(variable string, insertment string, body string) string {
	for true {
		n := strings.Count(body, variable)
		if n > 0 {
			pos := strings.Index(body, variable)
			if pos == 0 {
				body = insertment + body[pos:]
			} else if pos > 0 {
				body = body[0:pos] + insertment + body[pos:]
			}
		} else {
			break
		}
	}
	return body
}
func (v *basicReplacer) InsertInLines(variable string, replacement string, lines []string) {
	for i, line := range lines {
		lines[i] = v.InsertInLine(variable, replacement, line)
	}
}

// Variables is a container for variables (key, value)
type Variables interface {
	AddVar(key string, value string)
	GetVar(key string) (string, error)
	ReplaceInLine(replacer Replacer, line string) string
	ReplaceInLines(replacer Replacer, lines []string)
}

type basicVariables struct {
	vars map[string]string
}

// NewVars returns a new instance of Variables based on variables of the format ${VARIABLE}
func NewVars() Variables {
	return &basicVariables{vars: make(map[string]string)}
}

func (v *basicVariables) AddVar(key string, value string) {
	if strings.HasPrefix(key, "${") == false {
		key = fmt.Sprintf("${%s}", key)
	}
	v.vars[key] = value
}

func (v *basicVariables) GetVar(key string) (string, error) {
	if strings.HasPrefix(key, "${") == false {
		key = fmt.Sprintf("${%s}", key)
	}
	if value, ok := v.vars[key]; ok {
		return value, nil
	}
	return "", fmt.Errorf("Variables doesn't contain var with key %s", key)
}

func (v *basicVariables) ReplaceInLine(replacer Replacer, line string) string {
	for k, v := range v.vars {
		line = replacer.ReplaceInLine(k, v, line)
	}
	return line
}

func (v *basicVariables) ReplaceInLines(replacer Replacer, lines []string) {
	for i, line := range lines {
		lines[i] = v.ReplaceInLine(replacer, line)
	}
}
