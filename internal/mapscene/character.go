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

	"github.com/hajimehoshi/tsugunai/internal/assets"
	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/scene"
)

type character struct {
	imageName    string
	imageIndex   int
	dir          data.Dir
	dirFix       bool
	attitude     data.Attitude
	prevAttitude data.Attitude
	x            int
	y            int
	moveCount    int
	moveDir      data.Dir
	route        []data.RouteCommand
}

func (c *character) isMoving() bool {
	return c.moveCount > 0 || len(c.route) > 0
}

func (c *character) turn(dir data.Dir) {
	if c.dirFix {
		return
	}
	c.dir = dir
}

func (c *character) setRoute(route []data.RouteCommand) {
	if c.isMoving() {
		return
	}
	c.route = route
	c.consumeRoute()
}

func (c *character) consumeRoute() {
	for len(c.route) > 0 {
		switch c.route[0] {
		case data.RouteCommandMoveUp:
			c.move(data.DirUp)
			return
		case data.RouteCommandMoveRight:
			c.move(data.DirRight)
			return
		case data.RouteCommandMoveDown:
			c.move(data.DirDown)
			return
		case data.RouteCommandMoveLeft:
			c.move(data.DirLeft)
			return
		case data.RouteCommandTurnUp:
			c.turn(data.DirUp)
		case data.RouteCommandTurnRight:
			c.turn(data.DirRight)
		case data.RouteCommandTurnDown:
			c.turn(data.DirDown)
		case data.RouteCommandTurnLeft:
			c.turn(data.DirLeft)
		}
		c.route = c.route[1:]
	}
}

func (c *character) move(dir data.Dir) (bool, error) {
	c.turn(dir)
	c.moveDir = dir
	// TODO: Rename this
	c.moveCount = playerMaxMoveCount
	// TODO: Check passability
	return true, nil
}

type characterImageParts struct {
	charWidth  int
	charHeight int
	index      int
	dir        data.Dir
	attitude   data.Attitude
}

func (c *characterImageParts) Len() int {
	return 1
}

func (c *characterImageParts) Src(index int) (int, int, int, int) {
	x := ((c.index % 4) * 3) * c.charWidth
	y := (c.index / 4) * 2 * c.charHeight
	switch c.attitude {
	case data.AttitudeLeft:
	case data.AttitudeMiddle:
		x += c.charWidth
	case data.AttitudeRight:
		x += 2 * c.charHeight
	}
	switch c.dir {
	case data.DirUp:
	case data.DirRight:
		y += c.charHeight
	case data.DirDown:
		y += 2 * c.charHeight
	case data.DirLeft:
		y += 3 * c.charHeight
	}
	return x, y, x + c.charWidth, y + c.charHeight
}

func (c *characterImageParts) Dst(index int) (int, int, int, int) {
	return 0, 0, c.charWidth, c.charHeight
}

func (c *character) transferImmediately(x, y int) {
	c.x = x
	c.y = y
	c.moveCount = 0
}

func (c *character) update(passable func(x, y int) (bool, error)) error {
	if !c.isMoving() {
		return nil
	}
	if c.moveCount >= playerMaxMoveCount/2 {
		c.attitude = data.AttitudeMiddle
	} else if c.prevAttitude == data.AttitudeLeft {
		c.attitude = data.AttitudeRight
	} else {
		c.attitude = data.AttitudeLeft
	}
	c.moveCount--
	if c.moveCount == 0 {
		nx, ny := c.x, c.y
		switch c.moveDir {
		case data.DirLeft:
			nx--
		case data.DirRight:
			nx++
		case data.DirUp:
			ny--
		case data.DirDown:
			ny++
		}
		c.x = nx
		c.y = ny
		c.prevAttitude = c.attitude
		c.attitude = data.AttitudeMiddle
		if len(c.route) > 0 {
			c.route = c.route[1:]
		}
		c.consumeRoute()
	}
	return nil
}

func (c *character) draw(screen *ebiten.Image) error {
	if c.imageName == "" {
		return nil
	}
	imageW, imageH := assets.GetImage(c.imageName).Size()
	charW := imageW / 4 / 3
	charH := imageH / 2 / 4
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(c.x*scene.TileSize+scene.TileSize/2), float64((c.y+1)*scene.TileSize))
	op.GeoM.Translate(float64(-charW/2), float64(-charH))
	if c.moveCount > 0 {
		dx := 0
		dy := 0
		d := (playerMaxMoveCount - c.moveCount) * scene.TileSize / playerMaxMoveCount
		switch c.dir {
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
	op.ImageParts = &characterImageParts{
		charWidth:  charW,
		charHeight: charH,
		index:      c.imageIndex,
		dir:        c.dir,
		attitude:   c.attitude,
	}
	if err := screen.DrawImage(assets.GetImage(c.imageName), op); err != nil {
		return err
	}
	return nil
}
