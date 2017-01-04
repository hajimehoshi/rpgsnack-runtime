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
	"fmt"

	"github.com/hajimehoshi/ebiten"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/tsugunai/internal/assets"
	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

type event struct {
	data             *data.Event
	mapScene         *MapScene
	character        *character
	origDir          data.Dir
	currentPageIndex int
	commandIndex     *commandIndex
	chosenIndex      int
	steppingCount    int
	selfSwitches     [data.SelfSwitchNum]bool
}

func newEvent(eventData *data.Event, mapScene *MapScene) (*event, error) {
	c := &character{
		x: eventData.X,
		y: eventData.Y,
	}
	e := &event{
		data:             eventData,
		mapScene:         mapScene,
		character:        c,
		currentPageIndex: -1,
	}
	if err := e.updateCharacterIfNeeded(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *event) currentPage() *data.Page {
	if e.currentPageIndex == -1 {
		return nil
	}
	return e.data.Pages[e.currentPageIndex]
}

func (e *event) isPassable() bool {
	page := e.currentPage()
	if page == nil {
		return true
	}
	return page.Priority != data.PrioritySameAsCharacters
}

func (e *event) isRunnable() bool {
	page := e.currentPage()
	if page == nil {
		return true
	}
	return len(page.Commands) > 0
}

func (e *event) updateCharacterIfNeeded() error {
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
		c.image = nil
		c.imageIndex = 0
		c.dirFix = false
		c.turn(data.Dir(0))
		c.attitude = data.AttitudeMiddle
		return nil
	}
	page := e.data.Pages[i]
	c := e.character
	c.image = assets.GetImage(page.Image)
	c.imageIndex = page.ImageIndex
	c.dirFix = page.DirFix
	c.dir = page.Dir
	// page.Attitude is ignored so far.
	c.attitude = data.AttitudeMiddle
	return nil
}

func (e *event) meetsCondition(cond *data.Condition) (bool, error) {
	// TODO: Is it OK to allow null conditions?
	if cond == nil {
		return true, nil
	}
	switch cond.Type {
	case data.ConditionTypeSwitch:
		id := cond.ID
		v := e.mapScene.switchValue(id)
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeSelfSwitch:
		v := e.selfSwitches[cond.ID]
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeVariable:
		id := cond.ID
		v := e.mapScene.variableValue(id)
		rhs := int(cond.Value.(float64))
		switch cond.Comp {
		case data.ConditionCompEqual:
			return v == rhs, nil
		case data.ConditionCompGreaterThanOrEqualTo:
			return v >= rhs, nil
		case data.ConditionCompGreaterThan:
			return v > rhs, nil
		case data.ConditionCompLessThanOrEqualTo:
			return v <= rhs, nil
		case data.ConditionCompLessThan:
			return v < rhs, nil
		default:
			return false, fmt.Errorf("mapscene: invalid comp: %s", cond.Comp)
		}
	default:
		return false, fmt.Errorf("mapscene: invalid condition: %s", cond)
	}
	return false, nil
}

func (e *event) meetsPageCondition(page *data.Page) (bool, error) {
	for _, cond := range page.Conditions {
		m, err := e.meetsCondition(cond)
		if err != nil {
			return false, err
		}
		if !m {
			return false, nil
		}
	}
	return true, nil
}

func (e *event) calcPageIndex() (int, error) {
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

func (e *event) run(taskLine *task.TaskLine, trigger data.Trigger) bool {
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
	taskLine.PushFunc(func() error {
		e.origDir = e.character.dir
		var dir data.Dir
		ex, ey := e.character.x, e.character.y
		px, py := e.mapScene.player.character.x, e.mapScene.player.character.y
		switch {
		case trigger == data.TriggerAuto:
		case ex == px && ey == py:
			// The player and the event are at the same position.
		case ex > px && ey == py:
			dir = data.DirLeft
		case ex < px && ey == py:
			dir = data.DirRight
		case ex == px && ey > py:
			dir = data.DirUp
		case ex == px && ey < py:
			dir = data.DirDown
		default:
			panic("not reach")
		}
		e.character.turn(dir)
		page := e.data.Pages[e.currentPageIndex]
		if page == nil {
			e.commandIndex = nil
			return task.Terminated
		}
		// page.Attitude is ignored so far.
		e.character.attitude = data.AttitudeMiddle
		e.steppingCount = 0
		e.commandIndex = newCommandIndex(page)
		return task.Terminated
	})
	taskLine.Push(task.Sub(e.goOn))
	return true
}

func (e *event) goOn(sub *task.TaskLine) error {
	if e.commandIndex == nil {
		return task.Terminated
	}
	e.mapScene.closeAllBalloons(sub)
	if e.commandIndex.isTerminated() {
		sub.PushFunc(func() error {
			e.character.turn(e.origDir)
			return task.Terminated
		})
		return task.Terminated
	}
	c := e.commandIndex.command()
	switch c.Name {
	case data.CommandNameIf:
		println("not implemented yet")
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameWait:
		time := int(c.Args["time"].(float64))
		frames := time * 6
		sub.Push(task.Sleep(frames))
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameCallEvent:
		println("not implemented yet")
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameShowMessage:
		eventID := int(c.Args["eventId"].(float64))
		contentID, err := data.UUIDFromString(c.Args["content"].(string))
		if err != nil {
			return err
		}
		content := e.mapScene.gameData.Texts.Get(language.Und, contentID)
		e.showMessage(sub, content, eventID)
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameShowChoices:
		choices := []string{}
		for _, c := range c.Args["choices"].([]interface{}) {
			id, err := data.UUIDFromString(c.(string))
			if err != nil {
				return err
			}
			choice := e.mapScene.gameData.Texts.Get(language.Und, id)
			choices = append(choices, choice)
		}
		e.showChoices(sub, choices)
		sub.PushFunc(func() error {
			e.commandIndex.choose(e.chosenIndex)
			return task.Terminated
		})
	case data.CommandNameSetSwitch:
		number := int(c.Args["id"].(float64))
		value := c.Args["value"].(bool)
		e.setSwitch(sub, number, value)
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameSetSelfSwitch:
		number := int(c.Args["id"].(float64))
		value := c.Args["value"].(bool)
		e.setSelfSwitch(sub, number, value)
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameTransfer:
		roomID := int(c.Args["roomId"].(float64))
		x := int(c.Args["x"].(float64))
		y := int(c.Args["y"].(float64))
		e.transfer(sub, roomID, x, y)
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameSetRoute:
		println("not implemented yet")
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameTintScreen:
		r := c.Args["red"].(float64) / 255
		g := c.Args["green"].(float64) / 255
		b := c.Args["blue"].(float64) / 255
		gray := c.Args["gray"].(float64) / 255
		time := int(c.Args["time"].(float64))
		wait := c.Args["wait"].(bool)
		sub.PushFunc(func() error {
			e.mapScene.startTint(r, g, b, gray, time*6)
			return task.Terminated
		})
		sub.PushFunc(func() error {
			if wait && e.mapScene.isChangingTint() {
				return nil
			}
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNamePlaySE:
		println("not implemented yet")
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNamePlayBGM:
		println("not implemented yet")
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameStopBGM:
		println("not implemented yet")
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	default:
		return fmt.Errorf("command not implemented: %s", c.Name)
	}
	return nil
}

func (e *event) showMessage(taskLine *task.TaskLine, content string, eventID int) {
	var ch *character
	switch eventID {
	case -1:
		ch = e.mapScene.player.character
	case 0:
		ch = e.character
	default:
		panic("not implemented")
	}
	e.mapScene.showMessage(taskLine, content, ch)
}

func (e *event) showChoices(taskLine *task.TaskLine, choices []string) {
	e.mapScene.showChoices(taskLine, choices, func(index int) {
		e.chosenIndex = index
	})
}

func (e *event) setSwitch(taskLine *task.TaskLine, number int, value bool) {
	taskLine.PushFunc(func() error {
		e.mapScene.setSwitchValue(number, value)
		return task.Terminated
	})
}

func (e *event) setSelfSwitch(taskLine *task.TaskLine, number int, value bool) {
	taskLine.PushFunc(func() error {
		e.selfSwitches[number] = value
		return task.Terminated
	})
}

func (e *event) transfer(taskLine *task.TaskLine, roomID, x, y int) {
	count := 0
	const maxCount = 30
	taskLine.PushFunc(func() error {
		count++
		e.mapScene.fadingRate = float64(count) / maxCount
		if count == maxCount {
			return task.Terminated
		}
		return nil
	})
	taskLine.PushFunc(func() error {
		e.mapScene.player.transferImmediately(x, y)
		return task.Terminated
	})
	taskLine.PushFunc(func() error {
		count--
		e.mapScene.fadingRate = float64(count) / maxCount
		if count == 0 {
			return task.Terminated
		}
		return nil
	})
}

func (e *event) update() error {
	page := e.currentPage()
	if page == nil {
		return nil
	}
	if !page.Stepping {
		return nil
	}
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
	return nil
}

func (e *event) draw(screen *ebiten.Image) error {
	if err := e.character.draw(screen); err != nil {
		return err
	}
	return nil
}
