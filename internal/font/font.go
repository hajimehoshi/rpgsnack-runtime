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
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

const (
	renderingLineHeight = 18
)

func MeasureSize(text string) (int, int) {
	w := fixed.I(0)
	h := fixed.I(0)
	for _, l := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		b, _ := font.BoundString(face(1, lang.Get()), l)
		nw := b.Max.X - b.Min.X
		if nw > w {
			w = nw
		}
		h += fixed.I(renderingLineHeight)
	}
	return w.Ceil(), h.Ceil()
}

func DrawText(screen *ebiten.Image, str string, ox, oy int, scale int, textAlign data.TextAlign, color color.Color, displayTextRuneCount int) {
	DrawTextLang(screen, str, ox, oy, scale, textAlign, color, displayTextRuneCount, lang.Get())
}

func DrawTextLang(screen *ebiten.Image, str string, ox, oy int, scale int, textAlign data.TextAlign, color color.Color, displayTextRuneCount int, lang language.Tag) {
	str = string([]rune(str)[:displayTextRuneCount])

	f := face(scale, lang)
	m := f.Metrics()
	oy += (renderingLineHeight*scale - m.Height.Round()) / 2

	b, _, _ := f.GlyphBounds('M')
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
