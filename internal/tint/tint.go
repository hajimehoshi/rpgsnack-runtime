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

package tint

import (
	"fmt"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

type Tint struct {
	Red   float64
	Green float64
	Blue  float64
	Gray  float64
}

func (t *Tint) IsZero() bool {
	return t.Red == 0 && t.Green == 0 && t.Blue == 0 && t.Gray == 0
}

func (t *Tint) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()
	e.EncodeString("red")
	e.EncodeFloat64(t.Red)
	e.EncodeString("green")
	e.EncodeFloat64(t.Green)
	e.EncodeString("blue")
	e.EncodeFloat64(t.Blue)
	e.EncodeString("gray")
	e.EncodeFloat64(t.Gray)
	e.EndMap()
	return e.Flush()
}

func (t *Tint) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "red":
			t.Red = d.DecodeFloat64()
		case "green":
			t.Green = d.DecodeFloat64()
		case "blue":
			t.Blue = d.DecodeFloat64()
		case "gray":
			t.Gray = d.DecodeFloat64()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("gamestate: Tint.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (t *Tint) Apply(clr *ebiten.ColorM) {
	if t.IsZero() {
		return
	}
	if t.Gray != 0 {
		clr.ChangeHSV(0, 1-t.Gray, 1)
	}
	rs, gs, bs := 1.0, 1.0, 1.0
	if t.Red < 0 {
		rs = 1 - -t.Red
	}
	if t.Green < 0 {
		gs = 1 - -t.Green
	}
	if t.Blue < 0 {
		bs = 1 - -t.Blue
	}
	clr.Scale(rs, gs, bs, 1)
	rt, gt, bt := 0.0, 0.0, 0.0
	if t.Red > 0 {
		rt = t.Red
	}
	if t.Green > 0 {
		gt = t.Green
	}
	if t.Blue > 0 {
		bt = t.Blue
	}
	clr.Translate(rt, gt, bt, 0)
}
