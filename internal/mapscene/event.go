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
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/scene"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

type event struct {
	data         *data.Event
	character    *character
	currentIndex int
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
	task.Push(func() error {
		updated, err := taskLine.Update()
		if err != nil {
			return err
		}
		if updated {
			return nil
		}
		page := e.data.Pages[0]
		// TODO: Consider branches
		if len(page.Commands) <= e.currentIndex {
			e.character.dir = origDir
			return task.Terminated
		}
		c := page.Commands[e.currentIndex]
		switch c.Command {
		case "show_message":
			x := e.data.X*scene.TileSize + scene.TileSize/2
			y := e.data.Y * scene.TileSize
			mapScene.balloon.show(taskLine, x, y, c.Args["content"])
			taskLine.Push(func() error {
				e.currentIndex++
				return task.Terminated
			})
		default:
			// Ignore unknown commands so far.
			e.currentIndex++
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
