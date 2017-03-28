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

package ui

import (
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type Image struct {
	X     int
	Y     int
	Scale float64
	image *ebiten.Image
}

func NewImage(x, y int, scale float64, image *ebiten.Image) *Image {
	return &Image{
		X:     x,
		Y:     y,
		Scale: scale,
		image: image,
	}
}

func (i *Image) Update() {
}

func (i *Image) UpdateAsChild(visible bool, offsetX, offsetY int) {
}

func (i *Image) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(i.Scale*scene.TileScale, i.Scale*scene.TileScale)
	op.GeoM.Translate(float64(i.X)*scene.TileScale, float64(i.Y)*scene.TileScale)
	screen.DrawImage(i.image, op)
}