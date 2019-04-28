// Copyright 2019 The RPGSnack Authors
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

package ui

import (
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"math"
	"regexp"
	"unicode"

	"github.com/hajimehoshi/ebiten"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
)

var reColor = regexp.MustCompile(`^#([0-9a-fA-F]{6})$`)

func headerColor(c *data.CreditsSection) color.Color {
	if c.HeaderColor == "" {
		return color.White
	}
	m := reColor.FindStringSubmatch(c.HeaderColor)
	if len(m) < 2 {
		return color.White
	}
	bin, err := hex.DecodeString(m[1])
	if err != nil {
		panic(fmt.Sprintf("ui: invalid color: %s", m[1]))
	}

	r := bin[0]
	g := bin[1]
	b := bin[2]
	return color.RGBA{r, g, b, 0xff}
}

func columnNum(c *data.CreditsSection) int {
	if 15 <= len(c.Body) {
		return 2
	}
	return 1
}

func fontScale(c *data.CreditsSection) float64 {
	if 1 < columnNum(c) {
		return 1.5
	}
	return 2
}

type Credits struct {
	closeButton *Button
	visible     bool
	data        *data.Credits
	scrollY     int
	finished    bool
}

func NewCredits(useCloseButton bool) *Credits {
	if !useCloseButton {
		return &Credits{}
	}

	closeButton := NewImageButton(
		140,
		4,
		assets.GetImage("system/common/cancel_off.png"),
		assets.GetImage("system/common/cancel_on.png"),
		"system/cancel",
	)

	c := &Credits{
		closeButton: closeButton,
	}
	closeButton.SetOnPressed(func(_ *Button) {
		c.Hide()
	})
	return c
}

func (c *Credits) SetData(credits *data.Credits) {
	c.data = credits
}

func (c *Credits) Visible() bool {
	return c.visible
}

func (c *Credits) Show() {
	c.visible = true
	c.scrollY = 0
	c.finished = false
}

func (c *Credits) Hide() {
	c.visible = false
}

func (c *Credits) dash() bool {
	if !input.Pressed() {
		return false
	}
	if c.closeButton == nil {
		return true
	}
	return !c.closeButton.includesInput(0, 0)
}

func (c *Credits) Update() {
	if !c.visible {
		return
	}
	if c.closeButton != nil {
		c.closeButton.UpdateAsChild(c.visible, 0, 0)
	}
	if c.finished {
		c.Hide()
	}
	if c.dash() {
		c.scrollY += 8
	} else {
		c.scrollY++
	}
}

const creditsFontScale = 3
const numCharPerLine = 128
const maxCharCodePoint = 256

var creditsFont *ebiten.Image

func init() {
	var rs []rune
	for i := rune(0); i < maxCharCodePoint; i++ {
		if !unicode.IsPrint(i) {
			rs = append(rs, ' ')
		} else {
			rs = append(rs, i)
		}
		if (i+1)%numCharPerLine == 0 {
			rs = append(rs, '\n')
		}
	}
	str := string(rs)
	w, h := font.MeasureSize(str)

	creditsFont, _ = ebiten.NewImage(w*creditsFontScale, h*creditsFontScale, ebiten.FilterDefault)
	font.DrawTextLang(creditsFont, str, 0, 0, creditsFontScale, data.TextAlignLeft, color.White, maxCharCodePoint, language.English)
}

func drawCreditsText(img *ebiten.Image, str string, x, y int, scale float64, clr color.Color) {
	const (
		w = 6
		h = font.RenderingLineHeight
	)

	op := &ebiten.DrawImageOptions{}

	rf := 0.0
	gf := 0.0
	bf := 0.0
	af := 0.0
	if r, g, b, a := clr.RGBA(); a > 0 {
		af = float64(a) / 0xffff
		rf = float64(r) / float64(a)
		gf = float64(g) / float64(a)
		bf = float64(b) / float64(a)
	}
	op.ColorM.Scale(rf, gf, bf, af)
	op.Filter = ebiten.FilterLinear
	for i, r := range ([]rune)(str) {
		if !unicode.IsPrint(r) {
			continue
		}
		if r >= maxCharCodePoint {
			continue
		}
		op.GeoM.Reset()
		op.GeoM.Scale(scale/creditsFontScale, scale/creditsFontScale)
		op.GeoM.Translate(float64(x)+float64(i)*w*scale, float64(y))
		x := int(r%numCharPerLine) * w * creditsFontScale
		y := int(r/numCharPerLine) * h * creditsFontScale
		rect := image.Rect(x, y, x+w*creditsFontScale, y+h*creditsFontScale)
		img.DrawImage(creditsFont.SubImage(rect).(*ebiten.Image), op)
	}
}

func (c *Credits) Draw(screen *ebiten.Image) {
	if !c.visible {
		return
	}
	screen.Fill(color.Black)

	_, sy := screen.Size()
	const (
		sx         = 480 - 32
		ox         = 16
		oy         = 16
		baseScale  = 2
		lineHeight = 16
	)
	x := ox
	y := oy + sy - c.scrollY
	for _, s := range c.data.Sections {
		x = ox
		if -lineHeight*baseScale <= y && y < sy {
			drawCreditsText(screen, s.Header, x, y, baseScale, headerColor(&s))
		}
		y += lineHeight * baseScale
		x = ox
		for i, l := range s.Body {
			x = ox + (sx/columnNum(&s))*(i%columnNum(&s))
			if -lineHeight*baseScale <= y && y < sy {
				drawCreditsText(screen, l, x, y, fontScale(&s), color.White)
			}
			if i%columnNum(&s) == columnNum(&s)-1 || i == len(s.Body)-1 {
				if fontScale(&s) > 1 {
					y += int(math.Ceil(lineHeight * fontScale(&s)))
				} else {
					y += lineHeight + 2
				}
			}
		}
		y += lineHeight * baseScale / 2
	}
	if y <= 0 {
		c.finished = true
	}
	if c.closeButton != nil {
		c.closeButton.DrawAsChild(screen, 0, 0)
	}
}
