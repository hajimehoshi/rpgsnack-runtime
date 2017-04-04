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
	"encoding/json"

	"golang.org/x/text/language/display"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

const (
	buttonOffsetX = 4
	buttonDeltaY  = 24
)

type SettingsScene struct {
	settingsLabel          *ui.Label
	languageButton         *ui.Button
	creditButton           *ui.Button
	removeAdsButton        *ui.Button
	reviewThisAppButton    *ui.Button
	restorePurchasesButton *ui.Button
	moreGamesButton        *ui.Button
	closeButton            *ui.Button
	languageDialog         *ui.Dialog
	languageButtons        []*ui.Button
	waitingRequestID       int
	isAdsRemoved           bool
}

func NewSettingsScene() *SettingsScene {
	s := &SettingsScene{
		settingsLabel:          ui.NewLabel(4, 4),
		languageButton:         ui.NewButton(0, 0, 120, 20),
		creditButton:           ui.NewButton(0, 0, 120, 20),
		removeAdsButton:        ui.NewButton(0, 0, 120, 20),
		reviewThisAppButton:    ui.NewButton(0, 0, 120, 20),
		restorePurchasesButton: ui.NewButton(0, 0, 120, 20),
		moreGamesButton:        ui.NewButton(0, 0, 120, 20),
		closeButton:            ui.NewButton(0, 0, 120, 20),
		languageDialog:         ui.NewDialog(0, 4, 152, 232),
	}
	for i, l := range data.Current().Texts.Languages() {
		n := display.Self.Name(l)
		b := ui.NewButton(0, 8+i*buttonDeltaY, 120, 20)
		b.Text = n
		s.languageDialog.AddChild(b)
		s.languageButtons = append(s.languageButtons, b)
	}
	s.UpdatePurchasesState()
	return s
}

// TODO: Move this method to load.go?
func (s *SettingsScene) isPurchased(key string) bool {
	var purchases []string
	if data.Purchases() != nil {
		if err := json.Unmarshal(data.Purchases(), &purchases); err != nil {
			panic(err)
		}
	}

	for _, p := range purchases {
		if p == key {
			return true
		}
	}

	return false
}

func (s *SettingsScene) UpdatePurchasesState() {
	s.isAdsRemoved = s.isPurchased("ads_removal")
}

func (s *SettingsScene) Update(sceneManager *scene.Manager) error {
	if s.waitingRequestID != 0 {
		r := sceneManager.ReceiveResultIfExists(s.waitingRequestID)
		if r != nil {
			s.waitingRequestID = 0
			switch r.Type {
			case scene.RequestTypePurchase, scene.RequestTypeRestorePurchases:
				// Note: Ideally we should show a notification toast to notify users about the result
				// For now, the notifications are handled on the native platform side
				if r.Succeeded {
					s.UpdatePurchasesState()
				}
			}
		}
	}
	s.settingsLabel.Text = texts.Text(sceneManager.Language(), texts.TextIDSettings)
	s.languageButton.Text = texts.Text(sceneManager.Language(), texts.TextIDLanguage)
	s.creditButton.Text = texts.Text(sceneManager.Language(), texts.TextIDCredit)
	s.removeAdsButton.Text = texts.Text(sceneManager.Language(), texts.TextIDRemoveAds)
	s.reviewThisAppButton.Text = texts.Text(sceneManager.Language(), texts.TextIDReviewThisApp)
	s.restorePurchasesButton.Text = texts.Text(sceneManager.Language(), texts.TextIDRestorePurchases)
	s.moreGamesButton.Text = texts.Text(sceneManager.Language(), texts.TextIDMoreGames)
	s.closeButton.Text = texts.Text(sceneManager.Language(), texts.TextIDClose)

	buttonIndex := 1
	s.languageButton.Y = buttonOffsetX + buttonIndex*buttonDeltaY
	buttonIndex++
	s.creditButton.Y = buttonOffsetX + buttonIndex*buttonDeltaY
	buttonIndex++

	s.removeAdsButton.Visible = !s.isAdsRemoved
	if !s.isAdsRemoved {
		s.removeAdsButton.Y = buttonOffsetX + buttonIndex*buttonDeltaY
		buttonIndex++
	}
	s.reviewThisAppButton.Y = buttonOffsetX + buttonIndex*buttonDeltaY
	buttonIndex++
	s.restorePurchasesButton.Y = buttonOffsetX + buttonIndex*buttonDeltaY
	buttonIndex++
	s.moreGamesButton.Y = buttonOffsetX + buttonIndex*buttonDeltaY
	buttonIndex++
	s.closeButton.Y = buttonOffsetX + buttonIndex*buttonDeltaY

	w, _ := sceneManager.Size()
	s.languageButton.X = (w/scene.TileScale - s.languageButton.Width) / 2
	s.creditButton.X = (w/scene.TileScale - s.creditButton.Width) / 2
	s.removeAdsButton.X = (w/scene.TileScale - s.removeAdsButton.Width) / 2
	s.reviewThisAppButton.X = (w/scene.TileScale - s.reviewThisAppButton.Width) / 2
	s.restorePurchasesButton.X = (w/scene.TileScale - s.restorePurchasesButton.Width) / 2
	s.moreGamesButton.X = (w/scene.TileScale - s.moreGamesButton.Width) / 2
	s.closeButton.X = (w/scene.TileScale - s.closeButton.Width) / 2
	s.languageDialog.X = (w/scene.TileScale-160)/2 + 4
	for _, b := range s.languageButtons {
		b.X = (s.languageDialog.Width - b.Width) / 2
	}
	s.languageDialog.Update()
	if !s.languageDialog.Visible {
		s.languageButton.Update()
		s.creditButton.Update()
		s.removeAdsButton.Update()
		s.reviewThisAppButton.Update()
		s.restorePurchasesButton.Update()
		s.moreGamesButton.Update()
		s.closeButton.Update()
	}
	for i, b := range s.languageButtons {
		if b.Pressed() {
			s.languageDialog.Visible = false
			lang := data.Current().Texts.Languages()[i]
			sceneManager.SetLanguage(lang)
			return nil
		}
	}
	if s.languageButton.Pressed() {
		s.languageDialog.Visible = true
		return nil
	}
	if s.removeAdsButton.Pressed() {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestPurchase(s.waitingRequestID, "ads_removal")
		return nil
	}
	if s.reviewThisAppButton.Pressed() {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(s.waitingRequestID, "review", "")
		return nil
	}
	if s.restorePurchasesButton.Pressed() {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestRestorePurchases(s.waitingRequestID)
		return nil
	}
	if s.moreGamesButton.Pressed() {
		s.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(s.waitingRequestID, "more", "")
		return nil
	}
	if s.closeButton.Pressed() {
		sceneManager.GoTo(NewTitleScene())
		return nil
	}
	return nil
}

func (s *SettingsScene) Draw(screen *ebiten.Image) {
	s.settingsLabel.Draw(screen)
	s.languageButton.Draw(screen)
	s.creditButton.Draw(screen)
	s.removeAdsButton.Draw(screen)
	s.reviewThisAppButton.Draw(screen)
	s.restorePurchasesButton.Draw(screen)
	s.moreGamesButton.Draw(screen)
	s.closeButton.Draw(screen)
	s.languageDialog.Draw(screen)
}
