// Copyright 2018 Hajime Hoshi
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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
)

const (
	revealFrames          = 16
	headerHeight          = 16 * consts.TileScale
	closeFrames           = 100
	HeaderTouchAreaHeight = 48
)

type GameHeader struct {
	x              int
	y              int
	titleButton    *Button
	cameraButton   *Button
	blackImage     *ebiten.Image
	isClosing      bool
	isOpening      bool
	revealRatio    float64
	autoCloseTimer int

	onTitleButtonPressed  func()
	onCameraButtonPressed func()
}

func NewGameHeader() *GameHeader {
	titleButton := NewTextButton(10, 2, 24, 12, "system/click")
	l := lang.Get()
	titleButton.text = texts.Text(l, texts.TextIDMenu)
	titleButton.disabled = true

	cameraButton := NewImageButton(142, 0, assets.GetImage("system/common/camera_off.png"), assets.GetImage("system/common/camera_on.png"), "system/camera")
	cameraButton.disabled = true

	blackImage, _ := ebiten.NewImage(16, 16, ebiten.FilterNearest)
	blackImage.Fill(color.Black)

	g := &GameHeader{
		x:              0,
		y:              0,
		titleButton:    titleButton,
		cameraButton:   cameraButton,
		blackImage:     blackImage,
		isOpening:      false,
		isClosing:      false,
		revealRatio:    0.0,
		autoCloseTimer: 0,
	}

	titleButton.SetOnPressed(func(_ *Button) {
		g.onTitleButtonPressed()
	})
	cameraButton.SetOnPressed(func(_ *Button) {
		g.onCameraButtonPressed()
	})

	return g
}

func (g *GameHeader) SetOnTitleButtonPressed(f func()) {
	g.onTitleButtonPressed = f
}

func (g *GameHeader) SetOnCameraButtonPressed(f func()) {
	g.onCameraButtonPressed = f
}

func (g *GameHeader) Open() {
	g.titleButton.disabled = true
	g.cameraButton.disabled = true
	g.isOpening = true
	g.isClosing = false
	g.autoCloseTimer = 0
}

func (g *GameHeader) Close() {
	g.titleButton.disabled = true
	g.cameraButton.disabled = true
	g.isOpening = false
	g.isClosing = true
	g.autoCloseTimer = 0
}

func (g *GameHeader) Update(paused bool) {
	if paused {
		return
	}

	g.titleButton.UpdateAsChild(true, g.x, g.y)
	g.cameraButton.UpdateAsChild(true, g.x, g.y)

	if g.isOpening {
		g.revealRatio += 1 / float64(revealFrames)
		if g.revealRatio > 1.0 {
			g.revealRatio = 1.0
			g.isOpening = false
			g.titleButton.disabled = false
			g.cameraButton.disabled = false
		}
	}
	if g.isClosing {
		g.revealRatio -= 1 / float64(revealFrames)
		if g.revealRatio < 0.0 {
			g.revealRatio = 0.0
			g.isClosing = false
		}
	}

	if g.revealRatio == 0 && input.Pressed() {
		_, iy := input.Position()
		if iy < HeaderTouchAreaHeight {
			input.Cancel()
			g.Open()
		}
	}

	if g.revealRatio == 1 && g.revealRatio >= 1.0 {
		g.autoCloseTimer++
		if g.autoCloseTimer > closeFrames {
			g.Close()
		}
	}
}

func (g *GameHeader) Draw(screen *ebiten.Image) {
	sw, _ := screen.Size()
	w, h := g.blackImage.Size()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(sw)/float64(w), float64(headerHeight)/float64(h))
	dy := int((1.0 - g.revealRatio) * float64(headerHeight))
	op.GeoM.Translate(float64(0), float64(-dy))
	screen.DrawImage(g.blackImage, op)

	g.titleButton.DrawAsChild(screen, g.x, g.y-dy)
	g.cameraButton.DrawAsChild(screen, g.x, g.y-dy)
}
