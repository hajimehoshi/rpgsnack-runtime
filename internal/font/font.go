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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
)

var positions = map[rune]int{}

func init() {
	b := assets.MustAsset("images/mplus_positions")
	for i := 0; i < len(b)/4; i++ {
		r := rune(b[4*i]) + rune(b[4*i+1])<<8
		pos := int(b[4*i+2]) + int(b[4*i+3])<<8
		positions[r] = pos
	}
}

const (
	charHalfWidth       = 6
	charFullWidth       = 12
	charHeight          = 12
	lineHeight          = 16
	renderingLineHeight = 18
)

type textImageParts struct {
	runes []rune
}

func (t *textImageParts) Len() int {
	return len(t.runes)
}

func (t *textImageParts) Src(index int) (int, int, int, int) {
	r := t.runes[index]
	pos, ok := positions[r]
	if !ok {
		return 0, 0, 0, 0
	}
	x := pos % 256 * charFullWidth
	y := pos / 256 * lineHeight
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
	y := (renderingLineHeight - lineHeight) / 2
	for i := 0; i < index; i++ {
		if t.runes[i] == '\n' {
			x = 0
			y += renderingLineHeight
			continue
		}
		if t.runes[i] < 0x100 {
			x += charHalfWidth
			continue
		}
		x += charFullWidth
	}
	w := charFullWidth
	h := lineHeight
	if t.runes[index] < 0x100 {
		w = charHalfWidth
	}
	return x, y, x + w, y + h
}

func MeasureSize(text string) (int, int) {
	w := 0
	h := renderingLineHeight
	cw := 0
	for _, r := range text {
		if r == '\n' {
			if w < cw {
				w = cw
			}
			cw = 0
			h += renderingLineHeight
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
	op.ColorM.Scale(float64(r>>8)/255, float64(g>>8)/255, float64(b>>8)/255, float64(a>>8)/255)
	op.ImageParts = &textImageParts{[]rune(text)}
	mplusImage := assets.GetImage("mplus.compacted.png")
	if err := screen.DrawImage(mplusImage, op); err != nil {
		return err
	}
	return nil
}
