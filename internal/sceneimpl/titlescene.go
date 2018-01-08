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
	"golang.org/x/text/language"

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
	titleLine        *ui.ImageView
	waitingRequestID int
	initialized      bool
	lang             language.Tag
	err              error
}

func NewTitleScene() *TitleScene {
	t := &TitleScene{}
	return t
}

func (t *TitleScene) initUI(sceneManager *scene.Manager) {
	w, h := sceneManager.Size()

	settingsIcon := ui.NewImagePart(assets.GetImage("system/common/icon_settings.png"))
	moreGamesIcon := ui.NewImagePart(assets.GetImage("system/common/icon_moregames.png"))

	t.resumeGameButton = ui.NewTextButton((w/consts.TileScale-120)/2, 184, 120, 20, "click")
	t.titleLine = ui.NewImageView((w/consts.TileScale-120)/2+20, 206, 1.0, ui.NewImagePart(assets.GetImage("system/common/title_line.png")))
	t.newGameButton = ui.NewTextButton((w/consts.TileScale-120)/2, 208, 120, 20, "click")
	t.settingsButton = ui.NewImageButton(w/consts.TileScale-16, h/consts.TileScale-16, settingsIcon, settingsIcon, "click")
	t.moregamesButton = ui.NewImageButton(4, h/consts.TileScale-16, moreGamesIcon, moreGamesIcon, "click")
	t.warningDialog = ui.NewDialog((w/consts.TileScale-160)/2+4, 64, 152, 124)
	t.warningLabel = ui.NewLabel(16, 8)
	t.warningYesButton = ui.NewButton((152-120)/2, 72, 120, 20, "click")
	t.warningNoButton = ui.NewButton((152-120)/2, 96, 120, 20, "cancel")
	t.warningDialog.AddChild(t.warningLabel)
	t.warningDialog.AddChild(t.warningYesButton)
	t.warningDialog.AddChild(t.warningNoButton)

	t.quitDialog = ui.NewDialog((w/consts.TileScale-160)/2+4, 64, 152, 124)
	t.quitLabel = ui.NewLabel(16, 8)
	t.quitYesButton = ui.NewButton((152-120)/2, 72, 120, 20, "click")
	t.quitNoButton = ui.NewButton((152-120)/2, 96, 120, 20, "cancel")
	t.quitDialog.AddChild(t.quitLabel)
	t.quitDialog.AddChild(t.quitYesButton)
	t.quitDialog.AddChild(t.quitNoButton)

	t.warningYesButton.SetOnPressed(func(_ *ui.Button) {
		audio.StopBGM(0)
		sceneManager.GoToWithFading(NewMapScene(), 60)
	})
	t.warningNoButton.SetOnPressed(func(_ *ui.Button) {
		t.warningDialog.Hide()
	})
	t.quitYesButton.SetOnPressed(func(_ *ui.Button) {
		sceneManager.Requester().RequestTerminateGame()
	})
	t.quitNoButton.SetOnPressed(func(_ *ui.Button) {
		t.quitDialog.Hide()
	})
	t.newGameButton.SetOnPressed(func(_ *ui.Button) {
		if sceneManager.HasProgress() {
			t.warningDialog.Show()
		} else {
			audio.StopBGM(0)
			sceneManager.GoToWithFading(NewMapScene(), 60)
		}
	})
	t.resumeGameButton.SetOnPressed(func(_ *ui.Button) {
		var game *gamestate.Game
		if err := msgpack.Unmarshal(sceneManager.Progress(), &game); err != nil {
			t.err = err
			return
		}
		audio.StopBGM(0)
		sceneManager.GoToWithFading(NewMapSceneWithGame(game), 60)
	})
	t.settingsButton.SetOnPressed(func(_ *ui.Button) {
		sceneManager.GoTo(NewSettingsScene())
	})
	t.moregamesButton.SetOnPressed(func(_ *ui.Button) {
		t.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(t.waitingRequestID, "more", "")
	})
}

func (t *TitleScene) Update(sceneManager *scene.Manager) error {
	t.lang = sceneManager.Language()
	if t.err != nil {
		return t.err
	}
	if !t.initialized {
		t.initUI(sceneManager)
		t.initialized = true
	}
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
			audio.StopBGM(0)
		} else {
			audio.PlayBGM(titleBGM.Name, float64(titleBGM.Volume)/100, 0)
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

	if !sceneManager.HasProgress() {
		t.resumeGameButton.Visible = false
		t.newGameButton.SetY(184)
	} else {
		t.resumeGameButton.Visible = true
		t.resumeGameButton.SetY(184)
		t.newGameButton.SetY(208)
	}

	t.warningDialog.Update()
	t.quitDialog.Update()
	if !t.warningDialog.Visible() && !t.quitDialog.Visible() {
		t.newGameButton.Update()
		t.resumeGameButton.Update()
		t.settingsButton.Update()
		t.moregamesButton.Update()
	}
	return nil
}

func (t *TitleScene) handleBackButton() {
	if t.warningDialog.Visible() {
		audio.PlaySE("cancel", 1.0)
		t.warningDialog.Hide()
		return
	}
	if t.quitDialog.Visible() {
		audio.PlaySE("cancel", 1.0)
		t.quitDialog.Hide()
		return
	}

	audio.PlaySE("click", 1.0)
	t.quitDialog.Show()
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	if !t.initialized {
		return
	}
	timg := assets.GetLocalizeImage("titles/title", t.lang)
	tw, _ := timg.Size()
	sw, _ := screen.Size()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate((float64(sw)-float64(tw))/2, 0)
	screen.DrawImage(timg, op)

	// TODO: hide buttons to avoid visual conflicts between the dialog and the buttons
	if !t.warningDialog.Visible() && !t.quitDialog.Visible() {
		t.newGameButton.Draw(screen)
		t.titleLine.Draw(screen)
		t.resumeGameButton.Draw(screen)
		t.settingsButton.Draw(screen)
		t.moregamesButton.Draw(screen)
	}
	t.warningDialog.Draw(screen)
	t.quitDialog.Draw(screen)
}
