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

package game

import (
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/assets"
)

const playerMaxMoveCount = 4

type player struct {
	x               int
	y               int
	path            []dir
	moveCount       int
	charactersImage *ebiten.Image
}

func newPlayer() (*player, error) {
	charactersImage, err := assets.LoadImage("images/characters.png", ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	return &player{
		charactersImage: charactersImage,
	}, nil
}

func (p *player) isMoving() bool {
	return len(p.path) > 0
}

func passable(x, y int) bool {
	if x < 0 {
		return false
	}
	if y < 0 {
		return false
	}
	if tileXNum <= x {
		return false
	}
	if tileYNum <= y {
		return false
	}
	return true
}

func (p *player) move(x, y int) {
	if p.isMoving() {
		panic("not reach")
	}
	if p.x == x && p.y == y {
		return
	}
	p.path = calcPath(passable, p.x, p.y, x, y)
	p.moveCount = playerMaxMoveCount
}

func (p *player) update() error {
	if len(p.path) > 0 {
		if p.moveCount > 0 {
			p.moveCount--
		}
		if p.moveCount == 0 {
			switch p.path[0] {
			case dirLeft:
				p.x--
			case dirRight:
				p.x++
			case dirUp:
				p.y--
			case dirDown:
				p.y++
			}
			p.path = p.path[1:]
			if len(p.path) > 0 {
				p.moveCount = playerMaxMoveCount
			}
		}
	}
	return nil
}

type charactersImageParts struct {
	player *player
}

func (c *charactersImageParts) Len() int {
	return 1
}

func (c *charactersImageParts) Src(index int) (int, int, int, int) {
	x := characterSize
	y := characterSize * 2
	return x, y, x + characterSize, y + characterSize
}

func (c *charactersImageParts) Dst(index int) (int, int, int, int) {
	x := c.player.x * tileSize
	y := c.player.y * tileSize
	return x, y, x + characterSize, y + characterSize
}

func (p *player) draw(screen *ebiten.Image) error {
	op := &ebiten.DrawImageOptions{}
	if p.isMoving() {
		dx := 0
		dy := 0
		d := (playerMaxMoveCount - p.moveCount) * tileSize / playerMaxMoveCount
		switch p.path[0] {
		case dirLeft:
			dx -= d
		case dirRight:
			dx += d
		case dirUp:
			dy -= d
		case dirDown:
			dy += d
		}
		op.GeoM.Translate(float64(dx), float64(dy))
	}
	op.GeoM.Scale(tileScale, tileScale)
	op.ImageParts = &charactersImageParts{p}
	if err := screen.DrawImage(p.charactersImage, op); err != nil {
		return err
	}
	return nil
}
