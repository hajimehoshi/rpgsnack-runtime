// Copyright 2017 Hajime Hoshi
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
	"image/color"

	"github.com/hajimehoshi/ebiten"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

type Button struct {
	x             int
	y             int
	Width         int
	Height        int
	TouchExpand   int
	Visible       bool
	Text          string
	Disabled      bool
	Image         *ImagePart
	PressedImage  *ImagePart
	DisabledImage *ImagePart
	dropShadow    bool
	pressing      bool
	soundName     string
	showFrame     bool
	TextColor     color.Color
	Lang          language.Tag
	onPressed     func(button *Button)
}

func NewButton(x, y, width, height int, soundName string) *Button {
	return &Button{
		x:          x,
		y:          y,
		Width:      width,
		Height:     height,
		Visible:    true,
		soundName:  soundName,
		dropShadow: false,
		showFrame:  true,
		TextColor:  color.White,
	}
}

func NewTextButton(x, y, width, height int, soundName string) *Button {
	return &Button{
		x:          x,
		y:          y,
		Width:      width,
		Height:     height,
		Visible:    true,
		soundName:  soundName,
		dropShadow: false,
		showFrame:  false,
		TextColor:  color.White,
	}
}

func NewImageButton(x, y int, image *ImagePart, pressedImage *ImagePart, soundName string) *Button {
	w, h := image.Size()
	return &Button{
		x:             x,
		y:             y,
		Width:         w,
		Height:        h,
		Visible:       true,
		Image:         image,
		PressedImage:  pressedImage,
		DisabledImage: nil,
		soundName:     soundName,
		dropShadow:    true,
		showFrame:     true,
		TextColor:     color.White,
	}
}

func (b *Button) SetY(y int) {
	b.y = y
}

func (b *Button) SetOnPressed(onPressed func(button *Button)) {
	b.onPressed = onPressed
}

func (b *Button) includesInput(offsetX, offsetY int) bool {
	x, y := input.Position()
	x = int(float64(x) / consts.TileScale)
	y = int(float64(y) / consts.TileScale)
	x -= offsetX
	y -= offsetY

	buttonWidth := b.Width + b.TouchExpand*2
	buttonHeight := b.Height + b.TouchExpand*2
	buttonX := b.x - b.TouchExpand
	buttonY := b.y - b.TouchExpand

	if buttonX <= x && x < buttonX+buttonWidth && buttonY <= y && y < buttonY+buttonHeight {
		return true
	}
	return false
}

func (b *Button) update(visible bool, offsetX, offsetY int) {
	if !visible {
		return
	}
	if !b.Visible {
		return
	}
	if b.Disabled {
		return
	}
	if !b.pressing {
		if !input.Triggered() {
			return
		}
	}
	if !input.Pressed() {
		b.pressing = false
		audio.PlaySE(b.soundName, 1.0)
		if b.onPressed != nil {
			b.onPressed(b)
		}
		return
	}
	b.pressing = b.includesInput(offsetX, offsetY)
}

func (b *Button) Update() {
	b.update(true, 0, 0)
}

func (b *Button) UpdateAsChild(visible bool, offsetX, offsetY int) {
	b.update(visible, offsetX, offsetY)
}

func (b *Button) Draw(screen *ebiten.Image) {
	b.DrawAsChild(screen, 0, 0)
}

func (b *Button) DrawAsChild(screen *ebiten.Image, offsetX, offsetY int) {
	if !b.Visible {
		return
	}

	geoM := &ebiten.GeoM{}
	geoM.Translate(float64(b.x+offsetX), float64(b.y+offsetY))
	geoM.Scale(consts.TileScale, consts.TileScale)

	opacity := uint8(255)
	if b.showFrame {
		if b.Image != nil {
			var colorM *ebiten.ColorM
			image := b.Image
			if b.Disabled {
				if b.DisabledImage != nil {
					image = b.DisabledImage
				}
			} else {
				if b.pressing {
					if b.PressedImage == nil {
						colorM = &ebiten.ColorM{}
						colorM.ChangeHSV(0, 0, 1)
						colorM.Scale(0.5, 0.5, 0.5, 1)
					} else {
						image = b.PressedImage
					}
				}
			}
			image.Draw(screen, geoM, colorM)
		} else {
			img := assets.GetImage("system/common/9patch_frame_off.png")
			if b.pressing {
				img = assets.GetImage("system/common/9patch_frame_on.png")
			}

			var colorM *ebiten.ColorM
			if b.Disabled {
				colorM = &ebiten.ColorM{}
				colorM.ChangeHSV(0, 0, 1)
				colorM.Scale(0.5, 0.5, 0.5, 1)
			}
			drawNinePatches(screen, img, b.Width, b.Height, geoM, colorM)
		}
	} else {
		if b.pressing {
			opacity = uint8(127)
		}
	}

	_, th := font.MeasureSize(b.Text)
	tx := (b.x + offsetX) * consts.TileScale
	tx += b.Width * consts.TileScale / 2

	ty := (b.y + offsetY) * consts.TileScale
	ty += (b.Height*consts.TileScale - th*consts.TextScale) / 2

	cr, cg, cb, ca := b.TextColor.RGBA()
	r8 := uint8(cr >> 8)
	g8 := uint8(cg >> 8)
	b8 := uint8(cb >> 8)

	var c color.Color = color.RGBA{r8, g8, b8, uint8(ca * uint32(opacity) / 255)}
	if b.Disabled {
		c = color.RGBA{r8, g8, b8, uint8(ca * uint32(opacity) / (2 * 255))}
	}
	l := b.Lang
	if l == language.Und {
		l = lang.Get()
	}
	if b.dropShadow {
		font.DrawTextLang(screen, b.Text, tx+consts.TextScale, ty+consts.TextScale, consts.TextScale, data.TextAlignCenter, color.Black, len([]rune(b.Text)), l)
	}
	font.DrawTextLang(screen, b.Text, tx, ty, consts.TextScale, data.TextAlignCenter, c, len([]rune(b.Text)), l)
}
