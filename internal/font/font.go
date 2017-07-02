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
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

var positions map[rune]int

const (
	charHalfWidth       = 6
	charFullWidth       = 12
	charHeight          = 12
	lineHeight          = 16
	renderingLineHeight = 18
)

func runeWidth(r rune) int {
	if r < 0x100 {
		return charHalfWidth
	}
	return charFullWidth
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
		cw += runeWidth(r)
	}
	if w < cw {
		w = cw
	}
	return w, h
}

func DrawText(screen *ebiten.Image, text string, ox, oy int, scale int, textAlign data.TextAlign, color color.Color) {
	if positions == nil {
		positions = map[rune]int{}
		b := assets.GetResource("images/fonts/mplus_positions")
		for i := 0; i < len(b)/4; i++ {
			r := rune(b[4*i]) + rune(b[4*i+1])<<8
			pos := int(b[4*i+2]) + int(b[4*i+3])<<8
			positions[r] = pos
		}
	}
	op := &ebiten.DrawImageOptions{}
	r, g, b, a := color.RGBA()
	op.ColorM.Scale(float64(r)/65535, float64(g)/65535, float64(b)/65535, float64(a)/65535)

	x := 0
	lineWidths := []int{}
	for _, r := range text {
		if r == '\n' {
			lineWidths = append(lineWidths, x)
			x = 0
			continue
		}
		x += runeWidth(r)
	}
	lineWidths = append(lineWidths, x)

	dx := 0
	dy := (renderingLineHeight - lineHeight) / 2
	l := 0
	img := assets.GetImage("fonts/mplus.compacted.png")
	for _, r := range text {
		if r == '\n' {
			dx = 0
			dy += renderingLineHeight
			l++
			continue
		}
		// TODO: Use unicode package to detect space
		if r == ' ' || r == '　' {
			dx += runeWidth(r)
			continue
		}
		pos, ok := positions[r]
		if !ok {
			continue
		}
		x := pos % 256 * charFullWidth
		y := pos / 256 * lineHeight
		w := runeWidth(r)
		h := lineHeight
		src := image.Rect(x, y, x+w, y+h)
		op.SourceRect = &src

		lw := lineWidths[l]
		op.GeoM.Reset()
		switch textAlign {
		case data.TextAlignLeft:
			op.GeoM.Translate(float64(dx), float64(dy))
		case data.TextAlignCenter:
			op.GeoM.Translate(float64(dx-lw/2), float64(dy))
		case data.TextAlignRight:
			op.GeoM.Translate(float64(dx-lw), float64(dy))
		default:
			panic("not reached")
		}
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(ox), float64(oy))
		screen.DrawImage(img, op)
		dx += w
	}
}
