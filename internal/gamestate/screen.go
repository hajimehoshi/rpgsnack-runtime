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
	"encoding/json"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

var emptyImage *ebiten.Image

func init() {
	img, err := ebiten.NewImage(16, 16, ebiten.FilterNearest)
	if err != nil {
		panic(err)
	}
	emptyImage = img
}

type tint struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
	Gray  float64 `json:"gray"`
}

func (t *tint) isZero() bool {
	return t.Red == 0 && t.Green == 0 && t.Blue == 0 && t.Gray == 0
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

type tmpScreen struct {
	CurrentTint     tint `json:"currentTint"`
	OrigTint        tint `json:"origTint"`
	TargetTint      tint `json:"targetTint"`
	TintCount       int  `json:"tintCount"`
	TintMaxCount    int  `json:"tintMaxCount"`
	FadeInCount     int  `json:"fadeInCount"`
	FadeInMaxCount  int  `json:"fadeInMaxCount"`
	FadeOutCount    int  `json:"fadeOutCount"`
	FadeOutMaxCount int  `json:"fadeOutMaxCount"`
	FadedOut        bool `json:"fadedOut"`
}

func (s *Screen) MarshalJSON() ([]uint8, error) {
	tmp := &tmpScreen{
		CurrentTint:     s.currentTint,
		OrigTint:        s.origTint,
		TargetTint:      s.targetTint,
		TintCount:       s.tintCount,
		TintMaxCount:    s.tintMaxCount,
		FadeInCount:     s.fadeInCount,
		FadeInMaxCount:  s.fadeInMaxCount,
		FadeOutCount:    s.fadeOutCount,
		FadeOutMaxCount: s.fadeOutMaxCount,
		FadedOut:        s.fadedOut,
	}
	return json.Marshal(tmp)
}

func (s *Screen) UnmarshalJSON(data []uint8) error {
	var tmp *tmpScreen
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	s.currentTint = tmp.CurrentTint
	s.origTint = tmp.OrigTint
	s.targetTint = tmp.TargetTint
	s.tintCount = tmp.TintCount
	s.tintMaxCount = tmp.TintMaxCount
	s.fadeInCount = tmp.FadeInCount
	s.fadeInMaxCount = tmp.FadeInMaxCount
	s.fadeOutCount = tmp.FadeOutCount
	s.fadeOutMaxCount = tmp.FadeOutMaxCount
	s.fadedOut = tmp.FadedOut
	return nil
}

func (s *Screen) startTint(red, green, blue, gray float64, count int) {
	s.origTint.Red = s.currentTint.Red
	s.origTint.Green = s.currentTint.Green
	s.origTint.Blue = s.currentTint.Blue
	s.origTint.Gray = s.currentTint.Gray
	s.targetTint.Red = red
	s.targetTint.Green = green
	s.targetTint.Blue = blue
	s.targetTint.Gray = gray
	s.tintCount = count
	s.tintMaxCount = count
}

func (s *Screen) fadeIn(count int) {
	s.fadeInCount = count
	s.fadeInMaxCount = count
	s.fadedOut = false
}

func (s *Screen) fadeOut(count int) {
	s.fadeOutCount = count
	s.fadeOutMaxCount = count
}

func (s *Screen) isChangingTint() bool {
	return s.tintCount > 0
}

func (s *Screen) isFading() bool {
	return s.fadeInCount > 0 || s.fadeOutCount > 0
}

func (s *Screen) isFadedOut() bool {
	return s.fadedOut
}

func (s *Screen) Draw(screen *ebiten.Image, img *ebiten.Image, op *ebiten.DrawImageOptions) error {
	fadeRate := 0.0
	if s.fadedOut {
		fadeRate = 1
	} else {
		if !s.currentTint.isZero() {
			if s.currentTint.Gray != 0 {
				op.ColorM.ChangeHSV(0, 1-s.currentTint.Gray, 1)
			}
			rs, gs, bs := 1.0, 1.0, 1.0
			if s.currentTint.Red < 0 {
				rs = 1 - -s.currentTint.Red
			}
			if s.currentTint.Green < 0 {
				gs = 1 - -s.currentTint.Green
			}
			if s.currentTint.Blue < 0 {
				bs = 1 - -s.currentTint.Blue
			}
			op.ColorM.Scale(rs, gs, bs, 1)
			rt, gt, bt := 0.0, 0.0, 0.0
			if s.currentTint.Red > 0 {
				rt = s.currentTint.Red
			}
			if s.currentTint.Green > 0 {
				gt = s.currentTint.Green
			}
			if s.currentTint.Blue > 0 {
				bt = s.currentTint.Blue
			}
			op.ColorM.Translate(rt, gt, bt, 0)
		}
		if s.fadeInCount > 0 {
			fadeRate = float64(s.fadeInCount) / float64(s.fadeInMaxCount)
		}
		if s.fadeOutCount > 0 {
			fadeRate = 1 - float64(s.fadeOutCount)/float64(s.fadeOutMaxCount)
		}
	}
	if err := screen.DrawImage(img, op); err != nil {
		return err
	}
	if fadeRate > 0 {
		op := &ebiten.DrawImageOptions{}
		w, h := emptyImage.Size()
		targetW, targetH := img.Size()
		sx := float64(targetW) / float64(w)
		sy := float64(targetH) / float64(h)
		op.GeoM.Scale(sx, sy)
		op.GeoM.Scale(scene.TileScale, scene.TileScale)
		sw, _ := screen.Size()
		tx := (float64(sw) - scene.TileXNum*scene.TileSize*scene.TileScale) / 2
		op.GeoM.Translate(tx, scene.GameMarginTop)
		op.ColorM.Translate(0, 0, 0, 1)
		op.ColorM.Scale(1, 1, 1, fadeRate)
		if err := screen.DrawImage(emptyImage, op); err != nil {
			return err
		}
	}
	return nil
}

func (s *Screen) Update() error {
	if s.tintCount > 0 {
		s.tintCount--
		rate := 1 - float64(s.tintCount)/float64(s.tintMaxCount)
		s.currentTint.Red = s.origTint.Red*(1-rate) + s.targetTint.Red*rate
		s.currentTint.Green = s.origTint.Green*(1-rate) + s.targetTint.Green*rate
		s.currentTint.Blue = s.origTint.Blue*(1-rate) + s.targetTint.Blue*rate
		s.currentTint.Gray = s.origTint.Gray*(1-rate) + s.targetTint.Gray*rate
	} else {
		s.currentTint.Red = s.targetTint.Red
		s.currentTint.Green = s.targetTint.Green
		s.currentTint.Blue = s.targetTint.Blue
		s.currentTint.Gray = s.targetTint.Gray
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
