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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type SettingsScene struct {
	settingsLabel          *ui.Label
	languageButton         *ui.Button
	creditButton           *ui.Button
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

	w, _ := sceneManager.Size()

	s.settingsLabel = ui.NewLabel(16, 8)
	s.languageButton = ui.NewButton((w/consts.TileScale-120)/2, buttonOffsetX+1*buttonDeltaY, 120, 20, "click")
	s.creditButton = ui.NewButton((w/consts.TileScale-120)/2, buttonOffsetX+2*buttonDeltaY, 120, 20, "click")
	s.reviewThisAppButton = ui.NewButton((w/consts.TileScale-120)/2, buttonOffsetX+3*buttonDeltaY, 120, 20, "click")
	s.restorePurchasesButton = ui.NewButton((w/consts.TileScale-120)/2, buttonOffsetX+4*buttonDeltaY, 120, 20, "click")
	s.moreGamesButton = ui.NewButton((w/consts.TileScale-120)/2, buttonOffsetX+5*buttonDeltaY, 120, 20, "click")
	s.closeButton = ui.NewButton((w/consts.TileScale-120)/2, buttonOffsetX+6*buttonDeltaY, 120, 20, "cancel")

	s.languageDialog = ui.NewDialog((w/consts.TileScale-160)/2+4, 4, 152, 232)

	for i, l := range sceneManager.Game().Texts.Languages() {
		i := i // i is captured by the below closure and it is needed to copy here.
		n := display.Self.Name(l)
		b := ui.NewButton((152-120)/2, 8+i*buttonDeltaY, 120, 20, "click")
		b.Text = n
		s.languageDialog.AddChild(b)
		s.languageButtons = append(s.languageButtons, b)
		b.SetOnPressed(func(_ *ui.Button) {
			s.languageDialog.Hide()
			lang := sceneManager.Game().Texts.Languages()[i]
			lang = sceneManager.SetLanguage(lang)

			base, _ := lang.Base()
			s.waitingRequestID = sceneManager.GenerateRequestID()
			sceneManager.Requester().RequestChangeLanguage(s.waitingRequestID, base.String())
		})
	}

	s.languageButton.SetOnPressed(func(_ *ui.Button) {
		s.languageDialog.Show()
	})
	s.creditButton.SetOnPressed(func(_ *ui.Button) {
		sceneManager.Requester().RequestOpenLink(s.waitingRequestID, "show_credit", "menu")
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

func (s *SettingsScene) Update(sceneManager *scene.Manager) error {
	if !s.initialized {
		s.initUI(sceneManager)
		s.initialized = true
	}

	if input.BackButtonPressed() {
		s.handleBackButton(sceneManager)
	}

	s.settingsLabel.Text = texts.Text(sceneManager.Language(), texts.TextIDSettings)
	s.languageButton.Text = texts.Text(sceneManager.Language(), texts.TextIDLanguage)
	s.creditButton.Text = texts.Text(sceneManager.Language(), texts.TextIDCredit)
	s.reviewThisAppButton.Text = texts.Text(sceneManager.Language(), texts.TextIDReviewThisApp)
	s.restorePurchasesButton.Text = texts.Text(sceneManager.Language(), texts.TextIDRestorePurchases)
	s.moreGamesButton.Text = texts.Text(sceneManager.Language(), texts.TextIDMoreGames)
	s.closeButton.Text = texts.Text(sceneManager.Language(), texts.TextIDClose)

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
		s.reviewThisAppButton.Update()
		s.restorePurchasesButton.Update()
		s.moreGamesButton.Update()
		s.closeButton.Update()
	}

	return nil
}

func (s *SettingsScene) handleBackButton(sceneManager *scene.Manager) {
	if s.languageDialog.Visible() {
		audio.PlaySE("cancel", 1.0)
		s.languageDialog.Hide()
		return
	}

	audio.PlaySE("cancel", 1.0)
	sceneManager.GoTo(NewTitleScene())
}

func (s *SettingsScene) Draw(screen *ebiten.Image) {
	if !s.initialized {
		return
	}
	s.settingsLabel.Draw(screen)
	s.languageButton.Draw(screen)
	s.creditButton.Draw(screen)
	s.reviewThisAppButton.Draw(screen)
	s.restorePurchasesButton.Draw(screen)
	s.moreGamesButton.Draw(screen)
	s.closeButton.Draw(screen)
	s.languageDialog.Draw(screen)
}
