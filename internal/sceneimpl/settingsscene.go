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

package sceneimpl

import (
	"golang.org/x/text/language/display"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type SettingsScene struct {
	settingsLabel          *ui.Label
	languageButton         *ui.Button
	creditButton           *ui.Button
	updateCreditsButton    *ui.Button
	reviewThisAppButton    *ui.Button
	restorePurchasesButton *ui.Button
	moreGamesButton        *ui.Button
	closeButton            *ui.Button
	languageDialog         *ui.Dialog
	languageButtons        []*ui.Button
	waitingRequestID       int
	initialized            bool
}

func NewSettingsScene() *SettingsScene {
	return &SettingsScene{}
}

func (s *SettingsScene) initUI(sceneManager *scene.Manager) {
	const (
		buttonOffsetX = 4
		buttonDeltaY  = 24
	)

	w, h := sceneManager.Size()

	tx := (w/consts.TileScale - 120) / 2
	ty := (h - 640) / (2 * consts.TileScale)
	s.settingsLabel = ui.NewLabel(16, ty+8)
	s.languageButton = ui.NewButton(tx, ty+buttonOffsetX+1*buttonDeltaY, 120, 20, "system/click")
	s.creditButton = ui.NewButton(tx, ty+buttonOffsetX+2*buttonDeltaY, 120, 20, "system/click")
	s.updateCreditsButton = ui.NewButton(tx+80, ty+buttonOffsetX+2*buttonDeltaY, 40, 20, "system/click")

	s.reviewThisAppButton = ui.NewButton(tx, ty+buttonOffsetX+3*buttonDeltaY, 120, 20, "system/click")
	s.restorePurchasesButton = ui.NewButton(tx, ty+buttonOffsetX+4*buttonDeltaY, 120, 20, "system/click")
	s.moreGamesButton = ui.NewButton(tx, ty+buttonOffsetX+5*buttonDeltaY, 120, 20, "system/click")
	s.closeButton = ui.NewButton(tx, ty+buttonOffsetX+6*buttonDeltaY, 120, 20, "system/cancel")

	s.languageDialog = ui.NewDialog((w/consts.TileScale-160)/2+4, h/(2*consts.TileScale)-80, 152, 160)

	for i, l := range sceneManager.Game().Texts.Languages() {
		i := i // i is captured by the below closure and it is needed to copy here.
		n := display.Self.Name(l)
		b := ui.NewButton((152-120)/2, 8+i*buttonDeltaY, 120, 20, "system/click")
		b.Text = n
		b.Lang = l
		s.languageDialog.AddChild(b)
		s.languageButtons = append(s.languageButtons, b)
		b.SetOnPressed(func(_ *ui.Button) {
			s.languageDialog.Hide()
			lang := sceneManager.Game().Texts.Languages()[i]
			lang = sceneManager.SetLanguage(lang)
			s.waitingRequestID = sceneManager.GenerateRequestID()
			sceneManager.Requester().RequestChangeLanguage(s.waitingRequestID, lang.String())
			s.updateButtonTexts()
		})
	}

	s.languageButton.SetOnPressed(func(_ *ui.Button) {
		s.languageDialog.Show()
	})
	s.creditButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(s.waitingRequestID, "show_credit", "menu")
	})
	s.updateCreditsButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(s.waitingRequestID, "post_credit", "")
	})
	s.reviewThisAppButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(s.waitingRequestID, "review", "")
	})
	s.restorePurchasesButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestRestorePurchases(s.waitingRequestID)
	})
	s.moreGamesButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(s.waitingRequestID, "more", "")
	})
	s.closeButton.SetOnPressed(func(_ *ui.Button) {
		sceneManager.GoTo(NewTitleScene())
	})
}

func (s *SettingsScene) updateButtonTexts() {
	s.settingsLabel.Text = texts.Text(lang.Get(), texts.TextIDSettings)
	s.languageButton.Text = texts.Text(lang.Get(), texts.TextIDLanguage)
	s.creditButton.Text = texts.Text(lang.Get(), texts.TextIDCredit)
	s.updateCreditsButton.Text = texts.Text(lang.Get(), texts.TextIDCreditEntry)
	s.reviewThisAppButton.Text = texts.Text(lang.Get(), texts.TextIDReviewThisApp)
	s.restorePurchasesButton.Text = texts.Text(lang.Get(), texts.TextIDRestorePurchases)
	s.moreGamesButton.Text = texts.Text(lang.Get(), texts.TextIDMoreGames)
	s.closeButton.Text = texts.Text(lang.Get(), texts.TextIDClose)
}

func (s *SettingsScene) Update(sceneManager *scene.Manager) error {
	if !s.initialized {
		s.initUI(sceneManager)
		s.initialized = true
	}

	if input.BackButtonPressed() {
		s.handleBackButton(sceneManager)
	}

	s.updateButtonTexts()

	if sceneManager.MaxPurchaseTier() > 0 {
		s.updateCreditsButton.Visible = true
		s.creditButton.Width = 76
	} else {
		s.updateCreditsButton.Visible = false
		s.creditButton.Width = 120
	}

	if s.waitingRequestID != 0 {
		r := sceneManager.ReceiveResultIfExists(s.waitingRequestID)
		if r != nil {
			s.waitingRequestID = 0
		}
		return nil
	}

	s.languageDialog.Update()
	if !s.languageDialog.Visible() {
		s.languageButton.Update()
		s.creditButton.Update()
		s.updateCreditsButton.Update()
		s.reviewThisAppButton.Update()
		s.restorePurchasesButton.Update()
		s.moreGamesButton.Update()
		s.closeButton.Update()
	}

	return nil
}

func (s *SettingsScene) handleBackButton(sceneManager *scene.Manager) {
	if s.languageDialog.Visible() {
		audio.PlaySE("system/cancel", 1.0)
		s.languageDialog.Hide()
		return
	}

	audio.PlaySE("system/cancel", 1.0)
	sceneManager.GoTo(NewTitleScene())
}

func (s *SettingsScene) Draw(screen *ebiten.Image) {
	if !s.initialized {
		return
	}
	s.settingsLabel.Draw(screen)
	s.languageButton.Draw(screen)
	s.creditButton.Draw(screen)
	s.updateCreditsButton.Draw(screen)
	s.reviewThisAppButton.Draw(screen)
	s.restorePurchasesButton.Draw(screen)
	s.moreGamesButton.Draw(screen)
	s.closeButton.Draw(screen)
	s.languageDialog.Draw(screen)
}

func (s *SettingsScene) Resize() {
	s.initialized = false
}
