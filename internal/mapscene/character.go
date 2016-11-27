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
	"github.com/hajimehoshi/tsugunai/internal/scene"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

type attitude int

const (
	attitudeLeft attitude = iota
	attitudeRight
	attitudeMiddle
)

type character struct {
	image        *ebiten.Image
	imageIndex   int
	dir          data.Dir
	attitude     attitude
	prevAttitude attitude
	x            int
	y            int
	moveCount    int
}

type characterImageParts struct {
	charWidth  int
	charHeight int
	index      int
	dir        data.Dir
	attitude   attitude
}

func (c *characterImageParts) Len() int {
	return 1
}

func (c *characterImageParts) Src(index int) (int, int, int, int) {
	x := ((c.index % 4) * 3) * c.charWidth
	y := (c.index / 4) * 2 * c.charHeight
	switch c.attitude {
	case attitudeLeft:
	case attitudeMiddle:
		x += c.charWidth
	case attitudeRight:
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

func (c *character) move(taskLine *task.TaskLine, passable func(x, y int) bool, x, y int, player bool) {
	path := calcPath(passable, c.x, c.y, x, y)
	for _, d := range path {
		d := d
		init := false
		taskLine.PushFunc(func() error {
			if !init {
				c.dir = d
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
			if !passable(nx, ny) {
				nx = c.x
				ny = c.y
				c.moveCount = 0
				return task.Terminated
			}
			if c.moveCount > 0 {
				if c.moveCount >= playerMaxMoveCount/2 {
					c.attitude = attitudeMiddle
				} else if c.prevAttitude == attitudeLeft {
					c.attitude = attitudeRight
				} else {
					c.attitude = attitudeLeft
				}
				c.moveCount--
			}
			if c.moveCount == 0 {
				c.x = nx
				c.y = ny
				c.prevAttitude = c.attitude
				c.attitude = attitudeMiddle
				return task.Terminated
			}
			return nil
		})
	}
}

func (c *character) update(passable func(x, y int) bool) error {
	return nil
}

func (c *character) draw(screen *ebiten.Image) error {
	if c.image == nil {
		return nil
	}
	imageW, imageH := c.image.Size()
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
	if err := screen.DrawImage(c.image, op); err != nil {
		return err
	}
	return nil
}
