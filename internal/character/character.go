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
	"fmt"
	"image"
	"log"
	"regexp"
	"strconv"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

const (
	PlayerEventID = -1
	frameInerval  = 60
	iconWidth     = 16
	iconHeight    = 16
)

var characterFileRegexp = regexp.MustCompile(".*_([0-9]+)_([0-9]+)")

type StoredState struct {
	speed     data.Speed
	imageName string
	imageType data.ImageType
	stepping  bool
}

type Character struct {
	eventID         int
	speed           data.Speed
	imageName       string
	imageType       data.ImageType
	dir             data.Dir
	dirFix          bool
	stepping        bool
	walking         bool
	frame           int
	steppingDir     int
	steppingCount   int
	x               int
	y               int
	idleFrameCount  int
	moveCount       int
	moveDir         data.Dir
	visible         bool
	through         bool
	erased          bool
	opacity         int
	origOpacity     int
	targetOpacity   int
	opacityCount    int
	opacityMaxCount int

	// Not dumped
	sizeW       int
	sizeH       int
	dirCount    int
	frameCount  int
	imageW      int
	imageH      int
	storedState *StoredState
}

func NewPlayer(x, y int) *Character {
	return &Character{
		eventID:       PlayerEventID,
		speed:         data.Speed3,
		imageName:     "",
		x:             x,
		y:             y,
		dir:           data.DirDown,
		dirFix:        false,
		visible:       true,
		frame:         1,
		steppingDir:   1,
		walking:       true,
		opacity:       255,
		targetOpacity: 255,
	}
}

func NewEvent(id int, x, y int) *Character {
	return &Character{
		eventID:       id,
		speed:         data.Speed3,
		x:             x,
		y:             y,
		visible:       true,
		walking:       true,
		opacity:       255,
		targetOpacity: 255,
		steppingDir:   1,
	}
}

func (c *Character) SetSizeForTesting(w, h int) {
	c.imageW = w
	c.imageH = h
}

func (c *Character) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("eventId")
	e.EncodeInt(c.eventID)

	e.EncodeString("speed")
	e.EncodeInt(int(c.speed))

	e.EncodeString("imageName")
	e.EncodeString(c.imageName)

	e.EncodeString("imageType")
	e.EncodeString(string(c.imageType))

	e.EncodeString("dir")
	e.EncodeInt(int(c.dir))

	e.EncodeString("dirFix")
	e.EncodeBool(c.dirFix)

	e.EncodeString("stepping")
	e.EncodeBool(c.stepping)

	e.EncodeString("walking")
	e.EncodeBool(c.walking)

	e.EncodeString("steppingCount")
	e.EncodeInt(c.steppingCount)

	e.EncodeString("steppingDir")
	e.EncodeInt(c.steppingDir)

	e.EncodeString("frame")
	e.EncodeInt(c.frame)

	e.EncodeString("x")
	e.EncodeInt(c.x)

	e.EncodeString("y")
	e.EncodeInt(c.y)

	e.EncodeString("moveCount")
	e.EncodeInt(c.moveCount)

	e.EncodeString("idleFrameCount")
	e.EncodeInt(c.idleFrameCount)

	e.EncodeString("moveDir")
	e.EncodeInt(int(c.moveDir))

	e.EncodeString("visible")
	e.EncodeBool(c.visible)

	e.EncodeString("through")
	e.EncodeBool(c.through)

	e.EncodeString("erased")
	e.EncodeBool(c.erased)

	e.EncodeString("opacity")
	e.EncodeInt(c.opacity)
	e.EncodeString("origOpacity")
	e.EncodeInt(c.origOpacity)
	e.EncodeString("targetOpacity")
	e.EncodeInt(c.targetOpacity)
	e.EncodeString("opacityCount")
	e.EncodeInt(c.opacityCount)
	e.EncodeString("opacityMaxCount")
	e.EncodeInt(c.opacityMaxCount)

	e.EndMap()
	return e.Flush()
}

func (c *Character) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "eventId":
			c.eventID = d.DecodeInt()
		case "speed":
			c.speed = data.Speed(d.DecodeInt())
		case "imageName":
			c.imageName = d.DecodeString()
		case "imageType":
			c.imageType = data.ImageType(d.DecodeString())
		case "dir":
			c.dir = data.Dir(d.DecodeInt())
		case "dirFix":
			c.dirFix = d.DecodeBool()
		case "stepping":
			c.stepping = d.DecodeBool()
		case "walking":
			c.walking = d.DecodeBool()
		case "steppingCount":
			c.steppingCount = d.DecodeInt()
		case "steppingDir":
			c.steppingDir = d.DecodeInt()
		case "frame":
			c.frame = d.DecodeInt()
		case "x":
			c.x = d.DecodeInt()
		case "y":
			c.y = d.DecodeInt()
		case "moveCount":
			c.moveCount = d.DecodeInt()
		case "idleFrameCount":
			c.idleFrameCount = d.DecodeInt()
		case "moveDir":
			c.moveDir = data.Dir(d.DecodeInt())
		case "visible":
			c.visible = d.DecodeBool()
		case "through":
			c.through = d.DecodeBool()
		case "erased":
			c.erased = d.DecodeBool()
		case "opacity":
			c.opacity = d.DecodeInt()
		case "origOpacity":
			c.origOpacity = d.DecodeInt()
		case "targetOpacity":
			c.targetOpacity = d.DecodeInt()
		case "opacityCount":
			c.opacityCount = d.DecodeInt()
		case "opacityMaxCount":
			c.opacityMaxCount = d.DecodeInt()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("character: Character.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (c *Character) HasStoredState() bool {
	return c.storedState != nil
}

// StoreState stores a subset of character state. It can be restored by calling
// RestoreStoredState.
func (c *Character) StoreState() {
	c.storedState = &StoredState{
		speed:     c.speed,
		imageType: c.imageType,
		imageName: c.imageName,
		stepping:  c.stepping,
	}
}

// RestoreStoredState restores the last stored state. It is useful for reviving
// the character state after a temporary change.
func (c *Character) RestoreStoredState() {
	if c.storedState == nil {
		return
	}

	c.SetSpeed(c.storedState.speed)
	c.SetImage(c.storedState.imageType, c.storedState.imageName)
	c.SetStepping(c.storedState.stepping)
	c.storedState = nil
}

func (c *Character) EventID() int {
	return c.eventID
}

func (c *Character) getImage() *ebiten.Image {
	switch c.imageType {
	case data.ImageTypeCharacters:
		return assets.GetImage("characters/" + c.imageName + ".png")
	case data.ImageTypeIcons:
		return assets.GetImage("icons/" + c.imageName + ".png")
	}

	panic("character: invalid image type:" + c.imageType)
}

func (c *Character) ImageSize() (int, int) {
	if c.imageName == "" {
		return 0, 0
	}
	if c.imageType == data.ImageTypeIcons {
		return iconWidth, iconHeight
	}
	if c.imageW == 0 || c.imageH == 0 {
		c.imageW, c.imageH = c.getImage().Size()
	}
	return c.imageW, c.imageH
}

func (c *Character) Size() (int, int) {
	if c.imageName == "" {
		return 0, 0
	}
	if c.imageType == data.ImageTypeIcons {
		return iconWidth, iconHeight
	}
	if c.sizeW == 0 || c.sizeH == 0 {
		arr := characterFileRegexp.FindStringSubmatch(c.imageName)

		if len(arr) != 3 {
			log.Printf("Invalid image is loaded: %s", c.imageName)
			return 0, 0
		}
		c.sizeW, _ = strconv.Atoi(arr[1])
		c.sizeH, _ = strconv.Atoi(arr[2])

		// Validate to see if the character size is valid
		if c.sizeW == 0 || c.sizeH == 0 || c.imageW%c.sizeW != 0 || c.imageH%c.sizeH != 0 {
			panic(fmt.Sprintf("Invalid format imageName:%s imageW:%d imageH:%d sizeW:%d sizeH:%d", c.imageName, c.imageW, c.imageH, c.sizeW, c.sizeH))
		}
	}
	return c.sizeW, c.sizeH
}

func (c *Character) DirCount() int {
	if c.imageName == "" || c.imageType == data.ImageTypeIcons {
		return 1
	}
	if c.dirCount == 0 {
		_, imageH := c.ImageSize()
		_, h := c.Size()
		if h > 0 {
			c.dirCount = imageH / h
		}
	}

	return c.dirCount
}

func (c *Character) FrameCount() int {
	if c.imageName == "" || c.imageType == data.ImageTypeIcons {
		return 1
	}
	if c.frameCount == 0 {
		imageW, _ := c.ImageSize()
		w, _ := c.Size()
		if w > 0 {
			c.frameCount = imageW / w
		}
		return imageW / w
	}

	return c.frameCount
}

func (c *Character) BaseFrame() int {
	return c.FrameCount() / 2
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

func (c *Character) DrawFootPosition() (int, int) {
	x := c.x*consts.TileSize + consts.TileSize/2
	y := (c.y + 1) * consts.TileSize
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

func (c *Character) DrawPosition() (int, int) {
	x, y := c.DrawFootPosition()
	charW, charH := c.Size()
	return x - charW/2, y - charH
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

func (c *Character) SetImage(imageType data.ImageType, imageName string) {
	c.imageName = imageName
	c.imageType = imageType
	c.imageW = 0
	c.imageH = 0
	c.sizeW = 0
	c.sizeH = 0
	c.dirCount = 0
	c.frameCount = 0
}

func (c *Character) SetFrame(frame int) {
	c.frame = frame
}

func (c *Character) SetDir(dir data.Dir) {
	c.dir = dir
}

func (c *Character) ChangeOpacity(opacity int, count int) {
	c.opacityCount = count
	c.opacityMaxCount = count
	c.targetOpacity = opacity
	c.origOpacity = c.opacity
	if count == 0 {
		c.opacity = opacity
	}
}

func (c *Character) IsChangingOpacity() bool {
	return c.opacityCount > 0
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

func (c *Character) UpdateWithPage(page *data.Page) {
	c.imageW = 0
	c.imageH = 0
	c.sizeW = 0
	c.sizeH = 0
	if page == nil {
		c.imageName = ""
		c.dirFix = false
		c.dir = data.Dir(0)
		c.frame = 0
		c.stepping = false
		c.speed = data.Speed3
		return
	}
	c.imageName = page.Image
	c.imageType = page.ImageType
	c.dirFix = page.DirFix
	c.dir = page.Dir
	c.frame = page.Frame
	c.stepping = page.Stepping
	c.walking = page.Walking
	c.through = page.Through
	c.speed = page.Speed
	c.opacity = page.Opacity
	c.origOpacity = page.Opacity
	c.targetOpacity = page.Opacity
	c.steppingCount = 0
	c.dirCount = 0
}

func (c *Character) progressFrame(speedMultiplier int) {
	c.steppingCount += c.speed.SteppingIncrementFrames() * speedMultiplier
	if c.steppingCount > frameInerval {
		c.steppingCount %= frameInerval
		c.frame += c.steppingDir
	}

	if c.frame >= c.FrameCount()-1 {
		c.steppingDir = -1
	}
	if c.frame <= 0 {
		c.steppingDir = 1
	}
}

func (c *Character) Update() {
	if c.opacityCount > 0 {
		c.opacityCount--
		rate := 1 - float64(c.opacityCount)/float64(c.opacityMaxCount)
		c.opacity = int(float64(c.origOpacity)*(1-rate) + float64(c.targetOpacity)*rate)
	} else {
		c.opacity = c.targetOpacity
	}
	if c.erased {
		return
	}
	if c.stepping {
		c.progressFrame(1)
	}
	if !c.IsMoving() {
		// Reset the character state only if it is idle for one more frame
		if c.idleFrameCount > 0 && !c.stepping && c.walking && c.steppingCount > 0 {
			c.steppingCount = 0
			c.frame = c.BaseFrame()
		}
		c.idleFrameCount++
		return
	}
	if !c.stepping && c.walking {
		c.progressFrame(2)
	}

	c.idleFrameCount = 0
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
	}
}

func (c *Character) dirToIndex(dir data.Dir) int {
	switch c.dir {
	case data.DirUp:
		return 0
	case data.DirRight:
		return 1
	case data.DirDown:
		return 2
	case data.DirLeft:
		return 3
	}

	return 0
}

func (c *Character) Draw(screen *ebiten.Image, offsetX, offsetY int) {
	if c.imageName == "" || !c.visible || c.erased {
		return
	}
	op := &ebiten.DrawImageOptions{}
	x, y := c.DrawPosition()
	op.GeoM.Translate(float64(x), float64(y))
	charW, charH := c.Size()
	dirIndex := c.dirToIndex(c.dir)

	sx := c.frame * charW
	sy := 0
	scaleX := 1.0
	scaleY := 1.0
	switch c.DirCount() {
	case 1:
	case 2:
		sy = dirIndex / 2 * charH
	case 3:
		if dirIndex == 4 {
			// Reuse the second frame and mirror it
			sy = 2 * charH
			scaleX = -1.0
		} else {
			sy = dirIndex * charH
		}
	case 4:
		sy = dirIndex * charH
	default:
		panic(fmt.Sprintf("not supported DirCount %d", c.DirCount()))
	}

	r := image.Rect(sx, sy, sx+charW, sy+charH)
	op.SourceRect = &r
	op.ColorM.Scale(1, 1, 1, float64(c.opacity)/255)
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(float64(offsetX), float64(offsetY))
	screen.DrawImage(c.getImage(), op)
}
