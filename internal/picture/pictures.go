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

package picture

import (
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type Pictures struct {
	pictures []*picture
}

func (p *Pictures) Update() {
}

func (p *Pictures) Draw(screen *ebiten.Image) {
	for _, pic := range p.pictures {
		if pic == nil {
			continue
		}
		pic.Draw(screen)
	}
}

func (p *Pictures) Add(id int, image *ebiten.Image, x, y int, scaleX, scaleY, angle, opacity float64, origin data.ShowPictureOrigin, blendType data.ShowPictureBlendType) {
	if len(p.pictures) < id+1 {
		p.pictures = append(p.pictures, make([]*picture, id+1-len(p.pictures))...)
	}
	p.pictures[id] = &picture{
		image:     image,
		x:         x,
		y:         y,
		scaleX:    scaleX,
		scaleY:    scaleY,
		angle:     angle,
		opacity:   opacity,
		origin:    origin,
		blendType: blendType,
	}
}

func (p *Pictures) Remove(id int) {
	p.pictures[id] = nil
}

type picture struct {
	image     *ebiten.Image
	x         int
	y         int
	scaleX    float64
	scaleY    float64
	angle     float64
	opacity   float64
	origin    data.ShowPictureOrigin
	blendType data.ShowPictureBlendType
}

func (p *picture) Draw(screen *ebiten.Image) {
	sx, sy := p.image.Size()

	op := &ebiten.DrawImageOptions{}
	if p.origin == data.ShowPictureOriginCenter {
		op.GeoM.Translate(-float64(sx)/2, -float64(sy)/2)
	}
	op.GeoM.Scale(p.scaleX, p.scaleY)
	op.GeoM.Rotate(p.angle)
	op.GeoM.Translate(float64(p.x), float64(p.y))

	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	op.GeoM.Translate(0, consts.GameMarginTop)

	if p.opacity < 1 {
		op.ColorM.Scale(1, 1, 1, p.opacity)
	}
	switch p.blendType {
	case data.ShowPictureBlendTypeNormal:
		// Use default
	case data.ShowPictureBlendTypeAdd:
		op.CompositeMode = ebiten.CompositeModeLighter
	}

	screen.DrawImage(p.image, op)

	// TODO: Use blend type
}
