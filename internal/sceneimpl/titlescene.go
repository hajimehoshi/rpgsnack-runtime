// Copyright 2018 Hajime Hoshi
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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type TitleScene struct {
	view *ui.TitleView
}

type sceneMaker struct {}

func (s *sceneMaker) NewMapScene() scene.Scene {
	return NewMapScene()
}

func (s *sceneMaker) NewMapSceneWithGame(game *gamestate.Game) scene.Scene {
	return NewMapSceneWithGame(game)
}

func (s *sceneMaker) NewSettingsScene() scene.Scene {
	return NewSettingsScene()
}

func NewTitleScene() *TitleScene {
	return &TitleScene{
		view: ui.NewTitleView(&sceneMaker{}),
	}
}

func (t *TitleScene) Update(sceneManager *scene.Manager) error {
	return t.view.Update(sceneManager)
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	t.view.Draw(screen)
}

func (t *TitleScene) Resize() {
	t.view.Resize()
}
