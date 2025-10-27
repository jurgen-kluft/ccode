package corepkg

const (
	// KeyKey placeholder will be formatted to map key
	KeyKey = "key"
	// KeyValue placeholder will be formatted to map value
	KeyValue = "value"
)

// MapToString - format map keys and values according to format, joining parts with separator.
// Format should contain key and value placeholders which will be used for formatting, e.g.
// "{key} : {value}", or "{value}", or "{key} => {value}".
// Parts order in resulting string is not guranteed.
func MapToString[K string | int | uint | int32 | int64 | uint32 | uint64, V any](data map[K]V, format string, separator string, sb *StringBuilder) {
	if len(data) == 0 {
		return
	}
	sb.Grow(len(data)*len(format)*2 + (len(data)-1)*len(separator))

	isFirst := true
	for k, v := range data {
		if !isFirst {
			sb.WriteString(separator)
		}
		sb.FormatComplex(string(format), map[string]any{
			KeyKey:   k,
			KeyValue: v,
		})
		isFirst = false
	}
}

func KeyValueToString[K ~string | ~int | ~uint | ~int32 | ~int64 | ~uint32 | ~uint64, V any](key K, value V, format string, sb *StringBuilder) {
	sb.Grow(len(format) * 2)
	sb.FormatComplex(string(format), map[string]any{KeyKey: key, KeyValue: value})
}

func KeyValueToStringAppend[K ~string | ~int | ~uint | ~int32 | ~int64 | ~uint32 | ~uint64, V any](key K, value V, format string, sb *StringBuilder) {
	sb.Grow(sb.Len() + len(format)*2)
	sb.FormatComplex(string(format), map[string]any{KeyKey: key, KeyValue: value})
}
