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

	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/input"
	"github.com/hajimehoshi/tsugunai/internal/scene"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

type event struct {
	data         *data.Event
	mapScene     *MapScene
	character    *character
	origDir      data.Dir
	commandIndex *commandIndex
	chosenIndex  int
}

func newEvent(eventData *data.Event, mapScene *MapScene) *event {
	page := eventData.Pages[0]
	c := &character{
		image:      theImageCache.Get(page.Image),
		imageIndex: page.ImageIndex,
		dir:        page.Dir,
		attitude:   attitudeMiddle,
		x:          eventData.X,
		y:          eventData.Y,
	}
	return &event{
		data:      eventData,
		mapScene:  mapScene,
		character: c,
	}
}

func (e *event) trigger() data.Trigger {
	return e.data.Pages[0].Trigger
}

func (e *event) run(taskLine *task.TaskLine) {
	taskLine.PushFunc(func() error {
		e.origDir = e.character.dir
		var dir data.Dir
		ex, ey := e.character.x, e.character.y
		px, py := e.mapScene.player.character.x, e.mapScene.player.character.y
		switch {
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
		e.character.dir = dir
		page := e.data.Pages[0]
		e.commandIndex = newCommandIndex(page)
		return task.Terminated
	})
	taskLine.Push(task.Sub(e.goOn))
}

func (e *event) goOn(sub *task.TaskLine) error {
	if e.commandIndex.isTerminated() {
		sub.Push(task.CreateTaskLazily(func() (task.Task, error) {
			sub := []*task.TaskLine{}
			for _, b := range e.mapScene.balloons {
				if b == nil {
					continue
				}
				t := &task.TaskLine{}
				sub = append(sub, t)
				b.close(t)
				// mapScene.balloons will be cleared later.
			}
			return task.Parallel(sub...), nil
		}))
		sub.PushFunc(func() error {
			e.mapScene.balloons = nil
			e.character.dir = e.origDir
			return task.Terminated
		})
		return task.Terminated
	}
	c := e.commandIndex.command()
	switch c.Command {
	case "show_message":
		position := data.ShowMessagePositionSelf
		if c.Args["position"] != "" {
			position = data.ShowMessagePosition(c.Args["position"])
		}
		e.showMessage(sub, c.Args["content"], position)
		sub.PushFunc(func() error {
			e.commandIndex.advance()
			return task.Terminated
		})
	case "show_choices":
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
	default:
		return fmt.Errorf("command not implemented: %s", c.Command)
	}
	return nil
}

func (e *event) showMessage(taskLine *task.TaskLine, content string, position data.ShowMessagePosition) {
	taskLine.Push(task.CreateTaskLazily(func() (task.Task, error) {
		sub := []*task.TaskLine{}
		for _, b := range e.mapScene.balloons {
			if b == nil {
				continue
			}
			b := b
			t := &task.TaskLine{}
			sub = append(sub, t)
			b.close(t)
			t.PushFunc(func() error {
				e.mapScene.removeBalloon(b)
				return task.Terminated
			})
		}
		return task.Parallel(sub...), nil
	}))
	taskLine.Push(task.CreateTaskLazily(func() (task.Task, error) {
		sub := &task.TaskLine{}
		var b *balloon
		switch position {
		case data.ShowMessagePositionSelf:
			x := e.data.X*scene.TileSize + scene.TileSize/2 + scene.GameMarginX/scene.TileScale
			y := e.data.Y*scene.TileSize + scene.GameMarginTop/scene.TileScale
			b = newBalloonWithArrow(x, y, content)
		case data.ShowMessagePositionCenter:
			b = newBalloonCenter(content)
		default:
			return nil, fmt.Errorf("not implemented position: %s", string(position))
		}
		e.mapScene.balloons = []*balloon{b}
		e.mapScene.balloons[0].open(sub)
		return sub.ToTask(), nil
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
	taskLine.Push(task.CreateTaskLazily(func() (task.Task, error) {
		sub := []*task.TaskLine{}
		for i, choice := range choices {
			x := 0
			y := i*height + ymin
			width := scene.TileXNum * scene.TileSize
			b := newBalloon(x, y, width, height, choice)
			e.mapScene.balloons = append(e.mapScene.balloons, b)
			t := &task.TaskLine{}
			sub = append(sub, t)
			b.open(t)
			balloons = append(balloons, b)
		}
		return task.Parallel(sub...), nil
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
	taskLine.Push(task.CreateTaskLazily(func() (task.Task, error) {
		sub := []*task.TaskLine{}
		for i, b := range balloons {
			b := b
			if i == e.chosenIndex {
				continue
			}
			t := &task.TaskLine{}
			sub = append(sub, t)
			b.close(t)
			t.PushFunc(func() error {
				e.mapScene.removeBalloon(b)
				return task.Terminated
			})
		}
		return task.Parallel(sub...), nil
	}))
	taskLine.Push(task.Sleep(30))
}

func (e *event) draw(screen *ebiten.Image) error {
	if err := e.character.draw(screen); err != nil {
		return err
	}
	return nil
}
