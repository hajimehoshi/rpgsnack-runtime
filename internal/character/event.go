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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type Event struct {
	data      *data.Event
	character *character
}

func NewEvent(eventData *data.Event) (*Event, error) {
	c := &character{
		id:      eventData.ID,
		speed:   data.Speed3,
		x:       eventData.X,
		y:       eventData.Y,
		visible: true,
	}
	e := &Event{
		data:      eventData,
		character: c,
	}
	return e, nil
}

func (e *Event) Data() *data.Event {
	return e.data
}

func (e *Event) ID() int {
	return e.character.id
}

func (e *Event) Size() (int, int) {
	return e.character.size()
}

func (e *Event) Position() (int, int) {
	return e.character.position()
}

func (e *Event) DrawPosition() (int, int) {
	return e.character.drawPosition()
}

func (e *Event) Dir() data.Dir {
	return e.character.dir
}

func (e *Event) IsMoving() bool {
	return e.character.isMoving()
}

func (e *Event) Move(dir data.Dir) {
	e.character.move(dir)
}

func (e *Event) Turn(dir data.Dir) {
	e.character.turn(dir)
}

func (e *Event) SetSpeed(speed data.Speed) {
	e.character.speed = speed
}

func (e *Event) SetVisibility(visible bool) {
	e.character.visible = visible
}

func (e *Event) SetDirFix(dirFix bool) {
	e.character.dirFix = dirFix
}

func (e *Event) SetStepping(stepping bool) {
	e.character.stepping = stepping
}

func (e *Event) SetWalking(walking bool) {
	e.character.walking = walking
}

func (e *Event) SetImage(imageName string, imageIndex int, frame int, dir data.Dir, useFrameAndDir bool) {
	e.character.imageName = imageName
	e.character.imageIndex = imageIndex
	if useFrameAndDir {
		e.character.dir = dir
		e.character.frame = frame
		e.character.prevFrame = frame
	}
}

func (e *Event) UpdateCharacterIfNeeded(index int) error {
	if index == -1 {
		c := e.character
		c.imageName = ""
		c.imageIndex = 0
		c.dirFix = false
		c.dir = data.Dir(0)
		c.frame = 1
		c.stepping = false
		return nil
	}
	page := e.data.Pages[index]
	c := e.character
	c.imageName = page.Image
	c.imageIndex = page.ImageIndex
	c.dirFix = page.DirFix
	c.dir = page.Dir
	c.frame = page.Frame
	c.stepping = page.Stepping
	return nil
}

func (e *Event) Update() error {
	if err := e.character.update(); err != nil {
		return err
	}
	return nil
}

func (e *Event) Draw(screen *ebiten.Image) error {
	if err := e.character.draw(screen); err != nil {
		return err
	}
	return nil
}
