package foundation

type Arguments struct {
	Args []string
}

func NewArguments(init int) *Arguments {
	init = max(init, 16)
	return &Arguments{Args: make([]string, 0, init)}
}

func (a *Arguments) Clear() {
	a.Args = a.Args[:0]
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
