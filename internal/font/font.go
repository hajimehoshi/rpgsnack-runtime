// Copyright 2016 Hajime Hoshi
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

package font

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/assets"
)

var mplusImage *ebiten.Image

func init() {
	img, err := assets.LoadImage("images/mplus.png", ebiten.FilterNearest)
	if err != nil {
		panic(err)
	}
	mplusImage = img
}

const (
	charHalfWidth = 6
	charFullWidth = 12
	charHeight    = 12
	lineHeight    = 16
)

type textImageParts struct {
	runes []rune
}

func (t *textImageParts) Len() int {
	return len(t.runes)
}

func (t *textImageParts) Src(index int) (int, int, int, int) {
	r := t.runes[index]
	x := int(r%256) * charFullWidth
	y := int(r/256) * lineHeight
	w := charHalfWidth
	h := lineHeight
	if r == '\n' {
		return 0, 0, 0, 0
	}
	if 0x100 <= r {
		w = charFullWidth
	}
	return x, y, x + w, y + h
}

func (t *textImageParts) Dst(index int) (int, int, int, int) {
	x := 0
	y := 0
	w := charFullWidth
	h := lineHeight
	for i := 0; i < index; i++ {
		if t.runes[i] == '\n' {
			x = 0
			y += lineHeight
			continue
		}
		if t.runes[i] < 0x100 {
			x += charHalfWidth
			continue
		}
		x += charFullWidth
	}
	if t.runes[index] < 0x100 {
		w = charHalfWidth
	}
	return x, y, x + w, y + h
}

func MeasureSize(text string) (int, int) {
	w := 0
	h := lineHeight
	cw := 0
	for _, r := range text {
		if r == '\n' {
			if w < cw {
				w = cw
			}
			cw = 0
			h += lineHeight
			continue
		}
		if r < 0x100 {
			cw += charHalfWidth
			continue
		}
		cw += charFullWidth
	}
	if w < cw {
		w = cw
	}
	return w, h
}

func DrawText(screen *ebiten.Image, text string, x, y int, scale int, color color.Color) error {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x), float64(y))
	r, g, b, a := color.RGBA()
	op.ColorM.Scale(float64(r>>8), float64(g>>8), float64(b>>8), float64(a>>8))
	op.ImageParts = &textImageParts{[]rune(text)}
	if err := screen.DrawImage(mplusImage, op); err != nil {
		return err
	}
	return nil
}
