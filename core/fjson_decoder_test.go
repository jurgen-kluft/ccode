package corepkg

import "testing"

type JsonTestObject struct {
	BoolField    bool
	IntField     int
	StringField  string
	FloatField   float64
	ArrayField   []string
	MapField     map[string]string
	NestedObject *JsonTestObject
}

func DecodeTestObject(decoder *JsonDecoder, object *JsonTestObject) *JsonTestObject {
	fields := map[string]JsonDecode{
		"boolfield":    func(decoder *JsonDecoder) { object.BoolField = decoder.DecodeBool() },
		"intfield":     func(decoder *JsonDecoder) { object.IntField = int(decoder.DecodeInt32()) },
		"stringfield":  func(decoder *JsonDecoder) { object.StringField = decoder.DecodeString() },
		"floatfield":   func(decoder *JsonDecoder) { object.FloatField = decoder.DecodeFloat64() },
		"arrayfield":   func(decoder *JsonDecoder) { object.ArrayField = decoder.DecodeStringArray() },
		"mapfield":     func(decoder *JsonDecoder) { object.MapField = decoder.DecodeStringMapString() },
		"nestedobject": func(decoder *JsonDecoder) { object.NestedObject = DecodeTestObject(decoder, &JsonTestObject{}) },
	}
	decoder.Decode(fields)
	return object
}

func TestDecoding(t *testing.T) {
	json := `{
		"boolfield": true,
		"intfield": 42,
		"stringfield": "Hello, World!",
		"floatfield": 3.14,
		"arrayfield": ["item1", "item2", "item3"],
		"mapfield": {"key1": "value1", "key2": "value2"},
		"nestedobject": {
			"boolfield": false,
			"intfield": 7,
			"stringfield": "Nested",
			"floatfield": 1.618,
			"arrayfield": ["nested1", "nested2"],
			"mapfield": {"nestedKey1": "nestedValue1"}
		}
	}`

	decoder := NewJsonDecoder()
	if decoder.Begin(json) {
		object := DecodeTestObject(decoder, &JsonTestObject{})

		if !object.BoolField {
			t.Errorf("Expected BoolField to be true, got false")
		}
		if object.IntField != 42 {
			t.Errorf("Expected IntField to be 42, got %d", object.IntField)
		}
	} else {
		t.Error("Failed to begin decoding JSON")
	}
}
