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

package ui2

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
)

type Button struct {
	x                 int
	y                 int
	width             int
	height            int
	scale             float64
	touchExpand       int
	visible           bool
	text              string
	disabled          bool
	image             *ebiten.Image
	pressedImage      *ebiten.Image
	disabledImage     *ebiten.Image
	dropShadow        bool
	pressingCount     int
	pressing          bool
	soundName         string
	showFrame         bool
	textColor         color.Color
	textDisabledColor color.Color
	lang              language.Tag

	onPressed func(button *Button)
}

func NewButton(x, y, width, height int, soundName string) *Button {
	return &Button{
		x:          x,
		y:          y,
		width:      width,
		height:     height,
		scale:      1,
		visible:    true,
		soundName:  soundName,
		dropShadow: false,
		showFrame:  true,
		textColor:  color.White,
	}
}

func NewTextButton(x, y, width, height int, soundName string) *Button {
	return &Button{
		x:          x,
		y:          y,
		width:      width,
		height:     height,
		scale:      1,
		visible:    true,
		soundName:  soundName,
		dropShadow: false,
		showFrame:  false,
		textColor:  color.White,
	}
}

func NewImageButton(x, y int, image *ebiten.Image, pressedImage *ebiten.Image, soundName string) *Button {
	w, h := image.Size()
	return &Button{
		x:             x,
		y:             y,
		width:         w,
		height:        h,
		scale:         1,
		visible:       true,
		image:         image,
		pressedImage:  pressedImage,
		disabledImage: nil,
		soundName:     soundName,
		dropShadow:    true,
		showFrame:     true,
		textColor:     color.White,
	}
}

func (b *Button) SetX(x int) {
	b.x = x
}

func (b *Button) SetY(y int) {
	b.y = y
}

func (b *Button) SetWidth(width int) {
	b.width = width
}

func (b *Button) SetText(text string) {
	b.text = text
}

func (b *Button) SetScale(scale float64) {
	b.scale = scale
}

func (b *Button) SetColor(clr color.Color) {
	b.textColor = clr
}

func (b *Button) SetDisabledColor(clr color.Color) {
	b.textDisabledColor = clr
}

func (b *Button) Show() {
	b.visible = true
}

func (b *Button) Hide() {
	b.visible = false
}

func (b *Button) Enable() {
	b.disabled = false
}

func (b *Button) Disable() {
	b.disabled = true
}

func (b *Button) SetOnPressed(onPressed func(button *Button)) {
	b.onPressed = onPressed
}

func (b *Button) Region() image.Rectangle {
	if b.image != nil {
		w, h := b.image.Size()
		w = int(float64(w) * b.scale)
		h = int(float64(h) * b.scale)
		return image.Rect(b.x-b.touchExpand, b.y-b.touchExpand, b.x+w+b.touchExpand, b.y+h+b.touchExpand)
	}
	return image.Rect(b.x-b.touchExpand, b.y-b.touchExpand, b.x+b.width+b.touchExpand, b.y+b.height+b.touchExpand)
}

func (b *Button) HandleInput(offsetX, offsetY int) bool {
	if !b.visible {
		return false
	}
	if b.disabled {
		return false
	}

	if input.Released() {
		if !b.pressing {
			return false
		}
		b.pressing = false
		if !includesInput(offsetX, offsetY, b.Region()) {
			return false
		}
		if b.onPressed != nil {
			b.onPressed(b)
		}
		if b.soundName != "" {
			audio.PlaySE(b.soundName, 1.0)
		}
		return true
	}

	if !input.Pressed() {
		b.pressing = false
		b.pressingCount = 0
		return false
	}

	if !includesInput(offsetX, offsetY, b.Region()) {
		b.pressing = false
		b.pressingCount = 0
		return false
	}

	if input.Triggered() {
		b.pressingCount = 3
		return false
	}

	if b.pressingCount > 0 {
		b.pressingCount--
		b.pressing = b.pressingCount == 0
	}
	return b.pressing
}

func (b *Button) Update() {
}

func (b *Button) DrawAsChild(screen *ebiten.Image, offsetX, offsetY int) {
	if !b.visible {
		return
	}

	op := &ebiten.DrawImageOptions{}
	if b.image != nil {
		op.GeoM.Scale(float64(b.scale), float64(b.scale))
	}
	op.GeoM.Translate(float64(b.x+offsetX), float64(b.y+offsetY))

	opacity := uint8(255)
	if b.showFrame {
		if b.image != nil {
			image := b.image
			if b.disabled {
				if b.disabledImage != nil {
					image = b.disabledImage
				}
			} else {
				if b.pressing {
					if b.pressedImage == nil {
						op.ColorM.ChangeHSV(0, 0, 1)
						op.ColorM.Scale(0.5, 0.5, 0.5, 1)
					} else {
						image = b.pressedImage
					}
				}
			}
			screen.DrawImage(image, op)
		} else {
			img := assets.GetImage("system/common/9patch_frame_off.png")
			if b.pressing {
				img = assets.GetImage("system/common/9patch_frame_on.png")
			}

			if b.disabled {
				op.ColorM.ChangeHSV(0, 0, 1)
				op.ColorM.Scale(0.5, 0.5, 0.5, 1)
			}
			drawNinePatches(screen, img, b.width, b.height, &op.GeoM, &op.ColorM)
		}
	} else {
		if b.pressing {
			opacity = uint8(127)
		}
	}

	_, th := font.MeasureSize(b.text)
	tx := (b.x + offsetX)
	tx += b.width / 2

	ty := (b.y + offsetY)
	ty += (b.height - int(float64(th)*b.scale)) / 2

	cr, cg, cb, ca := b.textColor.RGBA()
	r8 := uint8(cr >> 8)
	g8 := uint8(cg >> 8)
	b8 := uint8(cb >> 8)
	a8 := uint8(ca >> 8)
	var c color.Color = color.RGBA{r8, g8, b8, uint8(uint16(a8) * uint16(opacity) / 255)}
	if b.disabled {
		if b.textDisabledColor != nil {
			c = b.textDisabledColor
		} else {
			c = color.RGBA{r8, g8, b8, uint8(uint16(a8) * uint16(opacity) / (2 * 255))}
		}
	}

	dtop := &font.DrawTextOptions{
		Scale:     b.scale,
		TextAlign: data.TextAlignCenter,
		Language:  b.lang,
	}
	if b.dropShadow {
		dtop.Color = color.Black
		font.DrawText(screen, b.text, tx+int(b.scale), ty+int(b.scale), dtop)
	}
	dtop.Color = c
	font.DrawText(screen, b.text, tx, ty, dtop)
}
