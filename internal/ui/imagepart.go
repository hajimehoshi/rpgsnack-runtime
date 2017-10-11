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

type ImagePart struct {
	image   *ebiten.Image
	srcRect *image.Rectangle
}

func NewImagePart(i *ebiten.Image) *ImagePart {
	return &ImagePart{
		image:   i,
		srcRect: nil,
	}
}

func NewImagePartWithRect(i *ebiten.Image, srcX, srcY, srcWidth, srcHeight int) *ImagePart {
	return &ImagePart{
		image:   i,
		srcRect: &image.Rectangle{Min: image.Point{X: srcX, Y: srcY}, Max: image.Point{X: srcX + srcWidth, Y: srcY + srcHeight}},
	}
}

func (i *ImagePart) Draw(screen *ebiten.Image, geoM *ebiten.GeoM, colorM *ebiten.ColorM) {
	// TODO support NinePatch Drawing
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Concat(*geoM)
	op.ColorM.Concat(*colorM)
	op.SourceRect = i.srcRect
	screen.DrawImage(i.image, op)
}
