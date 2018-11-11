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
	"image"

	"github.com/hajimehoshi/ebiten"
)

func drawNinePatches(dst, src *ebiten.Image, width, height int, geoM *ebiten.GeoM, colorM *ebiten.ColorM) {
	const partSize = 4

	parts := make([]*ebiten.Image, 9)
	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			x := i*partSize
			y := j*partSize
			parts[j*3+i] = src.SubImage(image.Rect(x, y, x+partSize, y+partSize)).(*ebiten.Image)
		}
	}

	xn, yn := width/partSize, height/partSize
	op := &ebiten.DrawImageOptions{}
	if colorM != nil {
		op.ColorM.Concat(*colorM)
	}
	for j := 0; j < yn; j++ {
		sy := 0
		switch j {
		case 0:
			sy = 0
		case yn - 1:
			sy = 2
		default:
			sy = 1
		}
		for i := 0; i < xn; i++ {
			sx := 0
			switch i {
			case 0:
				sx = 0
			case xn - 1:
				sx = 2
			default:
				sx = 1
			}
			op.GeoM.Reset()
			op.GeoM.Translate(float64(i*partSize), float64(j*partSize))
			op.GeoM.Concat(*geoM)
			dst.DrawImage(parts[sy*3+sx], op)
		}
	}
}
