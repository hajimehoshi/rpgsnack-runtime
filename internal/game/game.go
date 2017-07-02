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
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/sceneimpl"
)

type Game struct {
	projectPath  string
	width        int
	height       int
	requester    scene.Requester
	sceneManager *scene.Manager
	loadingCh    chan error
	loadedData   *data.LoadedData
	count        int
	prevScreen   *ebiten.Image
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

func (g *Game) loadGameData() {
	ch := make(chan error)
	go func() {
		defer close(ch)
		d, err := data.Load(g.projectPath)
		if err != nil {
			ch <- err
			return
		}
		g.loadedData = d
	}()
	g.loadingCh = ch
}

func (g *Game) Update(screen *ebiten.Image) error {
	if err := g.update(); err != nil {
		return err
	}
	if !ebiten.IsRunningSlowly() {
		switch g.count % 2 {
		case 0:
			if g.prevScreen == nil {
				w, h := screen.Size()
				g.prevScreen, _ = ebiten.NewImage(w, h, ebiten.FilterNearest)
			}
			g.prevScreen.Clear()
			g.draw(g.prevScreen)
			screen.DrawImage(g.prevScreen, nil)
		case 1:
			if g.prevScreen != nil {
				screen.DrawImage(g.prevScreen, nil)
			}
		}
	}
	g.count++
	return nil
}

func (g *Game) update() error {
	if g.loadingCh != nil {
		select {
		case err, ok := <-g.loadingCh:
			if err != nil {
				return err
			}
			if !ok {
				g.loadingCh = nil
			}
			d := g.loadedData
			assets.Set(d.Assets)
			g.sceneManager = scene.NewManager(g.width, g.height, g.requester, d.Game, d.Progress, d.Purchases, d.Language)
			g.sceneManager.InitScene(sceneimpl.NewTitleScene())
		default:
			return nil
		}
	}
	input.Update()
	if err := audio.Update(); err != nil {
		return err
	}
	return g.sceneManager.Update()
}

func (g *Game) draw(screen *ebiten.Image) {
	if g.loadingCh != nil {
		ebitenutil.DebugPrint(screen, "Now Loading...")
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
