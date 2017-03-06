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
	"encoding/json"
	"fmt"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type moveCharacterState struct {
	mapID         int
	roomID        int
	eventID       int
	args          *data.CommandArgsMoveCharacter
	routeSkip     bool
	distanceCount int
	path          []routeCommand
	waiting       bool
	terminated    bool

	// Field that is not dumped
	gameState *Game
}

func newMoveCharacterState(gameState *Game, mapID, roomID, eventID int, args *data.CommandArgsMoveCharacter, routeSkip bool) (*moveCharacterState, error) {
	m := &moveCharacterState{
		gameState: gameState,
		mapID:     mapID,
		roomID:    roomID,
		eventID:   eventID,
		args:      args,
		routeSkip: routeSkip,
	}
	switch m.args.Type {
	case data.MoveCharacterTypeDirection, data.MoveCharacterTypeForward, data.MoveCharacterTypeBackward:
		m.distanceCount = m.args.Distance
	case data.MoveCharacterTypeTarget:
		cx, cy := m.character().Position()
		x, y := args.X, args.Y
		path, lastX, lastY := calcPath(&passableOnMap{
			through:          m.character().Through(),
			m:                m.gameState.Map(),
			ignoreCharacters: true,
		}, cx, cy, x, y)
		m.path = path
		m.distanceCount = len(path)
		if x != lastX || y != lastY {
			if !m.routeSkip {
				return nil, fmt.Errorf("gamestate: route is not found")
			}
			m.terminated = true
		}
	case data.MoveCharacterTypeRandom, data.MoveCharacterTypeToward:
		m.distanceCount = 1

	default:
		panic("not reach")
	}
	return m, nil
}

type tmpMoveCharacterState struct {
	MapID         int                            `json:"mapId"`
	RoomID        int                            `json:"roomId"`
	EventID       int                            `json:"eventId"`
	Args          *data.CommandArgsMoveCharacter `json:"args"`
	RouteSkip     bool                           `json:"routeSkip"`
	DistanceCount int                            `json:"distanceCount"`
	Path          []routeCommand                 `json:"path"`
	Waiting       bool                           `json:"waiting"`
	Terminated    bool                           `json:"terminated"`
}

func (m *moveCharacterState) MarshalJSON() ([]uint8, error) {
	tmp := &tmpMoveCharacterState{
		MapID:         m.mapID,
		RoomID:        m.roomID,
		EventID:       m.eventID,
		Args:          m.args,
		RouteSkip:     m.routeSkip,
		DistanceCount: m.distanceCount,
		Path:          m.path,
		Waiting:       m.waiting,
		Terminated:    m.terminated,
	}
	return json.Marshal(tmp)
}

func (m *moveCharacterState) UnmarshalJSON(jsonData []uint8) error {
	var tmp *tmpMoveCharacterState
	if err := json.Unmarshal(jsonData, &tmp); err != nil {
		return err
	}
	m.mapID = tmp.MapID
	m.roomID = tmp.RoomID
	m.eventID = tmp.EventID
	m.args = tmp.Args
	m.routeSkip = tmp.RouteSkip
	m.distanceCount = tmp.DistanceCount
	m.path = tmp.Path
	m.waiting = tmp.Waiting
	m.terminated = tmp.Terminated
	return nil
}

func (m *moveCharacterState) setGame(game *Game) {
	m.gameState = game
}

func (m *moveCharacterState) character() *character.Character {
	return m.gameState.character(m.mapID, m.roomID, m.eventID)
}

func (m *moveCharacterState) IsTerminated() bool {
	c := m.character()
	if c == nil {
		return true
	}
	if c.IsMoving() {
		return false
	}
	return m.terminated
}

func (m *moveCharacterState) Update() error {
	c := m.character()
	if c == nil {
		return nil
	}
	// Check IsMoving() first since the character might be moving at this time.
	if c.IsMoving() {
		return nil
	}
	if m.terminated {
		return nil
	}
	if m.distanceCount > 0 && !m.waiting {
		dx, dy := c.Position()
		var dir data.Dir
		switch m.args.Type {
		case data.MoveCharacterTypeDirection:
			dir = m.args.Dir
		case data.MoveCharacterTypeTarget:
			switch m.path[len(m.path)-m.distanceCount] {
			case routeCommandMoveUp:
				dir = data.DirUp
			case routeCommandMoveRight:
				dir = data.DirRight
			case routeCommandMoveDown:
				dir = data.DirDown
			case routeCommandMoveLeft:
				dir = data.DirLeft
			default:
				panic("not reach")
			}
		case data.MoveCharacterTypeForward:
			dir = c.Dir()
		case data.MoveCharacterTypeBackward:
			println(fmt.Sprintf("not implemented yet (move_character): type %s", m.args.Type))
			dir = c.Dir()
		case data.MoveCharacterTypeToward:
			println(fmt.Sprintf("not implemented yet (move_character): type %s", m.args.Type))
			dir = c.Dir()
		case data.MoveCharacterTypeRandom:
			println(fmt.Sprintf("not implemented yet (move_character): type %s", m.args.Type))
			dir = c.Dir()
		default:
			panic("not reach")
		}
		switch dir {
		case data.DirUp:
			dy--
		case data.DirRight:
			dx++
		case data.DirDown:
			dy++
		case data.DirLeft:
			dx--
		default:
			panic("not reach")
		}
		if !m.gameState.Map().passable(c.Through(), dx, dy, false) {
			c.Turn(dir)
			if !m.routeSkip {
				return nil
			}
			// Skip
			m.terminated = true
			m.distanceCount = 0
			// TODO: Can continue Update.
			return nil
		}
		c.Move(dir)
		m.waiting = true
		return nil
	}
	m.distanceCount--
	m.waiting = false
	if m.distanceCount > 0 {
		return nil
	}
	m.terminated = true
	return nil
}
