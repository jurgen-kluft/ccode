package corepkg

import (
	"fmt"
	"strings"
)

/*
`Flag` can store 32 true/false (or on/off) values.

# Example:

	const (
		FlagA flag.Flag = 1 << iota // 0001
		FlagB 			   // 0010
		FlagC 			   // 0100
	)

	var f flag.Flag = 0
	f.Set(FlagB)
	fmt.Print(f.String())   // prints binary "0010"
	f.ClearAll()            // flag == 0 (no flags set on the "f" Flag variable)
	f.Toggle(FlagC, FlagA)  // you can pass multiple flags to Toggle
	f.IsSet(FlagA)          // now returns true because of the line above
*/

type Flag uint32

type BitFlags interface {
	Set(flags ...Flag)
	SetAll()
	Bits() uint32
	Has(flags ...Flag) bool
	Only(flags ...Flag) bool
	Toggle(flags ...Flag)
	ToggleAll()
	Clear(flags ...Flag)
	ClearAll()
	String() string
	EncodeJSON(encoder *JsonEncoder) error
	DecodeJSON(decoder *JsonDecoder) error
}

type BitFlagsInstance struct {
	Value Flag             // the underlying value of the flag
	Flags *map[string]Flag // a slice of BitFlag for named flags
}

// # New returns a Flag variable initialized to zero.
func NewBitFlags(value Flag, flags *map[string]Flag) BitFlags {
	return &BitFlagsInstance{
		Value: value,
		Flags: flags,
	}
}

// String returns the formatted string
func (b *BitFlagsInstance) String() string {
	if b.Flags == nil {
		return fmt.Sprintf("%032b", b.Value)
	}

	result := ""
	bits := b.Value
	for name, bit := range *b.Flags {
		if bits == 0 {
			break
		} else if bits&bit != 0 {
			bits &^= bit
			if len(result) > 0 {
				result += "|"
			}
			result += name
		}
	}
	return result
}

// Set sets one or more given flag(s) to be true/on
func (b *BitFlagsInstance) Set(flags ...Flag) {
	for _, flag := range flags {
		b.Value |= flag
	}
}

// SetAll sets all known bits
func (b *BitFlagsInstance) SetAll() {
	if b.Flags == nil {
		b.Value = 0xFFFFFFFF
	}
	b.Value = 0
	for _, flag := range *b.Flags {
		b.Value |= flag
	}
}

func (b *BitFlagsInstance) Bits() uint32 {
	return uint32(b.Value)
}

// Toggle toggles the provided flag, if the provided flag is off, Toggle turns it on.
func (b *BitFlagsInstance) Toggle(flag ...Flag) {
	for _, f := range flag {
		b.Value ^= f
	}
}

// ToggleAll toggles every bit/flag
func (b *BitFlagsInstance) ToggleAll() {
	b.Value = ^b.Value
}

// Clear sets one or more provided flags to `0` (false/off)
func (b *BitFlagsInstance) Clear(flag ...Flag) {
	for _, f := range flag {
		b.Value &^= f
	}
}

// ClearAll sets all bits back to `0` (all false/off)
func (b *BitFlagsInstance) ClearAll() {
	b.Value = 0
}

func (b *BitFlagsInstance) Has(flags ...Flag) bool {
	for _, flag := range flags {
		if b.Value&flag != flag {
			return false
		}
	}
	return true
}

func (b *BitFlagsInstance) Only(flags ...Flag) bool {
	all := Flag(0)
	for _, flag := range flags {
		all |= flag
	}
	return (b.Value & all) == all
}

// EncodeJSON returns a string representation of the active flags in JSON format.
// e.g. `{"flags": "FlagA|FlagB"}` if FlagA and FlagB are set.
func (f *BitFlagsInstance) EncodeJSON(encoder *JsonEncoder) error {
	if f.Flags == nil {
		encoder.WriteString(fmt.Sprintf("%d", f.Value))
		return nil
	}

	result := ""
	bits := f.Value
	for name, flag := range *f.Flags {
		if bits == 0 {
			break
		} else if bits&flag != 0 {
			bits &^= flag
			if len(result) > 0 {
				result += "|"
			}
			result += name
		}
	}
	encoder.WriteString(result)
	return nil
}

func (f *BitFlagsInstance) DecodeJSON(decoder *JsonDecoder) (err error) {
	f.Value = 0

	flagsStr := decoder.DecodeString()
	if f.Flags == nil {
		// If no flags are declared, just parse it as an integer value
		if _, err := fmt.Sscanf(flagsStr, "%d", &f.Value); err != nil {
			return fmt.Errorf("error parsing flags as integer: %v", err)
		}
		return nil
	}

	// Parse the string representation of flags
	for len(flagsStr) > 0 {
		if flagsStr[0] == '|' {
			flagsStr = flagsStr[1:]
		}
		index := strings.IndexByte(flagsStr, '|')
		if index == -1 {
			index = len(flagsStr)
		}
		flagName := strings.TrimSpace(flagsStr[:index])
		if flag, exists := (*f.Flags)[flagName]; exists {
			f.Value |= flag
		}
		flagsStr = flagsStr[index:]
	}

	return nil
}
