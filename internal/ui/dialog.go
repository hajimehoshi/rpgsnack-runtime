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
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
)

type Widget interface {
	UpdateAsChild(visible bool, offsetX, offsetY int)
	DrawAsChild(screen *ebiten.Image, offsetX, offsetY int)
}

type Dialog struct {
	X       int
	Y       int
	Width   int
	Height  int
	visible bool
	widgets []Widget
}

func NewDialog(x, y, width, height int) *Dialog {
	return &Dialog{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

func (d *Dialog) Visible() bool {
	return d.visible
}

func (d *Dialog) Show() {
	d.visible = true
}

func (d *Dialog) Hide() {
	d.visible = false
}

func (d *Dialog) AddChild(widget Widget) {
	d.widgets = append(d.widgets, widget)
}

func (d *Dialog) Update() {
	for _, w := range d.widgets {
		w.UpdateAsChild(d.visible, d.X, d.Y)
	}
}

func (d *Dialog) Draw(screen *ebiten.Image) {
	if !d.visible {
		return
	}
	if d.Width == 0 || d.Height == 0 {
		return
	}

	geoM := &ebiten.GeoM{}
	geoM.Translate(float64(d.X), float64(d.Y))
	geoM.Scale(consts.TileScale, consts.TileScale)
	drawNinePatches(screen, assets.GetImage("system/9patch_test_off.png"), d.Width, d.Height, geoM, nil)

	for _, w := range d.widgets {
		w.DrawAsChild(screen, d.X, d.Y)
	}
}
