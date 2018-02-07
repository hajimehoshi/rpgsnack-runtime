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
	"strings"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/text"
	"github.com/hajimehoshi/go-mplus-bitmap"
	"golang.org/x/image/font"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

const (
	renderingLineHeight = 18
)

var (
	fonts = map[int]font.Face{}
)

func MeasureSize(text string) (int, int) {
	w := 0
	h := 0
	for _, l := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		b, _ := font.BoundString(mplusbitmap.Gothic12r, l)
		w += (b.Max.X - b.Min.X).Ceil()
		h += renderingLineHeight
	}
	return w, h
}

func DrawText(screen *ebiten.Image, str string, ox, oy int, scale int, textAlign data.TextAlign, color color.Color, displayTextRuneCount int) {
	// Use the same instance to use text cache efficiently.
	f, ok := fonts[scale]
	if !ok {
		f = Scale(mplusbitmap.Gothic12r, scale)
		fonts[scale] = f
	}

	b, _, _ := f.GlyphBounds('.')
	dotX := -b.Min.X
	dotY := -b.Min.Y
	for _, l := range strings.Split(str, "\n") {
		x := ox + dotX.Floor()
		y := oy + dotY.Floor()
		_, a := font.BoundString(f, l)
		switch textAlign {
		case data.TextAlignLeft:
			// do nothing
		case data.TextAlignCenter:
			x -= a.Ceil() / 2
		case data.TextAlignRight:
			x -= a.Ceil()
		default:
			panic("not reached")
		}

		text.Draw(screen, l, f, x, y, color)
		oy += renderingLineHeight * scale
	}
}
