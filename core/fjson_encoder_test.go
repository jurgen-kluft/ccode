package corepkg

import (
	"testing"
)

type JsonTestObject2 struct {
	BoolField    bool
	IntField     int
	StringField  string
	FloatField   float64
	ArrayField   []string
	MapField     map[string]string
	NestedObject *JsonTestObject2
}

func decodeTestObject2(decoder *JsonDecoder, object *JsonTestObject2) *JsonTestObject2 {
	fields := map[string]JsonDecode{
		"boolfield":    func(decoder *JsonDecoder) { object.BoolField = decoder.DecodeBool() },
		"intfield":     func(decoder *JsonDecoder) { object.IntField = int(decoder.DecodeInt32()) },
		"stringfield":  func(decoder *JsonDecoder) { object.StringField = decoder.DecodeString() },
		"floatfield":   func(decoder *JsonDecoder) { object.FloatField = decoder.DecodeFloat64() },
		"arrayfield":   func(decoder *JsonDecoder) { object.ArrayField = decoder.DecodeStringArray() },
		"mapfield":     func(decoder *JsonDecoder) { object.MapField = decoder.DecodeStringMapString() },
		"nestedobject": func(decoder *JsonDecoder) { object.NestedObject = decodeTestObject2(decoder, &JsonTestObject2{}) },
	}
	decoder.Decode(fields)
	return object
}

func encodeTestObject2(encoder *JsonEncoder, key string, object *JsonTestObject2) {
	if object == nil {
		return
	}

	encoder.BeginObject(key)
	{
		encoder.WriteField("BoolField", object.BoolField)
		encoder.WriteField("IntField", object.IntField)
		encoder.WriteField("StringField", object.StringField)
		encoder.WriteField("FloatField", object.FloatField)
		encoder.BeginArray("ArrayField")
		{
			for _, item := range object.ArrayField {
				encoder.WriteArrayElement(item)
			}
		}
		encoder.EndArray()
		encoder.BeginMap("MapField")
		{
			for k, v := range object.MapField {
				encoder.WriteMapElement(k, v)
			}
		}
		encoder.EndMap()
		encodeTestObject2(encoder, "NestedObject", object.NestedObject)
	}
	encoder.EndObject()
}

func TestEncoding(t *testing.T) {
	encoder := NewJsonEncoder("    ")
	key := ""
	object1 := &JsonTestObject2{
		BoolField:   true,
		IntField:    42,
		StringField: "The Hitchhiker's Guide to the Galaxy",
		FloatField:  3.14159,
		ArrayField:  []string{"The answer", "to life", "the universe", "and everything"},
		MapField: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
		NestedObject: &JsonTestObject2{
			BoolField:   false,
			IntField:    7,
			StringField: "Nested Object",
			FloatField:  1.6180339887,
			ArrayField:  []string{"nested1", "nested2"},
			MapField: map[string]string{
				"nestedKey1": "nestedValue1",
				"nestedKey2": "nestedValue2",
			},
			NestedObject: nil, // No further nesting for this example
		},
	}

	encoder.Begin()
	encodeTestObject2(encoder, key, object1)
	json := encoder.End()

	if len(json) == 0 {
		t.Error("Failed to encode JSON")
	}

	decoder := NewJsonDecoder()
	if decoder.Begin(json) {
		object2 := decodeTestObject2(decoder, &JsonTestObject2{})

		if !object2.BoolField {
			t.Errorf("Expected BoolField to be true, got false")
		}
		if object2.IntField != 42 {
			t.Errorf("Expected IntField to be 42, got %d", object2.IntField)
		}

		// object1 and object2 should be equal
		if object1.BoolField != object2.BoolField {
			t.Errorf("Expected BoolField to be %v, got %v", object1.BoolField, object2.BoolField)
		}
		if object1.IntField != object2.IntField {
			t.Errorf("Expected IntField to be %d, got %d", object1.IntField, object2.IntField)
		}
		if object1.StringField != object2.StringField {
			t.Errorf("Expected StringField to be '%s', got '%s'", object1.StringField, object2.StringField)
		}
		if object1.FloatField != object2.FloatField {
			t.Errorf("Expected FloatField to be %f, got %f", object1.FloatField, object2.FloatField)
		}
		if len(object2.ArrayField) != len(object1.ArrayField) {
			t.Errorf("Expected ArrayField length to be %d, got %d", len(object1.ArrayField), len(object2.ArrayField))
		} else {
			for i, item := range object1.ArrayField {
				if item != object2.ArrayField[i] {
					t.Errorf("Expected ArrayField[%d] to be '%s', got '%s'", i, item, object2.ArrayField[i])
				}
			}
		}
		if len(object2.MapField) != len(object1.MapField) {
			t.Errorf("Expected MapField length to be %d, got %d", len(object1.MapField), len(object2.MapField))
		} else {
			for k, v := range object1.MapField {
				if object2.MapField[k] != v {
					t.Errorf("Expected MapField[%s] to be '%s', got '%s'", k, v, object2.MapField[k])
				}
			}
		}
		if object2.NestedObject == nil {
			t.Error("Expected NestedObject to not be nil")
		} else {
			if object2.NestedObject.BoolField {
				t.Errorf("Expected NestedObject.BoolField to be false, got %v", object2.NestedObject.BoolField)
			}
			if object2.NestedObject.IntField != 7 {
				t.Errorf("Expected NestedObject.IntField to be 7, got %d", object2.NestedObject.IntField)
			}
			if object2.NestedObject.StringField != "Nested Object" {
				t.Errorf("Expected NestedObject.StringField to be 'Nested Object', got '%s'", object2.NestedObject.StringField)
			}
			if object2.NestedObject.FloatField != 1.6180339887 {
				t.Errorf("Expected NestedObject.FloatField to be 1.6180339887, got %f", object2.NestedObject.FloatField)
			}
			if len(object2.NestedObject.ArrayField) != 2 || object2.NestedObject.ArrayField[0] != "nested1" || object2.NestedObject.ArrayField[1] != "nested2" {
				t.Errorf("Expected NestedObject.ArrayField to be ['nested1', 'nested2'], got %v", object2.NestedObject.ArrayField)
			}
			if len(object2.NestedObject.MapField) != 2 || object2.NestedObject.MapField["nestedKey1"] != "nestedValue1" || object2.NestedObject.MapField["nestedKey2"] != "nestedValue2" {
				t.Errorf("Expected NestedObject.MapField to be {'nestedKey1': 'nestedValue1', 'nestedKey2': 'nestedValue2'}, got %v", object2.NestedObject.MapField)
			}
		}
	} else {
		t.Error("Failed to begin decoding JSON")
	}

}
