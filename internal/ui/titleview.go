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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
)

type TitleView struct {
	initialized      bool
	initializedUI    bool
	startGameButton  *Button
	removeAdsButton  *Button
	settingsButton   *Button
	moregamesButton  *Button
	quitPopup        *Popup
	quitLabel        *Label
	quitYesButton    *Button
	quitNoButton     *Button
	waitingRequestID int // TODO: Can we remove this?
	bgImage          *ebiten.Image
	footerOffset     int

	sceneWidth  int
	sceneHeight int

	shakeStartGameButtonCount int

	onQuit      func()
	onStartGame func()
	onRemoveAds func()
	onSettings  func()
	onMoreGames func()
}

const (
	footerHeight = 192
)

func NewTitleView(sceneWidth, sceneHeight int) *TitleView {
	t := &TitleView{
		sceneWidth:  sceneWidth,
		sceneHeight: sceneHeight,
	}
	return t
}

func (t *TitleView) SetOnQuit(f func()) {
	t.onQuit = f
}

func (t *TitleView) SetOnStartGame(f func()) {
	t.onStartGame = f
}

func (t *TitleView) SetOnRemoveAds(f func()) {
	t.onRemoveAds = f
}

func (t *TitleView) SetOnSettings(f func()) {
	t.onSettings = f
}

func (t *TitleView) SetOnMoreGames(f func()) {
	t.onMoreGames = f
}

func (t *TitleView) startGameButtonX() int {
	return (t.sceneWidth/consts.TileScale - 120) / 2
}

const (
	shakeFrame = 15
)

func (t *TitleView) initUI() {
	w, h := t.sceneWidth, t.sceneHeight

	settingsIcon := assets.GetImage("system/common/icon_settings.png")
	moreGamesIcon := assets.GetImage("system/common/icon_moregames.png")

	by := 16
	t.footerOffset = 0
	if consts.HasExtraBottomGrid(t.sceneHeight) {
		by = 36
		t.footerOffset = 48
	}

	t.startGameButton = NewTextButton(t.startGameButtonX(), h/consts.TileScale-by-32, 120, 20, "system/start")
	t.removeAdsButton = NewTextButton((w/consts.TileScale-120)/2+20, h/consts.TileScale-by-4, 80, 20, "system/click")
	t.removeAdsButton.textColor = color.RGBA{0xc8, 0xc8, 0xc8, 0xff}
	t.settingsButton = NewImageButton(w/consts.TileScale-24, h/consts.TileScale-by, settingsIcon, settingsIcon, "system/click")
	t.settingsButton.touchExpand = 10
	t.moregamesButton = NewImageButton(12, h/consts.TileScale-by, moreGamesIcon, moreGamesIcon, "system/click")
	t.moregamesButton.touchExpand = 10

	t.quitPopup = NewPopup((w/consts.TileScale-160)/2+4, (h)/(2*consts.TileScale)-64, 152, 124)
	t.quitLabel = NewLabel(16, 8)
	t.quitYesButton = NewButton((152-120)/2, 72, 120, 20, "system/click")
	t.quitNoButton = NewButton((152-120)/2, 96, 120, 20, "system/cancel")
	t.quitPopup.AddChild(t.quitLabel)
	t.quitPopup.AddChild(t.quitYesButton)
	t.quitPopup.AddChild(t.quitNoButton)

	t.quitYesButton.SetOnPressed(func(_ *Button) {
		if t.onQuit != nil {
			t.onQuit()
		}
	})
	t.quitNoButton.SetOnPressed(func(_ *Button) {
		t.quitPopup.Hide()
	})
	t.startGameButton.SetOnPressed(func(_ *Button) {
		if t.onStartGame != nil {
			t.onStartGame()
		}
	})
	t.removeAdsButton.SetOnPressed(func(_ *Button) {
		if t.onRemoveAds != nil {
			t.onRemoveAds()
		}
	})
	t.settingsButton.SetOnPressed(func(_ *Button) {
		if t.onSettings != nil {
			t.onSettings()
		}
	})
	t.moregamesButton.SetOnPressed(func(_ *Button) {
		if t.onMoreGames != nil {
			t.onMoreGames()
		}
	})
}

func (t *TitleView) WaitingRequestID() int {
	return t.waitingRequestID
}

func (t *TitleView) SetWaitingRequestID(id int) {
	t.waitingRequestID = id
}

func (t *TitleView) ResetWaitingRequestID() {
	t.waitingRequestID = 0
}

func (t *TitleView) Update(game *data.Game, hasProgress bool, isAdsRemoved bool) error {
	if !t.initializedUI {
		t.initUI()
		t.initializedUI = true
	}

	if !t.initialized {
		if titleBGM := game.System.TitleBGM; titleBGM.Name == "" {
			audio.StopBGM(0)
		} else {
			audio.PlayBGM(titleBGM.Name, float64(titleBGM.Volume)/100, 0)
		}
		t.initialized = true
	}

	if input.BackButtonPressed() {
		t.handleBackButton()
	}

	if hasProgress {
		t.startGameButton.text = texts.Text(lang.Get(), texts.TextIDResumeGame)
	} else {
		t.startGameButton.text = texts.Text(lang.Get(), texts.TextIDNewGame)
	}
	if game.System.TitleTextColor == "black" {
		t.startGameButton.textColor = color.Black
	} else {
		t.startGameButton.textColor = color.White
	}

	t.removeAdsButton.text = texts.Text(lang.Get(), texts.TextIDRemoveAds)
	t.quitLabel.Text = texts.Text(lang.Get(), texts.TextIDQuitGame)
	t.quitYesButton.text = texts.Text(lang.Get(), texts.TextIDYes)
	t.quitNoButton.text = texts.Text(lang.Get(), texts.TextIDNo)

	t.quitPopup.Update()
	if !t.quitPopup.Visible() {
		t.startGameButton.Update()
		t.removeAdsButton.Update()
		t.settingsButton.Update()
		t.moregamesButton.Update()
	}

	t.removeAdsButton.visible = game.IsShopAvailable(data.ShopTypeHome) && !isAdsRemoved

	x := t.startGameButtonX()
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
	if t.quitPopup.Visible() {
		audio.PlaySE("system/cancel", 1.0)
		t.quitPopup.Hide()
		return
	}

	audio.PlaySE("system/click", 1.0)
	t.quitPopup.Show()
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
	timg := assets.GetLocalizedImage("titles/title")
	tw, th := timg.Size()
	sw, sh := screen.Size()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate((float64(sw)-float64(tw))/2, float64(sh-footerHeight-th)/2)
	screen.DrawImage(timg, op)
}

func (t *TitleView) Draw(screen *ebiten.Image) {
	if !t.initializedUI {
		return
	}

	t.drawFooter(screen)
	t.drawTitle(screen)

	// TODO: hide buttons to avoid visual conflicts between the popup and the buttons
	if !t.quitPopup.Visible() {
		t.startGameButton.Draw(screen)
		t.removeAdsButton.Draw(screen)
		t.settingsButton.Draw(screen)
		t.moregamesButton.Draw(screen)
	}
	t.quitPopup.Draw(screen)
}

func (t *TitleView) Resize(sceneWidth, sceneHeight int) {
	t.initializedUI = false
	t.sceneWidth = sceneWidth
	t.sceneHeight = sceneHeight
}

func (t *TitleView) ShakeStartGameButton() {
	t.shakeStartGameButtonCount = shakeFrame * 3
}
