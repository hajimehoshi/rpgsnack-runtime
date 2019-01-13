// Copyright 2017 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package easymsgpack

import (
	"reflect"

	"github.com/vmihailenco/msgpack"
)

type valueType int

const (
	valueTypeNil valueType = iota
	valueTypeBool
	valueTypeInt
	valueTypeInt64
	valueTypeFloat64
	valueTypeString
	valueTypeBytes
	valueTypeInterface
	valueTypeAny
	valueTypeArray
	valueTypeMap
)

type value struct {
	valueType      valueType // TODO: How about unifying to interfaceValue?
	boolValue      bool
	intValue       int64
	float64Value   float64
	stringValue    string
	bytesValue     []uint8
	interfaceValue msgpack.CustomEncoder
	anyValue       interface{}
	length         int
	indent         int
}

type Encoder struct {
	enc    *msgpack.Encoder
	vals   []*value
	indent int
}

func NewEncoder(enc *msgpack.Encoder) *Encoder {
	return &Encoder{
		enc: enc,
	}
}

func encodeValue(enc *msgpack.Encoder, val *value) error {
	switch val.valueType {
	case valueTypeNil:
		if err := enc.EncodeNil(); err != nil {
			return err
		}
	case valueTypeBool:
		if err := enc.EncodeBool(val.boolValue); err != nil {
			return err
		}
	case valueTypeInt:
		if err := enc.EncodeInt(val.intValue); err != nil {
			return err
		}
	case valueTypeInt64:
		if err := enc.EncodeInt64(val.intValue); err != nil {
			return err
		}
	case valueTypeFloat64:
		if err := enc.EncodeFloat64(val.float64Value); err != nil {
			return err
		}
	case valueTypeString:
		if err := enc.EncodeString(val.stringValue); err != nil {
			return err
		}
	case valueTypeBytes:
		if err := enc.EncodeBytes(val.bytesValue); err != nil {
			return err
		}
	case valueTypeInterface:
		if err := val.interfaceValue.EncodeMsgpack(enc); err != nil {
			return err
		}
	case valueTypeArray:
		if err := enc.EncodeArrayLen(val.length); err != nil {
			return err
		}
	case valueTypeMap:
		if err := enc.EncodeMapLen(val.length); err != nil {
			return err
		}
	case valueTypeAny:
		if err := enc.Encode(val.anyValue); err != nil {
			return err
		}
	default:
		panic("not reached")
	}
	return nil
}

func (e *Encoder) Flush() error {
	for _, v := range e.vals {
		if err := encodeValue(e.enc, v); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) BeginArray() {
	val := &value{
		valueType: valueTypeArray,
		indent:    e.indent,
	}
	e.vals = append(e.vals, val)
	e.indent++
}

func (e *Encoder) EndArray() {
	if e.indent == 0 {
		panic("msgpack: e.indent must be more than 0; forgot to call BeginMap?")
	}
	c := 0
	for i := len(e.vals) - 1; i >= 0; i-- {
		if e.vals[i].indent == e.indent-1 {
			if e.vals[i].valueType != valueTypeArray {
				panic("msgpack: invalid indent: forgot to call BeginArray?")
			}
			e.vals[i].length = c
			e.indent--
			return
		}
		if e.vals[i].indent == e.indent {
			c++
		}
	}
	panic("msgpack: no more values to seek: forgot to call BeginArray?")
}

func (e *Encoder) BeginMap() {
	val := &value{
		valueType: valueTypeMap,
		indent:    e.indent,
	}
	e.vals = append(e.vals, val)
	e.indent++
}

func (e *Encoder) EndMap() {
	if e.indent == 0 {
		panic("msgpack: e.indent must be more than 0; forgot to call BeginMap?")
	}
	c := 0
	for i := len(e.vals) - 1; i >= 0; i-- {
		if e.vals[i].indent == e.indent-1 {
			if e.vals[i].valueType != valueTypeMap {
				panic("msgpack: invalid indent: forgot to call BeginMap?")
			}
			e.vals[i].length = c / 2
			e.indent--
			return
		}
		if e.vals[i].indent == e.indent {
			c++
		}
	}
	panic("msgpack: no more values to seek: forgot to call BeginMap?")
}

func (e *Encoder) EncodeNil() {
	e.vals = append(e.vals, &value{
		valueType: valueTypeNil,
		indent:    e.indent,
	})
}

func (e *Encoder) EncodeBool(v bool) {
	e.vals = append(e.vals, &value{
		valueType: valueTypeBool,
		boolValue: v,
		indent:    e.indent,
	})
}

func (e *Encoder) EncodeInt(v int) {
	e.vals = append(e.vals, &value{
		valueType: valueTypeInt,
		intValue:  int64(v),
		indent:    e.indent,
	})
}

func (e *Encoder) EncodeInt64(v int64) {
	e.vals = append(e.vals, &value{
		valueType: valueTypeInt64,
		intValue:  v,
		indent:    e.indent,
	})
}

func (e *Encoder) EncodeFloat64(v float64) {
	e.vals = append(e.vals, &value{
		valueType:    valueTypeFloat64,
		float64Value: v,
		indent:       e.indent,
	})
}

func (e *Encoder) EncodeString(v string) {
	e.vals = append(e.vals, &value{
		valueType:   valueTypeString,
		stringValue: v,
		indent:      e.indent,
	})
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	switch rv := reflect.ValueOf(v); rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

func (e *Encoder) EncodeInterface(v msgpack.CustomEncoder) {
	if isNil(v) {
		e.EncodeNil()
		return
	}
	e.vals = append(e.vals, &value{
		valueType:      valueTypeInterface,
		interfaceValue: v,
		indent:         e.indent,
	})
}

func (e *Encoder) EncodeAny(v interface{}) {
	if isNil(v) {
		e.EncodeNil()
		return
	}
	e.vals = append(e.vals, &value{
		valueType: valueTypeAny,
		anyValue:  v,
		indent:    e.indent,
	})
}
