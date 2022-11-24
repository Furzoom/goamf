// Copyright (c) 2022 Furzoom.com, All rights reserved.
// Author: mn, mn@furzoom.com

package goamf

import "reflect"

const (
	AMF0 = uint(0)
	AMF3 = uint(3)
)

const (
	Amf0NumberMarker        = 0x00
	Amf0BooleanMarker       = 0x01
	Amf0StringMarker        = 0x02
	Amf0ObjectMarker        = 0x03
	Amf0MovieclipMarker     = 0x04
	Amf0NullMarker          = 0x05
	Amf0UndefinedMarker     = 0x06
	Amf0ReferenceMarker     = 0x07
	Amf0EcmaArrayMarker     = 0x08
	Amf0ObjectEndMarker     = 0x09
	Amf0StrictArrayMarker   = 0x0a
	Amf0DateMarker          = 0x0b
	Amf0LongStringMarker    = 0x0c
	Amf0UnsupportedMarker   = 0x0d
	Amf0RecordsetMarker     = 0x0e
	Amf0XMLDocumentMarker   = 0x0f
	Amf0TypedObjectMarker   = 0x10
	Amf0AvmplusObjectMarker = 0x11
)

const (
	Amf0MaxStringLen = 0xFFFF
)

const (
	Amf3UndefinedMarker = 0x00
	Amf3NullMarker      = 0x01
	Amf3FalseMarker     = 0x02
	Amf3TrueMarker      = 0x03
	Amf3IntegerMarker   = 0x04
	Amf3DoubleMarker    = 0x05
	Amf3StringMarker    = 0x06
	Amf3XMLDocMarker    = 0x07
	Amf3DateMarker      = 0x08
	Amf3ArrayMarker     = 0x09
	Amf3ObjectMarker    = 0x0a
	Amf3XMLMarker       = 0x0b
	Amf3ByteArrayMarker = 0x0c
)

type Writer interface {
	Write(p []byte) (n int, err error)
	WriteByte(c byte) error
}

type Reader interface {
	Read(p []byte) (n int, err error)
	ReadByte() (c byte, err error)
}

// Undefined type
type Undefined struct{}

// Object type
type Object map[string]interface{}

// stringValues is a slice of reflect.Value holding *reflect.StringValue.
// It implements the method to sort by string.
type stringValues []reflect.Value

func (sv stringValues) Len() int {
	return len(sv)
}

func (sv stringValues) Swap(i, j int) {
	sv[i], sv[j] = sv[j], sv[i]
}

func (sv stringValues) Less(i, j int) bool {
	return sv.get(i) < sv.get(j)
}

func (sv stringValues) get(i int) string {
	return sv[i].String()
}
