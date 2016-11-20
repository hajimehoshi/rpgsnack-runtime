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

package scene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/font"
	"github.com/hajimehoshi/tsugunai/internal/input"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

const (
	balloonMarginX     = 4
	balloonMarginY     = 4
	balloonMaxCount    = 20
	balloonArrowWidth  = 6
	balloonArrowHeight = 5
)

type balloon struct {
	x         int
	y         int
	arrowX    int
	arrowY    int
	arrowFlip bool
	content   string
	count     int
}

func (b *balloon) bodySize() (int, int) {
	w, h := font.MeasureSize(b.content)
	w = (w + 2*balloonMarginX) * textScale / tileScale
	h = (h + 2*balloonMarginY) * textScale / tileScale
	w = ((w + 3) / 4) * 4
	h = ((h + 3) / 4) * 4
	return w, h
}

func (b *balloon) show(x, y int, message string) {
	b.content = message
	b.arrowX = x
	b.arrowY = y - balloonArrowHeight
	b.arrowFlip = false
	w, h := b.bodySize()
	b.x = x - w/2
	if tileXNum*tileSize < b.x+w {
		b.arrowFlip = true
		b.x = tileXNum*tileSize - w
	}
	if b.x+w < 0 {
		b.x = 0
	}
	b.y = y - h - 4
	b.count = balloonMaxCount
	task.Push(func() error {
		b.count--
		if b.count == balloonMaxCount/2 {
			return task.Terminated
		}
		return nil
	})
	task.Push(func() error {
		if input.Triggered() {
			return task.Terminated
		}
		return nil
	})
	task.Push(func() error {
		b.count--
		if b.count == 0 {
			return task.Terminated
		}
		return nil
	})
}

type balloonImageParts struct {
	balloon *balloon
}

func (b *balloonImageParts) partsNum() (int, int) {
	width, height := b.balloon.bodySize()
	return width / 4, height / 4
}

func (b *balloonImageParts) Len() int {
	w, h := b.partsNum()
	return w*h + 1
}

func (b *balloonImageParts) Src(index int) (int, int, int, int) {
	if index == b.Len()-1 {
		return 12, 0, 12 + balloonArrowWidth, balloonArrowHeight
	}
	w, h := b.partsNum()
	x := index % w
	y := index / w
	sx := 0
	sy := 0
	switch {
	case x == 0:
	default:
		sx += 4
	case x == w-1:
		sx += 8
	}
	switch {
	case y == 0:
	default:
		sy += 4
	case y == h-1:
		sy += 8
	}
	return sx, sy, sx + 4, sy + 4
}

func (b *balloonImageParts) Dst(index int) (int, int, int, int) {
	if index == b.Len()-1 {
		x := b.balloon.arrowX
		y := b.balloon.arrowY
		if b.balloon.arrowFlip {
			x -= 4
			return x, y, x - balloonArrowWidth, y + balloonArrowHeight
		}
		x += 4
		return x, y, x + balloonArrowWidth, y + balloonArrowHeight
	}
	w, _ := b.partsNum()
	x := b.balloon.x + (index%w)*4
	y := b.balloon.y + (index/w)*4
	return x, y, x + 4, y + 4
}

func (b *balloon) draw(screen *ebiten.Image) error {
	if b.count == balloonMaxCount/2 {
		img := theImageCache.Get("balloon.png")
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(tileScale, tileScale)
		op.GeoM.Translate(gameMarginX, gameMarginY)
		op.ImageParts = &balloonImageParts{
			balloon: b,
		}
		if err := screen.DrawImage(img, op); err != nil {
			return err
		}
		x := (b.x+balloonMarginX)*tileScale + gameMarginX
		y := (b.y+balloonMarginY)*tileScale + gameMarginY
		if err := font.DrawText(screen, b.content, x, y, textScale, color.Black); err != nil {
			return err
		}
	}
	return nil
}
