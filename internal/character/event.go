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
	data             *data.Event
	character        *character
	currentPageIndex int
	steppingCount    int
}

func NewEvent(eventData *data.Event) (*Event, error) {
	c := &character{
		speed: speedNormal,
		x:     eventData.X,
		y:     eventData.Y,
	}
	e := &Event{
		data:             eventData,
		character:        c,
		currentPageIndex: -1,
	}
	return e, nil
}

func (e *Event) ID() int {
	return e.data.ID
}

func (e *Event) Size() (int, int) {
	return e.character.size()
}

func (e *Event) Position() (int, int) {
	return e.character.position()
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

func (e *Event) CurrentPage() *data.Page {
	if e.currentPageIndex == -1 {
		return nil
	}
	return e.data.Pages[e.currentPageIndex]
}

func (e *Event) CurrentPageIndex() int {
	return e.currentPageIndex
}

func (e *Event) IsPassable() bool {
	page := e.CurrentPage()
	if page == nil {
		return true
	}
	return page.Priority != data.PrioritySameAsCharacters
}

func (e *Event) IsRunnable() bool {
	page := e.CurrentPage()
	if page == nil {
		return true
	}
	return len(page.Commands) > 0
}

func (e *Event) UpdateCharacterIfNeeded(index int) (bool, error) {
	if e.currentPageIndex == index {
		return false, nil
	}
	e.currentPageIndex = index
	e.steppingCount = 0
	if index == -1 {
		c := e.character
		c.imageName = ""
		c.imageIndex = 0
		c.dirFix = false
		c.dir = data.Dir(0)
		c.attitude = data.AttitudeMiddle
		c.stepping = false
		return true, nil
	}
	page := e.data.Pages[index]
	c := e.character
	c.imageName = page.Image
	c.imageIndex = page.ImageIndex
	c.dirFix = page.DirFix
	c.dir = page.Dir
	// page.Attitude is ignored so far.
	c.attitude = data.AttitudeMiddle
	c.stepping = page.Stepping
	return true, nil
}

func (e *Event) Update() error {
	page := e.CurrentPage()
	if page == nil {
		return nil
	}
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
