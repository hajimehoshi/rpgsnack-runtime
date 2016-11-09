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
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/assets"
	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/font"
	"github.com/hajimehoshi/tsugunai/internal/input"
)

type mapScene struct {
	tilesBottomImage *ebiten.Image
	tilesTopImage    *ebiten.Image
	currentRoomID    int
	currentMap       *data.Map
	player           *player
}

func newMapScene() (*mapScene, error) {
	mapDataBytes := assets.MustAsset("data/map0.json")
	var mapData *data.Map
	if err := json.Unmarshal(mapDataBytes, &mapData); err != nil {
		return nil, err
	}
	// TODO: The image should be loaded asyncly.
	tileSet := tileSets[mapData.TileSetID]
	tilesBottomImage, err := assets.LoadImage("images/"+tileSet.Images[0], ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	tilesTopImage, err := assets.LoadImage("images/"+tileSet.Images[1], ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	player, err := newPlayer(0, 2)
	if err != nil {
		return nil, err
	}
	return &mapScene{
		tilesBottomImage: tilesBottomImage,
		tilesTopImage:    tilesTopImage,
		currentMap:       mapData,
		player:           player,
	}, nil
}

func (m *mapScene) passable(x, y int) bool {
	if x < 0 {
		return false
	}
	if y < 0 {
		return false
	}
	if tileXNum <= x {
		return false
	}
	if tileYNum <= y {
		return false
	}
	tileSet := tileSets[m.currentMap.TileSetID]
	for _, layer := range []int{1, 0} {
		tile := m.currentMap.Rooms[m.currentRoomID].Tiles[layer][y*tileXNum+x]
		if tileSet.PassageTypes[layer][tile] != data.PassageTypePassable {
			return false
		}
	}
	return true
}

func (m *mapScene) Update(sceneManager *sceneManager) error {
	if input.Triggered() {
		if !m.player.isMoving() {
			x, y := input.Position()
			tx := x / tileSize / tileScale
			ty := y / tileSize / tileScale
			m.player.move(m.passable, tx, ty)
		}
	}
	if err := m.player.update(); err != nil {
		return err
	}
	return nil
}

type tilesImageParts struct {
	room *data.Room
}

func (t *tilesImageParts) Len() int {
	return tileXNum * tileYNum
}

func (t *tilesImageParts) Src(index int) (int, int, int, int) {
	// TODO: 8 is a magic number and should be replaced.
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

func (m *mapScene) Draw(screen *ebiten.Image) error {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(tileScale, tileScale)
	op.ImageParts = &tilesImageParts{
		room: m.currentMap.Rooms[m.currentRoomID],
	}
	if err := screen.DrawImage(m.tilesBottomImage, op); err != nil {
		return err
	}
	if err := m.player.draw(screen); err != nil {
		return err
	}
	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	if err := font.DrawText(screen, msg, 0, 0, textScale, color.White); err != nil {
		return err
	}
	return nil
}
