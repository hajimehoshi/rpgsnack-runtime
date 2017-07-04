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
	"fmt"

	"github.com/vmihailenco/msgpack"
	"github.com/vmihailenco/msgpack/codes"
)

type Decoder struct {
	dec *msgpack.Decoder
	err error
}

func NewDecoder(dec *msgpack.Decoder) *Decoder {
	return &Decoder{
		dec: dec,
	}
}

func (d *Decoder) Error() error {
	return d.err
}

func (d *Decoder) DecodeArrayLen() int {
	if d.err != nil {
		return 0
	}
	n, err := d.dec.DecodeArrayLen()
	if err != nil {
		d.err = err
		return 0
	}
	return n
}

func (d *Decoder) DecodeMapLen() int {
	if d.err != nil {
		return 0
	}
	n, err := d.dec.DecodeMapLen()
	if err != nil {
		d.err = err
		return 0
	}
	return n
}

func (d *Decoder) DecodeNil() {
	if d.err != nil {
		return
	}
	if err := d.dec.DecodeNil(); err != nil {
		d.err = err
		return
	}
	return
}

func (d *Decoder) DecodeBool() bool {
	if d.err != nil {
		return false
	}
	v, err := d.dec.DecodeBool()
	if err != nil {
		d.err = err
		return false
	}
	return v
}

func (d *Decoder) DecodeInt() int {
	if d.err != nil {
		return 0
	}
	v, err := d.dec.DecodeInt()
	if err != nil {
		d.err = err
		return 0
	}
	return v
}

func (d *Decoder) DecodeFloat64() float64 {
	if d.err != nil {
		return 0
	}
	v, err := d.dec.DecodeFloat64()
	if err != nil {
		d.err = err
		return 0
	}
	return v
}

func (d *Decoder) DecodeString() string {
	if d.err != nil {
		return ""
	}
	v, err := d.dec.DecodeString()
	if err != nil {
		d.err = err
		return ""
	}
	return v
}

func (d *Decoder) SkipCodeIfNil() bool {
	if d.err != nil {
		return false
	}
	c, err := d.dec.PeekCode()
	if err != nil {
		d.err = err
		return false
	}
	if c == codes.Nil {
		if err := d.dec.DecodeNil(); err != nil {
			panic(fmt.Sprintf("not reached: %v", err))
		}
		return true
	}
	return false
}

func (d *Decoder) DecodeInterface(v msgpack.CustomDecoder) {
	if d.err != nil {
		return
	}
	if err := v.DecodeMsgpack(d.dec); err != nil {
		d.err = err
		return
	}
}

func (d *Decoder) DecodeAny(v interface{}) {
	if d.err != nil {
		return
	}
	if err := d.dec.Decode(v); err != nil {
		d.err = err
		return
	}
}
