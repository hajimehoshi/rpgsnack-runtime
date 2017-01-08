// Copyright 2017 Hajime Hoshi
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

package gamestate

import (
	"fmt"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type Map struct {
	game                  *Game
	player                *character.Player
	mapID                 int
	roomID                int
	events                []*character.Event
	continuingInterpreter *Interpreter
	autoInterpreter       *Interpreter
	playerMoving          *Interpreter
}

func NewMap(game *Game) (*Map, error) {
	pos := data.Current().System.InitialPosition
	x, y, roomID := 0, 0, 1
	if pos != nil {
		x, y, roomID = pos.X, pos.Y, pos.RoomID
	}
	player, err := character.NewPlayer(x, y)
	if err != nil {
		return nil, err
	}
	m := &Map{
		game:   game,
		player: player,
		mapID:  1,
	}
	m.setRoomID(roomID)
	return m, nil
}

func (m *Map) TileSet() (*data.TileSet, error) {
	id := m.currentMap().TileSetID
	for _, t := range data.Current().TileSets {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("mapscene: tile set not found: %d", id)
}

func (m *Map) setRoomID(id int) error {
	m.roomID = id
	m.events = nil
	for _, e := range m.CurrentRoom().Events {
		event, err := character.NewEvent(e)
		if err != nil {
			return err
		}
		m.events = append(m.events, event)
	}
	return nil
}

func (m *Map) IsEventExecuting() bool {
	if m.playerMoving != nil && m.playerMoving.IsExecuting() {
		return true
	}
	if m.continuingInterpreter != nil && m.continuingInterpreter.IsExecuting() {
		return true
	}
	if m.autoInterpreter != nil && m.autoInterpreter.IsExecuting() {
		return true
	}
	return false
}

func (m *Map) meetsPageCondition(page *data.Page, eventID int) (bool, error) {
	for _, cond := range page.Conditions {
		m, err := m.game.MeetsCondition(cond, eventID)
		if err != nil {
			return false, err
		}
		if !m {
			return false, nil
		}
	}
	return true, nil
}

func (m *Map) pageIndex(eventID int) (int, error) {
	var event *data.Event
	for _, e := range m.CurrentRoom().Events {
		if e.ID == eventID {
			event = e
			break
		}
	}
	if event == nil {
		panic("not reach")
	}
	for i := len(event.Pages) - 1; i >= 0; i-- {
		page := event.Pages[i]
		m, err := m.meetsPageCondition(page, event.ID)
		if err != nil {
			return 0, err
		}
		if m {
			return i, nil
		}
	}
	return -1, nil
}

func (m *Map) Update() error {
	if m.playerMoving != nil {
		if err := m.playerMoving.Update(); err != nil {
			return err
		}
		if !m.playerMoving.IsExecuting() {
			m.playerMoving = nil
		}
	}
	if m.autoInterpreter != nil {
		if err := m.autoInterpreter.Update(); err != nil {
			return err
		}
		if !m.autoInterpreter.IsExecuting() {
			m.autoInterpreter = nil
		}
	}
	if m.continuingInterpreter != nil {
		if err := m.continuingInterpreter.Update(); err != nil {
			return err
		}
		if !m.continuingInterpreter.IsExecuting() {
			m.continuingInterpreter = nil
		}
	}
	if err := m.player.Update(); err != nil {
		return err
	}
	if m.IsPlayerMovingByUserInput() {
		return nil
	}
	for _, e := range m.events {
		index, err := m.pageIndex(e.ID())
		if err != nil {
			return err
		}
		if err := e.UpdateCharacterIfNeeded(index); err != nil {
			return err
		}
	}
	for _, e := range m.events {
		if err := e.Update(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Map) eventAt(x, y int) *character.Event {
	for _, e := range m.events {
		ex, ey := e.Position()
		if ex == x && ey == y {
			return e
		}
	}
	return nil
}

func (m *Map) TryRunAutoEvent() {
	if m.IsEventExecuting() {
		return
	}
	if m.autoInterpreter != nil {
		return
	}
	for _, e := range m.events {
		page := e.CurrentPage()
		if page == nil {
			continue
		}
		if page.Trigger != data.TriggerAuto {
			continue
		}
		m.autoInterpreter = NewInterpreter(m.game, m.mapID, m.roomID, e.ID(), page.Commands)
		break
	}
}

func (m *Map) transferPlayerImmediately(roomID, x, y int, interpreter *Interpreter) {
	m.player.TransferImmediately(x, y)
	m.setRoomID(roomID)
	// TODO: What if this is not nil?
	m.continuingInterpreter = interpreter
}

func (m *Map) currentMap() *data.Map {
	for _, d := range data.Current().Maps {
		if d.ID == m.mapID {
			return d
		}
	}
	return nil
}

func (m *Map) CurrentRoom() *data.Room {
	for _, r := range m.currentMap().Rooms {
		if r.ID == m.roomID {
			return r
		}
	}
	return nil
}

func (m *Map) IsPlayerMovingByUserInput() bool {
	return m.game.variables.InnerVariableValue("is_player_moving_by_user_input") != 0
}

func (m *Map) passableTile(x, y int) (bool, error) {
	tileSet, err := m.TileSet()
	if err != nil {
		return false, err
	}
	layer := 1
	tile := m.CurrentRoom().Tiles[layer][y*scene.TileXNum+x]
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
	tile = m.CurrentRoom().Tiles[layer][y*scene.TileXNum+x]
	if tileSet.PassageTypes[layer][tile] == data.PassageTypePassable {
		return true, nil
	}
	return false, nil
}

func (m *Map) passable(x, y int) (bool, error) {
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
	return e.IsPassable(), nil
}

func (m *Map) MovePlayerByUserInput(x, y int) error {
	if m.playerMoving != nil {
		panic("not reach")
	}
	event := m.eventAt(x, y)
	p, err := m.passable(x, y)
	if err != nil {
		return err
	}
	if !p {
		if event == nil {
			return nil
		}
		if !event.IsRunnable() {
			return nil
		}
		if event.CurrentPage().Trigger != data.TriggerPlayer {
			return nil
		}
	}
	px, py := m.player.Position()
	path, lastPlayerX, lastPlayerY, err := calcPath(m.passable, px, py, x, y)
	if err != nil {
		return err
	}
	commands := []*data.Command{
		{
			Name: data.CommandNameSetInnerVariable,
			Args: &data.CommandArgsSetInnerVariable{
				Name:  "is_player_moving_by_user_input",
				Value: 1,
			},
		},
		{
			Name: data.CommandNameSetRoute,
			Args: &data.CommandArgsSetRoute{
				EventID:  -1,
				Repeat:   false,
				Skip:     false,
				Wait:     true,
				Commands: routeCommandsToEventCommands(path),
			},
		},
		{
			Name: data.CommandNameSetInnerVariable,
			Args: &data.CommandArgsSetInnerVariable{
				Name:  "is_player_moving_by_user_input",
				Value: 0,
			},
		},
	}
	if event != nil && event.CurrentPage() != nil && event.CurrentPage().Trigger == data.TriggerPlayer {
		origDir := event.Dir()
		var dir data.Dir
		ex, ey := event.Position()
		px, py := lastPlayerX, lastPlayerY
		switch {
		case ex == px && ey == py:
			// The player and the event are at the same position.
			dir = event.Dir()
		case ex > px && ey == py:
			dir = data.DirLeft
		case ex < px && ey == py:
			dir = data.DirRight
		case ex == px && ey > py:
			dir = data.DirUp
		case ex == px && ey < py:
			dir = data.DirDown
		default:
			panic("not reach")
		}
		commands = append(commands,
			&data.Command{
				Name: data.CommandNameSetRoute,
				Args: &data.CommandArgsSetRoute{
					EventID: event.ID(),
					Repeat:  false,
					Skip:    false,
					Wait:    true,
					Commands: []*data.Command{
						{
							Name: data.CommandNameTurnCharacter,
							Args: &data.CommandArgsTurnCharacter{
								Dir: dir,
							},
						},
					},
				},
			},
			&data.Command{
				Name: data.CommandNameCallEvent,
				Args: &data.CommandArgsCallEvent{
					EventID:   event.ID(),
					PageIndex: event.CurrentPageIndex(),
				},
			},
			&data.Command{
				Name: data.CommandNameSetRoute,
				Args: &data.CommandArgsSetRoute{
					EventID: event.ID(),
					Repeat:  false,
					Skip:    false,
					Wait:    true,
					Commands: []*data.Command{
						{
							Name: data.CommandNameTurnCharacter,
							Args: &data.CommandArgsTurnCharacter{
								Dir: origDir,
							},
						},
					},
				},
			})
	}
	m.playerMoving = NewInterpreter(m.game, m.mapID, m.roomID, -1, commands)
	return nil
}

func (m *Map) DrawPlayer(screen *ebiten.Image) error {
	return m.player.Draw(screen)
}

func (m *Map) DrawEvents(screen *ebiten.Image) error {
	for _, e := range m.events {
		if err := e.Draw(screen); err != nil {
			return err
		}
	}
	return nil
}
