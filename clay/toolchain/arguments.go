package toolchain

type Arguments struct {
	List []string
	Fix  int
}

func NewArguments(init int) Arguments {
	if init <= 3 {
		init = 4
	}
	return Arguments{List: make([]string, 0, init)}
}

func (a Arguments) Clear() {
	a.List = a.List[:0]
	a.Fix = 0
}

func (a Arguments) Add(arg ...string) {
	a.List = append(a.List, arg...)
}

func (a Arguments) Lock() {
	a.Fix = len(a.List)
}

func (a Arguments) Reset() {
	a.List = a.List[:a.Fix]
}
