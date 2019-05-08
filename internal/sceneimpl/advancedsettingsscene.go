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

package sceneimpl

import (
	"github.com/hajimehoshi/ebiten"
	"golang.org/x/text/language/display"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

// TODO
// We should rename setttingScene and AdvancedSettingsScene
// There is no stronger need to split scenes,
// so we probably should consider merging them into one.
type AdvancedSettingsScene struct {
	settingsLabel    *ui.Label
	languageButton   *ui.Button
	vibrationLabel   *ui.Label
	vibrationButton  *ui.SwitchButton
	resetGameButton  *ui.Button
	bgmLabel         *ui.Label
	bgmSlider        *ui.Slider
	seLabel          *ui.Label
	seSlider         *ui.Slider
	closeButton      *ui.Button
	warningDialog    *ui.Dialog
	warningLabel     *ui.Label
	warningYesButton *ui.Button
	warningNoButton  *ui.Button
	waitingRequestID int
	languageDialog   *ui.Dialog
	languageButtons  []*ui.Button

	initialized bool
	baseX       int
	baseY       int
}

func (s *AdvancedSettingsScene) calcButtonY(index int) int {
	return s.baseY + buttonOffsetY + index*buttonDeltaY
}

func (s *AdvancedSettingsScene) initUI(sceneManager *scene.Manager) {
	w, h := sceneManager.Size()
	s.baseX = (w/consts.TileScale - 120) / 2
	s.baseY = (h - 640) / (2 * consts.TileScale)

	s.settingsLabel = ui.NewLabel(16, s.baseY+8)
	s.languageButton = ui.NewButton(s.baseX, s.calcButtonY(1), 120, 20, "system/click")
	s.bgmLabel = ui.NewLabel(s.baseX, s.calcButtonY(2)+4)
	s.bgmSlider = ui.NewSlider(s.baseX+48, s.calcButtonY(2), 50, 0, 100)
	s.seLabel = ui.NewLabel(s.baseX, s.calcButtonY(3)+4)
	s.seSlider = ui.NewSlider(s.baseX+48, s.calcButtonY(3), 50, 0, 100)
	s.resetGameButton = ui.NewButton(s.baseX, s.calcButtonY(4), 120, 20, "system/click")
	s.vibrationLabel = ui.NewLabel(s.baseX, s.calcButtonY(4)+4)
	s.vibrationButton = ui.NewSwitchButton(s.baseX+72, s.calcButtonY(4))
	s.resetGameButton = ui.NewButton(s.baseX, s.calcButtonY(5), 120, 20, "system/click")
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

	s.vibrationButton.SetOnToggled(func(_ *ui.SwitchButton, value bool) {
		// TODO ON/OFF vibration
	})

	s.bgmSlider.SetOnValueChanged(func(slider *ui.Slider, value int) {
		// TODO set bgm volume
	})

	s.seSlider.SetOnValueChanged(func(slider *ui.Slider, value int) {
		// TODO set SE volume
	})

	s.resetGameButton.SetOnPressed(func(_ *ui.Button) {
		s.warningDialog.Show()
	})

	s.closeButton.SetOnPressed(func(_ *ui.Button) {
		sceneManager.GoTo(NewSettingsScene())
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
}

func (s *AdvancedSettingsScene) updateButtonTexts() {
	s.settingsLabel.Text = texts.Text(lang.Get(), texts.TextIDAdvancedSettings)
	s.vibrationLabel.Text = texts.Text(lang.Get(), texts.TextIDVibration)
	s.languageButton.SetText(texts.Text(lang.Get(), texts.TextIDLanguage))
	s.bgmLabel.Text = texts.Text(lang.Get(), texts.TextIDBGMVolume)
	s.seLabel.Text = texts.Text(lang.Get(), texts.TextIDSEVolume)
	s.closeButton.SetText(texts.Text(lang.Get(), texts.TextIDBack))
	s.resetGameButton.SetText(texts.Text(lang.Get(), texts.TextIDResetGame))
	s.warningLabel.Text = texts.Text(lang.Get(), texts.TextIDNewGameWarning)
	s.warningYesButton.SetText(texts.Text(lang.Get(), texts.TextIDYes))
	s.warningNoButton.SetText(texts.Text(lang.Get(), texts.TextIDNo))
}

func (s *AdvancedSettingsScene) Update(sceneManager *scene.Manager) error {
	if s.waitingRequestID != 0 {
		r := sceneManager.ReceiveResultIfExists(s.waitingRequestID)
		if r != nil {
			s.waitingRequestID = 0
		}
		return nil
	}

	if !s.initialized {
		s.initUI(sceneManager)
		s.initialized = true
	}

	if input.BackButtonPressed() {
		s.handleBackButton(sceneManager)
	}

	s.updateButtonTexts()

	s.languageDialog.Update()
	if !s.languageDialog.Visible() && !s.warningDialog.Visible() {
		s.languageButton.Update()
		s.closeButton.Update()
		s.resetGameButton.Update()
		s.vibrationLabel.Update()
		s.vibrationButton.Update()
		s.bgmLabel.Update()
		s.bgmSlider.Update()
		s.seLabel.Update()
		s.seSlider.Update()
	}

	if sceneManager.HasProgress() {
		s.resetGameButton.Enable()
	} else {
		s.resetGameButton.Disable()
	}

	return nil
}

func (s *AdvancedSettingsScene) handleBackButton(sceneManager *scene.Manager) {
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

	sceneManager.GoTo(NewSettingsScene())
}

func (s *AdvancedSettingsScene) Draw(screen *ebiten.Image) {
	if !s.initialized {
		return
	}
	s.settingsLabel.Draw(screen)
	s.languageButton.Draw(screen)
	s.resetGameButton.Draw(screen)
	s.bgmLabel.Draw(screen)
	s.bgmSlider.Draw(screen)
	s.seLabel.Draw(screen)
	s.seSlider.Draw(screen)
	s.warningDialog.Draw(screen)
	s.languageDialog.Draw(screen)
	s.closeButton.Draw(screen)
	s.vibrationLabel.Draw(screen)
	s.vibrationButton.Draw(screen)
}

func (s *AdvancedSettingsScene) Resize() {
	s.initialized = false
}
