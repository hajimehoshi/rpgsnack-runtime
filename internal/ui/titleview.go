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

package ui

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
)

type SceneMaker interface {
	NewMapScene() scene.Scene
	NewMapSceneWithGame(*gamestate.Game) scene.Scene
	NewSettingsScene() scene.Scene
}

type TitleView struct {
	init             bool
	startGameButton  *Button
	removeAdsButton  *Button
	settingsButton   *Button
	moregamesButton  *Button
	quitDialog       *Dialog
	quitLabel        *Label
	quitYesButton    *Button
	quitNoButton     *Button
	waitingRequestID int
	initialized      bool
	bgImage          *ebiten.Image
	footerOffset     int
	err              error

	sceneMaker SceneMaker

	shakeStartGameButtonCount int
}

const (
	footerHeight = 192
)

func NewTitleView(sceneMaker SceneMaker) *TitleView {
	t := &TitleView{
		sceneMaker: sceneMaker,
	}
	return t
}

func (t *TitleView) startGameButtonX(sceneManager *scene.Manager) int {
	w, _ := sceneManager.Size()
	return (w/consts.TileScale - 120) / 2
}

const (
	shakeFrame = 15
)

func (t *TitleView) initUI(sceneManager *scene.Manager) {
	w, h := sceneManager.Size()

	settingsIcon := assets.GetImage("system/common/icon_settings.png")
	moreGamesIcon := assets.GetImage("system/common/icon_moregames.png")

	by := 16
	t.footerOffset = 0
	if sceneManager.HasExtraBottomGrid() {
		by = 36
		t.footerOffset = 48
	}

	t.startGameButton = NewTextButton(t.startGameButtonX(sceneManager), h/consts.TileScale-by-32, 120, 20, "system/start")
	t.removeAdsButton = NewTextButton((w/consts.TileScale-120)/2+20, h/consts.TileScale-by-4, 80, 20, "system/click")
	t.removeAdsButton.textColor = color.RGBA{0xc8, 0xc8, 0xc8, 0xff}
	t.settingsButton = NewImageButton(w/consts.TileScale-24, h/consts.TileScale-by, settingsIcon, settingsIcon, "system/click")
	t.settingsButton.touchExpand = 10
	t.moregamesButton = NewImageButton(12, h/consts.TileScale-by, moreGamesIcon, moreGamesIcon, "system/click")
	t.moregamesButton.touchExpand = 10

	t.quitDialog = NewDialog((w/consts.TileScale-160)/2+4, (h)/(2*consts.TileScale)-64, 152, 124)
	t.quitLabel = NewLabel(16, 8)
	t.quitYesButton = NewButton((152-120)/2, 72, 120, 20, "system/click")
	t.quitNoButton = NewButton((152-120)/2, 96, 120, 20, "system/cancel")
	t.quitDialog.AddChild(t.quitLabel)
	t.quitDialog.AddChild(t.quitYesButton)
	t.quitDialog.AddChild(t.quitNoButton)

	t.quitYesButton.SetOnPressed(func(_ *Button) {
		sceneManager.Requester().RequestTerminateGame()
	})
	t.quitNoButton.SetOnPressed(func(_ *Button) {
		t.quitDialog.Hide()
	})
	t.startGameButton.SetOnPressed(func(_ *Button) {
		audio.Stop()
		if sceneManager.HasProgress() {
			// TODO: Remove this logic from UI.
			var game *gamestate.Game
			if err := msgpack.Unmarshal(sceneManager.Progress(), &game); err != nil {
				t.err = err
				return
			}
			sceneManager.GoToWithFading(t.sceneMaker.NewMapSceneWithGame(game), 60)
		} else {
			sceneManager.GoToWithFading(t.sceneMaker.NewMapScene(), 60)
		}
	})
	t.removeAdsButton.SetOnPressed(func(_ *Button) {
		if sceneManager.Game().IsShopAvailable(data.ShopTypeHome) {
			sceneManager.Requester().RequestShowShop(t.waitingRequestID, string(sceneManager.ShopProductsDataByShop(data.ShopTypeHome)))
		}
	})
	t.settingsButton.SetOnPressed(func(_ *Button) {
		sceneManager.GoTo(t.sceneMaker.NewSettingsScene())
	})
	t.moregamesButton.SetOnPressed(func(_ *Button) {
		t.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(t.waitingRequestID, "more", "")
	})
}

func (t *TitleView) Update(sceneManager *scene.Manager) error {
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

	if sceneManager.HasProgress() {
		t.startGameButton.text = texts.Text(lang.Get(), texts.TextIDResumeGame)
	} else {
		t.startGameButton.text = texts.Text(lang.Get(), texts.TextIDNewGame)
	}
	if sceneManager.Game().System.TitleTextColor == "black" {
		t.startGameButton.textColor = color.Black
	} else {
		t.startGameButton.textColor = color.White
	}

	t.removeAdsButton.text = texts.Text(lang.Get(), texts.TextIDRemoveAds)
	t.quitLabel.Text = texts.Text(lang.Get(), texts.TextIDQuitGame)
	t.quitYesButton.text = texts.Text(lang.Get(), texts.TextIDYes)
	t.quitNoButton.text = texts.Text(lang.Get(), texts.TextIDNo)

	t.quitDialog.Update()
	if !t.quitDialog.Visible() {
		t.startGameButton.Update()
		t.removeAdsButton.Update()
		t.settingsButton.Update()
		t.moregamesButton.Update()
	}

	t.removeAdsButton.visible = sceneManager.IsAdsRemovable() && !sceneManager.IsAdsRemoved()

	x := t.startGameButtonX(sceneManager)
	t.startGameButton.SetX(x)
	if t.shakeStartGameButtonCount > 0 {
		tx := 0
		switch {
		case t.shakeStartGameButtonCount >= shakeFrame*2:
			tx = rand.Intn(5) - 2
		case t.shakeStartGameButtonCount >= shakeFrame:
			// Do nothing
		default:
			tx = rand.Intn(5) - 2
		}
		t.startGameButton.SetX(x + tx)
		t.shakeStartGameButtonCount--
	}

	return nil
}

func (t *TitleView) handleBackButton() {
	if t.quitDialog.Visible() {
		audio.PlaySE("system/cancel", 1.0)
		t.quitDialog.Hide()
		return
	}

	audio.PlaySE("system/click", 1.0)
	t.quitDialog.Show()
}

func (t *TitleView) drawFooter(screen *ebiten.Image) {
	fimg := assets.GetImage("system/common/title_footer.png")
	_, sh := screen.Size()

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(sh-footerHeight-t.footerOffset))
	screen.DrawImage(fimg, op)
}

func (t *TitleView) drawTitle(screen *ebiten.Image) {
	// TODO: titles/title is used in games before 'title as map' was introduced.
	// Remove this usage in the future.
	if !assets.ImageExists("titles/title") {
		return
	}
	timg := assets.GetLocalizeImage("titles/title")
	tw, th := timg.Size()
	sw, sh := screen.Size()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate((float64(sw)-float64(tw))/2, float64(sh-footerHeight-th)/2)
	screen.DrawImage(timg, op)
}

func (t *TitleView) Draw(screen *ebiten.Image) {
	if !t.initialized {
		return
	}

	t.drawFooter(screen)
	t.drawTitle(screen)

	// TODO: hide buttons to avoid visual conflicts between the dialog and the buttons
	if !t.quitDialog.Visible() {
		t.startGameButton.Draw(screen)
		t.removeAdsButton.Draw(screen)
		t.settingsButton.Draw(screen)
		t.moregamesButton.Draw(screen)
	}
	t.quitDialog.Draw(screen)
}

func (t *TitleView) Resize() {
	t.initialized = false
}

func (t *TitleView) ShakeStartGameButton() {
	t.shakeStartGameButtonCount = shakeFrame * 3
}
