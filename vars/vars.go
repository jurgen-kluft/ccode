package vars

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Replacer is providing functionality to replace a variable in a large body of text
type Replacer interface {
	ReplaceInLine(variable string, replacement string, line string) string // Replaces occurences of @variable with @replace and thus will remove the variable from @line
	ReplaceInLines(variable string, replacement string, lines []string)    // Replaces occurences of @variable with @replace and thus will remove the variable from @lines
	InsertInLine(variable string, insertment string, line string) string   //Inserts @variable at places where @variable occurs without removing the variable in @line
	InsertInLines(variable string, insertment string, lines []string)      //Inserts @variable at places where @variable occurs without removing the variable in @lines
}

type basicReplacer struct {
}

// NewReplacer will return an instance of a Replacer object
func NewReplacer() Replacer {
	return &basicReplacer{}
}

func (v *basicReplacer) ReplaceInLine(variable string, replacement string, line string) string {
	for true {
		n := strings.Count(line, variable)
		if n > 0 {
			line = strings.Replace(line, variable, replacement, n)
		} else {
			break
		}
	}
	return line
}

func (v *basicReplacer) ReplaceInLines(variable string, replacement string, lines []string) {
	for i, line := range lines {
		lines[i] = v.ReplaceInLine(variable, replacement, line)
	}
}

func (v *basicReplacer) InsertInLine(variable string, insertment string, line string) string {
	insertment = strings.Trim(insertment, " ")
	if len(insertment) == 0 {
		return line
	}
	if strings.HasSuffix(insertment, ";") == false {
		insertment = insertment + ";"
	}
	piv := -1
	pos := strings.Index(line, variable)
	for pos > piv {
		if pos == 0 {
			line = insertment + line
		} else if pos > 0 {
			line = line[0:pos] + insertment + line[pos:]
		} else {
			break
		}
		pos += len(insertment)
		pos += len(variable)
		piv = pos
		pos = strings.Index(line, variable)
	}
	return line
}
func (v *basicReplacer) InsertInLines(variable string, replacement string, lines []string) {
	for i, line := range lines {
		lines[i] = v.InsertInLine(variable, replacement, line)
	}
}

type VariablesMerge func(key string, value string, vars Variables)
type VariablesIter func(key string, value string)

// Variables is a container for variables (key, value)
type Variables interface {
	HasVar(key string) bool
	SetVar(key string, value string) error
	AddVar(key string, value string)
	GetVar(key string) (string, error)
	DelVar(key string) error
	Iterate(iter VariablesIter)
	ReplaceInLine(replacer Replacer, line string) string
	ReplaceInLines(replacer Replacer, lines []string)

	Copy() Variables
	Print()
}

type basicVariables struct {
	vars map[string]string
}

// NewVars returns a new instance of Variables based on variables of the format ${VARIABLE}
func NewVars() Variables {
	return &basicVariables{vars: make(map[string]string)}
}

func MakeVarKey(key string) string {
	if strings.HasPrefix(key, "${") == false {
		key = fmt.Sprintf("${%s}", key)
	}
	return key
}
func UnmakeVarKey(key string) string {
	if strings.HasPrefix(key, "${") {
		key = strings.Trim(key, "${}")
	}
	return key
}

func (v *basicVariables) HasVar(key string) bool {
	key = MakeVarKey(key)
	_, ok := v.vars[key]
	return ok
}

func (v *basicVariables) SetVar(key string, value string) error {
	key = MakeVarKey(key)
	_, ok := v.vars[key]
	if ok {
		v.vars[key] = value
		return nil
	}
	return errors.New("key doesn't exist in var map")
}

func (v *basicVariables) AddVar(key string, value string) {
	key = MakeVarKey(key)
	v.vars[key] = value
}

func (v *basicVariables) GetVar(key string) (string, error) {
	key = MakeVarKey(key)
	if value, ok := v.vars[key]; ok {
		return value, nil
	}
	return "", fmt.Errorf("Variables doesn't contain var with key %s", key)
}

func (v *basicVariables) DelVar(key string) error {
	key = MakeVarKey(key)
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
	sortedkeys := []string{}
	for k := range v.vars {
		sortedkeys = append(sortedkeys, k)
	}
	sort.Strings(sortedkeys)
	for _, key := range sortedkeys {
		value := v.vars[key]
		fmt.Printf("Var: %s = %s\n", key, value)
	}
}

func (v *basicVariables) Iterate(iter VariablesIter) {
	for k, v := range v.vars {
		uk := UnmakeVarKey(k)
		iter(uk, v)
	}
}

func (v *basicVariables) Copy() Variables {
	newvars := NewVars()
	for k, v := range v.vars {
		newvars.AddVar(k, v)
	}
	return newvars
}

func MergeVars(master Variables, other Variables, merger VariablesMerge) {
	iter := func(key, value string) {
		merger(key, value, master)
	}
	other.Iterate(iter)
}
