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

	"github.com/hajimehoshi/tsugunai/internal/input"
	"github.com/hajimehoshi/tsugunai/internal/scene"
	"github.com/hajimehoshi/tsugunai/internal/task"
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

func (b *balloons) ShowMessage(taskLine *task.TaskLine, content string, character *character, mapScene *MapScene) {
	taskLine.Push(task.Sub(func(sub *task.TaskLine) error {
		// TODO: How to call newBalloonCenter?
		x := character.x*scene.TileSize + scene.TileSize/2 + scene.GameMarginX/scene.TileScale
		y := character.y*scene.TileSize + scene.GameMarginTop/scene.TileScale
		newBalloon := newBalloonWithArrow(x, y, content, mapScene)
		b.balloons = []*balloon{newBalloon}
		b.balloons[0].open(sub)
		return task.Terminated
	}))
	taskLine.PushFunc(func() error {
		if input.Triggered() {
			return task.Terminated
		}
		return nil
	})
	// TODO: close balloon here?
}

func (b *balloons) ShowChoices(taskLine *task.TaskLine, choices []string, chosenIndexSetter func(int), mapScene *MapScene) {
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
			balloon := newBalloon(x, y, width, height, choice, mapScene)
			b.balloons = append(b.balloons, balloon)
			t := &task.TaskLine{}
			sub2 = append(sub2, t)
			balloon.open(t)
			balloons = append(balloons, balloon)
		}
		sub.Push(task.Parallel(sub2...))
		return task.Terminated
	}))
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
	taskLine.Push(task.Sub(func(sub *task.TaskLine) error {
		subs := []*task.TaskLine{}
		for i, balloon := range balloons {
			balloon := balloon
			if i == chosenIndex {
				continue
			}
			t := &task.TaskLine{}
			subs = append(subs, t)
			balloon.close(t)
			t.PushFunc(func() error {
				b.removeBalloon(balloon)
				return task.Terminated
			})
		}
		sub.Push(task.Parallel(subs...))
		return task.Terminated
	}))
	taskLine.Push(task.Sleep(30))
}

func (b *balloons) CloseAll(taskLine *task.TaskLine) {
	taskLine.Push(task.Sub(func(sub *task.TaskLine) error {
		subs := []*task.TaskLine{}
		for _, balloon := range b.balloons {
			if balloon == nil {
				continue
			}
			balloon := balloon
			t := &task.TaskLine{}
			subs = append(subs, t)
			balloon.close(t)
			t.PushFunc(func() error {
				b.removeBalloon(balloon)
				return task.Terminated
			})
		}
		sub.Push(task.Parallel(subs...))
		return task.Terminated
	}))
	taskLine.PushFunc(func() error {
		b.balloons = nil
		return task.Terminated
	})
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
