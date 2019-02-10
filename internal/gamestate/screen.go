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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
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

	shakeCount     int
	shakeMaxCount  int
	shakePower     int
	shakeSpeed     int
	shakeDirection data.ShakeDirection
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

	e.EncodeString("shakeCount")
	e.EncodeInt(s.shakeCount)
	e.EncodeString("shakeMaxCount")
	e.EncodeInt(s.shakeMaxCount)
	e.EncodeString("shakePower")
	e.EncodeInt(s.shakePower)
	e.EncodeString("shakeSpeed")
	e.EncodeInt(s.shakeSpeed)
	e.EncodeString("shakeDirection")
	e.EncodeString(string(s.shakeDirection))

	e.EndMap()
	return e.Flush()
}

func (s *Screen) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
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
			n := d.DecodeArrayLen()
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
		case "shakeCount":
			s.shakeCount = d.DecodeInt()
		case "shakeMaxCount":
			s.shakeMaxCount = d.DecodeInt()
		case "shakePower":
			s.shakePower = d.DecodeInt()
		case "shakeSpeed":
			s.shakeSpeed = d.DecodeInt()
		case "shakeDirection":
			s.shakeDirection = data.ShakeDirection(d.DecodeString())
		default:
			if err := d.Error(); err != nil {
				return fmt.Errorf("gamestate: Screen.DecodeMsgpack failed: invalid key: %s", k)
			}
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

const infiniteCount = (1 << 31) - 1

func (s *Screen) startShaking(power, speed, count int, dir data.ShakeDirection) {
	s.shakePower = power
	s.shakeSpeed = speed
	if count > 0 {
		s.shakeCount = count
		s.shakeMaxCount = count
	} else {
		s.shakeCount = infiniteCount
		s.shakeMaxCount = infiniteCount
	}
	s.shakeDirection = dir
}

func (s *Screen) stopShaking() {
	s.shakePower = 0
	s.shakeSpeed = 0
	s.shakeCount = 0
	s.shakeMaxCount = 0
	s.shakeDirection = data.ShakeDirectionHorizontal
}

func (s *Screen) isShaking() bool {
	return s.shakeCount > 0
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

func (s *Screen) ApplyTintColor(c *ebiten.ColorM) {
	// TODO: When s.fadeOut is true, Apply should not be used for backward compatibility?
	s.tint.Apply(c)
}

func (s *Screen) ZeroTint() bool {
	return s.tint.Zero()
}

func (s *Screen) ApplyShake(g *ebiten.GeoM) {
	if s.shakeCount == 0 {
		return
	}

	duration := s.shakeMaxCount - s.shakeCount + 1
	amp := s.shakePower * 2
	delta := (s.shakePower * s.shakeSpeed * duration) / 10
	delta %= amp * 4
	switch {
	case delta < amp:
		// Do nothing
	case delta < amp*2:
		delta -= amp
		delta = amp - delta
	case delta < amp*3:
		delta -= amp * 2
		delta = -delta
	default:
		delta -= amp * 3
		delta = -amp + delta
	}
	if s.shakeDirection == data.ShakeDirectionVertical {
		g.Translate(0, float64(delta))
	} else {
		g.Translate(float64(delta), 0)
	}
	if s.shakeMaxCount == infiniteCount && delta == 0 {
		s.shakeCount = s.shakeMaxCount
	}
}

func (s *Screen) Draw(img *ebiten.Image) {
	fadeRate := 0.0
	if s.fadedOut {
		fadeRate = 1
	} else {
		if s.fadeInCount > 0 {
			fadeRate = float64(s.fadeInCount) / float64(s.fadeInMaxCount)
		}
		if s.fadeOutCount > 0 {
			fadeRate = 1 - float64(s.fadeOutCount)/float64(s.fadeOutMaxCount)
		}
	}

	if fadeRate > 0 {
		op := &ebiten.DrawImageOptions{}
		w, h := emptyImage.Size()
		targetW, targetH := img.Size()
		sx := float64(targetW) / float64(w)
		sy := float64(targetH) / float64(h)
		op.GeoM.Scale(sx, sy)
		op.ColorM.Translate(s.fadeColorTranslate())
		op.ColorM.Scale(1, 1, 1, fadeRate)
		img.DrawImage(emptyImage, op)
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
	if s.shakeCount > 0 {
		s.shakeCount--
	}
}
