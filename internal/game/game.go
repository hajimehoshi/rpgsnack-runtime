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
	sceneManager *scene.Manager
	loadingCh    chan error
}

func New(width, height int, requester scene.Requester) (*Game, error) {
	g := &Game{}
	g.loadGameData()
	initScene := sceneimpl.NewTitleScene()
	g.sceneManager = scene.NewManager(width, height, requester, initScene)
	return g, nil
}

func NewWithMockRequester(width, height int) (*Game, error) {
	g := &Game{}
	g.loadGameData()
	initScene := sceneimpl.NewTitleScene()
	g.sceneManager = scene.NewManager(width, height, &MockRequester{g}, initScene)
	return g, nil
}

func (g *Game) loadGameData() {
	ch := make(chan error)
	go func() {
		defer close(ch)
		if err := data.Load(); err != nil {
			ch <- err
			return
		}
	}()
	g.loadingCh = ch
}

func (g *Game) Update(screen *ebiten.Image) error {
	if err := g.update(); err != nil {
		return err
	}
	if ebiten.IsRunningSlowly() {
		return nil
	}
	if err := g.draw(screen); err != nil {
		return err
	}
	return nil

}

func (g *Game) update() error {
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
	if err := audio.Update(); err != nil {
		return err
	}
	return g.sceneManager.Update()
}

func (g *Game) draw(screen *ebiten.Image) error {
	if assets.IsLoading() || g.loadingCh != nil {
		if err := ebitenutil.DebugPrint(screen, "Now Loading..."); err != nil {
			return err
		}
		return nil
	}
	return g.sceneManager.Draw(screen)
}

func Title() string {
	return "Clock of Atonement"
}

func (g *Game) Size() (int, int) {
	return g.sceneManager.Size()
}

func (g *Game) FinishUnlockAchievement(id int, achievements string, err string) {
	g.sceneManager.FinishUnlockAchievement(id, achievements, err)
}

func (g *Game) FinishSaveProgress(id int, err string) {
	g.sceneManager.FinishSaveProgress(id, err)
}

func (g *Game) FinishLoadProgress(id int, data string, err string) {
	g.sceneManager.FinishLoadProgress(id, data, err)
}
