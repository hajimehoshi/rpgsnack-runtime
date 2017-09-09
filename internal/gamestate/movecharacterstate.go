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
	"log"

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/path"
)

type moveCharacterState struct {
	mapID         int
	roomID        int
	eventID       int
	args          *data.CommandArgsMoveCharacter
	routeSkip     bool
	distanceCount int
	path          []path.RouteCommand
	waiting       bool
	terminated    bool
}

type atFunc struct {
	f func(x, y int) bool
}

func (a *atFunc) At(x, y int) bool {
	return a.f(x, y)
}

func (m *moveCharacterState) setMoveTarget(gameState *Game, x int, y int, ignoreCharacters bool) bool {
	ch := m.character(gameState)
	cx, cy := ch.Position()
	path, lastX, lastY := path.Calc(&atFunc{
		f: func(x, y int) bool {
			return gameState.MapPassableAt(ch.Through(), x, y, ignoreCharacters)
		},
	}, cx, cy, x, y)
	m.path = path
	m.distanceCount = len(path)
	if x != lastX || y != lastY {
		if !m.routeSkip {
			return false
		}
		m.terminated = true
	}

	return true
}

func newMoveCharacterState(gameState *Game, mapID, roomID, eventID int, args *data.CommandArgsMoveCharacter, routeSkip bool) *moveCharacterState {
	m := &moveCharacterState{
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
		if args.ValueType == data.ValueTypeVariable {
			if !m.setMoveTarget(gameState, gameState.VariableValue(args.X), gameState.VariableValue(args.Y), args.IgnoreCharacters) {
				return nil
			}
		} else {
			if !m.setMoveTarget(gameState, args.X, args.Y, args.IgnoreCharacters) {
				return nil
			}
		}
	case data.MoveCharacterTypeRandom, data.MoveCharacterTypeToward:
		m.distanceCount = 1

	default:
		panic("not reach")
	}
	return m
}

func (m *moveCharacterState) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("mapId")
	e.EncodeInt(m.mapID)

	e.EncodeString("roomId")
	e.EncodeInt(m.roomID)

	e.EncodeString("eventId")
	e.EncodeInt(m.eventID)

	e.EncodeString("args")
	e.EncodeInterface(m.args)

	e.EncodeString("routeSkip")
	e.EncodeBool(m.routeSkip)

	e.EncodeString("distanceCount")
	e.EncodeInt(m.distanceCount)

	e.EncodeString("path")
	e.BeginArray()
	for _, r := range m.path {
		e.EncodeInt(int(r))
	}
	e.EndArray()

	e.EncodeString("waiting")
	e.EncodeBool(m.waiting)

	e.EncodeString("terminated")
	e.EncodeBool(m.terminated)

	e.EndMap()
	return e.Flush()
}

func (m *moveCharacterState) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "mapId":
			m.mapID = d.DecodeInt()
		case "roomId":
			m.roomID = d.DecodeInt()
		case "eventId":
			m.eventID = d.DecodeInt()
		case "args":
			if !d.SkipCodeIfNil() {
				m.args = &data.CommandArgsMoveCharacter{}
				d.DecodeInterface(m.args)
			}
		case "routeSkip":
			m.routeSkip = d.DecodeBool()
		case "distanceCount":
			m.distanceCount = d.DecodeInt()
		case "path":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				m.path = make([]path.RouteCommand, n)
				for i := 0; i < n; i++ {
					m.path[i] = path.RouteCommand(d.DecodeInt())
				}
			}
		case "waiting":
			m.waiting = d.DecodeBool()
		case "terminated":
			m.terminated = d.DecodeBool()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("gamestate: moveCharacterState.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (m *moveCharacterState) character(gameState *Game) *character.Character {
	return gameState.Character(m.mapID, m.roomID, m.eventID)
}

func (m *moveCharacterState) IsTerminated(gameState *Game) bool {
	c := m.character(gameState)
	if c == nil {
		return true
	}
	if c.IsMoving() {
		return false
	}
	return m.terminated
}

func (m *moveCharacterState) Update(gameState *Game) error {
	c := m.character(gameState)
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
			case path.RouteCommandMoveUp:
				dir = data.DirUp
			case path.RouteCommandMoveRight:
				dir = data.DirRight
			case path.RouteCommandMoveDown:
				dir = data.DirDown
			case path.RouteCommandMoveLeft:
				dir = data.DirLeft
			default:
				panic("not reach")
			}
		case data.MoveCharacterTypeForward:
			dir = c.Dir()
		case data.MoveCharacterTypeBackward:
			dir = (c.Dir() + 2) % 4
		case data.MoveCharacterTypeToward:
			log.Printf("not implemented yet (move_character): type %s", m.args.Type)
			dir = c.Dir()
		case data.MoveCharacterTypeAgainst:
			log.Printf("not implemented yet (move_character): type %s", m.args.Type)
			dir = c.Dir()
		case data.MoveCharacterTypeRandom:
			dir = data.Dir(gameState.RandomValue(0, 4))
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
		if !gameState.MapPassableAt(c.Through(), dx, dy, false) {
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
