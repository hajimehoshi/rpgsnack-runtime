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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

const (
	partSize = 4
)

type Button struct {
	X        int
	Y        int
	Width    int
	Height   int
	text     string
	pressing bool
	pressed  bool
}

func NewButton(x, y, width, height int, text string) *Button {
	return &Button{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		text:   text,
	}
}

func (b *Button) Pressed() bool {
	return b.pressed
}

func (b *Button) includesInput() bool {
	x, y := input.Position()
	x /= scene.TileScale
	y /= scene.TileScale
	if b.X <= x && x < b.X+b.Width && b.Y <= y && y < b.Y+b.Height {
		return true
	}
	return false
}

func (b *Button) Update() error {
	b.pressed = false
	if !b.pressing {
		if !input.Triggered() {
			return nil
		}
	}
	if !input.Pressed() {
		b.pressing = false
		b.pressed = true
		return nil
	}
	if b.includesInput() {
		b.pressing = true
	} else {
		b.pressing = false
	}
	return nil
}

type buttonParts struct {
	button *Button
}

func (b *buttonParts) Len() int {
	return (b.button.Width / partSize) * (b.button.Height / partSize)
}

func (b *buttonParts) Src(index int) (int, int, int, int) {
	xn := b.button.Width / partSize
	yn := b.button.Height / partSize
	sx, sy := 0, 0
	switch index % xn {
	case 0:
		sx = 0
	case xn - 1:
		sx = 2 * partSize
	default:
		sx = 1 * partSize
	}
	switch index / xn {
	case 0:
		sy = 0
	case yn - 1:
		sy = 2 * partSize
	default:
		sy = 1 * partSize
	}
	return sx, sy, sx + partSize, sy + partSize
}

func (b *buttonParts) Dst(index int) (int, int, int, int) {
	xn := b.button.Width / partSize
	dx := (index % xn) * partSize
	dy := (index / xn) * partSize
	return dx, dy, dx + partSize, dy + partSize
}

func (b *Button) Draw(screen *ebiten.Image) error {
	img := assets.GetImage("9patch_test_off.png")
	if b.pressing {
		img = assets.GetImage("9patch_test_on.png")
	}
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &buttonParts{b}
	op.GeoM.Translate(float64(b.X), float64(b.Y))
	op.GeoM.Scale(scene.TileScale, scene.TileScale)
	if err := screen.DrawImage(img, op); err != nil {
		return err
	}
	tw, th := font.MeasureSize(b.text)
	tx := b.X*scene.TileScale + (b.Width*scene.TileScale-tw*scene.TextScale)/2
	ty := b.Y*scene.TileScale + (b.Height*scene.TileScale-th*scene.TextScale)/2
	if err := font.DrawText(screen, b.text, tx, ty, scene.TextScale, color.White); err != nil {
		return err
	}
	return nil
}
