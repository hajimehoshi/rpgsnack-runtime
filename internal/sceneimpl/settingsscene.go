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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type SettingsScene struct {
	infoLabel              *ui.Label
	creditButton           *ui.Button
	removeAdsButton        *ui.Button
	reviewThisAppButton    *ui.Button
	restorePurchasesButton *ui.Button
	moreGamesButton        *ui.Button
	closeButton            *ui.Button
}

func NewSettingsScene() *SettingsScene {
	return &SettingsScene{
		infoLabel:              ui.NewLabel(4, 4, "Info"),
		creditButton:           ui.NewButton(0, 28, 120, 20, "Credit"),
		removeAdsButton:        ui.NewButton(0, 52, 120, 20, "Remove Ads"),
		reviewThisAppButton:    ui.NewButton(0, 76, 120, 20, "Review This App"),
		restorePurchasesButton: ui.NewButton(0, 100, 120, 20, "Restore Purchases"),
		moreGamesButton:        ui.NewButton(0, 124, 120, 20, "More Games"),
		closeButton:            ui.NewButton(0, 148, 120, 20, "Close"),
	}
}

func (s *SettingsScene) Update(sceneManager *scene.Manager) error {
	w, _ := sceneManager.Size()
	s.creditButton.X = (w/scene.TileScale - s.creditButton.Width) / 2
	s.removeAdsButton.X = (w/scene.TileScale - s.removeAdsButton.Width) / 2
	s.reviewThisAppButton.X = (w/scene.TileScale - s.reviewThisAppButton.Width) / 2
	s.restorePurchasesButton.X = (w/scene.TileScale - s.restorePurchasesButton.Width) / 2
	s.moreGamesButton.X = (w/scene.TileScale - s.moreGamesButton.Width) / 2
	s.closeButton.X = (w/scene.TileScale - s.closeButton.Width) / 2
	if err := s.creditButton.Update(0, 0); err != nil {
		return err
	}
	if err := s.removeAdsButton.Update(0, 0); err != nil {
		return err
	}
	if err := s.reviewThisAppButton.Update(0, 0); err != nil {
		return err
	}
	if err := s.restorePurchasesButton.Update(0, 0); err != nil {
		return err
	}
	if err := s.moreGamesButton.Update(0, 0); err != nil {
		return err
	}
	if err := s.closeButton.Update(0, 0); err != nil {
		return err
	}
	if s.closeButton.Pressed() {
		sceneManager.GoTo(NewTitleScene())
		return nil
	}
	return nil
}

func (s *SettingsScene) Draw(screen *ebiten.Image) error {
	if err := s.infoLabel.Draw(screen); err != nil {
		return err
	}
	if err := s.creditButton.Draw(screen); err != nil {
		return err
	}
	if err := s.removeAdsButton.Draw(screen); err != nil {
		return err
	}
	if err := s.reviewThisAppButton.Draw(screen); err != nil {
		return err
	}
	if err := s.restorePurchasesButton.Draw(screen); err != nil {
		return err
	}
	if err := s.moreGamesButton.Draw(screen); err != nil {
		return err
	}
	if err := s.closeButton.Draw(screen); err != nil {
		return err
	}
	return nil
}
