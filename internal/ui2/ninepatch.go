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

package ui2

import (
	"image"

	"github.com/hajimehoshi/ebiten"
)

func drawNinePatches(dst, src *ebiten.Image, width, height int, geoM *ebiten.GeoM, colorM *ebiten.ColorM) {
	const partSize = 4

	op := &ebiten.DrawImageOptions{}
	if colorM != nil {
		op.ColorM.Concat(*colorM)
	}
	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			x := i * partSize
			y := j * partSize

			tx := 0
			ty := 0
			sx := 1.0
			sy := 1.0

			switch i {
			case 0:
			case 1:
				tx = partSize
				sx = float64(width-2*partSize) / partSize
			case 2:
				tx = width - partSize
			}
			switch j {
			case 0:
			case 1:
				ty = partSize
				sy = float64(height-2*partSize) / partSize
			case 2:
				ty = height - partSize
			}

			op.GeoM.Reset()
			op.GeoM.Scale(sx, sy)
			op.GeoM.Translate(float64(tx), float64(ty))
			op.GeoM.Concat(*geoM)

			dst.DrawImage(src.SubImage(image.Rect(x, y, x+partSize, y+partSize)).(*ebiten.Image), op)
		}
	}
}
