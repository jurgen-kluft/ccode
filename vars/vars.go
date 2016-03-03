package vars

import (
	"errors"
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
	piv := -1
	pos := strings.Index(body, variable)
	for pos > piv {
		if pos == 0 {
			body = insertment + body
		} else if pos > 0 {
			body = body[0:pos] + insertment + body[pos:]
		} else {
			break
		}
		pos += len(insertment)
		pos += len(variable)
		piv = pos
		pos = strings.Index(body, variable)
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
	HasVar(key string) bool
	SetVar(key string, value string) error
	AddVar(key string, value string)
	GetVar(key string) (string, error)
	DelVar(key string) error
	ReplaceInLine(replacer Replacer, line string) string
	ReplaceInLines(replacer Replacer, lines []string)
	Print()
}

type basicVariables struct {
	vars map[string]string
}

// NewVars returns a new instance of Variables based on variables of the format ${VARIABLE}
func NewVars() Variables {
	return &basicVariables{vars: make(map[string]string)}
}

func correctVarKey(key string) string {
	if strings.HasPrefix(key, "${") == false {
		key = fmt.Sprintf("${%s}", key)
	}
	return key
}

func (v *basicVariables) HasVar(key string) bool {
	key = correctVarKey(key)
	_, ok := v.vars[key]
	return ok
}

func (v *basicVariables) SetVar(key string, value string) error {
	key = correctVarKey(key)
	_, ok := v.vars[key]
	if ok {
		v.vars[key] = value
		return nil
	}
	return errors.New("key doesn't exist in var map")
}

func (v *basicVariables) AddVar(key string, value string) {
	key = correctVarKey(key)
	v.vars[key] = value
}

func (v *basicVariables) GetVar(key string) (string, error) {
	key = correctVarKey(key)
	if value, ok := v.vars[key]; ok {
		return value, nil
	}
	return "", fmt.Errorf("Variables doesn't contain var with key %s", key)
}

func (v *basicVariables) DelVar(key string) error {
	key = correctVarKey(key)
	if _, ok := v.vars[key]; ok {
		delete(v.vars, key)
		return nil
	}
	return errors.New("key doesn't exist in var map")
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

func (v *basicVariables) Print() {
	for k, v := range v.vars {
		fmt.Printf("Var: %s = %s\n", k, v)
	}
}
