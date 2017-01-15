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

package window

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

type Character interface {
	DrawPosition() (int, int)
	Size() (int, int)
}

type Windows struct {
	nextBalloon               *balloon
	balloons                  []*balloon
	choiceBalloons            []*balloon
	chosenIndex               int
	choosing                  bool
	choosingInterpreterID     int
	chosenBalloonWaitingCount int
	hasChosenIndex            bool
}

func choiceBalloonsMinY(num int) int {
	return choiceBalloonsMaxY - num*choiceBalloonHeight
}

func (b *Windows) ChosenIndex() int {
	return b.chosenIndex
}

func (b *Windows) HasChosenIndex() bool {
	return b.hasChosenIndex
}

func (b *Windows) ShowMessage(content string, character Character, interpreterID int) {
	if b.nextBalloon != nil {
		panic("not reach")
	}
	// TODO: How to call newBalloonCenter?
	b.nextBalloon = newBalloonWithArrow(character, content, interpreterID)
}

func (b *Windows) ShowChoices(choices []string, interpreterID int) {
	ymin := choiceBalloonsMinY(len(choices))
	b.choiceBalloons = nil
	for i, choice := range choices {
		x := 0
		y := i*choiceBalloonHeight + ymin
		width := scene.TileXNum * scene.TileSize
		balloon := newBalloon(x, y, width, choiceBalloonHeight, choice, interpreterID)
		b.choiceBalloons = append(b.choiceBalloons, balloon)
		balloon.open()
	}
	b.chosenIndex = 0
	b.choosing = true
	b.choosingInterpreterID = interpreterID
}

func (b *Windows) CloseAll() {
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

func (b *Windows) IsBusy(interpreterID int) bool {
	if b.isAnimating(interpreterID) {
		return true
	}
	if b.choosingInterpreterID == interpreterID {
		if b.choosing || b.chosenBalloonWaitingCount > 0 {
			return true
		}
	}
	if b.isOpened(interpreterID) {
		return true
	}
	if b.nextBalloon != nil {
		return true
	}
	return false
}

func (b *Windows) CanProceed(interpreterID int) bool {
	if !b.IsBusy(interpreterID) {
		return true
	}
	if !b.isOpened(interpreterID) {
		return false
	}
	if !input.Triggered() {
		return false
	}
	return true
}

func (b *Windows) isOpened(interpreterID int) bool {
	for _, balloon := range b.balloons {
		if balloon == nil {
			continue
		}
		if interpreterID > 0 && balloon.interpreterID != interpreterID {
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
		if interpreterID > 0 && balloon.interpreterID != interpreterID {
			continue
		}
		if balloon.isOpened() {
			return true
		}
	}
	return false
}

func (b *Windows) isAnimating(interpreterID int) bool {
	for _, balloon := range b.balloons {
		if balloon == nil {
			continue
		}
		if interpreterID > 0 && balloon.interpreterID != interpreterID {
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
		if interpreterID > 0 && balloon.interpreterID != interpreterID {
			continue
		}
		if balloon.isAnimating() {
			return true
		}
	}
	return false
}

func (b *Windows) Update() error {
	if !b.choosing {
		// 0 means to check all balloons.
		// TODO: Don't use magic numbers.
		if b.nextBalloon != nil && !b.isAnimating(0) && !b.isOpened(0) {
			b.balloons = []*balloon{b.nextBalloon}
			b.balloons[0].open()
			b.nextBalloon = nil
		}
	}
	if b.chosenBalloonWaitingCount > 0 {
		b.chosenBalloonWaitingCount--
		if b.chosenBalloonWaitingCount == 0 {
			b.choiceBalloons[b.chosenIndex].close()
			for _, balloon := range b.balloons {
				if balloon == nil {
					continue
				}
				balloon.close()
			}
			b.hasChosenIndex = false
		}
	} else if b.choosing && b.isOpened(0) && input.Triggered() {
		ymax := choiceBalloonsMaxY
		ymin := choiceBalloonsMinY(len(b.choiceBalloons))
		_, y := input.Position()
		y /= scene.TileScale
		if y < ymin || ymax <= y {
			return nil
		}
		// Close regular balloons
		b.chosenIndex = (y - ymin) / choiceBalloonHeight
		for i, balloon := range b.choiceBalloons {
			if i == b.chosenIndex {
				continue
			}
			balloon.close()
		}
		b.chosenBalloonWaitingCount = 30
		b.choosing = false
		b.choosingInterpreterID = 0
		b.hasChosenIndex = true
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

func (b *Windows) Draw(screen *ebiten.Image) error {
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
