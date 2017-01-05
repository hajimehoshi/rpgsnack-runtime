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

package gamestate

import (
	"github.com/hajimehoshi/ebiten"
)

type tint struct {
	red   float64
	green float64
	blue  float64
	gray  float64
}

func (t *tint) isZero() bool {
	return t.red == 0 && t.green == 0 && t.blue == 0 && t.gray == 0
}

type Screen struct {
	tint         *tint
	origTint     *tint
	targetTint   *tint
	tintCount    int
	tintMaxCount int
}

func newScreen() *Screen {
	return &Screen{
		tint:       &tint{},
		origTint:   &tint{},
		targetTint: &tint{},
	}
}

func (s *Screen) StartTint(red, green, blue, gray float64, count int) {
	s.origTint.red = s.tint.red
	s.origTint.green = s.tint.green
	s.origTint.blue = s.tint.blue
	s.origTint.gray = s.tint.gray
	s.targetTint.red = red
	s.targetTint.green = green
	s.targetTint.blue = blue
	s.targetTint.gray = gray
	s.tintCount = count
	s.tintMaxCount = count
}

func (s *Screen) IsChangingTint() bool {
	return s.tintCount > 0
}

func (s *Screen) ApplyTint(colorM *ebiten.ColorM) {
	if s.tint.isZero() {
		return
	}
	if s.tint.gray != 0 {
		colorM.ChangeHSV(0, 1-s.tint.gray, 1)
	}
	rs, gs, bs := 1.0, 1.0, 1.0
	if s.tint.red < 0 {
		rs = 1 - -s.tint.red
	}
	if s.tint.green < 0 {
		gs = 1 - -s.tint.green
	}
	if s.tint.blue < 0 {
		bs = 1 - -s.tint.blue
	}
	colorM.Scale(rs, gs, bs, 1)
	rt, gt, bt := 0.0, 0.0, 0.0
	if s.tint.red > 0 {
		rt = s.tint.red
	}
	if s.tint.green > 0 {
		gt = s.tint.green
	}
	if s.tint.blue > 0 {
		bt = s.tint.blue
	}
	colorM.Translate(rt, gt, bt, 0)
}

func (s *Screen) Update() error {
	if s.tintCount > 0 {
		s.tintCount--
		rate := 1 - float64(s.tintCount)/float64(s.tintMaxCount)
		s.tint.red = s.origTint.red*(1-rate) + s.targetTint.red*rate
		s.tint.green = s.origTint.green*(1-rate) + s.targetTint.green*rate
		s.tint.blue = s.origTint.blue*(1-rate) + s.targetTint.blue*rate
		s.tint.gray = s.origTint.gray*(1-rate) + s.targetTint.gray*rate
		return nil
	}
	s.tint.red = s.targetTint.red
	s.tint.green = s.targetTint.green
	s.tint.blue = s.targetTint.blue
	s.tint.gray = s.targetTint.gray
	return nil
}
