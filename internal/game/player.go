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

type player struct {
	x               int
	y               int
	nextX           int
	nextY           int
	charactersImage *ebiten.Image
}

func newPlayer() (*player, error) {
	charactersImage, err := assets.LoadImage("images/characters.png", ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	return &player{
		x:               0,
		y:               0,
		nextX:           0,
		nextY:           0,
		charactersImage: charactersImage,
	}, nil
}

func (p *player) move(x, y int) {
	p.nextX = x
	p.nextY = y
}

func (p *player) update() error {
	if p.x != p.nextX || p.y != p.nextY {
		p.x = p.nextX
		p.y = p.nextY
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
	op.GeoM.Scale(tileScale, tileScale)
	op.ImageParts = &charactersImageParts{p}
	if err := screen.DrawImage(p.charactersImage, op); err != nil {
		return err
	}
	return nil
}
