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
	"time"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
)

const (
	saveIntervalFrames = 5 * 60 // 5 sec
	boostInterval      = 10     // 10 sec
)

type MinigamePopup struct {
	y                    int
	visible              bool
	fadeImage            *ebiten.Image
	rewardButton         *Button
	closeButton          *Button
	scoreLabel           *Label
	prevScore            int
	saveTimer            int
	lastBoostTime        int64
	adsLoaded            bool
	minigame             *collectingGame
	onSave               func()
	onProgress           func(int)
	onClose              func()
	onRequestRewardedAds func()
}

func NewMinigamePopup(y int) *MinigamePopup {
	closeButton := NewImageButton(
		128,
		5,
		assets.GetImage("system/common/cancel_off.png"),
		assets.GetImage("system/common/cancel_on.png"),
		"system/cancel",
	)

	rewardButton := NewButton(16, 112, 120, 20, "system/click")

	scoreLabel := NewLabel(16, 8)

	fadeImage, err := ebiten.NewImage(16, 16, ebiten.FilterNearest)
	if err != nil {
		panic(err)
	}

	m := &MinigamePopup{
		y:            y,
		fadeImage:    fadeImage,
		closeButton:  closeButton,
		rewardButton: rewardButton,
		scoreLabel:   scoreLabel,
		saveTimer:    saveIntervalFrames,
		visible:      false,
		minigame:     newCollectingGame(),
	}
	rewardButton.SetOnPressed(func(_ *Button) {
		m.showRewardedAds()
	})

	return m
}

func (m *MinigamePopup) checkProgress(score, reqScore int) {
	t1 := reqScore / 100     // 1%
	t2 := reqScore / 20      // 5%
	t3 := reqScore / 4       // 25%
	t4 := reqScore / 2       // 50%
	t5 := (reqScore * 3) / 4 // 75%
	if m.prevScore < t1 && score >= t1 {
		m.onProgress(1)
	}
	if m.prevScore < t2 && score >= t2 {
		m.onProgress(5)
	}
	if m.prevScore < t3 && score >= t3 {
		m.onProgress(25)
	}
	if m.prevScore < t4 && score >= t4 {
		m.onProgress(50)
	}
	if m.prevScore < t5 && score >= t5 {
		m.onProgress(75)
	}
}

func (m *MinigamePopup) Update(minigameState Minigame) {
	if !m.visible || minigameState == nil {
		return
	}

	m.scoreLabel.Text = fmt.Sprintf(texts.Text(lang.Get(), texts.TextIDMinigameProgress), minigameState.Score(), minigameState.ReqScore())

	// TODO: Separate this into Update and HandleInput
	m.minigame.UpdateAsChild(minigameState, 0, m.y)

	m.rewardButton.SetText(texts.Text(lang.Get(), texts.TextIDMinigameWatchAds))
	score := minigameState.Score()

	m.saveTimer -= 1
	if m.saveTimer <= 0 {
		if m.prevScore > 0 && score > m.prevScore {
			m.saveTimer = saveIntervalFrames
			m.onSave()
			m.checkProgress(score, minigameState.ReqScore())
		}
		m.prevScore = score
	}

	if !m.minigame.CanGetReward() || !m.adsLoaded || time.Now().Unix()-m.lastBoostTime < boostInterval {
		m.rewardButton.Disable()
	} else {
		m.rewardButton.Enable()
	}

	if minigameState.Success() {
		m.onProgress(100)
		m.onClose()
	}
}

func (m *MinigamePopup) HandleInput(offsetX, offsetY int) bool {
	if !m.visible {
		return false
	}
	if m.closeButton.HandleInput(0+offsetX, m.y+offsetY) {
		return true
	}
	if m.rewardButton.HandleInput(0+offsetX, m.y+offsetY) {
		return true
	}
	// If a popup is visible, do not propagate any input handling to parents.
	return true
}

func (m *MinigamePopup) ActivateBoostMode() {
	m.minigame.ActivateBoostMode()
	m.lastBoostTime = time.Now().Unix()
}

func (m *MinigamePopup) showRewardedAds() {
	m.onRequestRewardedAds()
}

func (m *MinigamePopup) Visible() bool {
	return m.visible
}

func (m *MinigamePopup) Show() {
	m.visible = true
}

func (m *MinigamePopup) Hide() {
	m.visible = false
}

func (m *MinigamePopup) SetOnClose(f func()) {
	m.onClose = f
	m.closeButton.SetOnPressed(func(_ *Button) {
		f()
	})
}

func (m *MinigamePopup) SetOnProgress(f func(int)) {
	m.onProgress = f
}

func (m *MinigamePopup) SetOnSave(f func()) {
	m.onSave = f
}

func (m *MinigamePopup) SetOnRequestRewardedAds(f func()) {
	m.onRequestRewardedAds = f
}

func (m *MinigamePopup) SetAdsLoaded(loaded bool) {
	m.adsLoaded = loaded
}

func (m *MinigamePopup) Draw(screen *ebiten.Image) {
	if !m.visible {
		return
	}

	w, h := m.fadeImage.Size()
	sw, sh := screen.Size()
	sx := float64(sw) / float64(w)
	sy := float64(sh) / float64(h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sx, sy)
	op.ColorM.Scale(1, 1, 1, 0.5)
	screen.DrawImage(m.fadeImage, op)

	geoM := &ebiten.GeoM{}
	geoM.Translate(6, float64(m.y))
	geoM.Scale(consts.TileScale, consts.TileScale)
	DrawNinePatches(screen, assets.GetImage("system/common/9patch_frame_off.png"), 140, 140, geoM, nil)

	m.closeButton.DrawAsChild(screen, 0, m.y)
	m.rewardButton.DrawAsChild(screen, 0, m.y)
	m.scoreLabel.DrawAsChild(screen, 0, m.y)

	if m.minigame != nil {
		m.minigame.DrawAsChild(screen, 0, m.y)
	}
}
