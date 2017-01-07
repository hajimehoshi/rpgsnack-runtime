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

package character

import (
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

const playerMaxMoveCount = 4

type Player struct {
	character         *character
	movingByUserInput bool
}

func NewPlayer(x, y int) (*Player, error) {
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
	return &Player{
		character: c,
	}, nil
}

func (p *Player) IsMovingByUserInput() bool {
	return p.movingByUserInput
}

func (p *Player) MoveByUserInput(passable func(x, y int) (bool, error), x, y int) error {
	c := p.character
	path, err := calcPath(passable, c.x, c.y, x, y)
	if err != nil {
		return err
	}
	if len(path) == 0 {
		return nil
	}
	c.setRoute(path)
	p.movingByUserInput = true
	return nil
}

func (p *Player) TransferImmediately(x, y int) {
	p.character.transferImmediately(x, y)
}

func (p *Player) Update(passable func(x, y int) (bool, error)) error {
	if err := p.character.update(passable); err != nil {
		return err
	}
	if p.movingByUserInput && !p.character.isMoving() {
		p.movingByUserInput = false
	}
	return nil
}

func (p *Player) Draw(screen *ebiten.Image) error {
	if err := p.character.draw(screen); err != nil {
		return err
	}
	return nil
}
