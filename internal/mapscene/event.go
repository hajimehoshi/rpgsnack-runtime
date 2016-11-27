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
	"regexp"
	"strconv"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/input"
	"github.com/hajimehoshi/tsugunai/internal/scene"
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
	c.image = theImageCache.Get(page.Image)
	c.imageIndex = page.ImageIndex
	c.dirFix = page.DirFix
	c.turn(page.Dir)
	c.attitude = page.Attitude
	return nil
}

func (e *event) calcPageIndex() (int, error) {
	reSwitches := regexp.MustCompile(`^\$switches\[(\d+)\]$`)
page:
	for i := len(e.data.Pages) - 1; i >= 0; i-- {
		page := e.data.Pages[i]
		for _, cond := range page.Conditions {
			if m := reSwitches.FindStringSubmatch(cond); m != nil {
				s, err := strconv.Atoi(m[1])
				if err != nil {
					return 0, err
				}
				if s < len(e.mapScene.switches) && e.mapScene.switches[s] {
					continue
				}
			} else {
				return 0, fmt.Errorf("invalid condition: %s", cond)
			}
			continue page
		}
		return i, nil
	}
	return -1, nil
}

func (e *event) runIfActionButtonTriggered(taskLine *task.TaskLine) {
	page := e.currentPage()
	if page == nil {
		return
	}
	if page.Trigger != data.TriggerActionButton {
		return
	}
	taskLine.PushFunc(func() error {
		e.origDir = e.character.dir
		var dir data.Dir
		ex, ey := e.character.x, e.character.y
		px, py := e.mapScene.player.character.x, e.mapScene.player.character.y
		switch {
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
		e.character.attitude = page.Attitude
		e.steppingCount = 0
		e.commandIndex = newCommandIndex(page)
		return task.Terminated
	})
	taskLine.Push(task.Sub(e.goOn))
}

func (e *event) goOn(sub *task.TaskLine) error {
	if e.commandIndex == nil {
		return task.Terminated
	}
	if e.commandIndex.isTerminated() {
		sub.Push(task.Sub(func(sub *task.TaskLine) error {
			subs := []*task.TaskLine{}
			for _, b := range e.mapScene.balloons {
				if b == nil {
					continue
				}
				t := &task.TaskLine{}
				subs = append(subs, t)
				b.close(t)
				// mapScene.balloons will be cleared later.
			}
			sub.Push(task.Parallel(subs...))
			return task.Terminated
		}))
		sub.PushFunc(func() error {
			e.mapScene.balloons = nil
			e.character.turn(e.origDir)
			return task.Terminated
		})
		return task.Terminated
	}
	c := e.commandIndex.command()
	switch c.Name {
	case data.CommandNameShowMessage:
		position := data.ShowMessagePositionSelf
		if c.Args["position"] != "" {
			position = data.ShowMessagePosition(c.Args["position"])
		}
		e.showMessage(sub, c.Args["content"], position)
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameShowChoices:
		i := 0
		choices := []string{}
		for {
			choice, ok := c.Args[fmt.Sprintf("choice%d", i)]
			if !ok {
				break
			}
			choices = append(choices, choice)
			i++
		}
		e.showChoices(sub, choices)
		sub.PushFunc(func() error {
			e.commandIndex.choose(e.chosenIndex)
			return task.Terminated
		})
	case data.CommandNameSetSwitch:
		number, err := strconv.Atoi(c.Args["number"])
		if err != nil {
			return err
		}
		value := false
		switch data.SwitchValue(c.Args["value"]) {
		case data.SwitchValueFalse:
			value = false
		case data.SwitchValueTrue:
			value = true
		default:
			panic("not reach")
		}
		e.setSwitch(sub, number, value)
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case data.CommandNameMove:
		x, err := strconv.Atoi(c.Args["x"])
		if err != nil {
			return err
		}
		y, err := strconv.Atoi(c.Args["y"])
		if err != nil {
			return err
		}
		e.move(sub, x, y)
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	default:
		return fmt.Errorf("command not implemented: %s", c.Name)
	}
	return nil
}

func (e *event) showMessage(taskLine *task.TaskLine, content string, position data.ShowMessagePosition) {
	taskLine.Push(task.Sub(func(sub *task.TaskLine) error {
		subs := []*task.TaskLine{}
		for _, b := range e.mapScene.balloons {
			if b == nil {
				continue
			}
			b := b
			t := &task.TaskLine{}
			subs = append(subs, t)
			b.close(t)
			t.PushFunc(func() error {
				e.mapScene.removeBalloon(b)
				return task.Terminated
			})
		}
		sub.Push(task.Parallel(subs...))
		return task.Terminated
	}))
	taskLine.Push(task.Sub(func(sub *task.TaskLine) error {
		var b *balloon
		switch position {
		case data.ShowMessagePositionSelf:
			x := e.data.X*scene.TileSize + scene.TileSize/2 + scene.GameMarginX/scene.TileScale
			y := e.data.Y*scene.TileSize + scene.GameMarginTop/scene.TileScale
			b = newBalloonWithArrow(x, y, content)
		case data.ShowMessagePositionCenter:
			b = newBalloonCenter(content)
		default:
			return fmt.Errorf("not implemented position: %s", string(position))
		}
		e.mapScene.balloons = []*balloon{b}
		e.mapScene.balloons[0].open(sub)
		return task.Terminated
	}))
	taskLine.PushFunc(func() error {
		if input.Triggered() {
			return task.Terminated
		}
		return nil
	})
}

func (e *event) showChoices(taskLine *task.TaskLine, choices []string) {
	const height = 20
	const ymax = scene.TileYNum*scene.TileSize + (scene.GameMarginTop+scene.GameMarginBottom)/scene.TileScale
	ymin := ymax - len(choices)*height
	balloons := []*balloon{}
	taskLine.Push(task.Sub(func(sub *task.TaskLine) error {
		sub2 := []*task.TaskLine{}
		for i, choice := range choices {
			x := 0
			y := i*height + ymin
			width := scene.TileXNum * scene.TileSize
			b := newBalloon(x, y, width, height, choice)
			e.mapScene.balloons = append(e.mapScene.balloons, b)
			t := &task.TaskLine{}
			sub2 = append(sub2, t)
			b.open(t)
			balloons = append(balloons, b)
		}
		sub.Push(task.Parallel(sub2...))
		return task.Terminated
	}))
	taskLine.PushFunc(func() error {
		if !input.Triggered() {
			return nil
		}
		_, y := input.Position()
		y /= scene.TileScale
		if y < ymin || ymax <= y {
			return nil
		}
		e.chosenIndex = (y - ymin) / height
		return task.Terminated
	})
	taskLine.Push(task.Sub(func(sub *task.TaskLine) error {
		subs := []*task.TaskLine{}
		for i, b := range balloons {
			b := b
			if i == e.chosenIndex {
				continue
			}
			t := &task.TaskLine{}
			subs = append(subs, t)
			b.close(t)
			t.PushFunc(func() error {
				e.mapScene.removeBalloon(b)
				return task.Terminated
			})
		}
		sub.Push(task.Parallel(subs...))
		return task.Terminated
	}))
	taskLine.Push(task.Sleep(30))
}

func (e *event) setSwitch(taskLine *task.TaskLine, number int, value bool) {
	taskLine.PushFunc(func() error {
		if len(e.mapScene.switches) < number+1 {
			zeros := make([]bool, number+1-len(e.mapScene.switches))
			e.mapScene.switches = append(e.mapScene.switches, zeros...)
		}
		e.mapScene.switches[number] = value
		return task.Terminated
	})
}

func (e *event) move(taskLine *task.TaskLine, x, y int) {
	taskLine.PushFunc(func() error {
		e.mapScene.player.moveImmediately(x, y)
		return task.Terminated
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
