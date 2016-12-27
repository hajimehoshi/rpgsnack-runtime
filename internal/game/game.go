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

	"github.com/hajimehoshi/tsugunai/internal/assets"
	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/input"
	"github.com/hajimehoshi/tsugunai/internal/scene"
	"github.com/hajimehoshi/tsugunai/internal/titlescene"
)

type Game struct {
	sceneManager *scene.SceneManager
	gameData     *data.Game
	loadingCh    chan error
}

func New() (*Game, error) {
	g := &Game{}
	g.startLoadingGameData()
	return g, nil
}

func (g *Game) startLoadingGameData() {
	ch := make(chan error)
	go func() {
		defer close(ch)
		gameData, err := data.Load()
		if err != nil {
			ch <- err
			return
		}
		g.gameData = gameData
		initScene := titlescene.New(gameData)
		g.sceneManager = scene.NewSceneManager(initScene)
	}()
	g.loadingCh = ch
}

func (g *Game) Update() error {
	if assets.IsLoading() {
		return nil
	}
	if g.loadingCh != nil {
		select {
		case err, ok := <-g.loadingCh:
			if err != nil {
				return err
			}
			if !ok {
				g.loadingCh = nil
			}
		default:
			return nil
		}
	}
	input.Update()
	return g.sceneManager.Update()
}

func (g *Game) Draw(screen *ebiten.Image) error {
	if assets.IsLoading() || g.loadingCh != nil {
		if err := ebitenutil.DebugPrint(screen, "Now Loading..."); err != nil {
			return err
		}
		return nil
	}
	return g.sceneManager.Draw(screen)
}

func (g *Game) Title() string {
	return "Clock of Atonement"
}

func (g *Game) Size() (int, int) {
	w := scene.TileXNum * scene.TileSize * scene.TileScale
	h := scene.TileYNum * scene.TileSize * scene.TileScale
	w += 2 * scene.GameMarginX
	h += scene.GameMarginTop + scene.GameMarginBottom
	return w, h
}
