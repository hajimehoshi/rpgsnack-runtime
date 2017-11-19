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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
)

type Button struct {
	x             int
	y             int
	width         int
	height        int
	Visible       bool
	Text          string
	Disabled      bool
	Image         *ImagePart
	PressedImage  *ImagePart
	DisabledImage *ImagePart
	dropShadow    bool
	pressing      bool
	pressed       bool
	soundName     string
}

func NewButton(x, y, width, height int, soundName string) *Button {
	return &Button{
		x:          x,
		y:          y,
		width:      width,
		height:     height,
		Visible:    true,
		soundName:  soundName,
		dropShadow: false,
	}
}

func NewImageButton(x, y int, image *ImagePart, pressedImage *ImagePart, soundName string) *Button {
	w, h := image.Size()
	return &Button{
		x:             x,
		y:             y,
		width:         w,
		height:        h,
		Visible:       true,
		Image:         image,
		PressedImage:  pressedImage,
		DisabledImage: nil,
		soundName:     soundName,
		dropShadow:    true,
	}
}

func (b *Button) SetY(y int) {
	b.y = y
}

func (b *Button) Width() int {
	return b.width
}

func (b *Button) Height() int {
	return b.height
}

func (b *Button) SetOriginalSize(width, height int) {
	// TODO: Better name
	b.width = width
	b.height = height
}

func (b *Button) Pressed() bool {
	return b.pressed
}

func (b *Button) includesInput(offsetX, offsetY int) bool {
	x, y := input.Position()
	x = int(float64(x) / consts.TileScale)
	y = int(float64(y) / consts.TileScale)
	x -= offsetX
	y -= offsetY

	buttonWidth := b.width
	buttonHeight := b.height
	buttonX := b.x
	buttonY := b.y

	if buttonX <= x && x < buttonX+buttonWidth && buttonY <= y && y < buttonY+buttonHeight {
		return true
	}
	return false
}

func (b *Button) update(visible bool, offsetX, offsetY int) {
	b.pressed = false
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
		b.pressed = true
		audio.PlaySE(b.soundName, 1.0)
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

	colorM := &ebiten.ColorM{}
	if b.Disabled {
		colorM.ChangeHSV(0, 0, 1)
		colorM.Scale(0.5, 0.5, 0.5, 1)
	}

	if b.Image != nil {
		image := b.Image
		if b.Disabled {
			if b.DisabledImage != nil {
				image = b.DisabledImage
			}
		} else {
			if b.pressing {
				if b.PressedImage == nil {
					colorM.ChangeHSV(0, 0, 1)
					colorM.Scale(0.5, 0.5, 0.5, 1)
				} else {
					image = b.PressedImage
				}
			}
		}
		image.Draw(screen, geoM, colorM)
	} else {
		img := assets.GetImage("system/9patch_test_off.png")
		if b.pressing {
			img = assets.GetImage("system/9patch_test_on.png")
		}
		drawNinePatches(screen, img, b.width, b.height, geoM, colorM)
	}

	_, th := font.MeasureSize(b.Text)
	tx := (b.x + offsetX) * consts.TileScale
	tx += b.width * consts.TileScale / 2

	ty := (b.y + offsetY) * consts.TileScale
	ty += (b.height*consts.TileScale - th*consts.TextScale) / 2

	var c color.Color = color.White
	if b.Disabled {
		c = color.RGBA{0x80, 0x80, 0x80, 0xff}
	}
	if b.dropShadow {
		font.DrawText(screen, b.Text, tx+consts.TextScale, ty+consts.TextScale, consts.TextScale, data.TextAlignCenter, color.Black)
	}
	font.DrawText(screen, b.Text, tx, ty, consts.TextScale, data.TextAlignCenter, c)
}
