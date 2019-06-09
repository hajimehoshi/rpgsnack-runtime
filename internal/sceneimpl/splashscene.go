// Copyright 2019 Hajime Hoshi
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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type Scene interface {
	Update(sceneManager *scene.Manager) error
	Draw(screen *ebiten.Image)
	Resize(width, height int)
}

func NewInitialScene(sceneManager *scene.Manager) (Scene, error) {
	if assets.ImageExists("system/splash/info") {
		return NewSplashScene(), nil
	}
	g, err := savedGame(sceneManager)
	if err != nil {
		return nil, err
	}
	return NewTitleMapScene(sceneManager, g), nil
}

const FadingCount = 30

type SplashScene struct {
	count int
}

func NewSplashScene() *SplashScene {
	const countMax = 180
	return &SplashScene{
		count: countMax,
	}
}

func savedGame(sceneManager *scene.Manager) (*gamestate.Game, error) {
	if sceneManager.HasProgress() {
		var savedGame *gamestate.Game
		if err := msgpack.Unmarshal(sceneManager.Progress(), &savedGame); err != nil {
			return nil, err
		}
		return savedGame, nil
	}
	return nil, nil
}

func (s *SplashScene) Update(sceneManager *scene.Manager) error {
	g, err := savedGame(sceneManager)
	if err != nil {
		return err
	}
	s.count--
	if input.Triggered() {
		s.count = 0
	}

	if s.count == 0 {
		sceneManager.GoToWithFading(NewTitleMapScene(sceneManager, g), FadingCount, FadingCount)
	}
	return nil
}

func (s *SplashScene) Draw(screen *ebiten.Image) {
	img := assets.GetLocalizedImage("system/splash/info")
	sw, sh := img.Size()
	dw, dh := screen.Size()

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(sw)/2, -float64(sh)/2)
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	op.GeoM.Translate(float64(dw)/2, float64(dh)/2)
	screen.DrawImage(img, op)
}

func (s *SplashScene) Resize(width, height int) {
}
