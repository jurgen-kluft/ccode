package ccode_gen

type Boolean bool

func (b Boolean) YesNo() string {
	if b.Bool() {
		return "YES"
	}
	return "NO"
}
func (b Boolean) TrueFalse() string {
	if b.Bool() {
		return "true"
	}
	return "false"
}
func (b Boolean) Int0or1() int {
	if b.Bool() {
		return 1
	}
	return 0
}
func (b Boolean) Bool() bool { return bool(b) }
func (b Boolean) Set(v bool) { b = Boolean(v) }
func (b Boolean) Get() bool  { return bool(b) }
