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

package easymsgpack_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/vmihailenco/msgpack"

	. "github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

func TestArray(t *testing.T) {
	vs := []int{1, 2, 3, 4, 5}

	var buf0 bytes.Buffer
	e0 := msgpack.NewEncoder(&buf0)
	e0.EncodeArrayLen(len(vs))
	for _, v := range vs {
		e0.EncodeInt(int64(v))
	}

	var buf1 bytes.Buffer
	e1 := NewEncoder(msgpack.NewEncoder(&buf1))
	e1.BeginArray()
	for _, v := range vs {
		e1.EncodeInt(v)
	}
	e1.EndArray()
	e1.Flush()

	b0 := buf0.Bytes()
	b1 := buf1.Bytes()
	if !reflect.DeepEqual(b0, b1) {
		t.Errorf("b0 (%v) != b1 (%v)", b0, b1)
	}
}

func TestArray2(t *testing.T) {
	vs := [][]int{
		{1, 2, 3, 4, 5},
		{6, 7, 8, 9},
		{10},
	}

	var buf0 bytes.Buffer
	e0 := msgpack.NewEncoder(&buf0)
	e0.EncodeArrayLen(len(vs))
	for _, v := range vs {
		e0.EncodeArrayLen(len(v))
		for _, vv := range v {
			e0.EncodeInt(int64(vv))
		}
	}

	var buf1 bytes.Buffer
	e1 := NewEncoder(msgpack.NewEncoder(&buf1))
	e1.BeginArray()
	for _, v := range vs {
		e1.BeginArray()
		for _, vv := range v {
			e1.EncodeInt(vv)
		}
		e1.EndArray()
	}
	e1.EndArray()
	e1.Flush()

	b0 := buf0.Bytes()
	b1 := buf1.Bytes()
	if !reflect.DeepEqual(b0, b1) {
		t.Errorf("b0 (%v) != b1 (%v)", b0, b1)
	}
}

func TestMap(t *testing.T) {
	m := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
	}
	// Key order matters.
	ks := []string{}
	for k := range m {
		ks = append(ks, k)
	}

	var buf0 bytes.Buffer
	e0 := msgpack.NewEncoder(&buf0)
	e0.EncodeMapLen(len(m))
	for _, k := range ks {
		e0.EncodeString(k)
		e0.EncodeInt(int64(m[k]))
	}

	var buf1 bytes.Buffer
	e1 := NewEncoder(msgpack.NewEncoder(&buf1))
	e1.BeginMap()
	for _, k := range ks {
		e1.EncodeString(k)
		e1.EncodeInt(m[k])
	}
	e1.EndMap()
	e1.Flush()

	b0 := buf0.Bytes()
	b1 := buf1.Bytes()
	if !reflect.DeepEqual(b0, b1) {
		t.Errorf("b0 (%v) != b1 (%v)", b0, b1)
	}
}

type invalid struct{}

func (i *invalid) EncodeMsgpack(enc *msgpack.Encoder) error {
	panic("not reached")
}

func TestNil(t *testing.T) {
	var val *invalid

	var buf0 bytes.Buffer
	e0 := msgpack.NewEncoder(&buf0)
	e0.EncodeNil()

	var buf1 bytes.Buffer
	e1 := NewEncoder(msgpack.NewEncoder(&buf1))
	e1.EncodeInterface(val)
	e1.Flush()

	b0 := buf0.Bytes()
	b1 := buf1.Bytes()
	if !reflect.DeepEqual(b0, b1) {
		t.Errorf("b0 (%v) != b1 (%v)", b0, b1)
	}
}

type Foo struct {
	Int int `msgpack:"int"`
}

func TestTags(t *testing.T) {
	val := &Foo{
		Int: 12345,
	}

	var buf0 bytes.Buffer
	e0 := msgpack.NewEncoder(&buf0)
	e0.Encode(val)

	var buf1 bytes.Buffer
	e1 := NewEncoder(msgpack.NewEncoder(&buf1))
	e1.BeginMap()
	e1.EncodeString("int")
	e1.EncodeInt(val.Int)
	e1.EndMap()
	e1.Flush()

	b0 := buf0.Bytes()
	b1 := buf1.Bytes()
	if !reflect.DeepEqual(b0, b1) {
		t.Errorf("b0 (%v) != b1 (%v)", b0, b1)
	}
}
