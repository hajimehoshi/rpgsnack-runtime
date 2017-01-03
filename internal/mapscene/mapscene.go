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

type tint struct {
	red   int
	green int
	blue  int
	gray  int
}

type MapScene struct {
	gameData      *data.Game
	currentRoomID int
	currentMap    *data.Map
	player        *player
	moveDstX      int
	moveDstY      int
	playerMoving  bool
	balloons      []*balloon
	tilesImage    *ebiten.Image
	emptyImage    *ebiten.Image
	events        []*event
	switches      []bool
	fadingRate    float64
	tint          *tint
}

func New(gameData *data.Game) (*MapScene, error) {
	pos := gameData.System.InitialPosition
	x, y := 0, 0
	if pos != nil {
		x, y = pos.X, pos.Y
	}
	player, err := newPlayer(x, y)
	if err != nil {
		return nil, err
	}
	tilesImage, err := ebiten.NewImage(scene.TileXNum*scene.TileSize, scene.TileYNum*scene.TileSize, ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	emptyImage, err := ebiten.NewImage(16, 16, ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	mapScene := &MapScene{
		gameData:   gameData,
		currentMap: gameData.Maps[0],
		player:     player,
		tilesImage: tilesImage,
		emptyImage: emptyImage,
	}
	for _, e := range mapScene.currentMap.Rooms[mapScene.currentRoomID].Events {
		event, err := newEvent(e, mapScene)
		if err != nil {
			return nil, err
		}
		mapScene.events = append(mapScene.events, event)
	}
	return mapScene, nil
}

func (m *MapScene) tileSet(id int) (*data.TileSet, error) {
	for _, t := range m.gameData.TileSets {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("mapscene: tile set not found: %d", id)
}

func (m *MapScene) passableTile(x, y int) (bool, error) {
	tileSet, err := m.tileSet(m.currentMap.TileSetID)
	if err != nil {
		return false, err
	}
	layer := 1
	tile := m.currentMap.Rooms[m.currentRoomID].Tiles[layer][y*scene.TileXNum+x]
	switch tileSet.PassageTypes[layer][tile] {
	case data.PassageTypeBlock:
		return false, nil
	case data.PassageTypePassable:
		return true, nil
	case data.PassageTypeWall:
		panic("not implemented")
	case data.PassageTypeOver:
	default:
		panic("not reach")
	}
	layer = 0
	tile = m.currentMap.Rooms[m.currentRoomID].Tiles[layer][y*scene.TileXNum+x]
	if tileSet.PassageTypes[layer][tile] == data.PassageTypePassable {
		return true, nil
	}
	return false, nil
}

func (m *MapScene) passable(x, y int) (bool, error) {
	if x < 0 {
		return false, nil
	}
	if y < 0 {
		return false, nil
	}
	if scene.TileXNum <= x {
		return false, nil
	}
	if scene.TileYNum <= y {
		return false, nil
	}
	p, err := m.passableTile(x, y)
	if err != nil {
		return false, err
	}
	if !p {
		return false, nil
	}
	e := m.eventAt(x, y)
	if e == nil {
		return true, nil
	}
	return e.isPassable(), nil
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

func (m *MapScene) movePlayerIfNeeded(taskLine *task.TaskLine) error {
	if !input.Triggered() {
		return nil
	}
	x, y := input.Position()
	tx := (x - scene.GameMarginX) / scene.TileSize / scene.TileScale
	ty := (y - scene.GameMarginTop) / scene.TileSize / scene.TileScale
	e := m.eventAt(tx, ty)
	p, err := m.passable(tx, ty)
	if err != nil {
		return err
	}
	if !p {
		if e == nil {
			return nil
		}
		if !e.isRunnable() {
			return nil
		}
	}
	m.playerMoving = true
	if err := m.player.move(taskLine, m.passable, tx, ty); err != nil {
		return err
	}
	m.moveDstX = tx
	m.moveDstY = ty
	taskLine.PushFunc(func() error {
		m.playerMoving = false
		return task.Terminated
	})
	if e == nil {
		return nil
	}
	e.run(taskLine, data.TriggerPlayer)
	return nil
}

func (m *MapScene) Update(subTasksUpdated bool, taskLine *task.TaskLine, sceneManager *scene.SceneManager) error {
	for _, e := range m.events {
		if err := e.updateCharacterIfNeeded(); err != nil {
			return err
		}
	}
	if subTasksUpdated {
		return nil
	}
	for _, e := range m.events {
		if e.run(taskLine, data.TriggerAuto) {
			return nil
		}
	}
	if err := m.movePlayerIfNeeded(taskLine); err != nil {
		return err
	}
	if err := m.player.update(m.passable); err != nil {
		return err
	}
	for _, e := range m.events {
		if err := e.update(); err != nil {
			return err
		}
	}
	return nil
}

func (m *MapScene) removeBalloon(balloon *balloon) {
	index := -1
	for i, b := range m.balloons {
		if b == balloon {
			index = i
			break
		}
	}
	if index != -1 {
		m.balloons[index] = nil
	}
}

func (m *MapScene) setTint(red, green, blue, gray int) {
	m.tint = &tint{
		red:   red,
		green: green,
		blue:  blue,
		gray:  gray,
	}
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
	tileset, err := m.tileSet(m.currentMap.TileSetID)
	if err != nil {
		return err
	}
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &tilesImageParts{
		room:    m.currentMap.Rooms[m.currentRoomID],
		tileSet: tileset,
		layer:   0,
	}
	if err := m.tilesImage.DrawImage(assets.GetImage(tileset.Images[0]), op); err != nil {
		return err
	}
	op.ImageParts = &tilesImageParts{
		room:     m.currentMap.Rooms[m.currentRoomID],
		tileSet:  tileset,
		layer:    1,
		overOnly: false,
	}
	if err := m.tilesImage.DrawImage(assets.GetImage(tileset.Images[1]), op); err != nil {
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
	tileSet, err := m.tileSet(m.currentMap.TileSetID)
	if err != nil {
		return err
	}
	op.ImageParts = &tilesImageParts{
		room:     m.currentMap.Rooms[m.currentRoomID],
		tileSet:  tileSet,
		layer:    1,
		overOnly: true,
	}
	if err := m.tilesImage.DrawImage(assets.GetImage(tileset.Images[1]), op); err != nil {
		return err
	}
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scene.TileScale, scene.TileScale)
	op.GeoM.Translate(scene.GameMarginX, scene.GameMarginTop)
	if m.tint != nil {
		if m.tint.gray != 0 {
			op.ColorM.ChangeHSV(0, float64(255-m.tint.gray)/255, 1)
		}
		rs, gs, bs := 1.0, 1.0, 1.0
		if m.tint.red < 0 {
			rs = float64(255 - -m.tint.red) / 255
		}
		if m.tint.green < 0 {
			gs = float64(255 - -m.tint.green) / 255
		}
		if m.tint.blue < 0 {
			bs = float64(255 - -m.tint.blue) / 255
		}
		op.ColorM.Scale(rs, gs, bs, 1)
		rt, gt, bt := 0.0, 0.0, 0.0
		if m.tint.red > 0 {
			rt = float64(m.tint.red) / 255
		}
		if m.tint.green > 0 {
			gt = float64(m.tint.green) / 255
		}
		if m.tint.blue > 0 {
			bt = float64(m.tint.blue) / 255
		}
		op.ColorM.Translate(rt, gt, bt, 0)
	}
	if err := screen.DrawImage(m.tilesImage, op); err != nil {
		return err
	}
	if m.playerMoving {
		x, y := m.moveDstX, m.moveDstY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*scene.TileSize), float64(y*scene.TileSize))
		op.GeoM.Scale(scene.TileScale, scene.TileScale)
		op.GeoM.Translate(scene.GameMarginX, scene.GameMarginTop)
		if err := screen.DrawImage(assets.GetImage("marker.png"), op); err != nil {
			return err
		}
	}
	if 0 < m.fadingRate {
		w, h := scene.TileXNum*scene.TileSize, scene.TileYNum*scene.TileSize
		ew, eh := m.emptyImage.Size()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(w)/float64(ew), float64(h)/float64(eh))
		op.GeoM.Scale(scene.TileScale, scene.TileScale)
		op.GeoM.Translate(scene.GameMarginX, scene.GameMarginTop)
		op.ColorM.Translate(0, 0, 0, m.fadingRate)
		if err := screen.DrawImage(m.emptyImage, op); err != nil {
			return err
		}
	}
	for _, b := range m.balloons {
		if b == nil {
			continue
		}
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
