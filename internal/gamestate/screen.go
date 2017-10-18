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
	"image/color"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/tint"
)

var emptyImage *ebiten.Image

func init() {
	emptyImage, _ = ebiten.NewImage(16, 16, ebiten.FilterNearest)
}

type Screen struct {
	tint            tint.Tint
	fadeInCount     int
	fadeInMaxCount  int
	fadeOutCount    int
	fadeOutMaxCount int
	fadedOut        bool
	fadeColor       color.RGBA
}

func (s *Screen) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("tint")
	e.EncodeInterface(&s.tint)
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

	e.EncodeString("fadeColor")
	e.BeginArray()
	e.EncodeInt(int(s.fadeColor.R))
	e.EncodeInt(int(s.fadeColor.G))
	e.EncodeInt(int(s.fadeColor.B))
	e.EncodeInt(int(s.fadeColor.A))
	e.EndArray()

	e.EndMap()
	return e.Flush()
}

func (s *Screen) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "tint":
			d.DecodeInterface(&s.tint)
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
		case "fadeColor":
			n = d.DecodeArrayLen()
			if n != 4 {
				for i := 0; i < n; i++ {
					d.Skip()
				}
				break
			}
			s.fadeColor.R = uint8(d.DecodeInt())
			s.fadeColor.G = uint8(d.DecodeInt())
			s.fadeColor.B = uint8(d.DecodeInt())
			s.fadeColor.A = uint8(d.DecodeInt())
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("gamestate: Screen.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (s *Screen) startTint(red, green, blue, gray float64, count int) {
	s.tint.Set(red, green, blue, gray, count)
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

func (s *Screen) setFadeColor(fadeColor color.Color) {
	r, g, b, a := fadeColor.RGBA()
	s.fadeColor.R = uint8(r >> 8)
	s.fadeColor.G = uint8(g >> 8)
	s.fadeColor.B = uint8(b >> 8)
	s.fadeColor.A = uint8(a >> 8)
}

func (s *Screen) isChangingTint() bool {
	return s.tint.IsChanging()
}

func (s *Screen) isFading() bool {
	return s.fadeInCount > 0 || s.fadeOutCount > 0
}

func (s *Screen) isFadedOut() bool {
	return s.fadedOut
}

func (s *Screen) fadeColorTranslate() (r, g, b, a float64) {
	if s.fadeColor.A == 0 {
		return 0, 0, 0, 0
	}
	r = float64(s.fadeColor.R) / float64(s.fadeColor.A)
	g = float64(s.fadeColor.G) / float64(s.fadeColor.A)
	b = float64(s.fadeColor.B) / float64(s.fadeColor.A)
	a = float64(s.fadeColor.A) / 255
	return
}

func (s *Screen) Draw(screen *ebiten.Image, img *ebiten.Image, op *ebiten.DrawImageOptions) {
	fadeRate := 0.0
	if s.fadedOut {
		fadeRate = 1
	} else {
		s.tint.Apply(&op.ColorM)
		if s.fadeInCount > 0 {
			fadeRate = float64(s.fadeInCount) / float64(s.fadeInMaxCount)
		}
		if s.fadeOutCount > 0 {
			fadeRate = 1 - float64(s.fadeOutCount)/float64(s.fadeOutMaxCount)
		}
	}
	screen.DrawImage(img, op)

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
		op.GeoM.Translate(tx, 0)
		op.ColorM.Translate(s.fadeColorTranslate())
		op.ColorM.Scale(1, 1, 1, fadeRate)
		screen.DrawImage(emptyImage, op)
	}
}

func (s *Screen) Update() {
	s.tint.Update()
	if s.fadeInCount > 0 {
		s.fadeInCount--
	}
	if s.fadeOutCount > 0 {
		s.fadeOutCount--
		if s.fadeOutCount == 0 {
			s.fadedOut = true
		}
	}
}
