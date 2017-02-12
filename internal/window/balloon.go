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

package window

import (
	"encoding/json"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

const (
	// TODO: Rename this to balloonUnitSize
	balloonMarginX     = 4
	balloonMarginY     = 4
	balloonMaxCount    = 4
	balloonArrowWidth  = 6
	balloonArrowHeight = 5
	balloonMinWidth    = 24
)

type balloon struct {
	interpreterID  int
	x              int
	y              int
	width          int
	height         int
	hasArrow       bool
	eventID        int
	content        string
	contentOffsetX int
	openingCount   int
	closingCount   int
	opened         bool
}

type tmpBalloon struct {
	InterpreterID  int    `json:"interpreterId"`
	X              int    `json:"x"`
	Y              int    `json:"y"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	HasArrow       bool   `json:"hasArrow"`
	EventID        int    `json:"eventId"`
	Content        string `json:"content"`
	ContentOffsetX int    `json:"contentOffsetX"`
	OpeningCount   int    `json:"openingCount"`
	ClosingCount   int    `json:"closingCount"`
	Opened         bool   `json:"opened"`
}

func (b *balloon) MarshalJSON() ([]uint8, error) {
	tmp := &tmpBalloon{
		InterpreterID:  b.interpreterID,
		X:              b.x,
		Y:              b.y,
		Width:          b.width,
		Height:         b.height,
		HasArrow:       b.hasArrow,
		EventID:        b.eventID,
		Content:        b.content,
		ContentOffsetX: b.contentOffsetX,
		OpeningCount:   b.openingCount,
		ClosingCount:   b.closingCount,
		Opened:         b.opened,
	}
	return json.Marshal(tmp)
}

func (b *balloon) UnmarshalJSON(data []uint8) error {
	var tmp *tmpBalloon
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	b.interpreterID = tmp.InterpreterID
	b.x = tmp.X
	b.y = tmp.Y
	b.width = tmp.Width
	b.height = tmp.Height
	b.hasArrow = tmp.HasArrow
	b.eventID = tmp.EventID
	b.content = tmp.Content
	b.contentOffsetX = tmp.ContentOffsetX
	b.openingCount = tmp.OpeningCount
	b.closingCount = tmp.ClosingCount
	b.opened = tmp.Opened
	return nil
}

func newBalloon(x, y, width, height int, content string, interpreterID int) *balloon {
	b := &balloon{
		interpreterID: interpreterID,
		content:       content,
		x:             x,
		y:             y,
		width:         ((width + 3) / 4) * 4,
		height:        ((height + 3) / 4) * 4,
	}
	return b
}

func balloonSizeFromContent(content string) (int, int, int) {
	// content is already parsed here.
	contentOffsetX := 0
	w, h := font.MeasureSize(content)
	w = (w + 2*balloonMarginX) * scene.TextScale / scene.TileScale
	h = (h + 2*balloonMarginY) * scene.TextScale / scene.TileScale
	w = ((w + 3) / 4) * 4
	h = ((h + 3) / 4) * 4
	if w < balloonMinWidth {
		contentOffsetX = (balloonMinWidth - w) / 2
		w = balloonMinWidth
	}
	return w, h, contentOffsetX
}

func newBalloonCenter(content string, interpreterID int) *balloon {
	sw := scene.TileXNum * scene.TileSize
	sh := scene.TileYNum*scene.TileSize + scene.GameMarginTop/scene.TileScale
	w, h, contentOffsetX := balloonSizeFromContent(content)
	x := (sw - w) / 2
	y := (sh - h) / 2
	b := &balloon{
		interpreterID:  interpreterID,
		content:        content,
		contentOffsetX: contentOffsetX,
		x:              x,
		y:              y,
		width:          w,
		height:         h,
	}
	return b
}

func newBalloonWithArrow(content string, eventID int, interpreterID int) *balloon {
	b := &balloon{
		interpreterID: interpreterID,
		content:       content,
		hasArrow:      true,
		eventID:       eventID,
	}
	w, h, contentOffsetX := balloonSizeFromContent(content)
	b.width = w
	b.height = h
	b.contentOffsetX = contentOffsetX
	return b
}

func (b *balloon) arrowPosition(screenWidth int, character *character.Character) (int, int) {
	if !b.hasArrow {
		panic("not reach")
	}
	x := (screenWidth/scene.TileScale - scene.TileXNum*scene.TileSize) / 2
	y := scene.GameMarginTop / scene.TileScale
	cx, cy := character.DrawPosition()
	w, _ := character.Size()
	x += cx + w/2
	y += cy
	return x, y
}

func (b *balloon) position(screenWidth int, character *character.Character) (int, int) {
	if !b.hasArrow {
		return b.x, b.y
	}
	ax, ay := b.arrowPosition(screenWidth, character)
	x := ax - b.width/2
	if scene.TileXNum*scene.TileSize < x+b.width {
		x = scene.TileXNum*scene.TileSize - b.width
	}
	if x < 0 {
		x = 0
	}
	y := ay - b.height - 4
	return x, y
}

func (b *balloon) arrowFlip(screenWidth int, character *character.Character) bool {
	if !b.hasArrow {
		return false
	}
	x, _ := b.position(screenWidth, character)
	return scene.TileXNum*scene.TileSize == x+b.width
}

func (b *balloon) isClosed() bool {
	return !b.opened && b.openingCount == 0 && b.closingCount == 0
}

func (b *balloon) isOpened() bool {
	return b.opened
}

func (b *balloon) isAnimating() bool {
	return b.openingCount > 0 || b.closingCount > 0
}

func (b *balloon) open() {
	// TODO: This should be called only in the constructor?
	b.openingCount = balloonMaxCount
}

func (b *balloon) close() {
	b.closingCount = balloonMaxCount
}

type balloonImageParts struct {
	balloon     *balloon
	screenWidth int
	character   *character.Character
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
		ax, ay := b.balloon.arrowPosition(b.screenWidth, b.character)
		x := ax
		y := ay - balloonArrowHeight
		if b.balloon.arrowFlip(b.screenWidth, b.character) {
			// TODO: 4 is an arbitrary number. Define a const.
			x -= 4
			return x, y, x - balloonArrowWidth, y + balloonArrowHeight
		}
		x += 4
		return x, y, x + balloonArrowWidth, y + balloonArrowHeight
	}
	w, _ := b.partsNum()
	x, y := b.balloon.position(b.screenWidth, b.character)
	x += (index % w) * 4
	y += (index / w) * 4
	return x, y, x + 4, y + 4
}

func (b *balloon) update() error {
	if b.closingCount > 0 {
		b.closingCount--
		b.opened = false
	}
	if b.openingCount > 0 {
		b.openingCount--
		if b.openingCount == 0 {
			b.opened = true
		}
	}
	return nil
}

func (b *balloon) draw(screen *ebiten.Image, character *character.Character) error {
	rate := 0.0
	switch {
	case b.opened:
		rate = 1
	case b.openingCount > 0:
		rate = 1 - float64(b.openingCount)/float64(balloonMaxCount)
	case b.closingCount > 0:
		rate = float64(b.closingCount) / float64(balloonMaxCount)
	}
	sw, _ := screen.Size()
	if rate > 0 {
		img := assets.GetImage("balloon.png")
		op := &ebiten.DrawImageOptions{}
		x, y := b.position(sw, character)
		dx := float64(x + b.width/2)
		dy := float64(y + b.height/2)
		if b.hasArrow {
			ax, ay := b.arrowPosition(sw, character)
			dx = float64(ax)
			dy = float64(ay) + balloonArrowHeight
			if b.arrowFlip(sw, character) {
				dx -= 4
			} else {
				dx += 4
			}
		}
		op.GeoM.Translate(-dx, -dy)
		op.GeoM.Scale(rate, rate)
		op.GeoM.Translate(dx, dy)
		op.GeoM.Scale(scene.TileScale, scene.TileScale)
		op.ImageParts = &balloonImageParts{
			balloon:     b,
			screenWidth: sw,
			character:   character,
		}
		if err := screen.DrawImage(img, op); err != nil {
			return err
		}
	}
	if b.opened {
		x, y := b.position(sw, character)
		x = (x + balloonMarginX + b.contentOffsetX) * scene.TileScale
		y = (y + balloonMarginY) * scene.TileScale
		if err := font.DrawText(screen, b.content, x, y, scene.TextScale, color.Black); err != nil {
			return err
		}
	}
	return nil
}
