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
	currentRoomID int
	currentMap    *data.Map
	player        *player
	moveDstX      int
	moveDstY      int
}

func newMapScene() (*mapScene, error) {
	mapDataBytes := assets.MustAsset("data/map0.json")
	var mapData *data.Map
	if err := json.Unmarshal(mapDataBytes, &mapData); err != nil {
		return nil, err
	}
	player, err := newPlayer(1, 2)
	if err != nil {
		return nil, err
	}
	return &mapScene{
		currentMap: mapData,
		player:     player,
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
	layer := 1
	tile := m.currentMap.Rooms[m.currentRoomID].Tiles[layer][y*tileXNum+x]
	switch tileSet.PassageTypes[layer][tile] {
	case data.PassageTypeBlock:
		return false
	case data.PassageTypePassable:
		return true
	case data.PassageTypeWall:
		panic("not implemented")
	case data.PassageTypeOver:
	default:
		panic("not reach")
	}
	layer = 0
	tile = m.currentMap.Rooms[m.currentRoomID].Tiles[layer][y*tileXNum+x]
	if tileSet.PassageTypes[layer][tile] == data.PassageTypePassable {
		return true
	}
	return false
}

func (m *mapScene) Update(sceneManager *sceneManager) error {
	if input.Triggered() {
		if !m.player.isMoving() {
			x, y := input.Position()
			tx := x / tileSize / tileScale
			ty := y / tileSize / tileScale
			m.player.move(m.passable, tx, ty)
			m.moveDstX = tx
			m.moveDstY = ty
		}
	}
	if err := m.player.update(); err != nil {
		return err
	}
	return nil
}

type tilesImageParts struct {
	room     *data.Room
	tileSet  *data.TileSet
	layer    int
	overOnly bool
}

func (t *tilesImageParts) Len() int {
	return tileXNum * tileYNum
}

func (t *tilesImageParts) Src(index int) (int, int, int, int) {
	tile := t.room.Tiles[t.layer][index]
	if t.layer == 1 {
		p := t.tileSet.PassageTypes[t.layer][tile]
		if !t.overOnly && p == data.PassageTypeOver {
			return 0, 0, 0, 0
		}
		if t.overOnly && p != data.PassageTypeOver {
			return 0, 0, 0, 0
		}
	}
	// TODO: 8 is a magic number and should be replaced.
	x := tile % 8 * tileSize
	y := tile / 8 * tileSize
	return x, y, x + tileSize, y + tileSize
}

func (t *tilesImageParts) Dst(index int) (int, int, int, int) {
	x := index % tileXNum * tileSize
	y := index / tileXNum * tileSize
	return x, y, x + tileSize, y + tileSize
}

type characterImageParts struct {
	charWidth  int
	charHeight int
	index      int
	dir        data.Dir
}

func (c *characterImageParts) Len() int {
	return 1
}

func (c *characterImageParts) Src(index int) (int, int, int, int) {
	x := ((c.index%4)*3 + 1) * c.charWidth
	y := (c.index / 4) * 2 * c.charHeight
	switch c.dir {
	case data.DirUp:
	case data.DirRight:
		y += c.charHeight
	case data.DirDown:
		y += 2 * c.charHeight
	case data.DirLeft:
		y += 3 * c.charHeight
	}
	return x, y, x + c.charWidth, y + c.charHeight
}

func (c *characterImageParts) Dst(index int) (int, int, int, int) {
	return 0, 0, c.charWidth, c.charHeight
}

func (m *mapScene) Draw(screen *ebiten.Image) error {
	tileset := tileSets[m.currentMap.TileSetID]
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(tileScale, tileScale)
	op.ImageParts = &tilesImageParts{
		room:    m.currentMap.Rooms[m.currentRoomID],
		tileSet: tileset,
		layer:   0,
	}
	if err := screen.DrawImage(theImageCache.Get(tileset.Images[0]), op); err != nil {
		return err
	}
	op.ImageParts = &tilesImageParts{
		room:     m.currentMap.Rooms[m.currentRoomID],
		tileSet:  tileset,
		layer:    1,
		overOnly: false,
	}
	if err := screen.DrawImage(theImageCache.Get(tileset.Images[1]), op); err != nil {
		return err
	}
	if err := m.player.draw(screen); err != nil {
		return err
	}
	room := m.currentMap.Rooms[m.currentRoomID]
	for _, e := range room.Events {
		page := e.Pages[0]
		image := theImageCache.Get(page.Image)
		imageW, imageH := image.Size()
		charW := imageW / 4 / 3
		charH := imageH / 2 / 4
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(e.X*tileSize+tileSize/2), float64((e.Y+1)*tileSize))
		op.GeoM.Translate(float64(-charW/2), float64(-charH))
		op.GeoM.Scale(tileScale, tileScale)
		op.ImageParts = &characterImageParts{
			charWidth:  charW,
			charHeight: charH,
			index:      page.ImageIndex,
			dir:        page.Dir,
		}
		if err := screen.DrawImage(image, op); err != nil {
			return err
		}
	}
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(tileScale, tileScale)
	op.ImageParts = &tilesImageParts{
		room:     m.currentMap.Rooms[m.currentRoomID],
		tileSet:  tileSets[m.currentMap.TileSetID],
		layer:    1,
		overOnly: true,
	}
	if err := screen.DrawImage(theImageCache.Get(tileset.Images[1]), op); err != nil {
		return err
	}
	if m.player.isMoving() {
		x, y := m.moveDstX, m.moveDstY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))
		op.GeoM.Scale(tileScale, tileScale)
		if err := screen.DrawImage(theImageCache.Get("marker.png"), op); err != nil {
			return err
		}
	}
	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	if err := font.DrawText(screen, msg, 0, 0, textScale, color.White); err != nil {
		return err
	}
	return nil
}
