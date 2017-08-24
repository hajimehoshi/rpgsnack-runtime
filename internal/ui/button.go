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
	X             int
	Y             int
	AnchorX       float64
	AnchorY       float64
	ScaleX        int
	ScaleY        int
	Width         int
	Height        int
	Visible       bool
	Text          string
	Disabled      bool
	Image         *ebiten.Image
	PressedImage  *ebiten.Image
	DisabledImage *ebiten.Image
	pressing      bool
	pressed       bool
	soundName     string
}

func NewButton(x, y, width, height int, soundName string) *Button {
	return &Button{
		X:         x,
		Y:         y,
		ScaleX:    1,
		ScaleY:    1,
		Width:     width,
		Height:    height,
		Visible:   true,
		soundName: soundName,
	}
}

func NewImageButton(x, y int, image *ebiten.Image, pressedImage *ebiten.Image, soundName string) *Button {
	w, h := image.Size()
	return &Button{
		X:             x,
		Y:             y,
		ScaleX:        1,
		ScaleY:        1,
		Width:         w,
		Height:        h,
		Visible:       true,
		Image:         image,
		PressedImage:  pressedImage,
		DisabledImage: nil,
		soundName:     soundName,
	}
}

func (b *Button) Pressing() bool {
	return b.pressing
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

	buttonWidth := b.ScaleX * b.Width
	buttonHeight := b.ScaleY * b.Height
	buttonX := b.X - int(float64(buttonWidth)*b.AnchorX)
	buttonY := b.Y - int(float64(buttonHeight)*b.AnchorY)

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
	if b.includesInput(offsetX, offsetY) {
		b.pressing = true
	} else {
		b.pressing = false
	}
}

func (b *Button) Update() {
	b.update(true, 0, 0)
}

func (b *Button) UpdateAsChild(visible bool, offsetX, offsetY int) {
	b.update(visible, offsetX, offsetY)
}

func (b *Button) Draw(screen *ebiten.Image) {
	if !b.Visible {
		return
	}
	if b.Image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-float64(b.Width)*b.AnchorX, -float64(b.Height)*b.AnchorY)
		op.GeoM.Scale(float64(b.ScaleX*consts.TileScale), float64(b.ScaleY*consts.TileScale))
		op.GeoM.Translate(float64(b.X*consts.TileScale), float64(b.Y*consts.TileScale))

		image := b.Image
		if b.Disabled {
			if b.DisabledImage == nil {
				op.ColorM.ChangeHSV(0, 0, 1)
				op.ColorM.Scale(0.5, 0.5, 0.5, 1)
			} else {
				image = b.DisabledImage
			}
		} else {
			if b.pressing {
				if b.PressedImage == nil {
					op.ColorM.ChangeHSV(0, 0, 1)
					op.ColorM.Scale(0.5, 0.5, 0.5, 1)
				} else {
					image = b.PressedImage
				}
			}
		}
		screen.DrawImage(image, op)
		return
	}
	img := assets.GetImage("system/9patch_test_off.png")
	if b.pressing {
		img = assets.GetImage("system/9patch_test_on.png")
	}
	geoM := &ebiten.GeoM{}
	geoM.Translate(float64(b.X)-float64(b.Width)*b.AnchorX, float64(b.Y)-float64(b.Height)*b.AnchorY)
	geoM.Scale(float64(b.ScaleX*consts.TileScale), float64(b.ScaleY*consts.TileScale))
	colorM := &ebiten.ColorM{}
	if b.Disabled {
		colorM.ChangeHSV(0, 0, 1)
		colorM.Scale(0.5, 0.5, 0.5, 1)
	}
	drawNinePatches(screen, img, b.Width, b.Height, geoM, colorM)

	_, th := font.MeasureSize(b.Text)
	tx := b.X*b.ScaleX*consts.TileScale + b.Width*consts.TileScale*b.ScaleX/2
	ty := b.Y*b.ScaleY*consts.TileScale + (b.Height*b.ScaleY*consts.TileScale-th*consts.TextScale*b.ScaleY)/2
	var c color.Color = color.White
	if b.Disabled {
		c = color.RGBA{0x80, 0x80, 0x80, 0xff}
	}
	font.DrawText(screen, b.Text, tx, ty, consts.TextScale, data.TextAlignCenter, c)
}
