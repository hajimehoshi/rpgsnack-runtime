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
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
)

const (
	sliderHeight     = 8
	sliderHandleSize = 12
)

type Slider struct {
	x        int
	y        int
	width    int
	min      int
	max      int
	value    int
	dragging bool

	onValueChanged func(slider *Slider, value int)
	onReleased     func(slider *Slider, value int)
}

func NewSlider(x, y, width, min, max, value int) *Slider {
	return &Slider{
		x:     x,
		y:     y + 6,
		width: width,
		min:   min,
		max:   max,
		value: value,
	}
}

func (s *Slider) Value() int {
	return s.value
}

func (s *Slider) SetOnValueChanged(onValueChanged func(slider *Slider, value int)) {
	s.onValueChanged = onValueChanged
}

func (s *Slider) SetOnReleased(onReleased func(slider *Slider, value int)) {
	s.onReleased = onReleased
}

func (s *Slider) region() image.Rectangle {
	const (
		w = sliderHandleSize
		h = sliderHandleSize
	)
	x := s.x
	y := s.y
	hx := s.width*s.value/(s.max-s.min) - sliderHandleSize/2
	x += hx
	y -= sliderHandleSize / 4
	return image.Rect(x, y, x+w, y+h)
}

func (s *Slider) valueFromInput(offsetX, offsetY int) int {
	x, y := input.Position()
	x = int(float64(x) / consts.TileScale)
	y = int(float64(y) / consts.TileScale)
	x -= offsetX
	y -= offsetY

	t := x - s.x
	if t <= 0 {
		return s.min
	}
	if t >= s.width {
		return s.max
	}
	return t * (s.max - s.min) / s.width
}

func (s *Slider) update(visible bool, offsetX, offsetY int) {
	if !visible {
		return
	}
	if input.Pressed() && s.dragging {
		if v := s.valueFromInput(offsetX, offsetY); v < 0 {
			s.dragging = false
		} else {
			s.value = v
			if s.onValueChanged != nil {
				s.onValueChanged(s, s.value)
			}
		}
		return
	}
	if input.Triggered() {
		s.dragging = includesInput(offsetX, offsetY, s.region())
		return
	}
	if input.Released() && s.dragging {
		if s.onReleased != nil {
			s.onReleased(s, s.value)
		}
		return
	}

	s.dragging = false
}

func (s *Slider) Update() {
	s.update(true, 0, 0)
}

func (s *Slider) UpdateAsChild(visible bool, offsetX, offsetY int) {
	s.update(visible, offsetX, offsetY)
}

func (s *Slider) Draw(screen *ebiten.Image) {
	s.DrawAsChild(screen, 0, 0)
}

func (s *Slider) DrawAsChild(screen *ebiten.Image, offsetX, offsetY int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.x+offsetX), float64(s.y+offsetY))
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	img := assets.GetImage("system/common/9patch_frame_off.png")
	DrawNinePatches(screen, img, s.width, sliderHeight, &op.GeoM, &op.ColorM)

	hx := s.width*s.value/(s.max-s.min) - sliderHandleSize/2
	hy := -(sliderHandleSize - sliderHeight) / 2
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.x+offsetX+hx), float64(s.y+offsetY+hy))
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	img = assets.GetImage("system/common/9patch_frame_off.png")
	if s.dragging {
		img = assets.GetImage("system/common/9patch_frame_on.png")
	}
	DrawNinePatches(screen, img, sliderHandleSize, sliderHandleSize, &op.GeoM, &op.ColorM)

	tx := (s.x + s.width + offsetX + offsetY + 8) * consts.TileScale
	ty := (s.y + offsetY - 2) * consts.TileScale
	text := fmt.Sprintf("%d%%", s.value)

	dtop := &font.DrawTextOptions{
		Color: color.White,
	}
	font.DrawText(screen, text, tx, ty, dtop)
}
