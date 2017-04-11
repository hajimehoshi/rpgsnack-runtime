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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type Button struct {
	X         int
	Y         int
	Width     int
	Height    int
	Visible   bool
	Text      string
	image     *ebiten.Image
	pressing  bool
	pressed   bool
	soundName string
}

func NewButton(x, y, width, height int, soundName string) *Button {
	return &Button{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Visible:   true,
		soundName: soundName,
	}
}

func NewImageButton(x, y int, image *ebiten.Image, soundName string) *Button {
	w, h := image.Size()
	return &Button{
		X:         x,
		Y:         y,
		Width:     w,
		Height:    h,
		Visible:   true,
		image:     image,
		soundName: soundName,
	}
}

func (b *Button) Pressed() bool {
	return b.pressed
}

func (b *Button) includesInput(offsetX, offsetY int) bool {
	x, y := input.Position()
	x /= scene.TileScale
	y /= scene.TileScale
	x -= offsetX
	y -= offsetY
	if b.X <= x && x < b.X+b.Width && b.Y <= y && y < b.Y+b.Height {
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
	if b.image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(b.X), float64(b.Y))
		op.GeoM.Scale(scene.TileScale, scene.TileScale)
		screen.DrawImage(b.image, op)
		return
	}
	img := assets.GetImage("9patch_test_off.png")
	if b.pressing {
		img = assets.GetImage("9patch_test_on.png")
	}
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &ninePatchParts{b.Width, b.Height}
	op.GeoM.Translate(float64(b.X), float64(b.Y))
	op.GeoM.Scale(scene.TileScale, scene.TileScale)
	screen.DrawImage(img, op)
	tw, th := font.MeasureSize(b.Text)
	tx := b.X*scene.TileScale + (b.Width*scene.TileScale-tw*scene.TextScale)/2
	ty := b.Y*scene.TileScale + (b.Height*scene.TileScale-th*scene.TextScale)/2
	font.DrawText(screen, b.Text, tx, ty, scene.TextScale, color.White)
}
