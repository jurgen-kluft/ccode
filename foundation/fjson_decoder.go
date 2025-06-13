package foundation

type JsonDecode func(decoder *JsonDecoder)

type JsonDecoder struct {
	reader *JsonReader
	Err    error
}

func NewJsonDecoder(json string) *JsonDecoder {
	reader := NewJsonReader()
	reader.Begin(json)
	return &JsonDecoder{
		reader: reader,
		Err:    nil,
	}
}

func (jd *JsonDecoder) GetField() string {
	return jd.reader.FieldStr(jd.reader.Key)
}

func (jd *JsonDecoder) Decode(fields map[string]JsonDecode) error {
	for !jd.reader.ReadUntilObjectEnd() {
		fname := jd.GetField()
		if fdecode, ok := fields[fname]; ok {
			fdecode(jd)
		}
	}
	return nil
}

func (jd *JsonDecoder) DecodeBool() bool {
	return jd.reader.ParseBool(jd.reader.Value)
}

func (jd *JsonDecoder) DecodeInt32() int32 {
	return int32(jd.reader.ParseInt(jd.reader.Value))
}

func (jd *JsonDecoder) DecodeInt64() int64 {
	return jd.reader.ParseLong(jd.reader.Value)
}

func (jd *JsonDecoder) DecodeFloat32() float32 {
	return jd.reader.ParseFloat32(jd.reader.Value)
}

func (jd *JsonDecoder) DecodeFloat64() float64 {
	return jd.reader.ParseFloat64(jd.reader.Value)
}

func (jd *JsonDecoder) DecodeString() string {
	return jd.reader.ParseString(jd.reader.Value)
}

func (jd *JsonDecoder) DecodeStringArray() (result []string) {
	result = make([]string, 0, 4)
	for !jd.IsEndOfArray() {
		str := jd.DecodeString()
		result = append(result, str)
	}
	return
}

func (jd *JsonDecoder) DecodeStringMapString() (result map[string]string) {
	result = make(map[string]string, 4)
	for !jd.IsEndOfObject() {
		key := jd.GetField()
		value := jd.DecodeString()
		result[key] = value
	}
	return result
}

func (jd *JsonDecoder) IsEndOfObject() bool {
	return jd.reader.ReadUntilObjectEnd()
}

func (jd *JsonDecoder) IsEndOfArray() bool {
	return jd.reader.ReadUntilArrayEnd()
}
