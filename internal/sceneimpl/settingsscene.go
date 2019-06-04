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
	advancedButton         *ui.Button
	creditsButton          *ui.Button
	updateCreditsButton    *ui.Button
	reviewThisAppButton    *ui.Button
	restorePurchasesButton *ui.Button
	privacyPolicyButton    *ui.Button
	shopButton             *ui.Button
	closeButton            *ui.Button
	waitingRequestID       int

	credits *ui.Credits

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

	// The buttons that Y is 0 are adjusted later.
	s.advancedButton = ui.NewButton(s.baseX, s.calcButtonY(1), 120, 20, "system/click")
	s.shopButton = ui.NewButton(s.baseX, s.calcButtonY(2), 120, 20, "system/click")
	s.restorePurchasesButton = ui.NewButton(s.baseX, s.calcButtonY(3), 120, 20, "system/click")
	s.creditsButton = ui.NewButton(s.baseX, 0, 120, 20, "system/click")
	s.updateCreditsButton = ui.NewButton(s.baseX+80, 0, 40, 20, "system/click")
	s.reviewThisAppButton = ui.NewButton(s.baseX, 0, 120, 20, "system/click")
	s.privacyPolicyButton = ui.NewButton(s.baseX, 0, 120, 20, "system/click")
	s.closeButton = ui.NewButton(s.baseX, s.calcButtonY(8), 120, 20, "system/cancel")

	s.advancedButton.SetOnPressed(func(_ *ui.Button) {
		sceneManager.GoTo(&AdvancedSettingsScene{})
		// TODO GameSettings Mode
	})

	s.creditsButton.SetOnPressed(func(_ *ui.Button) {
		s.credits.SetData(sceneManager.Credits())
		s.credits.Show()
	})
	s.updateCreditsButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		// TODO
		panic("post credit is not implemented")
	})
	s.reviewThisAppButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenReview(s.waitingRequestID)
	})
	s.restorePurchasesButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestRestorePurchases(s.waitingRequestID)
	})

	s.privacyPolicyButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		url := "" // TODO get privacy policy URL
		sceneManager.Requester().RequestOpenURL(s.waitingRequestID, url)
	})
	s.closeButton.SetOnPressed(func(_ *ui.Button) {
		g, err := savedGame(sceneManager)
		if err != nil {
			s.err = err
			return
		}
		sceneManager.GoTo(NewTitleMapScene(sceneManager, g))
	})
	s.shopButton.SetOnPressed(func(_ *ui.Button) {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestShowShop(s.waitingRequestID, string(sceneManager.ShopData(data.ShopTypeMain, []bool{true, true, true, true})))
	})

	s.credits = ui.NewCredits()
	s.credits.SetCloseButtonVisible(true)
}

func (s *SettingsScene) updateTexts() {
	s.advancedButton.SetText(texts.Text(lang.Get(), texts.TextIDAdvancedSettings))
	s.creditsButton.SetText(texts.Text(lang.Get(), texts.TextIDCredits))
	s.updateCreditsButton.SetText(texts.Text(lang.Get(), texts.TextIDCreditsEntry))
	s.reviewThisAppButton.SetText(texts.Text(lang.Get(), texts.TextIDReviewThisApp))
	s.restorePurchasesButton.SetText(texts.Text(lang.Get(), texts.TextIDRestorePurchases))
	s.privacyPolicyButton.SetText(texts.Text(lang.Get(), texts.TextIDPrivacyPolicy))
	s.shopButton.SetText(texts.Text(lang.Get(), texts.TextIDShop))
	s.closeButton.SetText(texts.Text(lang.Get(), texts.TextIDBack))
}

func (s *SettingsScene) Update(sceneManager *scene.Manager) error {
	if s.err != nil {
		return s.err
	}

	if !s.initialized {
		s.initUI(sceneManager)
		s.initialized = true
	}

	if input.BackButtonTriggered() {
		s.handleBackButton(sceneManager)
	}

	s.updateTexts()

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
	s.privacyPolicyButton.SetY(s.calcButtonY(itemOffset + 4))

	s.credits.Update()
	if !s.credits.Visible() {
		if s.shopButton.HandleInput(0, 0) {
			return nil
		}
		if s.creditsButton.HandleInput(0, 0) {
			return nil
		}
		if s.updateCreditsButton.HandleInput(0, 0) {
			return nil
		}
		if s.reviewThisAppButton.HandleInput(0, 0) {
			return nil
		}
		if s.restorePurchasesButton.HandleInput(0, 0) {
			return nil
		}
		if s.privacyPolicyButton.HandleInput(0, 0) {
			return nil
		}
		if s.closeButton.HandleInput(0, 0) {
			return nil
		}
		if s.advancedButton.HandleInput(0, 0) {
			return nil
		}
	}

	return nil
}

func (s *SettingsScene) handleBackButton(sceneManager *scene.Manager) {
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
	sceneManager.GoTo(NewTitleMapScene(sceneManager, g))
}

func (s *SettingsScene) Draw(screen *ebiten.Image) {
	if !s.initialized {
		return
	}
	s.advancedButton.Draw(screen)
	s.shopButton.Draw(screen)
	s.creditsButton.Draw(screen)
	s.updateCreditsButton.Draw(screen)
	s.reviewThisAppButton.Draw(screen)
	s.restorePurchasesButton.Draw(screen)
	s.privacyPolicyButton.Draw(screen)
	s.closeButton.Draw(screen)
	s.credits.Draw(screen)
}

func (s *SettingsScene) Resize(width, height int) {
	s.initialized = false
}
