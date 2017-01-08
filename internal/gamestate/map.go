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
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
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
	if err := m.player.Update(); err != nil {
		return err
	}
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

func (m *Map) EventAt(x, y int) *character.Event {
	for _, e := range m.events {
		ex, ey := e.Position()
		if ex == x && ey == y {
			return e
		}
	}
	return nil
}

func (m *Map) TryRunAutoEvent() {
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

func (m *Map) DrawEvents(screen *ebiten.Image) error {
	for _, e := range m.events {
		if err := e.Draw(screen); err != nil {
			return err
		}
	}
	return nil
}

func (m *Map) transferPlayerImmediately(roomID, x, y int, interpreter *Interpreter) {
	m.player.TransferImmediately(x, y)
	m.setRoomID(roomID)
	// TODO: What if this is not nil?
	m.continuingInterpreter = interpreter
}

func (m *Map) CurrentMap() *data.Map {
	for _, d := range data.Current().Maps {
		if d.ID == m.mapID {
			return d
		}
	}
	return nil
}

func (m *Map) CurrentRoom() *data.Room {
	for _, r := range m.CurrentMap().Rooms {
		if r.ID == m.roomID {
			return r
		}
	}
	return nil
}

func (m *Map) DrawPlayer(screen *ebiten.Image) error {
	return m.player.Draw(screen)
}

func (m *Map) IsPlayerMovingByUserInput() bool {
	return m.game.variables.InnerVariableValue("is_player_moving_by_user_input") != 0
}

func (m *Map) MovePlayerByUserInput(passable func(x, y int) (bool, error), x, y int, event *character.Event) error {
	if m.playerMoving != nil {
		panic("not reach")
	}
	px, py := m.player.Position()
	lastPlayerX, lastPlayerY := px, py
	path, err := calcPath(passable, px, py, x, y)
	if err != nil {
		return err
	}
	if len(path) == 0 {
		return nil
	}
	commands := []*data.Command{}
	for _, r := range path {
		switch r {
		case routeCommandMoveUp:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Dir:      data.DirUp,
					Distance: 1,
				},
			})
			lastPlayerY--
		case routeCommandMoveRight:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Dir:      data.DirRight,
					Distance: 1,
				},
			})
			lastPlayerX++
		case routeCommandMoveDown:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Dir:      data.DirDown,
					Distance: 1,
				},
			})
			lastPlayerY++
		case routeCommandMoveLeft:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Dir:      data.DirLeft,
					Distance: 1,
				},
			})
			lastPlayerX--
		case routeCommandTurnUp:
			commands = append(commands, &data.Command{
				Name: data.CommandNameTurnCharacter,
				Args: &data.CommandArgsTurnCharacter{
					Dir: data.DirUp,
				},
			})
		case routeCommandTurnRight:
			commands = append(commands, &data.Command{
				Name: data.CommandNameTurnCharacter,
				Args: &data.CommandArgsTurnCharacter{
					Dir: data.DirRight,
				},
			})
		case routeCommandTurnDown:
			commands = append(commands, &data.Command{
				Name: data.CommandNameTurnCharacter,
				Args: &data.CommandArgsTurnCharacter{
					Dir: data.DirDown,
				},
			})
		case routeCommandTurnLeft:
			commands = append(commands, &data.Command{
				Name: data.CommandNameTurnCharacter,
				Args: &data.CommandArgsTurnCharacter{
					Dir: data.DirLeft,
				},
			})
		default:
			panic("not reach")
		}
	}
	commands = []*data.Command{
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
				Commands: commands,
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
