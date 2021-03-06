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
	prevPressCount int
	x              int
	y              int

	backPressCount    int
	backButtonPressed bool
}

func IsMuteButtonTriggered() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyM)
}

func IsSwitchDebugButtonTriggered() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyS)
}

func IsVariableDebugButtonTriggered() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyV)
}

func IsTurboButtonTriggered() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyT)
}

func IsScreenshotButtonTriggered() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyP)
}

func Update(scaleX, scaleY float64) {
	theInput.Update(scaleX, scaleY)
}

func Wheel() (xoff, yoff float64) {
	return ebiten.Wheel()
}

func Pressed() bool {
	return theInput.Pressed()
}

func BackButtonTriggered() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyB) || theInput.BackButtonTriggered()
}

func PressBackButton() {
	theInput.PressBackButton()
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

func (i *input) updatePointerDevices(scaleX, scaleY float64) {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		i.pressCount++
		i.x, i.y = ebiten.CursorPosition()
		i.x = int(float64(i.x) / scaleX)
		i.y = int(float64(i.y) / scaleY)
		return
	}
	touches := ebiten.Touches()
	if len(touches) > 0 {
		i.pressCount++
		i.x, i.y = touches[0].Position()
		i.x = int(float64(i.x) / scaleX)
		i.y = int(float64(i.y) / scaleY)
		return
	}
	i.pressCount = 0
}

func (i *input) Update(scaleX, scaleY float64) {
	i.prevPressCount = i.pressCount
	i.updatePointerDevices(scaleX, scaleY)
	if i.backButtonPressed {
		i.backPressCount++
	} else {
		i.backPressCount = 0
	}
	i.backButtonPressed = false
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
	return i.x, i.y
}

func (i *input) BackButtonTriggered() bool {
	return i.backPressCount == 1
}

func (i *input) PressBackButton() {
	i.backButtonPressed = true
}
