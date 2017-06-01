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
	Draw(screen *ebiten.Image)
}

type Dialog struct {
	X         int
	Y         int
	Width     int
	Height    int
	Visible   bool
	widgets   []Widget
	offscreen *ebiten.Image
}

func NewDialog(x, y, width, height int) *Dialog {
	return &Dialog{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

func (d *Dialog) AddChild(widget Widget) {
	d.widgets = append(d.widgets, widget)
}

func (d *Dialog) Update() {
	for _, w := range d.widgets {
		w.UpdateAsChild(d.Visible, d.X, d.Y)
	}
}

func (d *Dialog) Draw(screen *ebiten.Image) {
	if !d.Visible {
		return
	}
	if d.Width == 0 || d.Height == 0 {
		return
	}
	if d.offscreen == nil {
		i, _ := ebiten.NewImage(d.Width*consts.TileScale, d.Height*consts.TileScale, ebiten.FilterNearest)
		d.offscreen = i
	} else {
		w, h := d.offscreen.Size()
		if d.Width != w || d.Height != h {
			d.offscreen.Dispose()
			i, _ := ebiten.NewImage(d.Width*consts.TileScale, d.Height*consts.TileScale, ebiten.FilterNearest)
			d.offscreen = i
		}
	}
	d.offscreen.Clear()
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &ninePatchParts{d.Width, d.Height}
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	d.offscreen.DrawImage(assets.GetImage("system/9patch_test_off.png"), op)
	for _, w := range d.widgets {
		w.Draw(d.offscreen)
	}
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(d.X)*consts.TileScale, float64(d.Y)*consts.TileScale)
	screen.DrawImage(d.offscreen, op)
}
