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
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/assets"
	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/font"
)

type mapScene struct {
	tilesImage      *ebiten.Image
	charactersImage *ebiten.Image
	currentMap      *data.Map
}

func newMapScene() (*mapScene, error) {
	// TODO: The image should be loaded asyncly.
	tilesImage, err := assets.LoadImage("images/tiles.png", ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	charactersImage, err := assets.LoadImage("images/characters.png", ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	mapDataBytes := assets.MustAsset("data/map0.json")
	var mapData *data.Map
	if err := json.Unmarshal(mapDataBytes, &mapData); err != nil {
		return nil, err
	}
	return &mapScene{
		tilesImage:      tilesImage,
		charactersImage: charactersImage,
		currentMap:      mapData,
	}, nil
}

func (m *mapScene) Update(sceneManager *sceneManager) error {
	return nil
}

type tilesImageParts struct {
	room *data.Room
}

func (t *tilesImageParts) Len() int {
	return tileXNum * tileYNum
}

func (t *tilesImageParts) Src(index int) (int, int, int, int) {
	tile := t.room.Tiles[0][index]
	x := tile % 8 * tileSize
	y := tile / 8 * tileSize
	return x, y, x + tileSize, y + tileSize
}

func (t *tilesImageParts) Dst(index int) (int, int, int, int) {
	x := index % tileXNum * tileSize
	y := index / tileXNum * tileSize
	return x, y, x + tileSize, y + tileSize
}

type charactersImageParts struct {
}

func (c *charactersImageParts) Len() int {
	return 1
}

func (c *charactersImageParts) Src(index int) (int, int, int, int) {
	x := characterSize
	y := characterSize * 2
	return x, y, x + characterSize, y + characterSize
}

func (c *charactersImageParts) Dst(index int) (int, int, int, int) {
	return 0, 0, characterSize, characterSize
}

func (m *mapScene) Draw(screen *ebiten.Image) error {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(tileScale, tileScale)
	op.ImageParts = &tilesImageParts{
		room: m.currentMap.Rooms[0],
	}
	if err := screen.DrawImage(m.tilesImage, op); err != nil {
		return err
	}
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(tileScale, tileScale)
	op.ImageParts = &charactersImageParts{}
	if err := screen.DrawImage(m.charactersImage, op); err != nil {
		return err
	}
	if err := font.DrawText(screen, "文字の大きさはこれくらい。", 0, 0, textScale, color.White); err != nil {
		return err
	}
	return nil
}
