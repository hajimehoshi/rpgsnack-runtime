// Copyright 2018 Hajime Hoshi
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

package sceneimpl

import (
	"image"

	"github.com/hajimehoshi/ebiten"
)

const animationInterval = 30

type animation struct {
	counter int
}

func (a *animation) Update() {
	a.counter++
}

func (a *animation) Draw(screen *ebiten.Image, texture *ebiten.Image, frameWidth, offsetX, offsetY int) {
	op := &ebiten.DrawImageOptions{}
	w, h := texture.Size()

	// This is a pingpong animation
	// We might add/change loop based animations in near future
	frameCount := 1
	baseFrameCount := w / frameWidth
	if baseFrameCount > 1 {
		frameCount = baseFrameCount*2 - 2
	}
	frames := a.counter / animationInterval
	frame := frames % frameCount
	if frame >= baseFrameCount {
		frame = frameCount - frame
	}
	op.GeoM.Translate(float64(offsetX), float64(offsetY))
	screen.DrawImage(texture.SubImage(image.Rect(frame*frameWidth, 0, w, h)).(*ebiten.Image), op)
}
