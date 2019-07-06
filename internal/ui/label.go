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
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
)

type Label struct {
	x         int
	y         int
	Text      string
	Scale     float64
	Color     color.Color
	TextAlign data.TextAlign
	visible   bool
}

func NewLabel(x, y int) *Label {
	return &Label{
		x:         x,
		y:         y,
		Color:     color.White,
		Scale:     1.0,
		TextAlign: data.TextAlignLeft,
		visible:   true,
	}
}

func (l *Label) Show() {
	l.visible = true
}

func (l *Label) Hide() {
	l.visible = false
}

func (l *Label) Region() image.Rectangle {
	w, h := font.MeasureSize(l.Text)
	wf, hf := float64(w), float64(h)
	wf *= l.Scale
	hf *= l.Scale
	wf *= consts.TextScale
	hf *= consts.TextScale
	wf /= consts.TileScale
	hf /= consts.TileScale
	return image.Rect(l.x, l.y, l.x+int(wf), l.y+int(hf))
}

func (l *Label) Update() {
}

func (l *Label) HandleInput(offsetX, offsetY int) bool {
	return false
}

func (l *Label) Draw(screen *ebiten.Image) {
	l.DrawAsChild(screen, 0, 0)
}

func (l *Label) DrawAsChild(screen *ebiten.Image, offsetX, offsetY int) {
	if !l.visible {
		return
	}

	tx := (l.x + offsetX) * consts.TileScale
	ty := (l.y + offsetY) * consts.TileScale
	op := &font.DrawTextOptions{
		Scale:     consts.TextScale * l.Scale,
		TextAlign: l.TextAlign,
		Color:     l.Color,
	}
	font.DrawText(screen, l.Text, tx, ty, op)
}
