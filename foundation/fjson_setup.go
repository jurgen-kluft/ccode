package foundation

type JsonObject struct {
	Name   string
	Fields map[string]JsonDecode
}

type JsonDecode func(decoder *JsonDecoder)

type JsonField struct {
	Name    string
	Decoder func(decoder *JsonDecoder)
}

type JsonDecoder struct {
	Err error
}

func (jd *JsonDecoder) HasError() bool {
	return jd.Err != nil
}

func (jd *JsonDecoder) GetKey() string {
	return "" // Example return value
}

func (jd *JsonDecoder) GetValueAsString() string {
	return "" // Example return value
}

func (jd *JsonDecoder) Decode(root *JsonObject) error {
	return nil // Example return value
}

// DecodeArray decodes an array of system types from JSON, this is a generics version
func DecodeArray[T string | int | uint | int32 | int64 | uint32 | uint64](decoder *JsonDecoder) (result []T) {
	var value T

	// First case, array of strings
	switch any(value).(type) {
	case string:
		var str string
		var strArray []string
		for !decoder.UntilEndOfArray() && !decoder.HasError() {
			str = decoder.DecodeString()
			strArray = append(strArray, str)
		}
		return any(strArray).([]T)
	}

	return result
}

func DecodeMapStringString(decoder *JsonDecoder) (result map[string]string) {
	result = make(map[string]string)

	for !decoder.UntilEndOfObject() && !decoder.HasError() {
		key := decoder.GetKey()
		value := decoder.GetValueAsString()
		result[key] = value
	}
	return result
}

func (jd *JsonDecoder) UntilEndOfObject() bool {
	return false
}

func (jd *JsonDecoder) UntilEndOfArray() bool {
	return false
}

func (jd *JsonDecoder) DecodeBool() bool {
	return true // Example return value
}

func (jd *JsonDecoder) DecodeString() string {
	return "example" // Example return value
}

type JsonTestObject struct {
	BoolField    bool
	IntField     int
	StringField  string
	FloatField   float64
	ArrayField   []string
	MapField     map[string]string
	NestedObject *JsonTestObject
}

// This is a declarative layout that instructs the json decoder how to
// decode the JSON into the JsonTestObject struct.

func TestJsonTestObjectLayout(object *JsonTestObject) *JsonTestObject {

	var JsonTestObjectLayout = &JsonObject{
		Name: "JsonTestObject",
		Fields: map[string]JsonDecode{
			"BoolField":    func(decoder *JsonDecoder) { object.BoolField = decoder.DecodeBool() },
			"ArrayField":   func(decoder *JsonDecoder) { object.ArrayField = DecodeArray[string](decoder) },
			"MapField":     func(decoder *JsonDecoder) { object.MapField = DecodeMapStringString(decoder) },
			"NestedObject": func(decoder *JsonDecoder) { object.NestedObject = TestJsonTestObjectLayout(&JsonTestObject{}) },
		},
	}

	decoder := &JsonDecoder{}
	decoder.Decode(JsonTestObjectLayout)
	return object
}
