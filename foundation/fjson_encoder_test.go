package foundation

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
	object := &JsonTestObject2{
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
	{
		encodeTestObject2(encoder, key, object)
	}
	json := encoder.End()

	if len(json) == 0 {
		t.Error("Failed to encode JSON")
	}

}
