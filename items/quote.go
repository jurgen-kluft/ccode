package items

import ()

type Quoter func(string, string, string) string

func NoQuoter(item string, pre string, post string) string {
	return item
}

func ItemQuoter(item string, pre string, post string) string {
	if len(item) == 0 {
		return item
	}
	return pre + item + post
}
