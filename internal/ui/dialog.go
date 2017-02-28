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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type Widget interface {
	Update(offsetX, offsetY int) error
	Draw(screen *ebiten.Image) error
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

func (d *Dialog) Update() error {
	if !d.Visible {
		return nil
	}
	for _, w := range d.widgets {
		if err := w.Update(d.X, d.Y); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dialog) Draw(screen *ebiten.Image) error {
	if !d.Visible {
		return nil
	}
	if d.Width == 0 || d.Height == 0 {
		return nil
	}
	if d.offscreen == nil {
		i, err := ebiten.NewImage(d.Width*scene.TileScale, d.Height*scene.TileScale, ebiten.FilterNearest)
		if err != nil {
			return err
		}
		d.offscreen = i
	} else {
		w, h := d.offscreen.Size()
		if d.Width != w || d.Height != h {
			if err := d.offscreen.Dispose(); err != nil {
				return err
			}
			i, err := ebiten.NewImage(d.Width*scene.TileScale, d.Height*scene.TileScale, ebiten.FilterNearest)
			if err != nil {
				return err
			}
			d.offscreen = i
		}
	}
	if err := d.offscreen.Clear(); err != nil {
		return err
	}
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &ninePatchParts{d.Width, d.Height}
	op.GeoM.Scale(scene.TileScale, scene.TileScale)
	if err := d.offscreen.DrawImage(assets.GetImage("9patch_test_off.png"), op); err != nil {
		return err
	}
	for _, w := range d.widgets {
		if err := w.Draw(d.offscreen); err != nil {
			return err
		}
	}
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(d.X)*scene.TileScale, float64(d.Y)*scene.TileScale)
	if err := screen.DrawImage(d.offscreen, op); err != nil {
		return err
	}
	return nil
}
