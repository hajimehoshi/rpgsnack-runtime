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
)

const (
	choiceBalloonHeight = 20
	choiceBalloonsMaxY  = scene.TileYNum*scene.TileSize +
		(scene.GameMarginTop+scene.GameMarginBottom)/scene.TileScale
)

type balloons struct {
	balloons                  []*balloon
	choiceBalloons            []*balloon
	chosenIndex               int
	choosing                  bool
	chosenBalloonWaitingCount int
}

func choiceBalloonsMinY(num int) int {
	return choiceBalloonsMaxY - num*choiceBalloonHeight
}

func (b *balloons) ChosenIndex() int {
	return b.chosenIndex
}

func (b *balloons) ShowMessage(content string, character *character) {
	// TODO: How to call newBalloonCenter?
	x := character.x*scene.TileSize + scene.TileSize/2 + scene.GameMarginX/scene.TileScale
	y := character.y*scene.TileSize + scene.GameMarginTop/scene.TileScale
	newBalloon := newBalloonWithArrow(x, y, content)
	b.balloons = []*balloon{newBalloon}
	b.balloons[0].open()
}

func (b *balloons) ShowChoices(choices []string) {
	ymin := choiceBalloonsMinY(len(choices))
	b.choiceBalloons = nil
	for i, choice := range choices {
		x := 0
		y := i*choiceBalloonHeight + ymin
		width := scene.TileXNum * scene.TileSize
		balloon := newBalloon(x, y, width, choiceBalloonHeight, choice)
		b.choiceBalloons = append(b.choiceBalloons, balloon)
		balloon.open()
	}
	b.chosenIndex = 0
	b.choosing = true
}

func (b *balloons) CloseAll() {
	for _, balloon := range b.balloons {
		if balloon == nil {
			continue
		}
		balloon.close()
	}
	for _, balloon := range b.choiceBalloons {
		if balloon == nil {
			continue
		}
		balloon.close()
	}
}

func (b *balloons) isBusy() bool {
	return b.isAnimating() || b.choosing || b.chosenBalloonWaitingCount > 0
}

func (b *balloons) isOpened() bool {
	for _, balloon := range b.balloons {
		if balloon == nil {
			continue
		}
		if balloon.isOpened() {
			return true
		}
	}
	for _, balloon := range b.choiceBalloons {
		if balloon == nil {
			continue
		}
		if balloon.isOpened() {
			return true
		}
	}
	return false
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
	for _, balloon := range b.choiceBalloons {
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
	if b.chosenBalloonWaitingCount > 0 {
		b.chosenBalloonWaitingCount--
		if b.chosenBalloonWaitingCount == 0 {
			b.choiceBalloons[b.chosenIndex].close()
		}
	} else if b.choosing && b.isOpened() && input.Triggered() {
		ymax := choiceBalloonsMaxY
		ymin := choiceBalloonsMinY(len(b.choiceBalloons))
		_, y := input.Position()
		y /= scene.TileScale
		if y < ymin || ymax <= y {
			return nil
		}
		b.chosenIndex = (y - ymin) / choiceBalloonHeight
		for i, balloon := range b.choiceBalloons {
			balloon := balloon
			if i == b.chosenIndex {
				continue
			}
			balloon.close()
		}
		b.chosenBalloonWaitingCount = 30
		b.choosing = false
	}
	for i, balloon := range b.balloons {
		if balloon == nil {
			continue
		}
		if err := balloon.update(); err != nil {
			return err
		}
		if balloon.isClosed() {
			b.balloons[i] = nil
		}
	}
	for i, balloon := range b.choiceBalloons {
		if balloon == nil {
			continue
		}
		if err := balloon.update(); err != nil {
			return err
		}
		if balloon.isClosed() {
			b.choiceBalloons[i] = nil
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
	for _, balloon := range b.choiceBalloons {
		if balloon == nil {
			continue
		}
		if err := balloon.draw(screen); err != nil {
			return err
		}
	}
	return nil
}
