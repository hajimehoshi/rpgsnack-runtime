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

type Node interface {
	UpdateAsChild(visible bool, offsetX, offsetY int)
	DrawAsChild(screen *ebiten.Image, offsetX, offsetY int)
}

type Dialog struct {
	x       int
	y       int
	width   int
	height  int
	visible bool
	nodes   []Node
}

func NewDialog(x, y, width, height int) *Dialog {
	return &Dialog{
		x:      x,
		y:      y,
		width:  width,
		height: height,
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

func (d *Dialog) AddChild(node Node) {
	d.nodes = append(d.nodes, node)
}

func (d *Dialog) Update() {
	if !d.visible {
		return
	}
	for _, n := range d.nodes {
		n.UpdateAsChild(d.visible, d.x, d.y)
	}
}

func (d *Dialog) Draw(screen *ebiten.Image) {
	if !d.visible {
		return
	}
	if d.width == 0 || d.height == 0 {
		return
	}

	geoM := &ebiten.GeoM{}
	geoM.Translate(float64(d.x), float64(d.y))
	geoM.Scale(consts.TileScale, consts.TileScale)
	DrawNinePatches(screen, assets.GetImage("system/common/9patch_frame_off.png"), d.width, d.height, geoM, nil)

	for _, n := range d.nodes {
		n.DrawAsChild(screen, d.x, d.y)
	}
}
