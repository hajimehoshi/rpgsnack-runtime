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
	"encoding/json"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/assets"
	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/input"
	"github.com/hajimehoshi/tsugunai/internal/scene"
	"github.com/hajimehoshi/tsugunai/internal/titlescene"
)

type Game struct {
	sceneManager *scene.SceneManager
	gameData     *data.Game
}

func New() (*Game, error) {
	mapDataJson := assets.MustAsset("data/map0.json")
	var mapData *data.Map
	if err := json.Unmarshal(mapDataJson, &mapData); err != nil {
		return nil, err
	}
	textsJson := assets.MustAsset("data/texts.json")
	var texts *data.Texts
	if err := json.Unmarshal(textsJson, &texts); err != nil {
		return nil, err
	}
	tileSetsJson := assets.MustAsset("data/tilesets.json")
	var tileSets []*data.TileSet
	if err := json.Unmarshal(tileSetsJson, &tileSets); err != nil {
		panic(err)
	}
	gameData := &data.Game{
		Maps:     []*data.Map{mapData},
		Texts:    texts,
		TileSets: tileSets,
	}
	initScene := titlescene.New(gameData)
	game := &Game{
		sceneManager: scene.NewSceneManager(initScene),
		gameData:     gameData,
	}
	return game, nil
}

func (g *Game) Update() error {
	input.Update()
	return g.sceneManager.Update()
}

func (g *Game) Draw(screen *ebiten.Image) error {
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
