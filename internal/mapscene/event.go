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
	data                *data.Event
	character           *character
	currentCommandIndex int
}

func newEvent(eventData *data.Event) *event {
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
		character: c,
	}
}

func (e *event) trigger() data.Trigger {
	return e.data.Pages[0].Trigger
}

func (e *event) run(taskLine *task.TaskLine, mapScene *MapScene) {
	origDir := e.character.dir
	taskLine.PushFunc(func() error {
		var dir data.Dir
		ex, ey := e.character.x, e.character.y
		px, py := mapScene.player.character.x, mapScene.player.character.y
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
		return task.Terminated
	})
	subTaskLine := &task.TaskLine{}
	terminated := false
	taskLine.PushFunc(func() error {
		if terminated {
			return task.Terminated
		}
		if updated, err := subTaskLine.Update(); err != nil {
			return err
		} else if updated {
			return nil
		}
		page := e.data.Pages[0]
		// TODO: Consider branches
		if len(page.Commands) <= e.currentCommandIndex {
			// TODO: Better variable name
			subTaskLines := []*task.TaskLine{}
			for _, b := range mapScene.balloons {
				t := &task.TaskLine{}
				subTaskLines = append(subTaskLines, t)
				b.close(t)
			}
			subTaskLine.Push(task.Parallel(subTaskLines...))
			subTaskLine.PushFunc(func() error {
				mapScene.balloons = nil
				e.character.dir = origDir
				e.currentCommandIndex = 0
				terminated = true
				return task.Terminated
			})
			return nil
		}
		c := page.Commands[e.currentCommandIndex]
		switch c.Command {
		case "show_message":
			e.showMessage(subTaskLine, mapScene, c.Args["content"])
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
			e.showChoices(subTaskLine, mapScene, choices)
		default:
			// Ignore unknown commands so far.
			e.currentCommandIndex++
		}
		return nil
	})
}

func (e *event) showMessage(taskLine *task.TaskLine, mapScene *MapScene, content string) {
	x := e.data.X*scene.TileSize + scene.TileSize/2
	y := e.data.Y * scene.TileSize
	// TODO: Better variable name
	subTaskLines := []*task.TaskLine{}
	for _, b := range mapScene.balloons {
		t := &task.TaskLine{}
		subTaskLines = append(subTaskLines, t)
		b.close(t)
	}
	taskLine.Push(task.Parallel(subTaskLines...))
	sub := &task.TaskLine{}
	taskLine.PushFunc(func() error {
		mapScene.balloons = []*balloon{newBalloonWithArrow(x, y, content)}
		mapScene.balloons[0].open(sub)
		return task.Terminated
	})
	taskLine.PushFunc(func() error {
		if updated, err := sub.Update(); err != nil {
			return err
		} else if updated {
			return nil
		}
		return task.Terminated
	})
	taskLine.PushFunc(func() error {
		if input.Triggered() {
			return task.Terminated
		}
		return nil
	})
	taskLine.PushFunc(func() error {
		e.currentCommandIndex++
		return task.Terminated
	})
}

func (e *event) showChoices(taskLine *task.TaskLine, mapScene *MapScene, choices []string) {
	const height = 20
	dy := scene.TileYNum*scene.TileSize + scene.GameMarginBottom/scene.TileScale - len(choices)*height
	// TODO: Better variable names
	subTaskLines := []*task.TaskLine{}
	for i, choice := range choices {
		x := 0
		y := i*height + dy
		width := scene.TileXNum * scene.TileSize
		// TODO: Show balloons as parallel
		b := newBalloon(x, y, width, height, choice)
		mapScene.balloons = append(mapScene.balloons, b)
		t := &task.TaskLine{}
		subTaskLines = append(subTaskLines, t)
		b.open(t)
	}
	taskLine.Push(task.Parallel(subTaskLines...))
	taskLine.PushFunc(func() error {
		if input.Triggered() {
			return task.Terminated
		}
		return nil
	})
	taskLine.PushFunc(func() error {
		e.currentCommandIndex++
		return task.Terminated
	})
}

func (e *event) draw(screen *ebiten.Image) error {
	if err := e.character.draw(screen); err != nil {
		return err
	}
	return nil
}
