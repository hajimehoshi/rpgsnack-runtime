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
	character *character
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

func (p *Player) Position() (int, int) {
	return p.character.position()
}

func (p *Player) Dir() data.Dir {
	return p.character.dir
}

func (p *Player) IsMoving() bool {
	return p.character.isMoving()
}

func (p *Player) Move(dir data.Dir) {
	p.character.move(dir)
}

func (p *Player) Turn(dir data.Dir) {
	p.character.turn(dir)
}

func (p *Player) TransferImmediately(x, y int) {
	p.character.transferImmediately(x, y)
}

func (p *Player) Update() error {
	if err := p.character.update(); err != nil {
		return err
	}
	return nil
}

func (p *Player) Draw(screen *ebiten.Image) error {
	if err := p.character.draw(screen); err != nil {
		return err
	}
	return nil
}
