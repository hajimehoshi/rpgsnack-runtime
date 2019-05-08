// Copyright 2019 The RPGSnack Authors
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
	"image/color"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
)

const (
	switchButtonWidth  = 48
	switchButtonHeight = 16
	switchHandleWidth  = 24
	switchTouchExpand  = 8
)

type SwitchButton struct {
	x       int
	y       int
	enabled bool

	onToggled func(SwitchButton *SwitchButton, value bool)
}

func NewSwitchButton(x, y int) *SwitchButton {
	return &SwitchButton{
		x:       x,
		y:       y + 2,
		enabled: false,
	}
}

func (s *SwitchButton) SetX(x int) {
	s.x = x
}

func (s *SwitchButton) SetY(y int) {
	s.y = y
}

func (s *SwitchButton) SetEnabled(enabled bool) {
	s.enabled = enabled
}

func (s *SwitchButton) Enabled() bool {
	return s.enabled
}

func (s *SwitchButton) SetOnToggled(onToggled func(SwitchButton *SwitchButton, value bool)) {
	s.onToggled = onToggled
}

func (s *SwitchButton) includesInput(offsetX, offsetY int) bool {
	x, y := input.Position()
	x = int(float64(x) / consts.TileScale)
	y = int(float64(y) / consts.TileScale)
	x -= offsetX
	y -= offsetY

	buttonWidth := switchButtonWidth + switchTouchExpand*2
	buttonHeight := switchButtonHeight + switchTouchExpand*2
	buttonX := s.x - switchTouchExpand
	buttonY := s.y - switchTouchExpand

	if buttonX <= x && x < buttonX+buttonWidth && buttonY <= y && y < buttonY+buttonHeight {
		return true
	}
	return false
}

func (s *SwitchButton) update(visible bool, offsetX, offsetY int) {
	if !visible {
		return
	}

	if input.Triggered() {
		if s.includesInput(offsetX, offsetY) {
			s.enabled = !s.enabled
			audio.PlaySE("system/cancel", 1.0)
		}
		return
	}
}

func (s *SwitchButton) Update() {
	s.update(true, 0, 0)
}

func (s *SwitchButton) UpdateAsChild(visible bool, offsetX, offsetY int) {
	s.update(visible, offsetX, offsetY)
}

func (s *SwitchButton) Draw(screen *ebiten.Image) {
	s.DrawAsChild(screen, 0, 0)
}

func (s *SwitchButton) DrawAsChild(screen *ebiten.Image, offsetX, offsetY int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.x+offsetX), float64(s.y+offsetY))
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	img := assets.GetImage("system/common/9patch_frame_off.png")
	DrawNinePatches(screen, img, switchButtonWidth, switchButtonHeight, &op.GeoM, &op.ColorM)

	hx := 0
	if s.enabled {
		hx += switchButtonWidth - switchHandleWidth
	}
	hy := 0
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.x+offsetX+hx), float64(s.y+offsetY+hy))
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	img = assets.GetImage("system/common/9patch_frame_on.png")
	DrawNinePatches(screen, img, switchHandleWidth, switchButtonHeight, &op.GeoM, &op.ColorM)

	ty := (s.y + offsetY + 2) * consts.TileScale
	if s.enabled {
		tx := (s.x + offsetX + 6) * consts.TileScale
		text := "ON"
		font.DrawText(screen, text, tx, ty, consts.TextScale, data.TextAlignLeft, color.White, len([]rune(text)))
	} else {
		tx := (s.x + switchButtonWidth + offsetX - 18) * consts.TileScale
		text := "OFF"
		font.DrawText(screen, text, tx, ty, consts.TextScale, data.TextAlignLeft, color.White, len([]rune(text)))
	}
}
