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
	"regexp"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/assets"
	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/font"
	"github.com/hajimehoshi/tsugunai/internal/gamestate"
	"github.com/hajimehoshi/tsugunai/internal/input"
	"github.com/hajimehoshi/tsugunai/internal/scene"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

type MapScene struct {
	gameState     *gamestate.Game
	currentMapID  int
	currentRoomID int
	player        *player
	moveDstX      int
	moveDstY      int
	playerMoving  bool
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

func (m *MapScene) Update(subTasksUpdated bool, taskLine *task.TaskLine, sceneManager *scene.SceneManager) error {
	for _, e := range m.events {
		if err := e.updateCharacterIfNeeded(); err != nil {
			return err
		}
	}
	if err := m.gameState.Screen().Update(); err != nil {
		return nil
	}
	if err := m.player.update(m.passable); err != nil {
		return err
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
	for _, e := range m.events {
		if err := e.update(); err != nil {
			return err
		}
	}
	return nil
}

var reMessage = regexp.MustCompile(`\\([a-zA-Z])\[(\d+)\]`)

func (m *MapScene) parseMessageSyntax(str string) string {
	return reMessage.ReplaceAllStringFunc(str, func(part string) string {
		name := strings.ToLower(part[1:2])
		id, err := strconv.Atoi(part[3 : len(part)-1])
		if err != nil {
			panic(fmt.Sprintf("not reach: %s", err))
		}
		switch name {
		case "v":
			return strconv.Itoa(m.state().Variables().VariableValue(id))
		}
		return str
	})
}

func (m *MapScene) showMessage(taskLine *task.TaskLine, content string, character *character) {
	content = m.parseMessageSyntax(content)
	m.balloons.ShowMessage(taskLine, content, character)
}

func (m *MapScene) showChoices(taskLine *task.TaskLine, choices []string, chosenIndexSetter func(int)) {
	for i, c := range choices {
		choices[i] = m.parseMessageSyntax(c)
	}
	m.balloons.ShowChoices(taskLine, choices, chosenIndexSetter)
}

func (m *MapScene) closeAllBalloons(taskLine *task.TaskLine) {
	m.balloons.CloseAll(taskLine)
}

func (m *MapScene) transferPlayerImmediately(roomID, x, y int) {
	m.player.transferImmediately(x, y)
	m.changeRoom(roomID)
}

func (m *MapScene) fadeOut(taskLine *task.TaskLine, count int) {
	taskLine.PushFunc(func() error {
		m.gameState.Screen().FadeOut(count)
		return task.Terminated
	})
	taskLine.PushFunc(func() error {
		if m.gameState.Screen().IsFading() {
			return nil
		}
		return task.Terminated
	})
}

func (m *MapScene) fadeIn(taskLine *task.TaskLine, count int) {
	taskLine.PushFunc(func() error {
		m.gameState.Screen().FadeIn(count)
		return task.Terminated
	})
	taskLine.PushFunc(func() error {
		if m.gameState.Screen().IsFading() {
			return nil
		}
		return task.Terminated
	})
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
	if err := m.balloons.Draw(screen); err != nil {
		return nil
	}
	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	if err := font.DrawText(screen, msg, 0, 0, scene.TextScale, color.White); err != nil {
		return err
	}
	return nil
}
