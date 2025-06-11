package clay

import "github.com/jurgen-kluft/ccode/foundation"

type IncludeMap = foundation.ValueSet

func NewIncludeMap() *IncludeMap {
	return foundation.NewValueSet()
}
