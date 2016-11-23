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

func (e *event) run(mapScene *MapScene) {
	origDir := e.character.dir
	task.Push(func() error {
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
	taskLine := &task.TaskLine{}
	terminated := false
	task.Push(func() error {
		if terminated {
			return task.Terminated
		}
		if updated, err := taskLine.Update(); err != nil {
			return err
		} else if updated {
			return nil
		}
		page := e.data.Pages[0]
		// TODO: Consider branches
		if len(page.Commands) <= e.currentCommandIndex {
			for _, b := range mapScene.balloons {
				b.close(taskLine)
			}
			mapScene.balloons = nil
			taskLine.Push(func() error {
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
			x := e.data.X*scene.TileSize + scene.TileSize/2
			y := e.data.Y * scene.TileSize
			for _, b := range mapScene.balloons {
				b.close(taskLine)
			}
			mapScene.balloons = nil
			mapScene.balloons = append(mapScene.balloons, newBalloonWithArrow(taskLine, x, y, c.Args["content"]))
			taskLine.Push(func() error {
				if input.Triggered() {
					return task.Terminated
				}
				return nil
			})
			taskLine.Push(func() error {
				e.currentCommandIndex++
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
			const height = 20
			dy := scene.TileYNum*scene.TileSize + scene.GameMarginY/scene.TileScale - len(choices)*height
			for i, choice := range choices {
				x := 0
				y := i*height + dy
				width := scene.TileXNum * scene.TileSize
				// TODO: Show balloons as parallel
				mapScene.balloons = append(mapScene.balloons, newBalloon(taskLine, x, y, width, height, choice))
			}
			taskLine.Push(func() error {
				if input.Triggered() {
					return task.Terminated
				}
				return nil
			})
			taskLine.Push(func() error {
				e.currentCommandIndex++
				return task.Terminated
			})
		default:
			// Ignore unknown commands so far.
			e.currentCommandIndex++
		}
		return nil
	})
}

func (e *event) draw(screen *ebiten.Image) error {
	if err := e.character.draw(screen); err != nil {
		return err
	}
	return nil
}
