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

package gamestate

import (
	"fmt"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

var emptyImage *ebiten.Image

func init() {
	img, err := ebiten.NewImage(16, 16, ebiten.FilterNearest)
	if err != nil {
		panic(err)
	}
	emptyImage = img
}

type tint struct {
	Red   float64
	Green float64
	Blue  float64
	Gray  float64
}

func (t *tint) isZero() bool {
	return t.Red == 0 && t.Green == 0 && t.Blue == 0 && t.Gray == 0
}

func (t *tint) EncodeMsgpack(enc *msgpack.Encoder) error {
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

func (t *tint) DecodeMsgpack(dec *msgpack.Decoder) error {
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
		return fmt.Errorf("gamestate: tint.DecodeMsgpack failed: %v", err)
	}
	return nil
}

type Screen struct {
	currentTint     tint
	origTint        tint
	targetTint      tint
	tintCount       int
	tintMaxCount    int
	fadeInCount     int
	fadeInMaxCount  int
	fadeOutCount    int
	fadeOutMaxCount int
	fadedOut        bool
}

func (s *Screen) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("currentTint")
	e.EncodeInterface(&s.currentTint)
	e.EncodeString("origTint")
	e.EncodeInterface(&s.origTint)
	e.EncodeString("targetTint")
	e.EncodeInterface(&s.targetTint)

	e.EncodeString("tintCount")
	e.EncodeInt(s.tintCount)
	e.EncodeString("tintMaxCount")
	e.EncodeInt(s.tintMaxCount)
	e.EncodeString("fadeInCount")
	e.EncodeInt(s.fadeInCount)
	e.EncodeString("fadeInMaxCount")
	e.EncodeInt(s.fadeInMaxCount)
	e.EncodeString("fadeOutCount")
	e.EncodeInt(s.fadeOutCount)
	e.EncodeString("fadeOutMaxCount")
	e.EncodeInt(s.fadeOutMaxCount)

	e.EncodeString("fadedOut")
	e.EncodeBool(s.fadedOut)

	e.EndMap()
	return e.Flush()
}

func (s *Screen) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "currentTint":
			d.DecodeInterface(&s.currentTint)
		case "origTint":
			d.DecodeInterface(&s.origTint)
		case "targetTint":
			d.DecodeInterface(&s.targetTint)
		case "tintCount":
			s.tintCount = d.DecodeInt()
		case "tintMaxCount":
			s.tintMaxCount = d.DecodeInt()
		case "fadeInCount":
			s.fadeInCount = d.DecodeInt()
		case "fadeInMaxCount":
			s.fadeInMaxCount = d.DecodeInt()
		case "fadeOutCount":
			s.fadeOutCount = d.DecodeInt()
		case "fadeOutMaxCount":
			s.fadeOutMaxCount = d.DecodeInt()
		case "fadedOut":
			s.fadedOut = d.DecodeBool()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("gamestate: Screen.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (s *Screen) startTint(red, green, blue, gray float64, count int) {
	s.origTint.Red = s.currentTint.Red
	s.origTint.Green = s.currentTint.Green
	s.origTint.Blue = s.currentTint.Blue
	s.origTint.Gray = s.currentTint.Gray
	s.targetTint.Red = red
	s.targetTint.Green = green
	s.targetTint.Blue = blue
	s.targetTint.Gray = gray
	s.tintCount = count
	s.tintMaxCount = count
}

func (s *Screen) fadeIn(count int) {
	s.fadeInCount = count
	s.fadeInMaxCount = count
	s.fadedOut = false
}

func (s *Screen) fadeOut(count int) {
	s.fadeOutCount = count
	s.fadeOutMaxCount = count
}

func (s *Screen) isChangingTint() bool {
	return s.tintCount > 0
}

func (s *Screen) isFading() bool {
	return s.fadeInCount > 0 || s.fadeOutCount > 0
}

func (s *Screen) isFadedOut() bool {
	return s.fadedOut
}

func (s *Screen) Draw(screen *ebiten.Image, img *ebiten.Image, op *ebiten.DrawImageOptions) error {
	fadeRate := 0.0
	if s.fadedOut {
		fadeRate = 1
	} else {
		if !s.currentTint.isZero() {
			if s.currentTint.Gray != 0 {
				op.ColorM.ChangeHSV(0, 1-s.currentTint.Gray, 1)
			}
			rs, gs, bs := 1.0, 1.0, 1.0
			if s.currentTint.Red < 0 {
				rs = 1 - -s.currentTint.Red
			}
			if s.currentTint.Green < 0 {
				gs = 1 - -s.currentTint.Green
			}
			if s.currentTint.Blue < 0 {
				bs = 1 - -s.currentTint.Blue
			}
			op.ColorM.Scale(rs, gs, bs, 1)
			rt, gt, bt := 0.0, 0.0, 0.0
			if s.currentTint.Red > 0 {
				rt = s.currentTint.Red
			}
			if s.currentTint.Green > 0 {
				gt = s.currentTint.Green
			}
			if s.currentTint.Blue > 0 {
				bt = s.currentTint.Blue
			}
			op.ColorM.Translate(rt, gt, bt, 0)
		}
		if s.fadeInCount > 0 {
			fadeRate = float64(s.fadeInCount) / float64(s.fadeInMaxCount)
		}
		if s.fadeOutCount > 0 {
			fadeRate = 1 - float64(s.fadeOutCount)/float64(s.fadeOutMaxCount)
		}
	}
	if err := screen.DrawImage(img, op); err != nil {
		return err
	}
	if fadeRate > 0 {
		op := &ebiten.DrawImageOptions{}
		w, h := emptyImage.Size()
		targetW, targetH := img.Size()
		sx := float64(targetW) / float64(w)
		sy := float64(targetH) / float64(h)
		op.GeoM.Scale(sx, sy)
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		sw, _ := screen.Size()
		tx := (float64(sw) - consts.TileXNum*consts.TileSize*consts.TileScale) / 2
		op.GeoM.Translate(tx, consts.GameMarginTop)
		op.ColorM.Translate(0, 0, 0, 1)
		op.ColorM.Scale(1, 1, 1, fadeRate)
		if err := screen.DrawImage(emptyImage, op); err != nil {
			return err
		}
	}
	return nil
}

func (s *Screen) Update() error {
	if s.tintCount > 0 {
		s.tintCount--
		rate := 1 - float64(s.tintCount)/float64(s.tintMaxCount)
		s.currentTint.Red = s.origTint.Red*(1-rate) + s.targetTint.Red*rate
		s.currentTint.Green = s.origTint.Green*(1-rate) + s.targetTint.Green*rate
		s.currentTint.Blue = s.origTint.Blue*(1-rate) + s.targetTint.Blue*rate
		s.currentTint.Gray = s.origTint.Gray*(1-rate) + s.targetTint.Gray*rate
	} else {
		s.currentTint.Red = s.targetTint.Red
		s.currentTint.Green = s.targetTint.Green
		s.currentTint.Blue = s.targetTint.Blue
		s.currentTint.Gray = s.targetTint.Gray
	}
	if s.fadeInCount > 0 {
		s.fadeInCount--
	}
	if s.fadeOutCount > 0 {
		s.fadeOutCount--
		if s.fadeOutCount == 0 {
			s.fadedOut = true
		}
	}
	return nil
}
