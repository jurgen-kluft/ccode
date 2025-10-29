package corepkg

// SliceToString function that converts slice of any type items to string in format {item}{sep}{item}...
// TODO(UMV): probably add one more param to wrap item in quotes if necessary
func (sb *StringBuilder) SliceToString(data *[]any, separator *string) {
	if len(*data) == 0 {
		return
	}
	sb.Grow(len(*data)*len(*separator)*2 + (len(*data)-1)*len(*separator))
	sb.Format("{0}", (*data)[0])
	for _, item := range (*data)[1:] {
		sb.WriteString(*separator)
		sb.Format("{0}", item)
	}
}

func SliceSameTypeToString[T any](sb *StringBuilder, data *[]T, separator *string) {
	if len(*data) == 0 {
		return
	}
	sb.Grow(len(*data)*len(*separator)*2 + (len(*data)-1)*len(*separator))
	sb.Format("{0}", (*data)[0])
	for _, item := range (*data)[1:] {
		sb.WriteString(*separator)
		sb.Format("{0}", item)
	}
}
