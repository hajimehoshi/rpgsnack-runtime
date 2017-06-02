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
	"encoding/json"
	"image"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
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
	erased        bool
}

func NewPlayer(x, y int) *Character {
	return &Character{
		eventID:    PlayerEventID,
		speed:      data.Speed3,
		imageName:  "",
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

type tmpCharacter struct {
	EventID       int        `json:"eventId"`
	Speed         data.Speed `json:"speed"`
	ImageName     string     `json:"imageName"`
	ImageIndex    int        `json:"imageIndex"`
	Dir           data.Dir   `json:"dir"`
	DirFix        bool       `json:"dirFix"`
	Stepping      bool       `json:"stepping"`
	SteppingCount int        `json:"steppingCount"`
	Walking       bool       `json:"walking"`
	WalkingCount  int        `json:"walkingCount"`
	Frame         int        `json:"frame"`
	PrevFrame     int        `json:"prevFrame"`
	X             int        `json:"x"`
	Y             int        `json:"y"`
	MoveCount     int        `json:"moveCount"`
	MoveDir       data.Dir   `json:"moveDir"`
	Visible       bool       `json:"visible"`
	Through       bool       `json:"through"`
	Erased        bool       `json:"erased"`
}

func (c *Character) MarshalJSON() ([]uint8, error) {
	tmp := &tmpCharacter{
		EventID:       c.eventID,
		Speed:         c.speed,
		ImageName:     c.imageName,
		ImageIndex:    c.imageIndex,
		Dir:           c.dir,
		DirFix:        c.dirFix,
		Stepping:      c.stepping,
		SteppingCount: c.steppingCount,
		Walking:       c.walking,
		WalkingCount:  c.walkingCount,
		Frame:         c.frame,
		PrevFrame:     c.prevFrame,
		X:             c.x,
		Y:             c.y,
		MoveCount:     c.moveCount,
		MoveDir:       c.moveDir,
		Visible:       c.visible,
		Through:       c.through,
		Erased:        c.erased,
	}
	return json.Marshal(tmp)
}

func (c *Character) UnmarshalJSON(data []uint8) error {
	var tmp *tmpCharacter
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	c.eventID = tmp.EventID
	c.speed = tmp.Speed
	c.imageName = tmp.ImageName
	c.imageIndex = tmp.ImageIndex
	c.dir = tmp.Dir
	c.dirFix = tmp.DirFix
	c.stepping = tmp.Stepping
	c.steppingCount = tmp.SteppingCount
	c.walking = tmp.Walking
	c.walkingCount = tmp.WalkingCount
	c.frame = tmp.Frame
	c.prevFrame = tmp.PrevFrame
	c.x = tmp.X
	c.y = tmp.Y
	c.moveCount = tmp.MoveCount
	c.moveDir = tmp.MoveDir
	c.visible = tmp.Visible
	c.through = tmp.Through
	c.erased = tmp.Erased
	return nil
}

func (c *Character) EventID() int {
	return c.eventID
}

func (c *Character) Size() (int, int) {
	if c.imageName == "" {
		return 0, 0
	}
	imageW, imageH := assets.GetImage("characters/" + c.imageName + ".png").Size()
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
	x := c.x*consts.TileSize + consts.TileSize/2 - charW/2
	y := (c.y+1)*consts.TileSize - charH
	if c.moveCount > 0 {
		d := (c.speed.Frames() - c.moveCount) * consts.TileSize / c.speed.Frames()
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
	return c.through || c.erased
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

func (c *Character) TransferImmediately(x, y int) {
	c.x = x
	c.y = y
	c.moveCount = 0
}

func (c *Character) Erase() {
	c.erased = true
}

func (c *Character) Erased() bool {
	return c.erased
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
	c.walking = page.Walking
	c.through = page.Through
	return nil
}

func (c *Character) Update() error {
	if c.erased {
		return nil
	}
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

func (c *Character) Draw(screen *ebiten.Image) {
	if c.imageName == "" || !c.visible || c.erased {
		return
	}
	op := &ebiten.DrawImageOptions{}
	x, y := c.DrawPosition()
	op.GeoM.Translate(float64(x), float64(y))
	charW, charH := c.Size()
	const characterXNum = 3
	const characterYNum = 4
	sx := (c.imageIndex % 4) * characterXNum * charW
	sy := (c.imageIndex / 4) * characterYNum * charH
	switch c.frame {
	case 0:
	case 1:
		sx += charW
	case 2:
		sx += 2 * charW
	default:
		panic("not reached")
	}
	switch c.dir {
	case data.DirUp:
	case data.DirRight:
		sy += charH
	case data.DirDown:
		sy += 2 * charH
	case data.DirLeft:
		sy += 3 * charH
	default:
		panic("not reached")
	}
	r := image.Rect(sx, sy, sx+charW, sy+charH)
	op.SourceRect = &r
	screen.DrawImage(assets.GetImage("characters/"+c.imageName+".png"), op)
}
