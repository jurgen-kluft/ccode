package foundation

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
		"BoolField":    func(decoder *JsonDecoder) { object.BoolField = decoder.DecodeBool() },
		"IntField":     func(decoder *JsonDecoder) { object.IntField = int(decoder.DecodeInt32()) },
		"StringField":  func(decoder *JsonDecoder) { object.StringField = decoder.DecodeString() },
		"FloatField":   func(decoder *JsonDecoder) { object.FloatField = decoder.DecodeFloat64() },
		"ArrayField":   func(decoder *JsonDecoder) { object.ArrayField = decoder.DecodeStringArray() },
		"MapField":     func(decoder *JsonDecoder) { object.MapField = decoder.DecodeStringMapString() },
		"NestedObject": func(decoder *JsonDecoder) { object.NestedObject = DecodeTestObject(decoder, &JsonTestObject{}) },
	}
	decoder.Decode(fields)
	return object
}

func TestDecoding(t *testing.T) {
	json := `{
		"BoolField": true,
		"IntField": 42,
		"StringField": "Hello, World!",
		"FloatField": 3.14,
		"ArrayField": ["item1", "item2", "item3"],
		"MapField": {"key1": "value1", "key2": "value2"},
		"NestedObject": {
			"BoolField": false,
			"IntField": 7,
			"StringField": "Nested",
			"FloatField": 1.618,
			"ArrayField": ["nested1", "nested2"],
			"MapField": {"nestedKey1": "nestedValue1"}
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
