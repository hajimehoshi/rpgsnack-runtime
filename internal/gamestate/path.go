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

package gamestate

import (
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type routeCommand int

const (
	routeCommandMoveUp routeCommand = iota
	routeCommandMoveRight
	routeCommandMoveDown
	routeCommandMoveLeft
	routeCommandTurnUp
	routeCommandTurnRight
	routeCommandTurnDown
	routeCommandTurnLeft
)

func calcPath(passable func(x, y int) (bool, error), startX, startY, goalX, goalY int) ([]routeCommand, int, int, error) {
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
				pa, err := passable(s.X, s.Y)
				if err != nil {
					return nil, 0, 0, err
				}
				if !pa {
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
			return nil, 0, 0, nil
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
	path := make([]routeCommand, len(dirs))
	for i, d := range dirs {
		switch d {
		case data.DirUp:
			path[len(dirs)-i-1] = routeCommandMoveUp
		case data.DirRight:
			path[len(dirs)-i-1] = routeCommandMoveRight
		case data.DirDown:
			path[len(dirs)-i-1] = routeCommandMoveDown
		case data.DirLeft:
			path[len(dirs)-i-1] = routeCommandMoveLeft
		default:
			panic("not reach")
		}
	}
	lastP, err := passable(goalX, goalY)
	if err != nil {
		return nil, 0, 0, err
	}
	lastX, lastY := goalX, goalY
	if !lastP && len(path) > 0 {
		switch path[len(path)-1] {
		case routeCommandMoveUp:
			path[len(path)-1] = routeCommandTurnUp
			lastY++
		case routeCommandMoveRight:
			path[len(path)-1] = routeCommandTurnRight
			lastX--
		case routeCommandMoveDown:
			path[len(path)-1] = routeCommandTurnDown
			lastY--
		case routeCommandMoveLeft:
			path[len(path)-1] = routeCommandTurnLeft
			lastX++
		default:
			panic("not reach")
		}
	}
	return path, lastX, lastY, nil
}

func routeCommandsToEventCommands(path []routeCommand) []*data.Command {
	commands := []*data.Command{}
	for _, r := range path {
		switch r {
		case routeCommandMoveUp:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Type:     data.MoveCharacterTypeDirection,
					Dir:      data.DirUp,
					Distance: 1,
				},
			})
		case routeCommandMoveRight:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Type:     data.MoveCharacterTypeDirection,
					Dir:      data.DirRight,
					Distance: 1,
				},
			})
		case routeCommandMoveDown:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Type:     data.MoveCharacterTypeDirection,
					Dir:      data.DirDown,
					Distance: 1,
				},
			})
		case routeCommandMoveLeft:
			commands = append(commands, &data.Command{
				Name: data.CommandNameMoveCharacter,
				Args: &data.CommandArgsMoveCharacter{
					Type:     data.MoveCharacterTypeDirection,
					Dir:      data.DirLeft,
					Distance: 1,
				},
			})
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
	return commands
}
