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
	"flag"
	"fmt"
	"image/color"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/sceneimpl"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/screenshot"
)

type Game struct {
	projectLocation   string
	width             int
	height            int
	requester         scene.Requester
	sceneManager      *scene.Manager
	loadProgressCh    chan data.LoadProgress
	loadProgressRate  float64
	setPlatformDataCh chan setPlatformDataArgs
	langs             []language.Tag
	screenshot        *screenshot.Screenshot
	screenshotDir     string

	pseudoScreen *ebiten.Image
	frozenScreen *ebiten.Image
}

type setPlatformDataArgs struct {
	key   scene.PlatformDataKey
	value string
}

func New(width, height int, requester scene.Requester) *Game {
	g := &Game{
		width:             width,
		height:            height,
		requester:         requester,
		setPlatformDataCh: make(chan setPlatformDataArgs, 1),
	}
	g.loadGameData()
	return g
}

// Rewrite this by specifying -ldflags='-X github.com/hajimehoshi/rpgsnack-runtime/internal/game.injectedProjectLocation=<project path>'
var injectedProjectLocation = ""

func projectLocation() string {
	if injectedProjectLocation != "" {
		return injectedProjectLocation
	}
	if flag.Arg(0) != "" {
		return flag.Arg(0)
	}
	if js.Global != nil {
		href := js.Global.Get("window").Get("location").Get("href").String()
		u, err := url.Parse(href)
		if err != nil {
			panic(err)
		}
		vals := u.Query()["project_location"]
		if len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}

func NewWithDefaultRequester(width, height int) (*Game, error) {
	p := projectLocation()

	g := &Game{
		projectLocation:   p,
		width:             width,
		height:            height,
		setPlatformDataCh: make(chan setPlatformDataArgs, 1),
	}
	g.loadGameData()
	g.requester = newRequester(g)
	return g, nil
}

func (g *Game) SetPseudoScreen(screen *ebiten.Image) {
	if g.loadProgressCh != nil {
		panic("game: a pseudo screen is not available during loading")
	}

	if g.frozenScreen == nil {
		g.frozenScreen, _ = ebiten.NewImage(g.width, g.height, ebiten.FilterDefault)
		g.sceneManager.Draw(g.frozenScreen)
	}

	g.pseudoScreen = screen
	g.sceneManager.SetScreenSize(g.pseudoScreen.Size())
}

func (g *Game) ResetPseudoScreen() {
	g.pseudoScreen = nil
	g.frozenScreen.Dispose()
	g.frozenScreen = nil
	g.sceneManager.SetScreenSize(g.width, g.height)
}

func (g *Game) SetScreenSize(width, height int, scale float64) {
	g.width = width
	g.height = height
	ebiten.SetScreenSize(width, height)
	ebiten.SetScreenScale(scale)
	if g.sceneManager != nil {
		g.sceneManager.SetScreenSize(width, height)
	}
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
	if ebiten.IsDrawingSkipped() {
		return nil
	}
	if err := g.draw(screen); err != nil {
		return err
	}
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
			g.sceneManager = scene.NewManager(g.width, g.height, g.requester, da.Game, da.Progress, da.Permanent, da.Purchases, sceneimpl.FadingCount)
			g.sceneManager.SetLanguage(da.Language)
			s, err := sceneimpl.NewInitialScene(g.sceneManager)
			if err != nil {
				return err
			}
			g.sceneManager.InitScene(s)
			g.langs = da.Game.Texts.Languages()
		default:
			return nil
		}
	}

	select {
	case a := <-g.setPlatformDataCh:
		g.sceneManager.SetPlatformData(a.key, a.value)
	default:
	}

	if err := audio.Update(); err != nil {
		return err
	}
	takeCPUProfileIfAvailable()

	if g.screenshot == nil && input.IsScreenshotButtonTriggered() {
		sizes := []screenshot.Size{
			{
				Width:  480, // 1242
				Height: 720, // 2208
			},
			{
				Width:  480, // 2048
				Height: 854, // 2732
			},
			{
				Width:  480,  // 1125
				Height: 1040, // 2436
			},
		}
		g.screenshot = screenshot.New(sizes, g.langs)
		g.screenshotDir = filepath.Join("screenshots", time.Now().Format("20060102_030405"))
	}
	if g.screenshot != nil {
		g.screenshot.Update(g, g.sceneManager)
		if g.screenshot.IsFinished() {
			g.screenshot = nil
		}
	}

	if err := g.sceneManager.Update(); err != nil {
		return err
	}
	return nil
}

func (g *Game) draw(screen *ebiten.Image) error {
	if g.loadProgressCh != nil {
		if runtime.GOARCH == "js" {
			const barHeight = 8
			w, h := screen.Size()
			barWidth := float64(w)
			y := float64(h-barHeight) / 2
			ebitenutil.DrawRect(screen, 0, y, barWidth, barHeight, color.RGBA{0x80, 0x80, 0x80, 0x80})
			activeWidth := barWidth * g.loadProgressRate
			ebitenutil.DrawRect(screen, 0, y, activeWidth, barHeight, color.RGBA{0xff, 0xff, 0xff, 0xff})
		}
		return nil
	}
	if g.pseudoScreen != nil {
		g.pseudoScreen.Clear()
		g.sceneManager.Draw(g.pseudoScreen)
		screen.DrawImage(g.frozenScreen, nil)
	} else {
		g.sceneManager.Draw(screen)
	}

	if g.screenshot != nil {
		img, size, lang, err := g.screenshot.TryDump()
		if err != nil {
			return err
		}
		if img != nil {
			if err := os.MkdirAll(g.screenshotDir, 0755); err != nil {
				return err
			}
			fn := filepath.Join(g.screenshotDir, fmt.Sprintf("%d-%d-%s.png", size.Width, size.Height, lang))
			fmt.Println(fn)
			ioutil.WriteFile(fn, img, 0666)
		}
	}
	return nil
}

func (g *Game) RespondUnlockAchievement(id int) {
	g.sceneManager.RespondUnlockAchievement(id)
}

func (g *Game) RespondSaveProgress(id int) {
	g.sceneManager.RespondSaveProgress(id)
}

func (g *Game) RespondSavePermanent(id int) {
	g.sceneManager.RespondSavePermanent(id)
}

func (g *Game) RespondPurchase(id int, success bool, purchases []uint8) {
	g.sceneManager.RespondPurchase(id, success, purchases)
}

func (g *Game) RespondShowShop(id int, success bool, purchases []uint8) {
	g.sceneManager.RespondShowShop(id, success, purchases)
}

func (g *Game) RespondRestorePurchases(id int, success bool, purchases []uint8) {
	g.sceneManager.RespondRestorePurchases(id, success, purchases)
}

func (g *Game) RespondInterstitialAds(id int, success bool) {
	g.sceneManager.RespondInterstitialAds(id, success)
}

func (g *Game) RespondRewardedAds(id int, success bool) {
	g.sceneManager.RespondRewardedAds(id, success)
}

func (g *Game) RespondOpenLink(id int) {
	g.sceneManager.RespondOpenLink(id)
}

func (g *Game) RespondShareImage(id int) {
	g.sceneManager.RespondShareImage(id)
}

func (g *Game) RespondChangeLanguage(id int) {
	g.sceneManager.RespondChangeLanguage(id)
}

func (g *Game) RespondAsset(id int, success bool, data []byte) {
	g.sceneManager.RespondAsset(id, success, data)
}

func (g *Game) SetPlatformData(key scene.PlatformDataKey, value string) {
	args := setPlatformDataArgs{
		key:   key,
		value: value,
	}
	go func() {
		g.setPlatformDataCh <- args
	}()
}
