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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

func ToValidContent(str string) string {
	return strings.Replace(str, "\r\n", "\n", -1)
}

type floatScaleImageCacheKey struct {
	text  string
	scale float64
	color color.Color
	lang  language.Tag
	align data.TextAlign
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

type DrawTextOptions struct {
	Scale        float64
	TextAlign    data.TextAlign
	Color        color.Color
	UseRuneCount bool
	RuneCount    int
	Language     language.Tag
}

func DrawText(screen *ebiten.Image, str string, ox, oy int, op *DrawTextOptions) {
	scale := op.Scale
	if scale == 0 {
		scale = consts.TextScale
	}

	str = ToValidContent(str)

	ta := op.TextAlign
	if ta == *new(data.TextAlign) {
		ta = data.TextAlignLeft
	}

	c := op.RuneCount
	if !op.UseRuneCount {
		c = len([]rune(str))
	}

	l := op.Language
	if l == language.Und {
		l = lang.Get()
	}

	if isInteger(scale) {
		drawTextLangIntScale(screen, str, ox, oy, int(scale), ta, op.Color, c, l)
		return
	}
	drawTextLangFloatScale(screen, str, ox, oy, scale, ta, op.Color, c, l)
}

func drawTextLangFloatScale(screen *ebiten.Image, str string, ox, oy int, scale float64, textAlign data.TextAlign, color color.Color, displayTextRuneCount int, lang language.Tag) {
	k := floatScaleImageCacheKey{
		text:  str,
		scale: scale,
		color: color,
		lang:  lang,
		align: textAlign,
	}
	var img *ebiten.Image
	w, h := MeasureSize(str)
	if cached, ok := floatScaleImageCache.Get(k); ok {
		img = cached.(*ebiten.Image)
	} else {
		scalei := int(math.Ceil(scale))

		// src is an image that has texts scaled by `ceil(scale)`.
		src, _ := ebiten.NewImage(w*scalei, h*scalei, ebiten.FilterDefault)
		x, y := 0, 0
		switch textAlign {
		case data.TextAlignLeft:
			// do nothing
		case data.TextAlignCenter:
			x += w * scalei / 2
		case data.TextAlignRight:
			x += w * scalei
		}
		drawTextLangIntScale(src, str, x, y, scalei, textAlign, color, displayTextRuneCount, lang)

		// dst is an image that has texts scaled by `scale`.
		dst, _ := ebiten.NewImage(int(math.Ceil(float64(w)*scale)), int(math.Ceil(float64(h)*scale)), ebiten.FilterDefault)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale/float64(scalei), scale/float64(scalei))
		op.Filter = ebiten.FilterLinear
		dst.DrawImage(src, op)
		floatScaleImageCache.Add(k, dst)

		img = dst
	}

	op := &ebiten.DrawImageOptions{}
	x, y := float64(ox), float64(oy)
	switch textAlign {
	case data.TextAlignLeft:
		// do nothing
	case data.TextAlignCenter:
		x -= float64(w) * scale / 2
	case data.TextAlignRight:
		x -= float64(w) * scale
	}
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func drawTextLangIntScale(screen *ebiten.Image, str string, ox, oy int, scale int, textAlign data.TextAlign, color color.Color, displayTextRuneCount int, lang language.Tag) {
	f := face(scale, lang)
	m := f.Metrics()
	oy += (RenderingLineHeight*scale - m.Height.Round()) / 2

	b, _, _ := f.GlyphBounds('.')
	dotX := (-b.Min.X).Floor()

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
