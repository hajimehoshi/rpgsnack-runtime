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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

const (
	partSize = 4
)

type Button struct {
	x      int
	y      int
	width  int
	height int
	text   string
}

func NewButton(x, y, width, height int, text string) *Button {
	return &Button{
		x:      x,
		y:      y,
		width:  width,
		height: height,
		text:   text,
	}
}

type buttonParts struct {
	button *Button
}

func (b *buttonParts) Len() int {
	return (b.button.width / partSize) * (b.button.height / partSize)
}

func (b *buttonParts) Src(index int) (int, int, int, int) {
	xn := b.button.width / partSize
	yn := b.button.height / partSize
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
	xn := b.button.width / partSize
	dx := (index % xn) * partSize
	dy := (index / xn) * partSize
	return dx, dy, dx + partSize, dy + partSize
}

func (b *Button) Draw(screen *ebiten.Image) error {
	off := assets.GetImage("9patch_test_off.png")
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &buttonParts{b}
	op.GeoM.Translate(float64(b.x), float64(b.y))
	op.GeoM.Scale(scene.TileScale, scene.TileScale)
	if err := screen.DrawImage(off, op); err != nil {
		return err
	}
	tw, th := font.MeasureSize(b.text)
	tx := b.x*scene.TileScale + (b.width*scene.TileScale-tw*scene.TextScale)/2
	ty := b.y*scene.TileScale + (b.height*scene.TileScale-th*scene.TextScale)/2
	if err := font.DrawText(screen, b.text, tx, ty, scene.TextScale, color.White); err != nil {
		return err
	}
	return nil
}
