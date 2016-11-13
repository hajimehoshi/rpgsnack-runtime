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

	"github.com/hajimehoshi/tsugunai/internal/data"
)

type attitude int

const (
	attitudeLeft attitude = iota
	attitudeMiddle
	attitudeRight
)

const playerMaxMoveCount = 4

type player struct {
	x            int
	y            int
	path         []data.Dir
	moveCount    int
	dir          data.Dir
	attitude     attitude
	prevAttitude attitude
}

func newPlayer(x, y int) (*player, error) {
	return &player{
		x:            x,
		y:            y,
		dir:          data.DirDown,
		attitude:     attitudeMiddle,
		prevAttitude: attitudeMiddle,
	}, nil
}

func (p *player) isMoving() bool {
	return len(p.path) > 0
}

func (p *player) move(passable func(x, y int) bool, x, y int) bool {
	if p.isMoving() {
		panic("not reach")
	}
	if p.x == x && p.y == y {
		return false
	}
	p.path = calcPath(passable, p.x, p.y, x, y)
	p.moveCount = playerMaxMoveCount
	return true
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
			case data.DirLeft:
				p.x--
			case data.DirRight:
				p.x++
			case data.DirUp:
				p.y--
			case data.DirDown:
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
	case data.DirUp:
	case data.DirRight:
		y += characterSize
	case data.DirDown:
		y += 2 * characterSize
	case data.DirLeft:
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
		case data.DirLeft:
			dx -= d
		case data.DirRight:
			dx += d
		case data.DirUp:
			dy -= d
		case data.DirDown:
			dy += d
		}
		op.GeoM.Translate(float64(dx), float64(dy))
	}
	op.GeoM.Scale(tileScale, tileScale)
	op.ImageParts = &charactersImageParts{p}
	if err := screen.DrawImage(theImageCache.Get("characters0.png"), op); err != nil {
		return err
	}
	return nil
}
