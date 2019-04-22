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
	"github.com/hajimehoshi/ebiten"
	"golang.org/x/text/language/display"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type SettingsScene struct {
	settingsLabel          *ui.Label
	languageButton         *ui.Button
	creditsButton          *ui.Button
	updateCreditsButton    *ui.Button
	reviewThisAppButton    *ui.Button
	restorePurchasesButton *ui.Button
	resetGameButton        *ui.Button
	privacyPolicyButton    *ui.Button
	shopButton             *ui.Button
	closeButton            *ui.Button

	languageDialog   *ui.Dialog
	languageButtons  []*ui.Button
	warningDialog    *ui.Dialog
	warningLabel     *ui.Label
	warningYesButton *ui.Button
	warningNoButton  *ui.Button
	waitingRequestID int
	credits          *ui.Credits

	initialized bool
	baseX       int
	baseY       int

	err error
}

const (
	buttonOffsetY = 4
	buttonDeltaY  = 24
)

func NewSettingsScene() *SettingsScene {
	return &SettingsScene{}
}

func (s *SettingsScene) calcButtonY(index int) int {
	return s.baseY + buttonOffsetY + index*buttonDeltaY
}

func (s *SettingsScene) initUI(sceneManager *scene.Manager) {
	w, h := sceneManager.Size()
	s.baseX = (w/consts.TileScale - 120) / 2
	s.baseY = (h - 640) / (2 * consts.TileScale)

	s.settingsLabel = ui.NewLabel(16, s.baseY+8)
	s.languageButton = ui.NewButton(s.baseX, s.calcButtonY(1), 120, 20, "system/click")
	s.shopButton = ui.NewButton(s.baseX, s.calcButtonY(2), 120, 20, "system/click")
	s.restorePurchasesButton = ui.NewButton(s.baseX, s.calcButtonY(3), 120, 20, "system/click")
	s.creditsButton = ui.NewButton(s.baseX, s.calcButtonY(4), 120, 20, "system/click")
	s.updateCreditsButton = ui.NewButton(s.baseX+80, s.calcButtonY(4), 40, 20, "system/click")
	s.reviewThisAppButton = ui.NewButton(s.baseX, s.calcButtonY(5), 120, 20, "system/click")
	s.resetGameButton = ui.NewButton(s.baseX, s.calcButtonY(6), 120, 20, "system/click")
	s.privacyPolicyButton = ui.NewButton(s.baseX, s.calcButtonY(7), 120, 20, "system/click")
	s.closeButton = ui.NewButton(s.baseX, s.calcButtonY(8), 120, 20, "system/cancel")

	s.languageDialog = ui.NewDialog((w/consts.TileScale-160)/2+4, h/(2*consts.TileScale)-80, 152, 160)

	for i, l := range sceneManager.Game().Texts.Languages() {
		i := i // i is captured by the below closure and it is needed to copy here.
		n := display.Self.Name(l)
		b := ui.NewButton((152-120)/2, 8+i*buttonDeltaY, 120, 20, "system/click")
		b.SetText(n)
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

	s.creditsButton.SetOnPressed(func(_ *ui.Button) {
		s.credits.SetData(sceneManager.Credits())
		s.credits.Show()
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

	s.resetGameButton.SetOnPressed(func(_ *ui.Button) {
		s.warningDialog.Show()
	})
	s.warningDialog = ui.NewDialog((w/consts.TileScale-160)/2+4, (h)/(2*consts.TileScale)-64, 152, 124)
	s.warningLabel = ui.NewLabel(16, 8)
	s.warningYesButton = ui.NewButton((152-120)/2, 72, 120, 20, "system/click")
	s.warningNoButton = ui.NewButton((152-120)/2, 96, 120, 20, "system/cancel")
	s.warningDialog.AddChild(s.warningLabel)
	s.warningDialog.AddChild(s.warningYesButton)
	s.warningDialog.AddChild(s.warningNoButton)
	s.warningYesButton.SetOnPressed(func(_ *ui.Button) {
		id := sceneManager.GenerateRequestID()
		s.waitingRequestID = id
		sceneManager.Requester().RequestSaveProgress(id, nil)
		sceneManager.SetProgress(nil)
		s.warningDialog.Hide()
	})
	s.warningNoButton.SetOnPressed(func(_ *ui.Button) {
		s.warningDialog.Hide()
	})

	s.privacyPolicyButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(s.waitingRequestID, "privacy", "")
	})
	s.closeButton.SetOnPressed(func(_ *ui.Button) {
		g, err := savedGame(sceneManager)
		if err != nil {
			s.err = err
			return
		}
		sceneManager.GoTo(NewTitleMapScene(g))
	})
	s.shopButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestShowShop(s.waitingRequestID, string(sceneManager.ShopData(data.ShopTypeMain, []bool{true, true, true, true})))
	})

	s.credits = ui.NewCredits(true)
}

func (s *SettingsScene) updateButtonTexts() {
	s.settingsLabel.Text = texts.Text(lang.Get(), texts.TextIDSettings)
	s.languageButton.SetText(texts.Text(lang.Get(), texts.TextIDLanguage))
	s.creditsButton.SetText(texts.Text(lang.Get(), texts.TextIDCredits))
	s.updateCreditsButton.SetText(texts.Text(lang.Get(), texts.TextIDCreditsEntry))
	s.reviewThisAppButton.SetText(texts.Text(lang.Get(), texts.TextIDReviewThisApp))
	s.restorePurchasesButton.SetText(texts.Text(lang.Get(), texts.TextIDRestorePurchases))
	s.resetGameButton.SetText(texts.Text(lang.Get(), texts.TextIDResetGame))
	s.warningLabel.Text = texts.Text(lang.Get(), texts.TextIDNewGameWarning)
	s.warningYesButton.SetText(texts.Text(lang.Get(), texts.TextIDYes))
	s.warningNoButton.SetText(texts.Text(lang.Get(), texts.TextIDNo))
	s.privacyPolicyButton.SetText(texts.Text(lang.Get(), texts.TextIDPrivacyPolicy))
	s.shopButton.SetText(texts.Text(lang.Get(), texts.TextIDShop))
	s.closeButton.SetText(texts.Text(lang.Get(), texts.TextIDClose))
}

func (s *SettingsScene) Update(sceneManager *scene.Manager) error {
	if s.err != nil {
		return s.err
	}

	if !s.initialized {
		s.initUI(sceneManager)
		s.initialized = true
	}

	if input.BackButtonPressed() {
		s.handleBackButton(sceneManager)
	}

	s.updateButtonTexts()

	if sceneManager.SponsorTier() > 0 {
		s.updateCreditsButton.Show()
		s.creditsButton.SetWidth(76)
	} else {
		s.updateCreditsButton.Hide()
		s.creditsButton.SetWidth(120)
	}

	if s.waitingRequestID != 0 {
		r := sceneManager.ReceiveResultIfExists(s.waitingRequestID)
		if r != nil {
			s.waitingRequestID = 0
		}
		return nil
	}

	itemOffset := 0
	// TODO: For now if there is no shop, we hide RestorePurchase button
	// But to implement this right, we should check whether the game contains
	// any non-consumeable
	if sceneManager.Game().IsShopAvailable(data.ShopTypeMain) {
		s.shopButton.Show()
		s.restorePurchasesButton.Show()
		itemOffset = 2
	} else {
		s.shopButton.Hide()
		s.restorePurchasesButton.Hide()
	}

	s.creditsButton.SetY(s.calcButtonY(itemOffset + 2))
	s.updateCreditsButton.SetY(s.calcButtonY(itemOffset + 2))
	s.reviewThisAppButton.SetY(s.calcButtonY(itemOffset + 3))
	s.resetGameButton.SetY(s.calcButtonY(itemOffset + 4))
	s.privacyPolicyButton.SetY(s.calcButtonY(itemOffset + 5))
	s.closeButton.SetY(s.calcButtonY(itemOffset + 6))

	if sceneManager.HasProgress() {
		s.resetGameButton.Enable()
	} else {
		s.resetGameButton.Disable()
	}

	s.languageDialog.Update()
	s.warningDialog.Update()
	s.credits.Update()
	if !s.languageDialog.Visible() && !s.warningDialog.Visible() && !s.credits.Visible() {
		s.languageButton.Update()
		s.shopButton.Update()
		s.creditsButton.Update()
		s.updateCreditsButton.Update()
		s.reviewThisAppButton.Update()
		s.restorePurchasesButton.Update()
		s.resetGameButton.Update()
		s.privacyPolicyButton.Update()
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
	if s.warningDialog.Visible() {
		audio.PlaySE("system/cancel", 1.0)
		s.warningDialog.Hide()
		return
	}
	if s.credits.Visible() {
		audio.PlaySE("system/cancel", 1.0)
		s.credits.Hide()
		return
	}

	audio.PlaySE("system/cancel", 1.0)
	g, err := savedGame(sceneManager)
	if err != nil {
		s.err = err
		return
	}
	sceneManager.GoTo(NewTitleMapScene(g))
}

func (s *SettingsScene) Draw(screen *ebiten.Image) {
	if !s.initialized {
		return
	}
	s.settingsLabel.Draw(screen)
	s.languageButton.Draw(screen)
	s.shopButton.Draw(screen)
	s.creditsButton.Draw(screen)
	s.updateCreditsButton.Draw(screen)
	s.reviewThisAppButton.Draw(screen)
	s.restorePurchasesButton.Draw(screen)
	s.resetGameButton.Draw(screen)
	s.privacyPolicyButton.Draw(screen)
	s.closeButton.Draw(screen)
	s.languageDialog.Draw(screen)
	s.warningDialog.Draw(screen)
	s.credits.Draw(screen)
}

func (s *SettingsScene) Resize() {
	s.initialized = false
}
