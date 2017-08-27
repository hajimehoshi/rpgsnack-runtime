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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/interpolation"
)

type Tint struct {
	red   interpolation.I
	green interpolation.I
	blue  interpolation.I
	gray  interpolation.I
}

func (t *Tint) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()
	e.EncodeString("red")
	e.EncodeInterface(&t.red)
	e.EncodeString("green")
	e.EncodeInterface(&t.green)
	e.EncodeString("blue")
	e.EncodeInterface(&t.blue)
	e.EncodeString("gray")
	e.EncodeInterface(&t.gray)
	e.EndMap()
	return e.Flush()
}

func (t *Tint) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "red":
			d.DecodeInterface(&t.red)
		case "green":
			d.DecodeInterface(&t.green)
		case "blue":
			d.DecodeInterface(&t.blue)
		case "gray":
			d.DecodeInterface(&t.gray)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("gamestate: Tint.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (t *Tint) isZero() bool {
	return t.red.Current() == 0 && t.green.Current() == 0 &&
		t.blue.Current() == 0 && t.gray.Current() == 0
}

func (t *Tint) IsChanging() bool {
	return t.red.IsChanging()
}

func (t *Tint) Set(red, green, blue, gray float64, count int) {
	t.red.Set(red, count)
	t.green.Set(green, count)
	t.blue.Set(blue, count)
	t.gray.Set(gray, count)
}

func (t *Tint) Update() {
	t.red.Update()
	t.green.Update()
	t.blue.Update()
	t.gray.Update()
}

func (t *Tint) Apply(clr *ebiten.ColorM) {
	if t.isZero() {
		return
	}
	if t.gray.Current() != 0 {
		clr.ChangeHSV(0, 1-t.gray.Current(), 1)
	}
	rs, gs, bs := 1.0, 1.0, 1.0
	if t.red.Current() < 0 {
		rs = 1 - -t.red.Current()
	}
	if t.green.Current() < 0 {
		gs = 1 - -t.green.Current()
	}
	if t.blue.Current() < 0 {
		bs = 1 - -t.blue.Current()
	}
	clr.Scale(rs, gs, bs, 1)
	rt, gt, bt := 0.0, 0.0, 0.0
	if t.red.Current() > 0 {
		rt = t.red.Current()
	}
	if t.green.Current() > 0 {
		gt = t.green.Current()
	}
	if t.blue.Current() > 0 {
		bt = t.blue.Current()
	}
	clr.Translate(rt, gt, bt, 0)
}
