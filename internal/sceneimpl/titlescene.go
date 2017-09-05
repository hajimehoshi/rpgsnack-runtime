// Copyright 2016 Hajime Hoshi
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
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type TitleScene struct {
	init             bool
	newGameButton    *ui.Button
	resumeGameButton *ui.Button
	settingsButton   *ui.Button
	moregamesButton  *ui.Button
	warningDialog    *ui.Dialog
	warningLabel     *ui.Label
	warningYesButton *ui.Button
	warningNoButton  *ui.Button
	quitDialog       *ui.Dialog
	quitLabel        *ui.Label
	quitYesButton    *ui.Button
	quitNoButton     *ui.Button
	waitingRequestID int
}

func NewTitleScene() *TitleScene {
	settingsIcon := assets.GetImage("system/icon_settings.png")
	moreGamesIcon := assets.GetImage("system/icon_moregames.png")
	t := &TitleScene{
		resumeGameButton: ui.NewButton(0, 184, 120, 20, "click"),
		newGameButton:    ui.NewButton(0, 208, 120, 20, "click"),
		settingsButton:   ui.NewImageButton(0, 0, settingsIcon, settingsIcon, "click"),
		moregamesButton:  ui.NewImageButton(0, 0, moreGamesIcon, moreGamesIcon, "click"),
		warningDialog:    ui.NewDialog(0, 64, 152, 124),
		warningLabel:     ui.NewLabel(16, 8),
		warningYesButton: ui.NewButton(0, 72, 120, 20, "click"),
		warningNoButton:  ui.NewButton(0, 96, 120, 20, "cancel"),
		quitDialog:       ui.NewDialog(0, 64, 152, 124),
		quitLabel:        ui.NewLabel(16, 8),
		quitYesButton:    ui.NewButton(0, 72, 120, 20, "click"),
		quitNoButton:     ui.NewButton(0, 96, 120, 20, "cancel"),
	}
	t.warningDialog.AddChild(t.warningLabel)
	t.warningDialog.AddChild(t.warningYesButton)
	t.warningDialog.AddChild(t.warningNoButton)

	t.quitDialog.AddChild(t.quitLabel)
	t.quitDialog.AddChild(t.quitYesButton)
	t.quitDialog.AddChild(t.quitNoButton)
	return t
}

func (t *TitleScene) Update(sceneManager *scene.Manager) error {
	if t.waitingRequestID != 0 {
		r := sceneManager.ReceiveResultIfExists(t.waitingRequestID)
		if r != nil {
			t.waitingRequestID = 0
		}
		return nil
	}

	if !t.init {
		var titleBGM = sceneManager.Game().System.TitleBGM
		if titleBGM.Name == "" {
			if err := audio.StopBGM(); err != nil {
				return err
			}
		} else {
			if err := audio.PlayBGM(titleBGM.Name, float64(titleBGM.Volume)/100); err != nil {
				return err
			}
		}
		t.init = true
	}

	if input.BackButtonPressed() {
		t.handleBackButton()
	}

	t.newGameButton.Text = texts.Text(sceneManager.Language(), texts.TextIDNewGame)
	t.resumeGameButton.Text = texts.Text(sceneManager.Language(), texts.TextIDResumeGame)
	t.warningLabel.Text = texts.Text(sceneManager.Language(), texts.TextIDNewGameWarning)
	t.warningYesButton.Text = texts.Text(sceneManager.Language(), texts.TextIDYes)
	t.warningNoButton.Text = texts.Text(sceneManager.Language(), texts.TextIDNo)
	t.quitLabel.Text = texts.Text(sceneManager.Language(), texts.TextIDQuitGame)
	t.quitYesButton.Text = texts.Text(sceneManager.Language(), texts.TextIDYes)
	t.quitNoButton.Text = texts.Text(sceneManager.Language(), texts.TextIDNo)

	w, h := sceneManager.Size()
	t.newGameButton.X = (w/consts.TileScale - t.newGameButton.Width) / 2
	t.resumeGameButton.X = (w/consts.TileScale - t.resumeGameButton.Width) / 2
	t.settingsButton.X = w/consts.TileScale - 16
	t.settingsButton.Y = h/consts.TileScale - 16
	t.moregamesButton.X = 4
	t.moregamesButton.Y = h/consts.TileScale - 16
	t.warningDialog.X = (w/consts.TileScale-160)/2 + 4
	t.warningYesButton.X = (t.warningDialog.Width - t.warningYesButton.Width) / 2
	t.warningNoButton.X = (t.warningDialog.Width - t.warningNoButton.Width) / 2
	t.quitDialog.X = (w/consts.TileScale-160)/2 + 4
	t.quitYesButton.X = (t.quitDialog.Width - t.quitYesButton.Width) / 2
	t.quitNoButton.X = (t.quitDialog.Width - t.quitNoButton.Width) / 2

	if !sceneManager.HasProgress() {
		t.resumeGameButton.Visible = false
		t.newGameButton.Y = 184
	} else {
		t.resumeGameButton.Visible = true
		t.resumeGameButton.Y = 184
		t.newGameButton.Y = 208
	}

	t.warningDialog.Update()
	if !t.warningDialog.Visible && !t.quitDialog.Visible {
		t.newGameButton.Update()
		t.resumeGameButton.Update()
		t.settingsButton.Update()
		t.moregamesButton.Update()
	}
	if t.warningYesButton.Pressed() {
		if err := audio.StopBGM(); err != nil {
			return err
		}
		sceneManager.GoToWithFading(NewMapScene(), 60)
		return nil
	}
	if t.warningNoButton.Pressed() {
		t.warningDialog.Visible = false
		return nil
	}
	if t.warningDialog.Visible {
		return nil
	}

	t.quitDialog.Update()
	if t.quitYesButton.Pressed() {
		sceneManager.Requester().RequestTerminateGame()
		return nil
	}
	if t.quitNoButton.Pressed() {
		t.quitDialog.Visible = false
		return nil
	}
	if t.quitDialog.Visible {
		return nil
	}

	if t.newGameButton.Pressed() {
		if sceneManager.HasProgress() {
			t.warningDialog.Visible = true
		} else {
			if err := audio.StopBGM(); err != nil {
				return err
			}
			sceneManager.GoToWithFading(NewMapScene(), 60)
		}
		return nil
	}
	if t.resumeGameButton.Pressed() {
		var game *gamestate.Game
		if err := msgpack.Unmarshal(sceneManager.Progress(), &game); err != nil {
			return err
		}
		if err := audio.StopBGM(); err != nil {
			return err
		}
		sceneManager.GoToWithFading(NewMapSceneWithGame(game), 60)
		return nil
	}
	if t.settingsButton.Pressed() {
		if err := audio.StopBGM(); err != nil {
			return err
		}
		sceneManager.GoTo(NewSettingsScene())
		return nil
	}
	if t.moregamesButton.Pressed() {
		t.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(t.waitingRequestID, "more", "")
		return nil
	}
	return nil
}

func (t *TitleScene) handleBackButton() {
	if t.warningDialog.Visible {
		audio.PlaySE("cancel", 1.0)
		t.warningDialog.Visible = false
		return
	}
	if t.quitDialog.Visible {
		audio.PlaySE("cancel", 1.0)
		t.quitDialog.Visible = false
		return
	}

	audio.PlaySE("click", 1.0)
	t.quitDialog.Visible = true
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	timg := assets.GetImage("titles/title.png")
	tw, _ := timg.Size()
	sw, _ := screen.Size()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate((float64(sw)-float64(tw))/2, 0)
	screen.DrawImage(timg, op)

	// TODO: hide buttons to avoid visual conflicts between the dialog and the buttons
	if !t.warningDialog.Visible && !t.quitDialog.Visible {
		t.newGameButton.Draw(screen)
		t.resumeGameButton.Draw(screen)
		t.settingsButton.Draw(screen)
		t.moregamesButton.Draw(screen)
	}
	t.warningDialog.Draw(screen)
	t.quitDialog.Draw(screen)
}
