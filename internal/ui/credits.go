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
	"image/color"
	"math"
	"regexp"
	"strings"

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

func NewCredits() *Credits {
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

func (c *Credits) boost() bool {
	if !input.Pressed() {
		return false
	}
	if c.closeButton == nil {
		return true
	}
	return !includesInput(0, 0, c.closeButton.region())
}

func (c *Credits) SetCloseButtonVisible(visible bool) {
	c.closeButton.visible = visible
}

func (c *Credits) Update() {
	if !c.visible {
		return
	}
	if c.closeButton != nil {
		// TODO: This function should return immediately when input is handled.
		c.closeButton.HandleInput(0, 0)
	}
	if c.finished {
		c.Hide()
	}
	if c.boost() {
		c.scrollY += 8
	} else {
		c.scrollY++
	}
}

func (c *Credits) Draw(screen *ebiten.Image) {
	if !c.visible {
		return
	}
	screen.Fill(color.Black)

	_, sy := screen.Size()
	const (
		sx        = 480 - 32
		ox        = 16
		oy        = 16
		baseScale = 2
	)
	x := ox
	y := oy + sy - c.scrollY
	for _, s := range c.data.Sections {
		x = ox
		h := font.RenderingLineHeight * baseScale
		if -h <= y && y < sy {
			font.DrawTextLang(screen, s.Header, x, y, baseScale, data.TextAlignLeft, headerColor(&s), len([]rune(s.Header)), language.English)
		}
		y += h

		n := ((len(s.Body)-1)/columnNum(&s) + 1)
		h = int(math.Ceil(float64(n) * font.RenderingLineHeight * fontScale(&s)))
		if -h <= y && y < sy {
			for i := 0; i < columnNum(&s); i++ {
				var body []string
				for j := i; j < len(s.Body); j += columnNum(&s) {
					body = append(body, s.Body[j])
				}
				x = ox + i*(sx/columnNum(&s))
				str := strings.Join(body, "\n")
				font.DrawTextLang(screen, str, x, y, fontScale(&s), data.TextAlignLeft, color.White, len([]rune(str)), language.English)
			}
		}

		y += h
		y += font.RenderingLineHeight * baseScale / 2
	}
	if y <= 0 {
		c.finished = true
	}
	if c.closeButton != nil {
		c.closeButton.DrawAsChild(screen, 0, 0)
	}
}
