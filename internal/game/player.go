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

type attitude int

const (
	attitudeLeft attitude = iota
	attitudeMiddle
	attitudeRight
)

const playerMaxMoveCount = 8

type player struct {
	x               int
	y               int
	path            []dir
	moveCount       int
	dir             dir
	attitude        attitude
	prevAttitude    attitude
	charactersImage *ebiten.Image
}

func newPlayer() (*player, error) {
	charactersImage, err := assets.LoadImage("images/characters.png", ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	return &player{
		dir:             dirDown,
		attitude:        attitudeMiddle,
		prevAttitude:    attitudeMiddle,
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
			p.dir = p.path[0]
			if p.moveCount >= playerMaxMoveCount/2 {
				p.attitude = attitudeMiddle
			} else if p.prevAttitude == attitudeLeft {
				p.attitude = attitudeRight
			} else {
				p.attitude = attitudeLeft
			}
			p.moveCount--
		}
		if p.moveCount == 0 {
			d := p.path[0]
			switch d {
			case dirLeft:
				p.x--
			case dirRight:
				p.x++
			case dirUp:
				p.y--
			case dirDown:
				p.y++
			}
			p.dir = d
			p.prevAttitude = p.attitude
			p.attitude = attitudeMiddle
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
	x := 0
	y := 0
	switch c.player.attitude {
	case attitudeLeft:
	case attitudeMiddle:
		x += characterSize
	case attitudeRight:
		x += 2 * characterSize
	}
	switch c.player.dir {
	case dirUp:
	case dirRight:
		y += characterSize
	case dirDown:
		y += 2 * characterSize
	case dirLeft:
		y += 3 * characterSize
	}
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
