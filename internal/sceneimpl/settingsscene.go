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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type SettingsScene struct {
	infoLabel              *ui.Label
	languageButton         *ui.Button
	creditButton           *ui.Button
	removeAdsButton        *ui.Button
	reviewThisAppButton    *ui.Button
	restorePurchasesButton *ui.Button
	moreGamesButton        *ui.Button
	closeButton            *ui.Button
	languageDialog         *ui.Dialog
	languageButtons        []*ui.Button
}

func NewSettingsScene() *SettingsScene {
	const d = 24
	s := &SettingsScene{
		infoLabel:              ui.NewLabel(4, 4, "Info"),
		languageButton:         ui.NewButton(0, 4+1*d, 120, 20, "Language"),
		creditButton:           ui.NewButton(0, 4+2*d, 120, 20, "Credit"),
		removeAdsButton:        ui.NewButton(0, 4+3*d, 120, 20, "Remove Ads"),
		reviewThisAppButton:    ui.NewButton(0, 4+4*d, 120, 20, "Review This App"),
		restorePurchasesButton: ui.NewButton(0, 4+5*d, 120, 20, "Restore Purchases"),
		moreGamesButton:        ui.NewButton(0, 4+6*d, 120, 20, "More Games"),
		closeButton:            ui.NewButton(0, 4+7*d, 120, 20, "Close"),
		languageDialog:         ui.NewDialog(0, 4, 152, 232),
	}
	for i, l := range data.Current().Texts.Languages() {
		b := ui.NewButton(0, 8+i*d, 120, 20, l.String())
		s.languageDialog.AddChild(b)
		s.languageButtons = append(s.languageButtons, b)
	}
	return s
}

func (s *SettingsScene) Update(sceneManager *scene.Manager) error {
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
	if s.closeButton.Pressed() {
		sceneManager.GoTo(NewTitleScene())
		return nil
	}
	return nil
}

func (s *SettingsScene) Draw(screen *ebiten.Image) {
	s.infoLabel.Draw(screen)
	s.languageButton.Draw(screen)
	s.creditButton.Draw(screen)
	s.removeAdsButton.Draw(screen)
	s.reviewThisAppButton.Draw(screen)
	s.restorePurchasesButton.Draw(screen)
	s.moreGamesButton.Draw(screen)
	s.closeButton.Draw(screen)
	s.languageDialog.Draw(screen)
}
