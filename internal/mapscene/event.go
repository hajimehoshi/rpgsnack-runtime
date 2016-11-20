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
	data      *data.Event
	character *character
}

func (e *event) position() (int, int) {
	return e.character.x, e.character.y
}

func (e *event) trigger() data.Trigger {
	return e.data.Pages[0].Trigger
}

func (e *event) run(mapScene *MapScene) {
	page := e.data.Pages[0]
	// TODO: Consider branches
	for _, c := range page.Commands {
		c := c
		switch c.Command {
		case "show_message":
			task.Push(func() error {
				x := e.data.X*scene.TileSize + scene.TileSize/2
				y := e.data.Y * scene.TileSize
				mapScene.balloon.show(x, y, c.Args["content"])
				return task.Terminated
			})
		}
	}
}

func (e *event) draw(screen *ebiten.Image) error {
	if err := e.character.draw(screen); err != nil {
		return err
	}
	return nil
}
