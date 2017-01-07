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

type Interpreter interface {
	SetEvent(event *Event)
	IsExecuting() bool
	MeetsCondition(cond *data.Condition) (bool, error)
	SetCommands(commands []*data.Command, trigger data.Trigger)
	Update() error
}

type Event struct {
	data             *data.Event
	interpreter      Interpreter
	character        *character
	currentPageIndex int
	steppingCount    int
	dirBeforeRunning data.Dir
}

func NewEvent(eventData *data.Event, interpreter Interpreter) (*Event, error) {
	c := &character{
		x: eventData.X,
		y: eventData.Y,
	}
	e := &Event{
		data:             eventData,
		interpreter:      interpreter,
		character:        c,
		currentPageIndex: -1,
	}
	e.interpreter.SetEvent(e)
	if err := e.UpdateCharacterIfNeeded(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Event) ID() int {
	return e.data.ID
}

func (e *Event) Position() (int, int) {
	return e.character.x, e.character.y
}

func (e *Event) Dir() data.Dir {
	return e.character.dir
}

func (e *Event) StartEvent(dir data.Dir) {
	e.dirBeforeRunning = e.character.dir
	e.character.turn(dir)
	// page.Attitude is ignored so far.
	e.character.attitude = data.AttitudeMiddle
	e.steppingCount = 0
}

func (e *Event) EndEvent() {
	e.character.turn(e.dirBeforeRunning)
}

func (e *Event) currentPage() *data.Page {
	if e.currentPageIndex == -1 {
		return nil
	}
	return e.data.Pages[e.currentPageIndex]
}

func (e *Event) IsPassable() bool {
	page := e.currentPage()
	if page == nil {
		return true
	}
	return page.Priority != data.PrioritySameAsCharacters
}

func (e *Event) IsRunnable() bool {
	page := e.currentPage()
	if page == nil {
		return true
	}
	return len(page.Commands) > 0
}

func (e *Event) IsExecutingCommands() bool {
	return e.interpreter.IsExecuting()
}

func (e *Event) UpdateCharacterIfNeeded() error {
	i, err := e.calcPageIndex()
	if err != nil {
		return err
	}
	if e.currentPageIndex == i {
		return nil
	}
	e.currentPageIndex = i
	e.steppingCount = 0
	if i == -1 {
		c := e.character
		c.imageName = ""
		c.imageIndex = 0
		c.dirFix = false
		c.dir = data.Dir(0)
		c.attitude = data.AttitudeMiddle
		return nil
	}
	page := e.data.Pages[i]
	c := e.character
	c.imageName = page.Image
	c.imageIndex = page.ImageIndex
	c.dirFix = page.DirFix
	c.dir = page.Dir
	// page.Attitude is ignored so far.
	c.attitude = data.AttitudeMiddle
	return nil
}

func (e *Event) meetsPageCondition(page *data.Page) (bool, error) {
	for _, cond := range page.Conditions {
		m, err := e.interpreter.MeetsCondition(cond)
		if err != nil {
			return false, err
		}
		if !m {
			return false, nil
		}
	}
	return true, nil
}

func (e *Event) calcPageIndex() (int, error) {
	for i := len(e.data.Pages) - 1; i >= 0; i-- {
		page := e.data.Pages[i]
		m, err := e.meetsPageCondition(page)
		if err != nil {
			return 0, err
		}
		if m {
			return i, nil
		}
	}
	return -1, nil
}

func (e *Event) TryRun(trigger data.Trigger) bool {
	if e.interpreter.IsExecuting() {
		return false
	}
	if trigger == data.TriggerNever {
		return false
	}
	page := e.currentPage()
	if page == nil {
		return false
	}
	if page.Trigger != trigger {
		return false
	}
	e.interpreter.SetCommands(page.Commands, page.Trigger)
	return true
}

func (e *Event) Update() error {
	if !e.interpreter.IsExecuting() {
		page := e.currentPage()
		if page == nil {
			return nil
		}
		if page.Stepping {
			switch {
			case e.steppingCount < 30:
				e.character.attitude = data.AttitudeMiddle
			case e.steppingCount < 60:
				e.character.attitude = data.AttitudeLeft
			case e.steppingCount < 90:
				e.character.attitude = data.AttitudeMiddle
			default:
				e.character.attitude = data.AttitudeRight
			}
			e.steppingCount++
			e.steppingCount %= 120
		}
	}
	if err := e.interpreter.Update(); err != nil {
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
