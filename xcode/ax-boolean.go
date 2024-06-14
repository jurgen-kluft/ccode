package xcode

var BooleanToYesNo = map[bool]string{
	true:  "YES",
	false: "NO",
}

var BooleanToTrueFalse = map[bool]string{
	true:  "true",
	false: "false",
}

var BooleanTo01 = map[bool]int{
	true:  1,
	false: 0,
}

type Boolean bool

func (b Boolean) YesNo() string     { return BooleanToYesNo[bool(b)] }
func (b Boolean) TrueFalse() string { return BooleanToTrueFalse[bool(b)] }
func (b Boolean) Int0or1() int      { return BooleanTo01[bool(b)] }
func (b Boolean) Bool() bool        { return bool(b) }
func (b Boolean) Set(v bool)        { b = Boolean(v) }
func (b Boolean) Get() bool         { return bool(b) }
