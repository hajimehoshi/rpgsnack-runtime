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

package interpolation

import (
	"fmt"

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

type I struct {
	src      float64
	dst      float64
	count    int
	maxCount int
}

func New(val float64) *I {
	return &I{
		src: val,
		dst: val,
	}
}

func (i *I) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("src")
	e.EncodeFloat64(i.src)

	e.EncodeString("dst")
	e.EncodeFloat64(i.dst)

	e.EncodeString("count")
	e.EncodeInt(i.count)

	e.EncodeString("maxCount")
	e.EncodeInt(i.maxCount)

	e.EndMap()
	return e.Flush()
}

func (i *I) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)

	n := d.DecodeMapLen()
	for j := 0; j < n; j++ {
		switch k := d.DecodeString(); k {
		case "src":
			i.src = d.DecodeFloat64()
		case "dst":
			i.dst = d.DecodeFloat64()
		case "count":
			i.count = d.DecodeInt()
		case "maxCount":
			i.maxCount = d.DecodeInt()
		}
	}

	if err := d.Error(); err != nil {
		return fmt.Errorf("pictures: Pictures.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (i *I) Current() float64 {
	if i.maxCount == 0 {
		return i.dst
	}
	rate := float64(i.count) / float64(i.maxCount)
	return rate*i.src + (1-rate)*i.dst
}

func (i *I) Set(value float64, count int) {
	i.src = i.Current()
	i.dst = value
	i.count = count
	i.maxCount = count
}

func (i *I) SetDiff(value float64, count int) {
	i.src = i.Current()
	i.dst += value
	i.count = count
	i.maxCount = count
}

func (i *I) Update() {
	if i.count > 0 {
		i.count--
	}
}
