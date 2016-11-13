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

package game

import (
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/data"
)

type character struct {
	image      *ebiten.Image
	imageIndex int
	dir        data.Dir
	x          int
	y          int
}

type characterImageParts struct {
	charWidth  int
	charHeight int
	index      int
	dir        data.Dir
}

func (c *characterImageParts) Len() int {
	return 1
}

func (c *characterImageParts) Src(index int) (int, int, int, int) {
	x := ((c.index%4)*3 + 1) * c.charWidth
	y := (c.index / 4) * 2 * c.charHeight
	switch c.dir {
	case data.DirUp:
	case data.DirRight:
		y += c.charHeight
	case data.DirDown:
		y += 2 * c.charHeight
	case data.DirLeft:
		y += 3 * c.charHeight
	}
	return x, y, x + c.charWidth, y + c.charHeight
}

func (c *characterImageParts) Dst(index int) (int, int, int, int) {
	return 0, 0, c.charWidth, c.charHeight
}

func (c *character) draw(screen *ebiten.Image) error {
	imageW, imageH := c.image.Size()
	charW := imageW / 4 / 3
	charH := imageH / 2 / 4
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(c.x*tileSize+tileSize/2), float64((c.y+1)*tileSize))
	op.GeoM.Translate(float64(-charW/2), float64(-charH))
	op.GeoM.Scale(tileScale, tileScale)
	op.ImageParts = &characterImageParts{
		charWidth:  charW,
		charHeight: charH,
		index:      c.imageIndex,
		dir:        c.dir,
	}
	if err := screen.DrawImage(c.image, op); err != nil {
		return err
	}
	return nil
}
