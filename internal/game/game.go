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
)

const (
	tileSize      = 16
	characterSize = 16
	tileXNum      = 10
	tileYNum      = 10
	textScale     = 2
	tileScale     = 3
)

const (
	gameWidth   = tileXNum * tileSize * tileScale
	gameHeight  = tileYNum * tileSize * tileScale
	gameMarginX = 0
	gameMarginY = 2.5 * tileSize * tileScale
)

// TODO: This variable should belong to a struct.
var (
	tileSets []*data.TileSet
)

type Game struct {
	sceneManager *sceneManager
}

func New() (*Game, error) {
	initScene := &titleScene{}
	game := &Game{
		sceneManager: newSceneManager(initScene),
	}
	mapDataBytes := assets.MustAsset("data/tilesets.json")
	if err := json.Unmarshal(mapDataBytes, &tileSets); err != nil {
		return nil, err
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
	const w = gameWidth + 2*gameMarginX
	const h = gameHeight + 2*gameMarginY
	return w, h
}
