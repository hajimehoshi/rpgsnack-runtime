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
	balloons      []*balloon
	tilesImage    *ebiten.Image
	events        []*event
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
	tilesImage, err := ebiten.NewImage(scene.TileXNum*scene.TileSize, scene.TileYNum*scene.TileSize, ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	mapScene := &MapScene{
		currentMap: mapData,
		player:     player,
		tilesImage: tilesImage,
	}
	for _, e := range mapScene.currentMap.Rooms[mapScene.currentRoomID].Events {
		mapScene.events = append(mapScene.events, newEvent(e))
	}
	return mapScene, nil
}

func (m *MapScene) passableTile(x, y int) bool {
	tileSet := tileSets[m.currentMap.TileSetID]
	layer := 1
	tile := m.currentMap.Rooms[m.currentRoomID].Tiles[layer][y*scene.TileXNum+x]
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
	tile = m.currentMap.Rooms[m.currentRoomID].Tiles[layer][y*scene.TileXNum+x]
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
	if scene.TileXNum <= x {
		return false
	}
	if scene.TileYNum <= y {
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

func (m *MapScene) eventAt(x, y int) *event {
	for _, e := range m.events {
		ex, ey := e.character.x, e.character.y
		if ex == x && ey == y {
			return e
		}
	}
	return nil
}

func (m *MapScene) movePlayerIfNeeded(taskLine *task.TaskLine) {
	if !input.Triggered() {
		return
	}
	x, y := input.Position()
	tx := (x - scene.GameMarginX) / scene.TileSize / scene.TileScale
	ty := (y - scene.GameMarginTop) / scene.TileSize / scene.TileScale
	e := m.eventAt(tx, ty)
	if !m.passable(tx, ty) && e == nil {
		return
	}
	m.playerMoving = true
	m.player.move(taskLine, m.passable, tx, ty)
	m.moveDstX = tx
	m.moveDstY = ty
	taskLine.Push(func() error {
		m.playerMoving = false
		return task.Terminated
	})
	if e == nil {
		return
	}
	if e.trigger() != data.TriggerActionButton {
		return
	}
	e.run(taskLine, m)
}

func (m *MapScene) Update(taskLine *task.TaskLine, sceneManager *scene.SceneManager) error {
	m.movePlayerIfNeeded(taskLine)
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
	return scene.TileXNum * scene.TileYNum
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
	x := tile % 8 * scene.TileSize
	y := tile / 8 * scene.TileSize
	return x, y, x + scene.TileSize, y + scene.TileSize
}

func (t *tilesImageParts) Dst(index int) (int, int, int, int) {
	x := index % scene.TileXNum * scene.TileSize
	y := index / scene.TileXNum * scene.TileSize
	return x, y, x + scene.TileSize, y + scene.TileSize
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
	for _, e := range m.events {
		if err := e.draw(m.tilesImage); err != nil {
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
		op.GeoM.Translate(float64(x*scene.TileSize), float64(y*scene.TileSize))
		if err := m.tilesImage.DrawImage(theImageCache.Get("marker.png"), op); err != nil {
			return err
		}
	}
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scene.TileScale, scene.TileScale)
	op.GeoM.Translate(scene.GameMarginX, scene.GameMarginTop)
	if err := screen.DrawImage(m.tilesImage, op); err != nil {
		return err
	}
	for _, b := range m.balloons {
		if err := b.draw(screen); err != nil {
			return err
		}
	}
	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	if err := font.DrawText(screen, msg, 0, 0, scene.TextScale, color.White); err != nil {
		return err
	}
	return nil
}
