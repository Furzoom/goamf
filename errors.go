// Copyright (c) 2022 Furzoom.com, All rights reserved.
// Author: mn, mn@furzoom.com

package goamf

import (
	"fmt"
	"reflect"
	"strconv"
)

type UnsupportedTypeError struct {
	Type string
}

func (e *UnsupportedTypeError) Error() string {
	return fmt.Sprintf("unsupported type: %s", e.Type)
}

type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "unsupported value: " + e.Str
}

type InvalidUTF8Error struct {
	S string
}

func (e *InvalidUTF8Error) Error() string {
	return "invalid UTF-8 in string: " + strconv.Quote(e.S)
}

type NameLengthOverflowError struct {
	Name string
}

func (e *NameLengthOverflowError) Error() string {
	return "field[" + e.Name + "] is too long to encode into AMF0"
}

type UnexpectedTypeError struct {
	Type byte
}

func (e *UnexpectedTypeError) Error() string {
	return fmt.Sprintf("unexpected type: %#x", e.Type)
}

type ExpectedTypeError struct {
	Type byte
}

func (e *ExpectedTypeError) Error() string {
	return fmt.Sprintf("expected type: %#x here", e.Type)
}

type PropertyExistError struct {
	Name string
}

func (e *PropertyExistError) Error() string {
	return fmt.Sprintf("object-property '%s' exists", e.Name)
}

type ReadDateError struct {
	err string
}

func (e *ReadDateError) Error() string {
	return fmt.Sprintf("read date error: %s", e.err)
}

type OutOfRangeError struct {
	s string
}

func (e *OutOfRangeError) Error() string {
	return "out of range"
}

type LengthError struct {
	name string
}

func (e *LengthError) Error() string {
	return "length error: " + e.name
}
