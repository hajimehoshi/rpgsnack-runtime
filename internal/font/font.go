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
	"fmt"
	"image/color"
	"math"
	"strings"
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

type floatScaleImageCacheKey struct {
	text  string
	scale float64
	color color.Color
	lang  language.Tag
}

// floatScaleImageCache is an image cache with scales and its text.
var floatScaleImageCache = lru.New(10)

const (
	RenderingLineHeight = 18

	// These values are copied from github.com/hajimehoshi/bitmap's private values.
	mplusDotX = 4
	mplusDotY = 12
)

func MeasureSize(text string) (int, int) {
	w := fixed.I(0)
	h := fixed.I(0)
	for _, l := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		b, _ := boundString(face(1, lang.Get()), l)
		nw := b.Max.X - b.Min.X
		if nw > w {
			w = nw
		}
		h += fixed.I(RenderingLineHeight)
	}
	return w.Ceil(), h.Ceil()
}

func DrawText(screen *ebiten.Image, str string, ox, oy int, scale float64, textAlign data.TextAlign, color color.Color, displayTextRuneCount int) {
	DrawTextLang(screen, str, ox, oy, scale, textAlign, color, displayTextRuneCount, lang.Get())
}

var scratchPad *ebiten.Image

func init() {
	scratchPad, _ = ebiten.NewImage(16, 16, ebiten.FilterDefault)
}

var scratchPadM sync.Mutex

func DrawTextToScratchPad(str string, scale float64, lang language.Tag) {
	scratchPadM.Lock()
	f := face(int(math.Ceil(scale)), lang)
	text.Draw(scratchPad, str, f, 0, 0, color.White)
	scratchPadM.Unlock()
}

func isInteger(x float64) bool {
	return x == math.Floor(x)
}

func DrawTextLang(screen *ebiten.Image, str string, ox, oy int, scale float64, textAlign data.TextAlign, color color.Color, displayTextRuneCount int, lang language.Tag) {
	if isInteger(scale) {
		drawTextLangIntScale(screen, str, ox, oy, int(scale), textAlign, color, displayTextRuneCount, lang)
		return
	}
	drawTextLangFloatScale(screen, str, ox, oy, scale, textAlign, color, displayTextRuneCount, lang)
}

func drawTextLangFloatScale(screen *ebiten.Image, str string, ox, oy int, scale float64, textAlign data.TextAlign, color color.Color, displayTextRuneCount int, lang language.Tag) {
	k := floatScaleImageCacheKey{
		text:  str,
		scale: scale,
		color: color,
		lang:  lang,
	}
	if img, ok := floatScaleImageCache.Get(k); ok {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(ox), float64(oy))
		screen.DrawImage(img.(*ebiten.Image), op)
		return
	}

	scalei := int(math.Ceil(scale))
	w, h := MeasureSize(str)

	// src is an image that has texts scaled by `ceil(scale)`.
	src, _ := ebiten.NewImage(w*scalei, h*scalei, ebiten.FilterDefault)
	drawTextLangIntScale(src, str, 0, 0, scalei, textAlign, color, displayTextRuneCount, lang)

	// dst is an image that has texts scaled by `scale`.
	dst, _ := ebiten.NewImage(int(math.Ceil(float64(w)*scale)), int(math.Ceil(float64(h)*scale)), ebiten.FilterDefault)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale/float64(scalei), scale/float64(scalei))
	op.Filter = ebiten.FilterLinear
	dst.DrawImage(src, op)
	floatScaleImageCache.Add(k, dst)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(ox), float64(oy))
	screen.DrawImage(dst, op)
}

func drawTextLangIntScale(screen *ebiten.Image, str string, ox, oy int, scale int, textAlign data.TextAlign, color color.Color, displayTextRuneCount int, lang language.Tag) {
	f := face(scale, lang)
	m := f.Metrics()
	oy += (RenderingLineHeight*scale - m.Height.Round()) / 2

	b, _, _ := f.GlyphBounds('.')
	dotX := (-b.Min.X).Floor()

	str = strings.Replace(str, "\r\n", "\n", -1)
	lines := strings.Split(str, "\n")
	linesToShow := strings.Split(string([]rune(str)[:displayTextRuneCount]), "\n")

	for i, l := range linesToShow {
		x := ox + dotX
		y := oy + mplusDotY*scale
		_, a := boundString(f, lines[i])
		switch textAlign {
		case data.TextAlignLeft:
			// do nothing
		case data.TextAlignCenter:
			x -= a.Ceil() / 2
		case data.TextAlignRight:
			x -= a.Ceil()
		default:
			panic(fmt.Sprintf("font: invalid text align: %d", textAlign))
		}

		text.Draw(screen, l, f, x, y, color)
		oy += RenderingLineHeight * scale
	}
}
