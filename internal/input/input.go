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

package input

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
)

var theInput = &input{}

type input struct {
	pressCount     int
	x              int
	y              int
	offsetX        int
	offsetY        int
	backPressCount int
	prevPressCount int
}

func IsTurboButtonTriggered() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyT)
}

func SetOffset(offsetX, offsetY int) {
	theInput.offsetX = offsetX
	theInput.offsetY = offsetY
}

func Update() {
	theInput.Update()
}

func Pressed() bool {
	return theInput.Pressed()
}

func BackButtonPressed() bool {
	return theInput.BackButtonPressed()
}

func TriggerBackButton() {
	theInput.TriggerBackButton()
}

func Triggered() bool {
	return theInput.Triggered()
}

func Released() bool {
	return theInput.Released()
}

func Position() (int, int) {
	return theInput.Position()
}

func (i *input) updatePointerDevices() {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		i.pressCount++
		i.x, i.y = ebiten.CursorPosition()
		return
	}
	touches := ebiten.Touches()
	if len(touches) > 0 {
		i.pressCount++
		i.x, i.y = touches[0].Position()
		return
	}
	i.pressCount = 0
}

func (i *input) Update() {
	i.prevPressCount = i.pressCount
	i.updatePointerDevices()
	if i.backPressCount > 0 {
		i.backPressCount--
	}
}

func (i *input) Pressed() bool {
	return i.pressCount > 0
}

func (i *input) Released() bool {
	return i.pressCount == 0 && i.prevPressCount > 0
}

func (i *input) Triggered() bool {
	return i.pressCount == 1
}

func (i *input) Position() (int, int) {
	return i.x - i.offsetX, i.y - i.offsetY
}

func (i *input) BackButtonPressed() bool {
	return i.backPressCount > 0
}

func (i *input) TriggerBackButton() {
	i.backPressCount = 1
}
