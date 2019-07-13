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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
)

type Label struct {
	x         int
	y         int
	text      string
	scale     float64
	color     color.Color
	textAlign data.TextAlign
	visible   bool
}

func NewLabel(x, y int) *Label {
	return &Label{
		x:         x,
		y:         y,
		color:     color.White,
		scale:     1.0,
		textAlign: data.TextAlignLeft,
		visible:   true,
	}
}

func (l *Label) Show() {
	l.visible = true
}

func (l *Label) Hide() {
	l.visible = false
}

func (l *Label) SetText(text string) {
	l.text = text
}

func (l *Label) SetScale(scale float64) {
	l.scale = scale
}

func (l *Label) SetTextAlign(align data.TextAlign) {
	l.textAlign = align
}

func (l *Label) SetColor(clr color.Color) {
	l.color = clr
}

func (l *Label) Region() image.Rectangle {
	w, h := font.MeasureSize(l.text)
	wf, hf := float64(w), float64(h)
	wf *= l.scale
	hf *= l.scale
	return image.Rect(l.x, l.y, l.x+int(wf), l.y+int(hf))
}

func (l *Label) HandleInput(offsetX, offsetY int) bool {
	return false
}

func (l *Label) Update() {
}

func (l *Label) DrawAsChild(screen *ebiten.Image, offsetX, offsetY int) {
	if !l.visible {
		return
	}

	tx := l.x + offsetX
	ty := l.y + offsetY
	op := &font.DrawTextOptions{
		Scale:     l.scale,
		TextAlign: l.textAlign,
		Color:     l.color,
	}
	font.DrawText(screen, l.text, tx, ty, op)
}
