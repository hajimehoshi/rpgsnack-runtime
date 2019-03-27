// Copyright 2019 Hajime Hoshi
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
	"bytes"
	"compress/gzip"
	"image"
	"image/color"
	"io/ioutil"

	"github.com/hajimehoshi/bitmapfont"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type binaryImage struct {
	bits   []byte
	width  int
	height int
	bounds image.Rectangle
}

func newBinaryImage(bits []byte, width, height int) *binaryImage {
	return &binaryImage{
		bits:   bits,
		width:  width,
		height: height,
		bounds: image.Rect(0, 0, width, height),
	}
}

func (b *binaryImage) At(i, j int) color.Color {
	if i < b.bounds.Min.X || j < b.bounds.Min.Y || i >= b.bounds.Max.X || j >= b.bounds.Max.Y {
		return color.Alpha{0}
	}
	idx := b.width*j + i
	if (b.bits[idx/8]>>uint(7-idx%8))&1 != 0 {
		return color.Alpha{0xff}
	}
	return color.Alpha{0}
}

func (b *binaryImage) ColorModel() color.Model {
	return color.AlphaModel
}

func (b *binaryImage) Bounds() image.Rectangle {
	return b.bounds
}

func (b *binaryImage) SubImage(r image.Rectangle) image.Image {
	bounds := r.Intersect(b.bounds)
	if bounds.Empty() {
		return &binaryImage{}
	}
	return &binaryImage{
		bits:   b.bits,
		width:  b.width,
		height: b.height,
		bounds: bounds,
	}
}

type tag int

const (
	tagZhHans tag = iota
	tagZhHant
)

type zhBitmap struct {
	tag   tag
	image *binaryImage
}

var (
	gothic12r_sc *zhBitmap
	gothic12r_tc *zhBitmap
)

func init() {
	s, err := gzip.NewReader(bytes.NewReader(zhglyphs))
	if err != nil {
		panic(err)
	}
	defer s.Close()

	bits, err := ioutil.ReadAll(s)
	if err != nil {
		panic(err)
	}

	img := newBinaryImage(bits, 12*256, 16*256)
	gothic12r_sc = &zhBitmap{tagZhHans, img}
	gothic12r_tc = &zhBitmap{tagZhHant, img}
}

func (z *zhBitmap) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	if 0x4e00 <= r && r <= 0x9fff {
		rect, a, _ := z.GlyphBounds(r)
		w := (rect.Max.X - rect.Min.X).Floor()
		h := (rect.Max.Y - rect.Min.Y).Floor()

		dotX := fixed.I(mplusDotX)
		dotY := fixed.I(mplusDotY)
		dx := (dot.X - dotX).Floor()
		dy := (dot.Y - dotY).Floor()
		dr = image.Rect(dx, dy, dx+w, dy+h)

		mx := (int(r) % 256) * w
		my := (int(r) / 256) * h
		mask = z.image.SubImage(image.Rect(mx, my, mx+w, my+h))
		maskp = image.Pt(mx, my)
		advance = a
		ok = true
		return
	}

	ox := 0
	oy := 0
	if z.tag == tagZhHant {
		if r == '、' {
			r = '，'
		}
		if r == '，' || r == '。' {
			ox = 3
			oy = -3
		}
	}

	dr, mask, maskp, advance, ok = bitmapfont.Gothic12r.Glyph(dot, r)
	dr.Min.X += ox
	dr.Min.Y += oy
	dr.Max.X += ox
	dr.Max.Y += oy
	return
}

func (*zhBitmap) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	return bitmapfont.Gothic12r.GlyphBounds(r)
}

func (*zhBitmap) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	return bitmapfont.Gothic12r.GlyphAdvance(r)
}

func (*zhBitmap) Kern(r0, r1 rune) fixed.Int26_6 {
	return bitmapfont.Gothic12r.Kern(r0, r1)
}

func (*zhBitmap) Metrics() font.Metrics {
	return bitmapfont.Gothic12r.Metrics()
}

func (*zhBitmap) Close() error {
	return nil
}
