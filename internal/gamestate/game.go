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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/window"
)

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Game struct {
	variables             *Variables
	screen                *Screen
	windows               *window.Windows
	player                *character.Player
	mapID                 int
	roomID                int
	events                []*character.Event
	continuingInterpreter *Interpreter
	playerMoving          *Interpreter
}

func NewGame() (*Game, error) {
	pos := data.Current().System.InitialPosition
	x, y, roomID := 0, 0, 1
	if pos != nil {
		x, y, roomID = pos.X, pos.Y, pos.RoomID
	}
	player, err := character.NewPlayer(x, y)
	if err != nil {
		return nil, err
	}
	g := &Game{
		variables: &Variables{},
		screen:    &Screen{},
		windows:   &window.Windows{},
		player:    player,
		mapID:     1,
	}
	g.setRoomID(roomID)
	return g, nil
}

func (g *Game) setRoomID(id int) error {
	g.roomID = id
	g.events = nil
	for _, e := range g.CurrentRoom().Events {
		i := NewInterpreter(g, g.mapID, g.roomID, e.ID)
		event, err := character.NewEvent(e, i)
		if err != nil {
			return err
		}
		g.events = append(g.events, event)
	}
	return nil
}

func (g *Game) Events() []*character.Event {
	return g.events
}

func (g *Game) IsEventExecuting() bool {
	if g.playerMoving != nil && g.playerMoving.IsExecuting() {
		return true
	}
	if g.continuingInterpreter != nil && g.continuingInterpreter.IsExecuting() {
		return true
	}
	for _, e := range g.events {
		if e.IsExecutingCommands() {
			return true
		}
	}
	return false
}

func (g *Game) meetsPageCondition(page *data.Page, eventID int) (bool, error) {
	for _, cond := range page.Conditions {
		m, err := g.MeetsCondition(cond, eventID)
		if err != nil {
			return false, err
		}
		if !m {
			return false, nil
		}
	}
	return true, nil
}

func (g *Game) pageIndex(eventID int) (int, error) {
	var event *data.Event
	for _, e := range g.CurrentRoom().Events {
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
		m, err := g.meetsPageCondition(page, event.ID)
		if err != nil {
			return 0, err
		}
		if m {
			return i, nil
		}
	}
	return -1, nil
}

func (g *Game) UpdateEvents() error {
	if g.playerMoving != nil {
		if err := g.playerMoving.Update(); err != nil {
			return err
		}
		if !g.playerMoving.IsExecuting() {
			g.playerMoving = nil
		}
	}
	for _, e := range g.events {
		index, err := g.pageIndex(e.ID())
		if err != nil {
			return err
		}
		if err := e.UpdateCharacterIfNeeded(index); err != nil {
			return err
		}
	}
	for _, e := range g.events {
		if err := e.Update(); err != nil {
			return err
		}
	}
	if g.continuingInterpreter != nil {
		if err := g.continuingInterpreter.Update(); err != nil {
			return err
		}
		if !g.continuingInterpreter.IsExecuting() {
			g.continuingInterpreter = nil
		}
	}
	return nil
}

func (g *Game) transferPlayerImmediately(roomID, x, y int, interpreter *Interpreter) {
	g.player.TransferImmediately(x, y)
	g.setRoomID(roomID)
	// TODO: What if this is not nil?
	g.continuingInterpreter = interpreter
}

func (g *Game) CurrentMap() *data.Map {
	for _, d := range data.Current().Maps {
		if d.ID == g.mapID {
			return d
		}
	}
	return nil
}

func (g *Game) CurrentRoom() *data.Room {
	for _, r := range g.CurrentMap().Rooms {
		if r.ID == g.roomID {
			return r
		}
	}
	return nil
}

func (g *Game) Screen() *Screen {
	return g.screen
}

func (g *Game) Windows() *window.Windows {
	return g.windows
}

func (g *Game) UpdatePlayer() error {
	return g.player.Update()
}

func (g *Game) DrawPlayer(screen *ebiten.Image) error {
	return g.player.Draw(screen)
}

func (g *Game) IsPlayerMovingByUserInput() bool {
	return g.variables.InnerVariableValue("is_player_moving_by_user_input") != 0
}

func (g *Game) MovePlayerByUserInput(passable func(x, y int) (bool, error), x, y int, event *character.Event) error {
	if g.playerMoving != nil {
		panic("not reach")
	}
	px, py := g.player.Position()
	lastPlayerX, lastPlayerY := px, py
	path, err := calcPath(passable, px, py, x, y)
	if err != nil {
		return err
	}
	if len(path) == 0 {
		return nil
	}
	g.playerMoving = NewInterpreter(g, g.mapID, g.roomID, -1)
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
	if event != nil {
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
	g.playerMoving.SetCommands(commands)
	return nil
}

var reMessage = regexp.MustCompile(`\\([a-zA-Z])\[(\d+)\]`)

func (g *Game) ParseMessageSyntax(str string) string {
	return reMessage.ReplaceAllStringFunc(str, func(part string) string {
		name := strings.ToLower(part[1:2])
		id, err := strconv.Atoi(part[3 : len(part)-1])
		if err != nil {
			panic(fmt.Sprintf("not reach: %s", err))
		}
		switch name {
		case "v":
			return strconv.Itoa(g.variables.VariableValue(id))
		}
		return str
	})
}

func (g *Game) MeetsCondition(cond *data.Condition, eventID int) (bool, error) {
	// TODO: Is it OK to allow null conditions?
	if cond == nil {
		return true, nil
	}
	switch cond.Type {
	case data.ConditionTypeSwitch:
		id := cond.ID
		v := g.variables.SwitchValue(id)
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeSelfSwitch:
		m, r := g.mapID, g.roomID
		v := g.variables.SelfSwitchValue(m, r, eventID, cond.ID)
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeVariable:
		id := cond.ID
		v := g.variables.VariableValue(id)
		rhs := cond.Value.(int)
		switch cond.ValueType {
		case data.ConditionValueTypeConstant:
		case data.ConditionValueTypeVariable:
			rhs = g.variables.VariableValue(rhs)
		default:
			return false, fmt.Errorf("mapscene: invalid value type: %s", cond.ValueType)
		}
		switch cond.Comp {
		case data.ConditionCompEqualTo:
			return v == rhs, nil
		case data.ConditionCompNotEqualTo:
			return v != rhs, nil
		case data.ConditionCompGreaterThanOrEqualTo:
			return v >= rhs, nil
		case data.ConditionCompGreaterThan:
			return v > rhs, nil
		case data.ConditionCompLessThanOrEqualTo:
			return v <= rhs, nil
		case data.ConditionCompLessThan:
			return v < rhs, nil
		default:
			return false, fmt.Errorf("mapscene: invalid comp: %s", cond.Comp)
		}
	default:
		return false, fmt.Errorf("mapscene: invalid condition: %s", cond)
	}
	return false, nil
}
