// Copyright 2017 Hajime Hoshi
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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/task"
)

type balloons struct {
	balloons []*balloon
}

func (b *balloons) removeBalloon(balloon *balloon) {
	index := -1
	for i, cb := range b.balloons {
		if cb == balloon {
			index = i
			break
		}
	}
	if index != -1 {
		b.balloons[index] = nil
	}
}

func (b *balloons) ShowMessage(content string, character *character) {
	b.balloons = nil
	// TODO: How to call newBalloonCenter?
	x := character.x*scene.TileSize + scene.TileSize/2 + scene.GameMarginX/scene.TileScale
	y := character.y*scene.TileSize + scene.GameMarginTop/scene.TileScale
	newBalloon := newBalloonWithArrow(x, y, content)
	b.balloons = []*balloon{newBalloon}
	b.balloons[0].open()
}

func (b *balloons) ShowChoices(taskLine *task.TaskLine, choices []string, chosenIndexSetter func(int)) {
	const height = 20
	const ymax = scene.TileYNum*scene.TileSize + (scene.GameMarginTop+scene.GameMarginBottom)/scene.TileScale
	ymin := ymax - len(choices)*height
	balloons := []*balloon{}
	taskLine.PushFunc(func() error {
		for i, choice := range choices {
			x := 0
			y := i*height + ymin
			width := scene.TileXNum * scene.TileSize
			balloon := newBalloon(x, y, width, height, choice)
			b.balloons = append(b.balloons, balloon)
			balloon.open()
			balloons = append(balloons, balloon)
		}
		return task.Terminated
	})
	chosenIndex := 0
	taskLine.PushFunc(func() error {
		if !input.Triggered() {
			return nil
		}
		_, y := input.Position()
		y /= scene.TileScale
		if y < ymin || ymax <= y {
			return nil
		}
		chosenIndex = (y - ymin) / height
		chosenIndexSetter(chosenIndex)
		return task.Terminated
	})
	taskLine.PushFunc(func() error {
		for i, balloon := range balloons {
			balloon := balloon
			if i == chosenIndex {
				continue
			}
			balloon.close()
		}
		return task.Terminated
	})
	taskLine.PushFunc(func() error {
		for i, balloon := range balloons {
			if i == chosenIndex {
				continue
			}
			if balloon == nil {
				continue
			}
			if balloon.isAnimating() {
				return nil
			}
			b.removeBalloon(balloon)
		}
		return task.Terminated
	})
	taskLine.Push(task.Sleep(30))
}

func (b *balloons) CloseAll() {
	for _, balloon := range b.balloons {
		if balloon == nil {
			continue
		}
		balloon.close()
	}
}

func (b *balloons) isAnimating() bool {
	for _, balloon := range b.balloons {
		if balloon == nil {
			continue
		}
		if balloon.isAnimating() {
			return true
		}
	}
	return false
}

func (b *balloons) Update() error {
	for _, balloon := range b.balloons {
		if balloon == nil {
			continue
		}
		if err := balloon.update(); err != nil {
			return err
		}
	}
	return nil
}

func (b *balloons) Draw(screen *ebiten.Image) error {
	for _, balloon := range b.balloons {
		if balloon == nil {
			continue
		}
		if err := balloon.draw(screen); err != nil {
			return err
		}
	}
	return nil
}
