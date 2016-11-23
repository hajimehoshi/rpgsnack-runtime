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

package mapscene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/font"
	"github.com/hajimehoshi/tsugunai/internal/scene"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

const (
	// TODO: Rename this to balloonUnitSize
	balloonMarginX     = 4
	balloonMarginY     = 4
	balloonMaxCount    = 8
	balloonArrowWidth  = 6
	balloonArrowHeight = 5
)

type balloon struct {
	x         int
	y         int
	width     int
	height    int
	hasArrow  bool
	arrowX    int
	arrowY    int
	arrowFlip bool
	content   string
	count     int
	maxCount  int
}

func newBalloon(x, y, width, height int, content string) *balloon {
	b := &balloon{
		content: content,
		x:       x,
		y:       y,
		width:   ((width + 3) / 4) * 4,
		height:  ((height + 3) / 4) * 4,
		count:   balloonMaxCount,
	}
	return b
}

func newBalloonWithArrow(arrowX, arrowY int, content string) *balloon {
	b := &balloon{
		content:  content,
		hasArrow: true,
		arrowX:   arrowX,
		arrowY:   arrowY - balloonArrowHeight,
		count:    balloonMaxCount,
	}
	w, h := font.MeasureSize(b.content)
	w = (w + 2*balloonMarginX) * scene.TextScale / scene.TileScale
	h = (h + 2*balloonMarginY) * scene.TextScale / scene.TileScale
	w = ((w + 3) / 4) * 4
	h = ((h + 3) / 4) * 4
	b.width = w
	b.height = h
	b.x = arrowX - w/2
	if scene.TileXNum*scene.TileSize < b.x+w {
		b.arrowFlip = true
		b.x = scene.TileXNum*scene.TileSize - w
	}
	if b.x+w < 0 {
		b.x = 0
	}
	b.y = arrowY - h - 4
	return b
}

func (b *balloon) open(taskLine *task.TaskLine) {
	taskLine.Push(func() error {
		b.count--
		if b.count == balloonMaxCount/2 {
			return task.Terminated
		}
		return nil
	})
}

func (b *balloon) close(taskLine *task.TaskLine) {
	taskLine.Push(func() error {
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
	return b.balloon.width / 4, b.balloon.height / 4
}

func (b *balloonImageParts) Len() int {
	w, h := b.partsNum()
	return w*h + 1
}

func (b *balloonImageParts) Src(index int) (int, int, int, int) {
	if index == b.Len()-1 {
		if !b.balloon.hasArrow {
			return 0, 0, 0, 0
		}
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
		if !b.balloon.hasArrow {
			return 0, 0, 0, 0
		}
		x := b.balloon.arrowX
		y := b.balloon.arrowY
		if b.balloon.arrowFlip {
			// TODO: 4 is an arbitrary number. Define a const.
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
	if b.count > 0 {
		img := theImageCache.Get("balloon.png")
		op := &ebiten.DrawImageOptions{}
		rate := 1.0
		if balloonMaxCount/2 < b.count {
			rate = 1 - float64(b.count-balloonMaxCount/2)/float64(balloonMaxCount/2)
		} else if balloonMaxCount/2 > b.count {
			rate = float64(b.count) / float64(balloonMaxCount/2)
		}
		if rate != 1.0 {
			dx := float64(b.x + b.width/2)
			dy := float64(b.y + b.height/2)
			if b.hasArrow {
				dx = float64(b.arrowX)
				dy = float64(b.arrowY) + balloonArrowHeight
				if b.arrowFlip {
					dx -= 4
				} else {
					dx += 4
				}
			}
			op.GeoM.Translate(-dx, -dy)
			op.GeoM.Scale(rate, rate)
			op.GeoM.Translate(dx, dy)
		}
		op.GeoM.Scale(scene.TileScale, scene.TileScale)
		op.GeoM.Translate(scene.GameMarginX, scene.GameMarginY)
		op.ImageParts = &balloonImageParts{
			balloon: b,
		}
		if err := screen.DrawImage(img, op); err != nil {
			return err
		}
	}
	if b.count == balloonMaxCount/2 {
		x := (b.x+balloonMarginX)*scene.TileScale + scene.GameMarginX
		y := (b.y+balloonMarginY)*scene.TileScale + scene.GameMarginY
		if err := font.DrawText(screen, b.content, x, y, scene.TextScale, color.Black); err != nil {
			return err
		}
	}
	return nil
}
