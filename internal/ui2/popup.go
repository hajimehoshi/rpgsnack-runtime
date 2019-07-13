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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
)

// TODO: Unify similar usages of empty images for tinting the screen.

var shadowImage *ebiten.Image

func init() {
	shadowImage, _ = ebiten.NewImage(1, 1, ebiten.FilterDefault)
	shadowImage.Fill(color.RGBA{0, 0, 0, 0x80})
}

const (
	popupMargin         = 12
	PopupWidth          = 480 - popupMargin*2
	popupFadingCountMax = 5
)

type Node interface {
	HandleInput(offsetX, offsetY int) bool
	Update()
	DrawAsChild(screen *ebiten.Image, offsetX, offsetY int)
	Region() image.Rectangle
}

type Popup struct {
	x       int
	y       int
	width   int
	height  int
	visible bool
	nodes   []Node

	fadingInCount  int
	fadingOutCount int

	offscreen *ebiten.Image
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
	if p.visible {
		return
	}
	p.fadingInCount = popupFadingCountMax
}

func (p *Popup) Hide() {
	if !p.visible {
		return
	}
	p.fadingOutCount = popupFadingCountMax
}

func (p *Popup) AddChild(node Node) {
	p.nodes = append(p.nodes, node)
}

func (p *Popup) Update() {
	if !p.visible && p.fadingInCount > 0 {
		p.fadingInCount--
		if p.fadingInCount == 0 {
			p.visible = true
		}
	}
	if p.visible && p.fadingOutCount > 0 {
		p.fadingOutCount--
		if p.fadingOutCount == 0 {
			p.visible = false
		}
	}
	if !p.visible || p.fadingInCount > 0 || p.fadingOutCount > 0 {
		return
	}
	for _, n := range p.nodes {
		n.Update()
	}
}

func (p *Popup) HandleInput(offsetX, offsetY int) bool {
	if !p.visible {
		return false
	}

	if input.BackButtonTriggered() {
		audio.PlaySE("system/cancel", 1)
		p.Hide()
		return true
	}

	for _, n := range p.nodes {
		if n.HandleInput(p.x+offsetX, p.y+offsetY) {
			return true
		}
	}
	// If a popup is visible, do not propagate any input handling to parents.
	return true
}

func (p *Popup) Draw(screen *ebiten.Image) {
	if !p.visible && p.fadingInCount == 0 && p.fadingOutCount == 0 {
		return
	}
	if p.width == 0 || p.height == 0 {
		return
	}

	alpha := 1.0
	if p.fadingInCount > 0 {
		alpha = 1.0 - float64(p.fadingInCount)/popupFadingCountMax
	}
	if p.fadingOutCount > 0 {
		alpha = float64(p.fadingOutCount) / popupFadingCountMax
	}

	op := &ebiten.DrawImageOptions{}
	w, h := shadowImage.Size()
	sw, sh := screen.Size()
	op.GeoM.Scale(float64(sw)/float64(w), float64(sh)/float64(h))
	op.ColorM.Scale(1, 1, 1, alpha)
	screen.DrawImage(shadowImage, op)

	offscreen := screen
	if alpha < 1.0 {
		if p.offscreen != nil {
			w, h := p.offscreen.Size()
			sw, sh := screen.Size()
			if w != sw || h != sh {
				p.offscreen.Dispose()
				p.offscreen = nil
			}
		}
		if p.offscreen == nil {
			sw, sh := screen.Size()
			p.offscreen, _ = ebiten.NewImage(sw, sh, ebiten.FilterNearest)
		}
		offscreen = p.offscreen
	}

	geoM := &ebiten.GeoM{}
	geoM.Translate(float64(p.x), float64(p.y))
	drawNinePatches(offscreen, assets.GetImage("system/common/9patch_frame_off.png"), p.width, p.height, geoM, nil)

	for _, n := range p.nodes {
		n.DrawAsChild(offscreen, p.x, p.y)
	}

	if offscreen != screen {
		op := &ebiten.DrawImageOptions{}
		op.ColorM.Scale(1, 1, 1, alpha)
		screen.DrawImage(offscreen, op)
	}
}
