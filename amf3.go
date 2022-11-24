// Copyright (c) 2022 Furzoom.com, All rights reserved.
// Author: mn, mn@furzoom.com

package goamf

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"
)

// AMF3 write functions

// Amf3WriteU29 writes a U29
//
//	(hex)           : (binary)
//
// 0x00000000 - 0x0000007F : 0xxxxxxx
// 0x00000080 - 0x00003FFF : 1xxxxxxx 0xxxxxxx
// 0x00004000 - 0x001FFFFF : 1xxxxxxx 1xxxxxxx 0xxxxxxx
// 0x00200000 - 0x3FFFFFFF : 1xxxxxxx 1xxxxxxx 1xxxxxxx xxxxxxxx
// 0x40000000 - 0xFFFFFFFF : throw range exception
func Amf3WriteU29(w Writer, n uint32) (num int, err error) {
	if n <= 0x7F {
		err = w.WriteByte(byte(n))
		if err != nil {
			return 0, err
		} else {
			return 1, nil
		}
	} else if n <= 0x3FFF {
		return w.Write([]byte{byte(n>>7 | 0x80), byte(n & 0x7F)})
	} else if n <= 0x1FFFFF {
		return w.Write([]byte{byte(n>>14 | 0x80), byte(n>>7&0x7F | 0x80), byte(n & 0x7F)})
	} else if n <= 0x3FFFFFFF {
		return w.Write([]byte{byte(n>>22 | 0x80), byte(n>>15&0x7f | 0x80), byte(n>>8&0x7F | 0x80), byte(n & 0xFF)})
	}
	return 0, &OutOfRangeError{}
}

func Amf3WriteString(w Writer, str string) (n int, err error) {
	err = w.WriteByte(Amf3StringMarker)
	if err != nil {
		return 0, err
	}
	n, err = Amf3WriteUTF8(w, str)
	if err != nil {
		return 1, err
	}
	return 1 + n, err
}

// Amf3WriteUTF8
//
// U29S-ref = U29     ; The first (low) bit is a flag with
//
//	; value 0. The remaining 1 to 28
//	; significant bits are used to encode a
//	; string reference table index (an
//	; integer).
//
// U29S-value = U29   ; The first (low) bit is a flag with
//
//	; value 1. The remaining 1 to 28
//	; significant bits are used to encode the
//	; byte-length of the UTF-8 encoded
//	; representation of the string
//
// UTF-8-empty = 0x01 ; The UTF-8-vr empty string which is
//
//	; never sent by reference.
//
// UTF-8-vr = U29S-ref | (U29S-value *(UTF8-char))
// string-type = string-marker UTF-8-vr
func Amf3WriteUTF8(w Writer, str string) (n int, err error) {
	length := len(str)
	if length == 0 {
		err = w.WriteByte(0x01)
		if err != nil {
			return 0, err
		} else {
			return 1, nil
		}
	}
	u := uint32(length<<1) | 0x01
	n, err = Amf3WriteU29(w, u)
	if err != nil {
		return 0, nil
	}
	m, err := w.Write([]byte(str))
	if err != nil {
		return n, err
	}
	return m + n, nil
}

func Amf3WriteDouble(w Writer, num float64) (n int, err error) {
	err = w.WriteByte(Amf3DoubleMarker)
	if err != nil {
		return 0, err
	}
	err = binary.Write(w, binary.BigEndian, num)
	if err != nil {
		return 1, err
	}
	return 9, nil
}

func Amf3WriteBoolean(w Writer, b bool) (n int, err error) {
	if b {
		err = w.WriteByte(Amf3TrueMarker)
	} else {
		err = w.WriteByte(Amf3FalseMarker)
	}
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func Amf3WriteNull(w Writer) (n int, err error) {
	err = w.WriteByte(Amf3NullMarker)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func Amf3WriteUndefined(w Writer) (n int, err error) {
	err = w.WriteByte(Amf3UndefinedMarker)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func Amf3WriteObjectMarker(w Writer) (n int, err error) {
	return WriteMarker(w, Amf3ObjectMarker)
}

func Amf3WriteObjectEndMarker(w Writer) (n int, err error) {
	err = w.WriteByte(0x01) // empty string
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func Amf3WriteObjectName(w Writer, name string) (n int, err error) {
	return Amf3WriteUTF8(w, name)
}

func Amf3WriteObject(w Writer, obj Object) (n int, err error) {
	n, err = Amf3WriteObjectMarker(w)
	if err != nil {
		return
	}
	m := 0
	err = w.WriteByte(0x0b) // traits class
	if err != nil {
		return
	}
	n += 1
	// Write empty class name
	m, err = Amf3WriteUTF8(w, "")
	if err != nil {
		return
	}
	n += m
	for key, value := range obj {
		m, err = Amf3WriteObjectName(w, key)
		if err != nil {
			return
		}
		n += m
		m, err = Amf3WriteValue(w, value)
		if err != nil {
			return
		}
		n += m
	}
	m, err = Amf3WriteObjectEndMarker(w)
	return n + m, err
}

func Amf3WriteValue(w Writer, value interface{}) (n int, err error) {
	if value == nil {
		return Amf3WriteNull(w)
	}
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return Amf3WriteNull(w)
	}
	switch v.Kind() {
	case reflect.String:
		return Amf3WriteString(w, v.String())
	case reflect.Bool:
		return Amf3WriteBoolean(w, v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return Amf3WriteDouble(w, float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Amf3WriteDouble(w, float64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		return Amf3WriteDouble(w, v.Float())
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// Byte array
			n, err = WriteMarker(w, Amf3ByteArrayMarker)
			if err != nil {
				return
			}
			b := v.Bytes()
			length := len(b)
			u := uint32((length << 1) | 0x01)
			var m int
			m, err = Amf3WriteU29(w, u)
			if err != nil {
				return
			}
			n += m
			m, err = w.Write(b)
			if err != nil {
				return
			}
			n += m
			return
		} else {
			n, err = WriteMarker(w, Amf3ArrayMarker)
			if err != nil {
				return
			}
			length := v.Len()
			u := uint32((length << 1) | 0x01)
			var m int
			m, err = Amf3WriteU29(w, u)
			if err != nil {
				return
			}
			n += m
			err = w.WriteByte(0x01) // empty string
			if err != nil {
				return
			}
			n += 1
			for i := 0; i < length; i++ {
				m, err = Amf3WriteValue(w, v.Index(i).Interface())
				if err != nil {
					return
				}
				n += m
			}
			return
		}
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return 0, &UnsupportedTypeError{v.Type().Key().Kind().String()}
		}
		n, err = Amf3WriteObjectMarker(w)
		if err != nil {
			return
		}
		m := 0
		err = w.WriteByte(0x0b) // traits class
		if err != nil {
			return
		}
		n += 1
		// Write empty class name
		m, err = Amf3WriteUTF8(w, "")
		if err != nil {
			return
		}
		n += m
		var sv stringValues = v.MapKeys()
		sort.Sort(sv)
		for _, k := range sv {
			m, err = Amf3WriteObjectName(w, k.String())
			if err != nil {
				return
			}
			n += m
			m, err = Amf3WriteValue(w, v.MapIndex(k).Interface())
			if err != nil {
				return
			}
			n += m
		}
		m, err = Amf3WriteObjectEndMarker(w)
		return n + m, err
	}
	if _, ok := value.(Undefined); ok {
		return Amf3WriteUndefined(w)
	} else if vt, ok := value.(Object); ok {
		return Amf3WriteObject(w, vt)
	} else if vt, ok := value.([]interface{}); ok {
		return 0, &UnsupportedTypeError{fmt.Sprintf("%+v", vt)}
	}
	return 0, &UnsupportedTypeError{v.Kind().String()}
}

// AMF3 write functions

func Amf3ReadU29(r Reader) (n uint32, err error) {
	var b byte
	for i := 0; i < 3; i++ {
		b, err = r.ReadByte()
		if err != nil {
			return
		}
		n = (n << 7) + uint32(b&0x7f)
		if (b & 0x80) == 0 {
			return
		}
	}
	b, err = r.ReadByte()
	if err != nil {
		return
	}
	return (n << 8) + uint32(b), nil
}

func Amf3ReadUTF8(r Reader) (string, error) {
	var length uint32
	var err error
	length, err = Amf3ReadU29(r)
	if err != nil {
		return "", err
	}
	if length&uint32(0x01) != uint32(0x01) {
		return "", &UnsupportedTypeError{"AMF3 string reference"}
	}
	length = length >> 1
	if length == 0 {
		return "", nil
	}
	data := make([]byte, length)
	_, err = r.Read(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func Amf3ReadString(r Reader) (str string, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return "", err
	}
	if marker != Amf3StringMarker {
		return "", &UnexpectedTypeError{marker}
	}
	return Amf3ReadUTF8(r)
}

func Amf3ReadInteger(r Reader) (num uint32, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return 0, err
	}
	if marker != Amf3IntegerMarker {
		return 0, &UnexpectedTypeError{marker}
	}
	return Amf3ReadU29(r)
}

func Amf3ReadDouble(r Reader) (num float64, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return 0, err
	}
	if marker != Amf3DoubleMarker {
		return 0, &UnexpectedTypeError{marker}
	}
	err = binary.Read(r, binary.BigEndian, &num)
	return
}

func Amf3ReadObjectName(r Reader) (name string, err error) {
	return Amf3ReadUTF8(r)
}

func Amf3ReadObject(r Reader) (obj Object, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return nil, err
	}
	if marker != Amf3ObjectMarker {
		return nil, &UnexpectedTypeError{marker}
	}
	return Amf3ReadObjectProperty(r)
}

func Amf3ReadObjectProperty(r Reader) (Object, error) {
	obj := make(Object)
	// Read traits flag
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0x0b {
		return nil, &UnsupportedTypeError{"traits object"}
	}
	// Read empty string
	b, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0x01 {
		return nil, &UnsupportedTypeError{"traits object"}
	}
	for {
		name, err := Amf3ReadObjectName(r)
		if err != nil {
			return nil, err
		}
		if name == "" {
			break
		}
		if _, ok := obj[name]; ok {
			return nil, &PropertyExistError{name}
		}
		value, err := Amf3ReadValue(r)
		if err != nil {
			return nil, err
		}
		obj[name] = value
	}
	return obj, nil
}

func Amf3ReadByteArray(r Reader) ([]byte, error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return nil, err
	}
	if marker != Amf3ByteArrayMarker {
		return nil, &UnexpectedTypeError{marker}
	}
	return Amf3readByteArray(r)
}

func Amf3readByteArray(r Reader) ([]byte, error) {
	length, err := Amf3ReadU29(r)
	if err != nil {
		return nil, err
	}
	if length&uint32(0x01) != uint32(0x01) {
		return nil, &UnsupportedTypeError{"AMF3 byte array reference"}
	}
	length = length >> 1
	buf := make([]byte, length)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != int(length) {
		return nil, &LengthError{fmt.Sprintf("expect %d, got %d", length, n)}
	}
	return buf, nil
}

func Amf3ReadValue(r Reader) (value interface{}, err error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return 0, err
	}
	switch marker {
	case Amf3UndefinedMarker:
		return Undefined{}, nil
	case Amf3NullMarker:
		return nil, nil
	case Amf3FalseMarker:
		return false, nil
	case Amf3TrueMarker:
		return true, nil
	case Amf3IntegerMarker:
		return Amf3ReadU29(r)
	case Amf3DoubleMarker:
		var num float64
		err = binary.Read(r, binary.BigEndian, &num)
		return num, err
	case Amf3StringMarker:
		return Amf3ReadUTF8(r)
	case Amf3ArrayMarker:
		// Todo: read array
	case Amf3ObjectMarker:
		return Amf3ReadObjectProperty(r)
	case Amf3ByteArrayMarker:
		return Amf3readByteArray(r)
	}
	return nil, &UnsupportedTypeError{fmt.Sprintf("%x", marker)}
}
