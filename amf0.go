// Copyright (c) 2022 Furzoom.com, All rights reserved.
// Author: mn, mn@furzoom.com

package goamf

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// Write functions

func WriteMarker(w Writer, mark byte) (n int, err error) {
	err = w.WriteByte(mark)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func WriteString(w Writer, str string) (n int, err error) {
	length := uint32(len(str))
	if length > Amf0MaxStringLen {
		err = w.WriteByte(Amf0LongStringMarker)
		if err != nil {
			return 0, err
		}
		err = WriteUTF8Long(w, str, length)
		length += 5
	} else {
		err = w.WriteByte(Amf0StringMarker)
		if err != nil {
			return 0, err
		}
		err = WriteUTF8(w, str, uint16(length))
		length += 3
	}
	if err != nil {
		return 1, err
	}
	return int(length), nil
}

func WriteUTF8(w Writer, s string, length uint16) error {
	err := binary.Write(w, binary.BigEndian, &length)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	return err
}

func WriteUTF8Long(w Writer, s string, length uint32) error {
	err := binary.Write(w, binary.BigEndian, &length)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	return err
}

func WriteDouble(w Writer, num float64) (n int, err error) {
	err = w.WriteByte(Amf0NumberMarker)
	if err != nil {
		return 0, err
	}
	err = binary.Write(w, binary.BigEndian, num)
	if err != nil {
		return 1, err
	}
	return 9, nil
}

func WriteBoolean(w Writer, b bool) (n int, err error) {
	err = w.WriteByte(Amf0BooleanMarker)
	if err != nil {
		return 0, err
	}
	if b {
		err = w.WriteByte(0x01)
	} else {
		err = w.WriteByte(0x00)
	}
	if err != nil {
		return 1, err
	}
	return 2, nil
}

func WriteNull(w Writer) (n int, err error) {
	err = w.WriteByte(Amf0NullMarker)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func WriteUndefined(w Writer) (n int, err error) {
	err = w.WriteByte(Amf0UndefinedMarker)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func WriteEcmaArray(w Writer, arr []interface{}) (n int, err error) {
	n, err = WriteMarker(w, Amf0EcmaArrayMarker)
	if err != nil {
		return
	}
	length := len(arr)
	err = binary.Write(w, binary.BigEndian, &length)
	if err != nil {
		return
	}
	n += 4
	m := 0
	for index, value := range arr {
		m, err = WriteObjectName(w, fmt.Sprintf("%d", index))
		if err != nil {
			return
		}
		n += m
		m, err = WriteValue(w, value)
		if err != nil {
			return
		}
		n += m
	}
	m, err = WriteObjectEndMarker(w)
	return n + m, err
}

func WriteObjectMarker(w Writer) (n int, err error) {
	return WriteMarker(w, Amf0ObjectMarker)
}

func WriteObjectEndMarker(w Writer) (n int, err error) {
	return w.Write([]byte{0x00, 0x00, Amf0ObjectEndMarker})
}

func WriteObjectName(w Writer, name string) (n int, err error) {
	length := uint16(len(name))
	err = WriteUTF8(w, name, length)
	return int(length + 2), err
}

func WriteObject(w Writer, obj Object) (n int, err error) {
	n, err = WriteObjectMarker(w)
	if err != nil {
		return
	}
	m := 0
	for key, value := range obj {
		m, err = WriteObjectName(w, key)
		if err != nil {
			return
		}
		n += m
		m, err = WriteValue(w, value)
		if err != nil {
			return
		}
		n += m
	}
	m, err = WriteObjectEndMarker(w)
	return n + m, err
}

func WriteStruct(w Writer, value reflect.Value) (n int, err error) {
	var m int
	for i := 0; i < value.NumField(); i++ {
		skip := false
		structField := value.Type().Field(i)
		if structField.Anonymous {
			m, err = WriteStruct(w, value.Field(i))
			if err != nil {
				return
			}
			n += m
		} else {
			name := structField.Tag.Get("amf")
			switch name {
			case "":
				name = structField.Name
			case "-":
				skip = true
			default:
				if strings.HasSuffix(name, ",omitempty") {
					if value.IsNil() {
						skip = true
					}

					name = strings.Split(name, ",")[0]
					if len(name) == 0 {
						name = structField.Name
					}
				}
			}
			if skip {
				continue
			}
			m, err = WriteObjectName(w, name)
			if err != nil {
				return
			}
			n += m
			field := value.Field(i)
			m, err = writeValue(w, field)
			if err != nil {
				return
			}
			n += m
		}
	}
	return n, nil
}

func WriteValue(w Writer, value interface{}) (n int, err error) {
	if value == nil {
		return WriteNull(w)
	}
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return WriteNull(w)
	}
	return writeValue(w, v)
}

func writeValue(w Writer, v reflect.Value) (n int, err error) {
	switch v.Kind() {
	case reflect.String:
		return WriteString(w, v.String())
	case reflect.Bool:
		return WriteBoolean(w, v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return WriteDouble(w, float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return WriteDouble(w, float64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		return WriteDouble(w, v.Float())
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		n, err = WriteMarker(w, Amf0EcmaArrayMarker)
		if err != nil {
			return
		}
		length := int32(v.Len())
		err = binary.Write(w, binary.BigEndian, &length)
		if err != nil {
			return
		}
		n += 4
		m := 0
		for index := int32(0); index < length; index++ {
			m, err = WriteObjectName(w, fmt.Sprintf("%d", index))
			if err != nil {
				return
			}
			n += m
			m, err = WriteValue(w, v.Index(int(index)).Interface())
			if err != nil {
				return
			}
			n += m
		}
		m, err = WriteObjectEndMarker(w)
		return n + m, err
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return 0, &UnsupportedTypeError{v.Type().Name()}
		}
		if v.IsNil() {
			return WriteNull(w)
		}
		n, err = WriteObjectMarker(w)
		if err != nil {
			return
		}
		m := 0
		var sv stringValues = v.MapKeys()
		sort.Sort(sv)
		for _, k := range sv {
			m, err = WriteObjectName(w, k.String())
			if err != nil {
				return
			}
			n += m
			m, err = WriteValue(w, v.MapIndex(k).Interface())
			if err != nil {
				return
			}
			n += m
		}
		m, err = WriteObjectEndMarker(w)
		return n + m, err
	case reflect.Ptr:
		if v.IsNil() || !v.IsValid() {
			return WriteNull(w)
		}
		return WriteValue(w, v.Elem().Interface())
	case reflect.Struct:
		n, err = WriteObjectMarker(w)
		if err != nil {
			return
		}
		m := 0
		m, err = WriteStruct(w, v)
		if err != nil {
			return
		}
		n += m
		m, err = WriteObjectEndMarker(w)
		return n + m, err
	}
	value := v.Interface()
	if value != nil {
		if _, ok := value.(Undefined); ok {
			return WriteUndefined(w)
		} else if vt, ok := value.(Object); ok {
			return WriteObject(w, vt)
		} else if vt, ok := value.([]interface{}); ok {
			return WriteEcmaArray(w, vt)
		}
	}
	return 0, &UnsupportedTypeError{v.Type().Name()}
}

// Read functions

func ReadMarker(r Reader) (mark byte, err error) {
	return r.ReadByte()
}

func ReadString(r Reader) (str string, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return
	}
	switch marker {
	case Amf0StringMarker:
		return ReadUTF8(r)
	case Amf0LongStringMarker:
		return ReadUTF8Long(r)
	}
	return str, &UnexpectedTypeError{marker}
}

func ReadUTF8(r Reader) (s string, err error) {
	var stringLength uint16
	err = binary.Read(r, binary.BigEndian, &stringLength)
	if err != nil {
		return
	}
	if stringLength == 0 {
		return s, nil
	}
	data := make([]byte, stringLength)
	_, err = r.Read(data)
	if err != nil {
		return
	}
	return string(data), nil
}

func ReadUTF8Long(r Reader) (s string, err error) {
	var stringLength uint32
	err = binary.Read(r, binary.BigEndian, &stringLength)
	if err != nil {
		return
	}
	if stringLength == 0 {
		return s, nil
	}
	data := make([]byte, stringLength)
	_, err = r.Read(data)
	if err != nil {
		return
	}
	return string(data), nil
}

func ReadDouble(r Reader) (num float64, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return
	}
	if marker != Amf0NumberMarker {
		return 0, &UnexpectedTypeError{marker}
	}
	err = binary.Read(r, binary.BigEndian, &num)
	return
}

func ReadBoolean(r Reader) (b bool, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return
	}
	if marker != Amf0BooleanMarker {
		return false, &UnexpectedTypeError{marker}
	}
	value, err := r.ReadByte()
	return bool(value != 0), err
}

func ReadObjectName(r Reader) (name string, err error) {
	return ReadUTF8(r)
}

func ReadObject(r Reader) (obj Object, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return
	}
	if marker != Amf0ObjectMarker {
		return nil, &UnexpectedTypeError{marker}
	}
	return ReadObjectProperty(r)
}

func ReadObjectProperty(r Reader) (Object, error) {
	obj := make(Object)
	for {
		name, err := ReadUTF8(r)
		if err != nil {
			return nil, err
		}
		if name == "" {
			b, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			if b == Amf0ObjectEndMarker {
				break
			} else {
				return nil, &ExpectedTypeError{b}
			}
		}
		if _, ok := obj[name]; ok {
			return nil, &PropertyExistError{name}
		}
		value, err := ReadValue(r)
		if err != nil {
			return nil, err
		}
		obj[name] = value
	}
	return obj, nil
}

func ReadStrictArray(r Reader) (arr []interface{}, err error) {
	var arrayCount uint32
	err = binary.Read(r, binary.BigEndian, &arrayCount)
	if err != nil {
		return nil, err
	}
	if arrayCount == 0 {
		return nil, nil
	}
	arr = make([]interface{}, arrayCount)

	for i := uint32(0); i < arrayCount; i++ {
		arr[i], err = ReadValue(r)
		if err != nil {
			return nil, err
		}
	}
	return
}

func ReadDate(r Reader) (t time.Time, err error) {
	var d float64
	var timezone int16
	err = binary.Read(r, binary.BigEndian, &d)
	if err != nil {
		return t, &ReadDateError{fmt.Sprintf("read double %s", err)}
	}
	err = binary.Read(r, binary.BigEndian, &timezone)
	if err != nil {
		return t, &ReadDateError{fmt.Sprintf("read time zone %s", err)}
	}

	d /= 1000.0
	sec := int64(d)
	nsec := int64((d - float64(sec)) * 10e9)
	t = time.Unix(sec, nsec)
	return
}

func ReadValue(r Reader) (value interface{}, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return nil, err
	}
	switch marker {
	case Amf0NumberMarker:
		var num float64
		err = binary.Read(r, binary.BigEndian, &num)
		return num, err
	case Amf0BooleanMarker:
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		return bool(b != 0), nil
	case Amf0StringMarker:
		return ReadUTF8(r)
	case Amf0ObjectMarker:
		return ReadObjectProperty(r)
	case Amf0MovieclipMarker:
		return nil, &UnsupportedTypeError{"Movieclip"}
	case Amf0NullMarker:
		return nil, nil
	case Amf0UndefinedMarker:
		return Undefined{}, nil
	case Amf0EcmaArrayMarker:
		// Decode ECMA Array to object
		arrLen := make([]byte, 4)
		_, err = r.Read(arrLen)
		if err != nil {
			return nil, err
		}
		obj, err := ReadObjectProperty(r)
		if err != nil {
			return nil, err
		}
		return obj, nil
	case Amf0ObjectEndMarker:
		return nil, &UnexpectedTypeError{marker}
	case Amf0StrictArrayMarker:
		return ReadStrictArray(r)
	case Amf0DateMarker:
		return ReadDate(r)
	case Amf0LongStringMarker:
		return ReadUTF8Long(r)
	case Amf0UnsupportedMarker:
		return nil, &UnexpectedTypeError{marker}
	case Amf0RecordsetMarker:
		return nil, &UnexpectedTypeError{marker}
	case Amf0XMLDocumentMarker:
		return nil, &UnexpectedTypeError{marker}
	case Amf0TypedObjectMarker:
		return nil, &UnexpectedTypeError{marker}
	case Amf0AvmplusObjectMarker:
		return Amf3ReadValue(r)
	}
	return nil, &UnsupportedTypeError{string(marker)}
}
