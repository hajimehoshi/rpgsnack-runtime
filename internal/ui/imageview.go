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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
)

type ImageView struct {
	X     int
	Y     int
	Scale float64
	image *ImagePart
}

func NewImageView(x, y int, scale float64, image *ImagePart) *ImageView {
	return &ImageView{
		X:     x,
		Y:     y,
		Scale: scale,
		image: image,
	}
}

func (i *ImageView) Update() {
}

func (i *ImageView) UpdateAsChild(visible bool, offsetX, offsetY int) {
}

func (i *ImageView) Draw(screen *ebiten.Image) {
	geoM := &ebiten.GeoM{}
	geoM.Scale(i.Scale, i.Scale)
	geoM.Translate(float64(i.X), float64(i.Y))
	geoM.Scale(consts.TileScale, consts.TileScale)
	i.image.Draw(screen, geoM, &ebiten.ColorM{})
}
