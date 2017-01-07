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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type MapScene struct {
	gameState     *gamestate.Game
	currentMapID  int
	currentRoomID int
	player        *player
	moveDstX      int
	moveDstY      int
	balloons      *balloons
	tilesImage    *ebiten.Image
	events        []*event
}

func New() (*MapScene, error) {
	pos := data.Current().System.InitialPosition
	x, y, roomID := 0, 0, 1
	if pos != nil {
		x, y, roomID = pos.X, pos.Y, pos.RoomID
	}
	player, err := newPlayer(x, y)
	if err != nil {
		return nil, err
	}
	tilesImage, err := ebiten.NewImage(scene.TileXNum*scene.TileSize, scene.TileYNum*scene.TileSize, ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	mapScene := &MapScene{
		currentMapID: 1,
		gameState:    gamestate.NewGame(),
		balloons:     &balloons{},
		tilesImage:   tilesImage,
		player:       player,
	}
	mapScene.changeRoom(roomID)
	return mapScene, nil
}

func (m *MapScene) state() *gamestate.Game {
	return m.gameState
}

func (m *MapScene) currentMap() *data.Map {
	for _, d := range data.Current().Maps {
		if d.ID == m.currentMapID {
			return d
		}
	}
	return nil
}

func (m *MapScene) currentRoom() *data.Room {
	for _, r := range m.currentMap().Rooms {
		if r.ID == m.currentRoomID {
			return r
		}
	}
	return nil
}

func (m *MapScene) changeRoom(roomID int) error {
	m.currentRoomID = roomID
	m.events = nil
	for _, e := range m.currentRoom().Events {
		event, err := newEvent(e, m)
		if err != nil {
			return err
		}
		m.events = append(m.events, event)
	}
	return nil
}

func (m *MapScene) tileSet(id int) (*data.TileSet, error) {
	for _, t := range data.Current().TileSets {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("mapscene: tile set not found: %d", id)
}

func (m *MapScene) passableTile(x, y int) (bool, error) {
	tileSet, err := m.tileSet(m.currentMap().TileSetID)
	if err != nil {
		return false, err
	}
	layer := 1
	tile := m.currentRoom().Tiles[layer][y*scene.TileXNum+x]
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
	tile = m.currentRoom().Tiles[layer][y*scene.TileXNum+x]
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

func (m *MapScene) movePlayerIfNeeded() error {
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
	if err := m.player.moveByUserInput(m.passable, tx, ty); err != nil {
		return err
	}
	m.moveDstX = tx
	m.moveDstY = ty
	if e == nil {
		return nil
	}
	e.tryRun(data.TriggerPlayer)
	return nil
}

func (m *MapScene) character(id int, self *event) *character {
	var ch *character
	switch id {
	case -1:
		ch = m.player.character
	case 0:
		ch = self.character
	default:
		for _, e := range m.events {
			if id == e.data.ID {
				return e.character
			}
		}
		return nil
	}
	return ch
}

func (m *MapScene) Update(sceneManager *scene.SceneManager) error {
	if err := m.gameState.Screen().Update(); err != nil {
		return err
	}
	if err := m.player.update(m.passable); err != nil {
		return err
	}
	if err := m.balloons.Update(); err != nil {
		return err
	}
	for _, e := range m.events {
		if err := e.update(); err != nil {
			return err
		}
	}
	for _, e := range m.events {
		if e.executingPage != nil {
			return nil
		}
	}
	for _, e := range m.events {
		if e.tryRun(data.TriggerAuto) {
			break
		}
	}
	if err := m.movePlayerIfNeeded(); err != nil {
		return err
	}
	return nil
}

func (m *MapScene) showMessage(content string, character *character) {
	content = m.gameState.ParseMessageSyntax(content)
	m.balloons.ShowMessage(content, character)
}

func (m *MapScene) showChoices(choices []string) {
	for i, c := range choices {
		choices[i] = m.gameState.ParseMessageSyntax(c)
	}
	m.balloons.ShowChoices(choices)
}

func (m *MapScene) transferPlayerImmediately(roomID, x, y int) {
	m.player.transferImmediately(x, y)
	m.changeRoom(roomID)
}

func (m *MapScene) fadeOut(count int) {
	m.gameState.Screen().FadeOut(count)
}

func (m *MapScene) fadeIn(count int) {
	m.gameState.Screen().FadeIn(count)
}

func (m *MapScene) isFadedOut() bool {
	return m.gameState.Screen().IsFadedOut()
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
	tileset, err := m.tileSet(m.currentMap().TileSetID)
	if err != nil {
		return err
	}
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &tilesImageParts{
		room:    m.currentRoom(),
		tileSet: tileset,
		layer:   0,
	}
	if err := m.tilesImage.DrawImage(assets.GetImage(tileset.Images[0]), op); err != nil {
		return err
	}
	op.ImageParts = &tilesImageParts{
		room:     m.currentRoom(),
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
	tileSet, err := m.tileSet(m.currentMap().TileSetID)
	if err != nil {
		return err
	}
	op.ImageParts = &tilesImageParts{
		room:     m.currentRoom(),
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
	m.gameState.Screen().Apply(&op.ColorM)
	if err := screen.DrawImage(m.tilesImage, op); err != nil {
		return err
	}
	if m.player.character.isMoving() {
		x, y := m.moveDstX, m.moveDstY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*scene.TileSize), float64(y*scene.TileSize))
		op.GeoM.Scale(scene.TileScale, scene.TileScale)
		op.GeoM.Translate(scene.GameMarginX, scene.GameMarginTop)
		if err := screen.DrawImage(assets.GetImage("marker.png"), op); err != nil {
			return err
		}
	}
	if err := m.balloons.Draw(screen); err != nil {
		return nil
	}
	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	if err := font.DrawText(screen, msg, 0, 0, scene.TextScale, color.White); err != nil {
		return err
	}
	return nil
}
