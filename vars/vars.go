package vars

import (
	"fmt"
	"strings"
)

// Replacer is providing functionality to replace a variable in a large body of text
type Replacer interface {
	Replace(variable string, replacement string, body string) string // Replaces occurences of @variable with @replace and thus will remove the variable from @body
	Insert(variable string, insertment string, body string) string   //Inserts @variable at places where @variable occurs without removing the variable in @body
}

type basicReplacer struct {
}

// NewReplacer will return an instance of a Replacer object
func NewReplacer() Replacer {
	return &basicReplacer{}
}

func (v *basicReplacer) Replace(variable string, replacement string, body string) string {
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

func (v *basicReplacer) Insert(variable string, insertment string, body string) string {
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

// Variables is a container for variables (key, value)
type Variables interface {
	AddVar(key string, value string)
	GetVar(key string) (string, error)
	Replace(replacer Replacer, body string) string
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

func (v *basicVariables) Replace(replacer Replacer, body string) string {
	for k, v := range v.vars {
		body = replacer.Replace(k, v, body)
	}
	return body
}
