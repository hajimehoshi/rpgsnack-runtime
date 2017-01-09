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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type speed int

const (
	speedFastest speed = 4
	speedFast          = 8
	speedNormal        = 16
	speedSlow          = 32
	speedSlowest       = 64
)

type character struct {
	speed         speed
	imageName     string
	imageIndex    int
	dir           data.Dir
	dirFix        bool
	stepping      bool
	steppingCount int
	attitude      data.Attitude
	prevAttitude  data.Attitude
	x             int
	y             int
	moveCount     int
	moveDir       data.Dir
}

func (c *character) position() (int, int) {
	if c.moveCount > 0 {
		x, y := c.x, c.y
		switch c.moveDir {
		case data.DirLeft:
			x--
		case data.DirRight:
			x++
		case data.DirUp:
			y--
		case data.DirDown:
			y++
		}
		return x, y
	}
	return c.x, c.y
}

func (c *character) isMoving() bool {
	return c.moveCount > 0
}

func (c *character) turn(dir data.Dir) {
	if c.dirFix {
		return
	}
	c.dir = dir
}

func (c *character) move(dir data.Dir) bool {
	c.turn(dir)
	c.moveDir = dir
	// TODO: Rename this
	c.moveCount = int(c.speed)
	return true
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
		x += 2 * c.charWidth
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

func (c *character) update() error {
	if c.stepping {
		switch {
		case c.steppingCount < 15:
			c.attitude = data.AttitudeMiddle
		case c.steppingCount < 30:
			c.attitude = data.AttitudeLeft
		case c.steppingCount < 45:
			c.attitude = data.AttitudeMiddle
		default:
			c.attitude = data.AttitudeRight
		}
		c.steppingCount++
		c.steppingCount %= 60
	}
	if !c.isMoving() {
		return nil
	}
	if !c.stepping {
		if c.moveCount >= int(c.speed)/2 {
			c.attitude = data.AttitudeMiddle
		} else if c.prevAttitude == data.AttitudeLeft {
			c.attitude = data.AttitudeRight
		} else {
			c.attitude = data.AttitudeLeft
		}
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
		if !c.stepping {
			c.prevAttitude = c.attitude
			c.attitude = data.AttitudeMiddle
		}
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
		d := (int(c.speed) - c.moveCount) * scene.TileSize / int(c.speed)
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
