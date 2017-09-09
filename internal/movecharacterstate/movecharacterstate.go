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

package movecharacterstate

import (
	"fmt"
	"log"

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/path"
)

type State struct {
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

type GameState interface {
	MapPassableAt(through bool, x, y int, ignoreCharacters bool) bool
	VariableValue(id int) int
	RandomValue(min, max int) int
	Character(mapID, roomID, eventID int) *character.Character
}

func (s *State) setMoveTarget(gameState GameState, x int, y int, ignoreCharacters bool) bool {
	ch := s.character(gameState)
	cx, cy := ch.Position()
	path, lastX, lastY := path.Calc(&atFunc{
		f: func(x, y int) bool {
			return gameState.MapPassableAt(ch.Through(), x, y, ignoreCharacters)
		},
	}, cx, cy, x, y)
	s.path = path
	s.distanceCount = len(path)
	if x != lastX || y != lastY {
		if !s.routeSkip {
			return false
		}
		s.terminated = true
	}

	return true
}

func New(gameState GameState, mapID, roomID, eventID int, args *data.CommandArgsMoveCharacter, routeSkip bool) *State {
	s := &State{
		mapID:     mapID,
		roomID:    roomID,
		eventID:   eventID,
		args:      args,
		routeSkip: routeSkip,
	}
	switch s.args.Type {
	case data.MoveCharacterTypeDirection, data.MoveCharacterTypeForward, data.MoveCharacterTypeBackward:
		s.distanceCount = s.args.Distance
	case data.MoveCharacterTypeTarget:
		if args.ValueType == data.ValueTypeVariable {
			if !s.setMoveTarget(gameState, gameState.VariableValue(args.X), gameState.VariableValue(args.Y), args.IgnoreCharacters) {
				return nil
			}
		} else {
			if !s.setMoveTarget(gameState, args.X, args.Y, args.IgnoreCharacters) {
				return nil
			}
		}
	case data.MoveCharacterTypeRandom, data.MoveCharacterTypeToward:
		s.distanceCount = 1

	default:
		panic("not reach")
	}
	return s
}

func (s *State) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("mapId")
	e.EncodeInt(s.mapID)

	e.EncodeString("roomId")
	e.EncodeInt(s.roomID)

	e.EncodeString("eventId")
	e.EncodeInt(s.eventID)

	e.EncodeString("args")
	e.EncodeInterface(s.args)

	e.EncodeString("routeSkip")
	e.EncodeBool(s.routeSkip)

	e.EncodeString("distanceCount")
	e.EncodeInt(s.distanceCount)

	e.EncodeString("path")
	e.BeginArray()
	for _, r := range s.path {
		e.EncodeInt(int(r))
	}
	e.EndArray()

	e.EncodeString("waiting")
	e.EncodeBool(s.waiting)

	e.EncodeString("terminated")
	e.EncodeBool(s.terminated)

	e.EndMap()
	return e.Flush()
}

func (s *State) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "mapId":
			s.mapID = d.DecodeInt()
		case "roomId":
			s.roomID = d.DecodeInt()
		case "eventId":
			s.eventID = d.DecodeInt()
		case "args":
			if !d.SkipCodeIfNil() {
				s.args = &data.CommandArgsMoveCharacter{}
				d.DecodeInterface(s.args)
			}
		case "routeSkip":
			s.routeSkip = d.DecodeBool()
		case "distanceCount":
			s.distanceCount = d.DecodeInt()
		case "path":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				s.path = make([]path.RouteCommand, n)
				for i := 0; i < n; i++ {
					s.path[i] = path.RouteCommand(d.DecodeInt())
				}
			}
		case "waiting":
			s.waiting = d.DecodeBool()
		case "terminated":
			s.terminated = d.DecodeBool()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("movecharacterstate: State.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (s *State) character(gameState GameState) *character.Character {
	return gameState.Character(s.mapID, s.roomID, s.eventID)
}

func (s *State) IsTerminated(gameState GameState) bool {
	c := s.character(gameState)
	if c == nil {
		return true
	}
	if c.IsMoving() {
		return false
	}
	return s.terminated
}

func (s *State) Update(gameState GameState) error {
	c := s.character(gameState)
	if c == nil {
		return nil
	}
	// Check IsMoving() first since the character might be moving at this time.
	if c.IsMoving() {
		return nil
	}
	if s.terminated {
		return nil
	}
	if s.distanceCount > 0 && !s.waiting {
		dx, dy := c.Position()
		var dir data.Dir
		switch s.args.Type {
		case data.MoveCharacterTypeDirection:
			dir = s.args.Dir
		case data.MoveCharacterTypeTarget:
			switch s.path[len(s.path)-s.distanceCount] {
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
			log.Printf("not implemented yet (move_character): type %s", s.args.Type)
			dir = c.Dir()
		case data.MoveCharacterTypeAgainst:
			log.Printf("not implemented yet (move_character): type %s", s.args.Type)
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
			if !s.routeSkip {
				return nil
			}
			// Skip
			s.terminated = true
			s.distanceCount = 0
			// TODO: Can continue Update.
			return nil
		}
		c.Move(dir)
		s.waiting = true
		return nil
	}
	s.distanceCount--
	s.waiting = false
	if s.distanceCount > 0 {
		return nil
	}
	s.terminated = true
	return nil
}
