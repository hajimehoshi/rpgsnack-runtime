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

package mapscene

import (
	"encoding/json"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/assets"
	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/font"
	"github.com/hajimehoshi/tsugunai/internal/input"
	"github.com/hajimehoshi/tsugunai/internal/scene"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

type MapScene struct {
	currentRoomID int
	currentMap    *data.Map
	player        *player
	moveDstX      int
	moveDstY      int
	playerMoving  bool
	balloon       *balloon
	tilesImage    *ebiten.Image
}

func New() (*MapScene, error) {
	mapDataBytes := assets.MustAsset("data/map0.json")
	var mapData *data.Map
	if err := json.Unmarshal(mapDataBytes, &mapData); err != nil {
		return nil, err
	}
	player, err := newPlayer(1, 2)
	if err != nil {
		return nil, err
	}
	tilesImage, err := ebiten.NewImage(tileXNum * tileSize, tileYNum * tileSize, ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	return &MapScene{
		currentMap: mapData,
		player:     player,
		balloon:    &balloon{},
		tilesImage: tilesImage,
	}, nil
}

func (m *MapScene) passableTile(x, y int) bool {
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

func (m *MapScene) passable(x, y int) bool {
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
	if !m.passableTile(x, y) {
		return false
	}
	if m.eventAt(x, y) != nil {
		return false
	}
	return true
}

func (m *MapScene) eventAt(x, y int) *data.Event {
	// TODO: Fix this when an event starts to move
	room := m.currentMap.Rooms[m.currentRoomID]
	for _, e := range room.Events {
		if e.X == x && e.Y == y {
			return e
		}
	}
	return nil
}

func (m *MapScene) runEvent(event *data.Event) {
	page := event.Pages[0]
	// TODO: Consider branches
	for _, c := range page.Commands {
		c := c
		switch c.Command {
		case "show_message":
			task.Push(func() error {
				x := event.X*tileSize + tileSize/2
				y := event.Y * tileSize
				m.balloon.show(x, y, c.Args["content"])
				return task.Terminated
			})
		}
	}
}

func (m *MapScene) Update(sceneManager *scene.SceneManager) error {
	if input.Triggered() {
		x, y := input.Position()
		tx := (x - gameMarginX) / tileSize / tileScale
		ty := (y - gameMarginY) / tileSize / tileScale
		e := m.eventAt(tx, ty)
		if m.passable(tx, ty) || e != nil {
			m.playerMoving = true
			m.player.move(m.passable, tx, ty)
			m.moveDstX = tx
			m.moveDstY = ty
			task.Push(func() error {
				m.playerMoving = false
				return task.Terminated
			})
			if e != nil && e.Pages[0].Trigger == data.TriggerActionButton {
				m.runEvent(e)
			}
		}
	}
	if err := m.player.update(m.passable); err != nil {
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

func (m *MapScene) Draw(screen *ebiten.Image) error {
	m.tilesImage.Clear()
	tileset := tileSets[m.currentMap.TileSetID]
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &tilesImageParts{
		room:    m.currentMap.Rooms[m.currentRoomID],
		tileSet: tileset,
		layer:   0,
	}
	if err := m.tilesImage.DrawImage(theImageCache.Get(tileset.Images[0]), op); err != nil {
		return err
	}
	op.ImageParts = &tilesImageParts{
		room:     m.currentMap.Rooms[m.currentRoomID],
		tileSet:  tileset,
		layer:    1,
		overOnly: false,
	}
	if err := m.tilesImage.DrawImage(theImageCache.Get(tileset.Images[1]), op); err != nil {
		return err
	}
	if err := m.player.draw(m.tilesImage); err != nil {
		return err
	}
	room := m.currentMap.Rooms[m.currentRoomID]
	for _, e := range room.Events {
		page := e.Pages[0]
		image := theImageCache.Get(page.Image)
		c := &character{
			image:      image,
			imageIndex: page.ImageIndex,
			dir:        page.Dir,
			attitude:   attitudeMiddle,
			x:          e.X,
			y:          e.Y,
		}
		if err := c.draw(m.tilesImage); err != nil {
			return err
		}
	}
	op = &ebiten.DrawImageOptions{}
	op.ImageParts = &tilesImageParts{
		room:     m.currentMap.Rooms[m.currentRoomID],
		tileSet:  tileSets[m.currentMap.TileSetID],
		layer:    1,
		overOnly: true,
	}
	if err := m.tilesImage.DrawImage(theImageCache.Get(tileset.Images[1]), op); err != nil {
		return err
	}
	if m.playerMoving {
		x, y := m.moveDstX, m.moveDstY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))
		if err := m.tilesImage.DrawImage(theImageCache.Get("marker.png"), op); err != nil {
			return err
		}
	}
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(tileScale, tileScale)
	op.GeoM.Translate(gameMarginX, gameMarginY)
	if err := screen.DrawImage(m.tilesImage, op); err != nil {
		return err
	}
	if err := m.balloon.draw(screen); err != nil {
		return err
	}
	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	if err := font.DrawText(screen, msg, 0, 0, scene.TextScale, color.White); err != nil {
		return err
	}
	return nil
}
