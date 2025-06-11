/*
# Provides a simple `Flag` (uint32) type that can store 32 true-false values.

Provides variadic forms of functions that end in "V", but these are not as efficient as their normal counterparts.

In other words, `func (b *Flag) Clear(flag Flag)` is preferrable to `func (b *Flag) ClearV(flags ...Flag)`.

But, you can use the variadic ones if you want and are okay with a little slice allocation overhead (which isn't much because these are uint32 anyways).

# Example:

	import (
		"github.com/chasecarlson1/go-bitflags/flag"
	)

	const (
		FlagA flag.Flag = 1 << iota // 0001
		FlagB 			   // 0010
		FlagC 			   // 0100
	)

	var f flag.Flag = 0
	f.Set(FlagB)
	fmt.Print(f.String()) // prints binary "0010"
	f.ClearAll() // flag == 0 (no flags set on the "f" Flag variable)
	f.ToggleV(FlagC, FlagA) // variadic forms of functions end in "V". These have performance overhead though but allow for multiple arguments.
	f.IsSet(FlagA) // now returns true because of the line above
*/
package foundation

import (
	"encoding/json"
	"fmt"
	"strings"
)

/*
`Flag` can store 32 true/false (or on/off) values.

# Example:

	import (
		"github.com/chasecarlson1/go-bitflags/flag"
	)

	const (
		FlagA flag.Flag = 1 << iota // 0001
		FlagB 			   // 0010
		FlagC 			   // 0100
	)

	var f flag.Flag = 0
	f.Set(FlagB)
	fmt.Print(f.String()) // prints binary "0010"
	f.ClearAll() // flag == 0 (no flags set on the "f" Flag variable)
	f.ToggleV(FlagC, FlagA) // variadic forms of functions end in "V". These have performance overhead though but allow for multiple arguments.
	f.IsSet(FlagA) // now returns true because of the line above
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
	MarshalJSON() ([]byte, error)
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

// MarshalJSON returns a string representation of the active flags in JSON format.
// e.g. `{"flags": "FlagA|FlagB"}` if FlagA and FlagB are set.
func (f *BitFlagsInstance) MarshalJSON() ([]byte, error) {
	if f.Flags == nil {
		return []byte(fmt.Sprintf("%d", f.Value)), nil
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
	return []byte(fmt.Sprintf("\"%s\"", result)), nil
}

func UnmarshalBitFlagsFromJSON(data []byte, flags map[string]Flag) (value Flag, err error) {
    var flagsStr string
    if err := json.Unmarshal(data, &flagsStr); err != nil {
        return 0, fmt.Errorf("error unmarshalling flags: %v", err)
    }

    value = 0 // Reset the value

    if flags == nil {
        // If no flags are declared, just parse the integer value
        if _, err := fmt.Sscanf(flagsStr, "%d", &value); err != nil {
            return 0, fmt.Errorf("error parsing flags as integer: %v", err)
        }
        return value, nil
    }

    // Parse the string representation of flags
    flagStrs := strings.Split(flagsStr, "|")
    for _, flagName := range flagStrs {
        flagName = strings.TrimSpace(flagName)
        if flag, exists := flags[flagName]; exists {
            value |= flag
        }
    }

    return value, nil
}
