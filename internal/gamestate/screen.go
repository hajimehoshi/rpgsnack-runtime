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
	currentTint     tint
	origTint        tint
	targetTint      tint
	tintCount       int
	tintMaxCount    int
	fadeInCount     int
	fadeInMaxCount  int
	fadeOutCount    int
	fadeOutMaxCount int
	fadedOut        bool
}

func (s *Screen) StartTint(red, green, blue, gray float64, count int) {
	s.origTint.red = s.currentTint.red
	s.origTint.green = s.currentTint.green
	s.origTint.blue = s.currentTint.blue
	s.origTint.gray = s.currentTint.gray
	s.targetTint.red = red
	s.targetTint.green = green
	s.targetTint.blue = blue
	s.targetTint.gray = gray
	s.tintCount = count
	s.tintMaxCount = count
}

func (s *Screen) FadeIn(count int) {
	s.fadeInCount = count
	s.fadeInMaxCount = count
	s.fadedOut = false
}

func (s *Screen) FadeOut(count int) {
	s.fadeOutCount = count
	s.fadeOutMaxCount = count
}

func (s *Screen) IsChangingTint() bool {
	return s.tintCount > 0
}

func (s *Screen) IsFading() bool {
	return s.fadeInCount > 0 || s.fadeOutCount > 0
}

func (s *Screen) Apply(colorM *ebiten.ColorM) {
	if s.fadedOut {
		colorM.Scale(0, 0, 0, 1)
		return
	}
	if !s.currentTint.isZero() {
		if s.currentTint.gray != 0 {
			colorM.ChangeHSV(0, 1-s.currentTint.gray, 1)
		}
		rs, gs, bs := 1.0, 1.0, 1.0
		if s.currentTint.red < 0 {
			rs = 1 - -s.currentTint.red
		}
		if s.currentTint.green < 0 {
			gs = 1 - -s.currentTint.green
		}
		if s.currentTint.blue < 0 {
			bs = 1 - -s.currentTint.blue
		}
		colorM.Scale(rs, gs, bs, 1)
		rt, gt, bt := 0.0, 0.0, 0.0
		if s.currentTint.red > 0 {
			rt = s.currentTint.red
		}
		if s.currentTint.green > 0 {
			gt = s.currentTint.green
		}
		if s.currentTint.blue > 0 {
			bt = s.currentTint.blue
		}
		colorM.Translate(rt, gt, bt, 0)
	}
	if s.fadeInCount > 0 {
		rate := 1 - float64(s.fadeInCount)/float64(s.fadeInMaxCount)
		colorM.Scale(rate, rate, rate, 1)
	}
	if s.fadeOutCount > 0 {
		rate := float64(s.fadeOutCount) / float64(s.fadeOutMaxCount)
		colorM.Scale(rate, rate, rate, 1)
	}
}

func (s *Screen) Update() error {
	if s.tintCount > 0 {
		s.tintCount--
		rate := 1 - float64(s.tintCount)/float64(s.tintMaxCount)
		s.currentTint.red = s.origTint.red*(1-rate) + s.targetTint.red*rate
		s.currentTint.green = s.origTint.green*(1-rate) + s.targetTint.green*rate
		s.currentTint.blue = s.origTint.blue*(1-rate) + s.targetTint.blue*rate
		s.currentTint.gray = s.origTint.gray*(1-rate) + s.targetTint.gray*rate
	} else {
		s.currentTint.red = s.targetTint.red
		s.currentTint.green = s.targetTint.green
		s.currentTint.blue = s.targetTint.blue
		s.currentTint.gray = s.targetTint.gray
	}
	if s.fadeInCount > 0 {
		s.fadeInCount--
	}
	if s.fadeOutCount > 0 {
		s.fadeOutCount--
		if s.fadeOutCount == 0 {
			s.fadedOut = true
		}
	}
	return nil
}
