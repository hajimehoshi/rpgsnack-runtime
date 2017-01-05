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
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

const playerMaxMoveCount = 4

type player struct {
	character *character
}

func newPlayer(x, y int) (*player, error) {
	c := &character{
		imageName:    "characters0.png",
		imageIndex:   0,
		x:            x,
		y:            y,
		dir:          data.DirDown,
		dirFix:       false,
		attitude:     data.AttitudeMiddle,
		prevAttitude: data.AttitudeMiddle,
	}
	return &player{
		character: c,
	}, nil
}

func (p *player) move(taskLine *task.TaskLine, passable func(x, y int) (bool, error), x, y int) error {
	c := p.character
	path, err := calcPath(passable, c.x, c.y, x, y)
	if err != nil {
		return err
	}
	for _, d := range path {
		d := d
		init := false
		taskLine.PushFunc(func() error {
			if !init {
				c.turn(d)
				c.moveCount = playerMaxMoveCount
				init = true
			}
			nx, ny := c.x, c.y
			switch d {
			case data.DirLeft:
				nx--
			case data.DirRight:
				nx++
			case data.DirUp:
				ny--
			case data.DirDown:
				ny++
			}
			p, err := passable(nx, ny)
			if err != nil {
				return err
			}
			if !p {
				nx = c.x
				ny = c.y
				c.moveCount = 0
				return task.Terminated
			}
			if c.moveCount > 0 {
				if c.moveCount >= playerMaxMoveCount/2 {
					c.attitude = data.AttitudeMiddle
				} else if c.prevAttitude == data.AttitudeLeft {
					c.attitude = data.AttitudeRight
				} else {
					c.attitude = data.AttitudeLeft
				}
				c.moveCount--
			}
			if c.moveCount == 0 {
				c.x = nx
				c.y = ny
				c.prevAttitude = c.attitude
				c.attitude = data.AttitudeMiddle
				return task.Terminated
			}
			return nil
		})
	}
	return nil
}

func (p *player) transferImmediately(x, y int) {
	p.character.transferImmediately(x, y)
}

func (p *player) update(passable func(x, y int) (bool, error)) error {
	if err := p.character.update(passable); err != nil {
		return err
	}
	return nil
}

func (p *player) draw(screen *ebiten.Image) error {
	if err := p.character.draw(screen); err != nil {
		return err
	}
	return nil
}
