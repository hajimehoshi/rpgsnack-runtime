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

package path

import (
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type RouteCommand int

const (
	RouteCommandMoveUp RouteCommand = iota
	RouteCommandMoveRight
	RouteCommandMoveDown
	RouteCommandMoveLeft
	RouteCommandTurnUp
	RouteCommandTurnRight
	RouteCommandTurnDown
	RouteCommandTurnLeft
)

type Passable interface {
	At(x, y int) bool
}

func Calc(passable Passable, startX, startY, goalX, goalY int) ([]RouteCommand, int, int) {
	type pos struct {
		X, Y int
	}
	current := []pos{{startX, startY}}
	parents := map[pos]pos{}
	for 0 < len(current) {
		next := []pos{}
		for _, p := range current {
			successors := []pos{
				{p.X + 1, p.Y},
				{p.X - 1, p.Y},
				{p.X, p.Y + 1},
				{p.X, p.Y - 1},
			}
			for _, s := range successors {
				if !passable.At(s.X, s.Y) {
					// It's OK even if the final destination is not passable so far.
					if s.X != goalX || s.Y != goalY {
						continue
					}
				}
				if _, ok := parents[s]; ok {
					continue
				}
				parents[s] = p
				if s.X == goalX && s.Y == goalY {
					break
				}
				next = append(next, s)
			}
		}
		current = next
	}
	p := pos{goalX, goalY}
	dirs := []data.Dir{}
	for p.X != startX || p.Y != startY {
		parent, ok := parents[p]
		// There is no path.
		if !ok {
			return nil, 0, 0
		}
		switch {
		case parent.X == p.X-1:
			dirs = append(dirs, data.DirRight)
		case parent.X == p.X+1:
			dirs = append(dirs, data.DirLeft)
		case parent.Y == p.Y-1:
			dirs = append(dirs, data.DirDown)
		case parent.Y == p.Y+1:
			dirs = append(dirs, data.DirUp)
		default:
			panic("not reach")
		}
		p = parent
	}
	path := make([]RouteCommand, len(dirs))
	for i, d := range dirs {
		switch d {
		case data.DirUp:
			path[len(dirs)-i-1] = RouteCommandMoveUp
		case data.DirRight:
			path[len(dirs)-i-1] = RouteCommandMoveRight
		case data.DirDown:
			path[len(dirs)-i-1] = RouteCommandMoveDown
		case data.DirLeft:
			path[len(dirs)-i-1] = RouteCommandMoveLeft
		default:
			panic("not reach")
		}
	}
	lastP := passable.At(goalX, goalY)
	lastX, lastY := goalX, goalY
	if !lastP && len(path) > 0 {
		switch path[len(path)-1] {
		case RouteCommandMoveUp:
			path[len(path)-1] = RouteCommandTurnUp
			lastY++
		case RouteCommandMoveRight:
			path[len(path)-1] = RouteCommandTurnRight
			lastX--
		case RouteCommandMoveDown:
			path[len(path)-1] = RouteCommandTurnDown
			lastY--
		case RouteCommandMoveLeft:
			path[len(path)-1] = RouteCommandTurnLeft
			lastX++
		default:
			panic("not reach")
		}
	}
	return path, lastX, lastY
}

func RouteCommandsToEventCommands(path []RouteCommand) []*data.Command {
	commands := []*data.Command{}
	for _, r := range path {
		switch r {
		case RouteCommandMoveUp:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Type:     data.MoveCharacterTypeDirection,
					Dir:      data.DirUp,
					Distance: 1,
				},
			})
		case RouteCommandMoveRight:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Type:     data.MoveCharacterTypeDirection,
					Dir:      data.DirRight,
					Distance: 1,
				},
			})
		case RouteCommandMoveDown:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Type:     data.MoveCharacterTypeDirection,
					Dir:      data.DirDown,
					Distance: 1,
				},
			})
		case RouteCommandMoveLeft:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Type:     data.MoveCharacterTypeDirection,
					Dir:      data.DirLeft,
					Distance: 1,
				},
			})
		case RouteCommandTurnUp:
			commands = append(commands, &data.Command{
				Name: data.CommandNameTurnCharacter,
				Args: &data.CommandArgsTurnCharacter{
					Dir: data.DirUp,
				},
			})
		case RouteCommandTurnRight:
			commands = append(commands, &data.Command{
				Name: data.CommandNameTurnCharacter,
				Args: &data.CommandArgsTurnCharacter{
					Dir: data.DirRight,
				},
			})
		case RouteCommandTurnDown:
			commands = append(commands, &data.Command{
				Name: data.CommandNameTurnCharacter,
				Args: &data.CommandArgsTurnCharacter{
					Dir: data.DirDown,
				},
			})
		case RouteCommandTurnLeft:
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
	return commands
}
