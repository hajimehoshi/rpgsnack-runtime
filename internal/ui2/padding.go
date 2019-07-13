// Copyright 2019 The RPGSnack Authors
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

package ui2

import (
	"image"

	"github.com/hajimehoshi/ebiten"
)

type Padding struct {
	x      int
	y      int
	width  int
	height int
}

func NewPadding(x, y, width, height int) *Padding {
	return &Padding{
		x:      x,
		y:      y,
		width:  width,
		height: height,
	}
}

func (p *Padding) Region() image.Rectangle {
	return image.Rect(p.x, p.y, p.x+p.width, p.y+p.height)
}

func (p *Padding) Update() {
}

func (p *Padding) HandleInput(offsetX, offsetY int) bool {
	return false
}

func (p *Padding) DrawAsChild(screen *ebiten.Image, offsetX, offsetY int) {
}
