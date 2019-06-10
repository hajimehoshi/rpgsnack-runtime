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

const (
	popupMargin = 4
	PopupWidth  = 160 - popupMargin*2
)

type Node interface {
	UpdateAsChild(offsetX, offsetY int)
	DrawAsChild(screen *ebiten.Image, offsetX, offsetY int)
}

type Popup struct {
	x       int
	y       int
	width   int
	height  int
	visible bool
	nodes   []Node
}

func NewPopup(y, height int) *Popup {
	return &Popup{
		x:      popupMargin,
		y:      y,
		width:  PopupWidth,
		height: height,
	}
}

func (p *Popup) Visible() bool {
	return p.visible
}

func (p *Popup) Show() {
	p.visible = true
}

func (p *Popup) Hide() {
	p.visible = false
}

func (p *Popup) AddChild(node Node) {
	p.nodes = append(p.nodes, node)
}

func (p *Popup) Update() {
	if !p.visible {
		return
	}
	for _, n := range p.nodes {
		n.UpdateAsChild(p.x, p.y)
	}
}

func (p *Popup) Draw(screen *ebiten.Image) {
	if !p.visible {
		return
	}
	if p.width == 0 || p.height == 0 {
		return
	}

	geoM := &ebiten.GeoM{}
	geoM.Translate(float64(p.x), float64(p.y))
	geoM.Scale(consts.TileScale, consts.TileScale)
	DrawNinePatches(screen, assets.GetImage("system/common/9patch_frame_off.png"), p.width, p.height, geoM, nil)

	for _, n := range p.nodes {
		n.DrawAsChild(screen, p.x, p.y)
	}
}
