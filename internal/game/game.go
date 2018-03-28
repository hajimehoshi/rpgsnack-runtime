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

package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/sceneimpl"
)

type Game struct {
	projectLocation  string
	width            int
	height           int
	requester        scene.Requester
	sceneManager     *scene.Manager
	loadProgressCh   chan data.LoadProgress
	loadProgressRate float64
}

func New(width, height int, requester scene.Requester) *Game {
	g := &Game{
		width:     width,
		height:    height,
		requester: requester,
	}
	g.loadGameData()
	return g
}

var lang language.Tag

func Language() language.Tag {
	return lang
}

func (g *Game) loadGameData() {
	ch := make(chan data.LoadProgress, 4)
	go func() {
		data.Load(g.projectLocation, ch)
	}()
	g.loadProgressCh = ch
}

func (g *Game) Update(screen *ebiten.Image) error {
	if err := g.update(); err != nil {
		return err
	}
	if ebiten.IsRunningSlowly() {
		return nil
	}
	g.draw(screen)
	return nil
}

func (g *Game) update() error {
	if g.loadProgressCh != nil {
		select {
		case d := <-g.loadProgressCh:
			if d.Error != nil {
				return d.Error
			}
			g.loadProgressRate = d.Progress

			if d.LoadedData == nil {
				return nil
			}
			g.loadProgressCh = nil
			da := d.LoadedData
			assets.Set(da.Assets, da.AssetsMetadata)
			g.sceneManager = scene.NewManager(g.width, g.height, g.requester, da.Game, da.Progress, da.Purchases)
			g.sceneManager.SetLanguage(da.Language)
			g.sceneManager.InitScene(sceneimpl.NewTitleScene())
		default:
			return nil
		}
	}
	input.Update()
	if err := audio.Update(); err != nil {
		return err
	}
	takeCPUProfileIfAvailable()
	if err := g.sceneManager.Update(); err != nil {
		return err
	}
	return nil
}

func (g *Game) draw(screen *ebiten.Image) {
	if g.loadProgressCh != nil {
		const barHeight = 8
		w, h := screen.Size()
		barWidth := float64(w)
		y := float64(h-barHeight) / 2
		ebitenutil.DrawRect(screen, 0, y, barWidth, barHeight, color.RGBA{0x80, 0x80, 0x80, 0x80})
		activeWidth := barWidth * g.loadProgressRate
		ebitenutil.DrawRect(screen, 0, y, activeWidth, barHeight, color.RGBA{0xff, 0xff, 0xff, 0xff})
		return
	}
	g.sceneManager.Draw(screen)
}

func Title() string {
	return "Clock of Atonement"
}

func (g *Game) Size() (int, int) {
	return g.sceneManager.Size()
}

func (g *Game) FinishUnlockAchievement(id int) {
	g.sceneManager.FinishUnlockAchievement(id)
}

func (g *Game) FinishSaveProgress(id int) {
	g.sceneManager.FinishSaveProgress(id)
}

func (g *Game) FinishPurchase(id int, success bool, purchases []uint8) {
	g.sceneManager.FinishPurchase(id, success, purchases)
}

func (g *Game) FinishRestorePurchases(id int, success bool, purchases []uint8) {
	g.sceneManager.FinishRestorePurchases(id, success, purchases)
}

func (g *Game) FinishInterstitialAds(id int) {
	g.sceneManager.FinishInterstitialAds(id)
}

func (g *Game) FinishRewardedAds(id int, success bool) {
	g.sceneManager.FinishRewardedAds(id, success)
}

func (g *Game) FinishOpenLink(id int) {
	g.sceneManager.FinishOpenLink(id)
}

func (g *Game) FinishShareImage(id int) {
	g.sceneManager.FinishShareImage(id)
}

func (g *Game) FinishChangeLanguage(id int) {
	g.sceneManager.FinishChangeLanguage(id)
}

func (g *Game) FinishGetIAPPrices(id int, success bool, prices []uint8) {
	g.sceneManager.FinishGetIAPPrices(id, success, prices)
}

func (g *Game) SetPlatformData(key scene.PlatformDataKey, value string) {
	g.sceneManager.SetPlatformData(key, value)
}
