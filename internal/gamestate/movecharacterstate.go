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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type moveCharacterState struct {
	gameState     *Game
	character     *character.Character
	args          *data.CommandArgsMoveCharacter
	routeSkip     bool
	distanceCount int
	path          []routeCommand
	waiting       bool
	terminated    bool
}

func newMoveCharacterState(gameState *Game, character *character.Character, args *data.CommandArgsMoveCharacter, routeSkip bool) (*moveCharacterState, error) {
	m := &moveCharacterState{
		gameState: gameState,
		character: character,
		args:      args,
		routeSkip: routeSkip,
	}
	switch m.args.Type {
	case data.MoveCharacterTypeDirection, data.MoveCharacterTypeForward, data.MoveCharacterTypeBackward:
		m.distanceCount = m.args.Distance
	case data.MoveCharacterTypeTarget:
		cx, cy := m.character.Position()
		x, y := args.X, args.Y
		path, lastX, lastY, err := calcPath(&passableOnMap{
			self:             m.character,
			m:                m.gameState.Map(),
			ignoreCharacters: true,
		}, cx, cy, x, y)
		if err != nil {
			return nil, err
		}
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

func (m *moveCharacterState) IsTerminated() bool {
	if m.character.IsMoving() {
		return false
	}
	return m.terminated
}

func (m *moveCharacterState) Update() error {
	// Check IsMoving() first since the character might be moving at this time.
	if m.character.IsMoving() {
		return nil
	}
	if m.terminated {
		return nil
	}
	if m.distanceCount > 0 && !m.waiting {
		dx, dy := m.character.Position()
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
			dir = m.character.Dir()
		case data.MoveCharacterTypeBackward:
			println(fmt.Sprintf("not implemented yet (move_character): type %s", m.args.Type))
			dir = m.character.Dir()
		case data.MoveCharacterTypeToward:
			println(fmt.Sprintf("not implemented yet (move_character): type %s", m.args.Type))
			dir = m.character.Dir()
		case data.MoveCharacterTypeRandom:
			println(fmt.Sprintf("not implemented yet (move_character): type %s", m.args.Type))
			dir = m.character.Dir()
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
		p, err := m.gameState.Map().passable(m.character, dx, dy, false)
		if err != nil {
			return err
		}
		if !p {
			m.character.Turn(dir)
			if !m.routeSkip {
				return nil
			}
			// Skip
			m.terminated = true
			m.distanceCount = 0
			// TODO: Can continue Update.
			return nil
		}
		m.character.Move(dir)
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