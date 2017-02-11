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

const (
	PlayerEventID = -1
)

type Character struct {
	eventID       int
	speed         data.Speed
	imageName     string
	imageIndex    int
	dir           data.Dir
	dirFix        bool
	stepping      bool
	steppingCount int
	walking       bool
	walkingCount  int
	frame         int
	prevFrame     int
	x             int
	y             int
	moveCount     int
	moveDir       data.Dir
	visible       bool
	through       bool
}

func NewPlayer(x, y int) *Character {
	return &Character{
		eventID:    PlayerEventID,
		speed:      data.Speed3,
		imageName:  "characters0.png",
		imageIndex: 0,
		x:          x,
		y:          y,
		dir:        data.DirDown,
		dirFix:     false,
		visible:    true,
		frame:      1,
		prevFrame:  1,
		walking:    true,
	}
}

func NewEvent(id int, x, y int) *Character {
	return &Character{
		eventID: id,
		speed:   data.Speed3,
		x:       x,
		y:       y,
		visible: true,
		walking: true,
	}
}

func (c *Character) EventID() int {
	return c.eventID
}

func (c *Character) Size() (int, int) {
	if c.imageName == "" {
		return 0, 0
	}
	imageW, imageH := assets.GetImage(c.imageName).Size()
	w := imageW / 4 / 3
	h := imageH / 2 / 4
	return w, h
}

func (c *Character) Position() (int, int) {
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
		default:
			panic("not reach")
		}
		return x, y
	}
	return c.x, c.y
}

func (c *Character) DrawPosition() (int, int) {
	charW, charH := c.Size()
	x := c.x*scene.TileSize + scene.TileSize/2 - charW/2
	y := (c.y+1)*scene.TileSize - charH
	if c.moveCount > 0 {
		d := (c.speed.Frames() - c.moveCount) * scene.TileSize / c.speed.Frames()
		switch c.moveDir {
		case data.DirLeft:
			x -= d
		case data.DirRight:
			x += d
		case data.DirUp:
			y -= d
		case data.DirDown:
			y += d
		default:
			panic("not reach")
		}
	}
	return x, y
}

func (c *Character) Dir() data.Dir {
	return c.dir
}

func (c *Character) IsMoving() bool {
	return c.moveCount > 0
}

func (c *Character) Move(dir data.Dir) {
	c.Turn(dir)
	c.moveDir = dir
	// TODO: Rename this
	c.moveCount = c.speed.Frames()
}

func (c *Character) Turn(dir data.Dir) {
	if c.dirFix {
		return
	}
	c.dir = dir
}

func (c *Character) Speed() data.Speed {
	return c.speed
}

func (c *Character) DirFix() bool {
	return c.dirFix
}

func (c *Character) Through() bool {
	return c.through
}

func (c *Character) SetSpeed(speed data.Speed) {
	c.speed = speed
}

func (c *Character) SetVisibility(visible bool) {
	c.visible = visible
}

func (c *Character) SetDirFix(dirFix bool) {
	c.dirFix = dirFix
}

func (c *Character) SetStepping(stepping bool) {
	c.stepping = stepping
}

func (c *Character) SetWalking(walking bool) {
	c.walking = walking
}

func (c *Character) SetThrough(through bool) {
	c.through = through
}

func (c *Character) SetImage(imageName string, imageIndex int) {
	c.imageName = imageName
	c.imageIndex = imageIndex
}

func (c *Character) SetFrame(frame int) {
	c.frame = frame
	c.prevFrame = frame
}

func (c *Character) SetDir(dir data.Dir) {
	c.dir = dir
}

type characterImageParts struct {
	charWidth  int
	charHeight int
	index      int
	frame      int
	dir        data.Dir
}

func (c *characterImageParts) Len() int {
	return 1
}

func (c *characterImageParts) Src(index int) (int, int, int, int) {
	const characterXNum = 3
	const characterYNum = 4
	x := (c.index % 4) * characterXNum * c.charWidth
	y := (c.index / 4) * characterYNum * c.charHeight
	switch c.frame {
	case 0:
	case 1:
		x += c.charWidth
	case 2:
		x += 2 * c.charWidth
	default:
		panic("not reach")
	}
	switch c.dir {
	case data.DirUp:
	case data.DirRight:
		y += c.charHeight
	case data.DirDown:
		y += 2 * c.charHeight
	case data.DirLeft:
		y += 3 * c.charHeight
	default:
		panic("not reach")
	}
	return x, y, x + c.charWidth, y + c.charHeight
}

func (c *characterImageParts) Dst(index int) (int, int, int, int) {
	return 0, 0, c.charWidth, c.charHeight
}

func (c *Character) TransferImmediately(x, y int) {
	c.x = x
	c.y = y
	c.moveCount = 0
}

func (c *Character) UpdateWithPage(page *data.Page) error {
	if page == nil {
		c.imageName = ""
		c.imageIndex = 0
		c.dirFix = false
		c.dir = data.Dir(0)
		c.frame = 1
		c.stepping = false
		return nil
	}
	c.imageName = page.Image
	c.imageIndex = page.ImageIndex
	c.dirFix = page.DirFix
	c.dir = page.Dir
	c.frame = page.Frame
	c.stepping = page.Stepping
	c.through = page.Through
	return nil
}

func (c *Character) Update() error {
	if c.stepping {
		switch {
		case c.steppingCount < 15:
			c.frame = 1
		case c.steppingCount < 30:
			c.frame = 0
		case c.steppingCount < 45:
			c.frame = 1
		default:
			c.frame = 2
		}
		c.steppingCount++
		c.steppingCount %= 60
	}
	if !c.IsMoving() {
		return nil
	}
	if !c.stepping && c.walking {
		if c.walkingCount < 8 {
			c.frame = 1
		} else if c.prevFrame == 0 {
			c.frame = 2
		} else {
			c.frame = 0
		}
		c.walkingCount++
		c.walkingCount %= 16
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
		default:
			panic("not reach")
		}
		c.x = nx
		c.y = ny
		if !c.stepping && c.walking {
			c.prevFrame = c.frame
			c.frame = 1
		}
	}
	return nil
}

func (c *Character) Draw(screen *ebiten.Image) error {
	if c.imageName == "" || !c.visible {
		return nil
	}
	op := &ebiten.DrawImageOptions{}
	x, y := c.DrawPosition()
	op.GeoM.Translate(float64(x), float64(y))
	charW, charH := c.Size()
	op.ImageParts = &characterImageParts{
		charWidth:  charW,
		charHeight: charH,
		index:      c.imageIndex,
		dir:        c.dir,
		frame:      c.frame,
	}
	if err := screen.DrawImage(assets.GetImage(c.imageName), op); err != nil {
		return err
	}
	return nil
}
