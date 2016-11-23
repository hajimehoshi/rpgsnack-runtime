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
		image:        theImageCache.Get("characters0.png"),
		imageIndex:   0,
		x:            x,
		y:            y,
		dir:          data.DirDown,
		attitude:     attitudeMiddle,
		prevAttitude: attitudeMiddle,
	}
	return &player{
		character: c,
	}, nil
}

func (p *player) move(taskLine *task.TaskLine, passable func(x, y int) bool, x, y int) {
	p.character.move(taskLine, passable, x, y, true)
}

func (p *player) update(passable func(x, y int) bool) error {
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