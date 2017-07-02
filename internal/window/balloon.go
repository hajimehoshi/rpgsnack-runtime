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
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
)

const (
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
	contentOffsetY int
	openingCount   int
	closingCount   int
	opened         bool
	balloonType    data.BalloonType
}

type tmpBalloon struct {
	InterpreterID  int              `json:"interpreterId"`
	X              int              `json:"x"`
	Y              int              `json:"y"`
	Width          int              `json:"width"`
	Height         int              `json:"height"`
	HasArrow       bool             `json:"hasArrow"`
	EventID        int              `json:"eventId"`
	Content        string           `json:"content"`
	ContentOffsetX int              `json:"contentOffsetX"`
	ContentOffsetY int              `json:"contentOffsetY"`
	OpeningCount   int              `json:"openingCount"`
	ClosingCount   int              `json:"closingCount"`
	Opened         bool             `json:"opened"`
	BalloonType    data.BalloonType `json:"balloonType"`
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
		ContentOffsetY: b.contentOffsetY,
		OpeningCount:   b.openingCount,
		ClosingCount:   b.closingCount,
		Opened:         b.opened,
		BalloonType:    b.balloonType,
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
	b.contentOffsetY = tmp.ContentOffsetY
	b.openingCount = tmp.OpeningCount
	b.closingCount = tmp.ClosingCount
	b.opened = tmp.Opened
	b.balloonType = tmp.BalloonType
	return nil
}

func newBalloon(x, y, width, height int, content string, balloonType data.BalloonType, interpreterID int) *balloon {
	b := &balloon{
		interpreterID: interpreterID,
		content:       content,
		x:             x,
		y:             y,
		balloonType:   balloonType,
	}
	s := b.partSize()
	b.width = ((width + (s - 1)) / s) * s
	b.height = ((height + (s - 1)) / s) * s
	return b
}

func balloonPartSize(balloonType data.BalloonType) int {
	if balloonType == data.BalloonTypeShout {
		return 8
	}
	return 4
}

func balloonMargin(balloonType data.BalloonType) (int, int) {
	if balloonType == data.BalloonTypeShout {
		return 8, 8
	}
	return 4, 4
}

func balloonSizeFromContent(content string, balloonType data.BalloonType) (int, int, int, int) {
	// content is already parsed here.
	tw, th := font.MeasureSize(content)
	tw = tw * consts.TextScale / consts.TileScale
	th = th * consts.TextScale / consts.TileScale
	mx, my := balloonMargin(balloonType)
	w := tw + 2*mx
	h := th + 2*my
	s := balloonPartSize(balloonType)
	w = ((w + (s - 1)) / s) * s
	h = ((h + (s - 1)) / s) * s
	contentOffsetX := 0
	if w < balloonMinWidth {
		contentOffsetX = (balloonMinWidth - w) / 2
		w = balloonMinWidth
	}
	contentOffsetY := ((h - 2*my) - th) / 2
	return w, h, contentOffsetX, contentOffsetY
}

func newBalloonWithArrow(content string, balloonType data.BalloonType, eventID int, interpreterID int) *balloon {
	b := &balloon{
		interpreterID: interpreterID,
		content:       content,
		hasArrow:      true,
		eventID:       eventID,
		balloonType:   balloonType,
	}
	w, h, contentOffsetX, contentOffsetY := balloonSizeFromContent(content, balloonType)
	b.width = w
	b.height = h
	b.contentOffsetX = contentOffsetX
	b.contentOffsetY = contentOffsetY
	return b
}

func (b *balloon) arrowPosition(screenWidth int, character *character.Character) (int, int) {
	if !b.hasArrow {
		panic("not reach")
	}
	cx, cy := character.DrawPosition()
	w, _ := character.Size()
	x := cx + w/2
	y := cy
	return x, y
}

func (b *balloon) position(screenWidth int, character *character.Character) (int, int) {
	if !b.hasArrow {
		return b.x, b.y
	}
	ax, ay := b.arrowPosition(screenWidth, character)
	x := ax - b.width/2
	if consts.TileXNum*consts.TileSize < x+b.width {
		x = consts.TileXNum*consts.TileSize - b.width
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
	return consts.TileXNum*consts.TileSize == x+b.width
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

func (b *balloon) partSize() int {
	return balloonPartSize(b.balloonType)
}

func (b *balloon) margin() (int, int) {
	return balloonMargin(b.balloonType)
}

func (b *balloon) update() {
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
}

func (b *balloon) geoMForRate(screen *ebiten.Image, character *character.Character) *ebiten.GeoM {
	sw, _ := screen.Size()
	x, y := b.position(sw, character)
	cx := float64(x + b.width/2)
	cy := float64(y + b.height/2)
	if b.hasArrow {
		ax, ay := b.arrowPosition(sw, character)
		cx = float64(ax)
		cy = float64(ay) + balloonArrowHeight
		if b.arrowFlip(sw, character) {
			cx -= 4
		} else {
			cx += 4
		}
	}
	g := ebiten.GeoM{}
	g.Translate(-cx, -cy)
	rate := b.openingRate()
	g.Scale(rate, rate)
	g.Translate(cx, cy)
	g.Scale(consts.TileScale, consts.TileScale)
	return &g
}

func (b *balloon) openingRate() float64 {
	switch {
	case b.opened:
		return 1
	case b.openingCount > 0:
		return 1 - float64(b.openingCount)/float64(balloonMaxCount)
	case b.closingCount > 0:
		return float64(b.closingCount) / float64(balloonMaxCount)
	default:
		return 0
	}
}

func (b *balloon) draw(screen *ebiten.Image, character *character.Character) {
	sw, _ := screen.Size()
	dx := (sw - consts.TileXNum*consts.TileSize*consts.TileScale) / 2
	dy := consts.GameMarginTop
	if b.openingRate() > 0 {
		img := assets.GetImage("system/balloon.png")
		if b.balloonType == data.BalloonTypeShout {
			img = assets.GetImage("system/shout.png")
		}
		op := &ebiten.DrawImageOptions{}
		g := b.geoMForRate(screen, character)
		g.Translate(float64(dx), float64(dy))
		pw, ph := b.width/b.partSize(), b.height/b.partSize()
		s := b.partSize()
		for j := 0; j < ph; j++ {
			for i := 0; i < pw; i++ {
				op.GeoM.Reset()
				sx, sy := 0, 0
				switch {
				case i == 0:
				default:
					sx += s
				case i == pw-1:
					sx += s * 2
				}
				switch {
				case j == 0:
				default:
					sy += s
				case j == ph-1:
					sy += s * 2
				}
				r := image.Rect(sx, sy, sx+s, sy+s)
				op.SourceRect = &r
				dx, dy := b.position(sw, character)
				dx += i * s
				dy += j * s
				op.GeoM.Translate(float64(dx), float64(dy))
				op.GeoM.Concat(*g)
				screen.DrawImage(img, op)
			}
		}
		if b.hasArrow && (b.balloonType == data.BalloonTypeNormal ||
			b.balloonType == data.BalloonTypeThink) {
			op := &ebiten.DrawImageOptions{}
			switch b.balloonType {
			case data.BalloonTypeNormal:
				r := image.Rect(12, 0, 12+balloonArrowWidth, balloonArrowHeight)
				op.SourceRect = &r
			case data.BalloonTypeThink:
				r := image.Rect(18, 0, 18+balloonArrowWidth, balloonArrowHeight)
				op.SourceRect = &r
			default:
				panic("not reached")
			}
			ax, ay := b.arrowPosition(sw, character)
			dx := ax
			dy := ay - balloonArrowHeight
			if b.arrowFlip(sw, character) {
				// TODO: 4 is an arbitrary number. Define a const.
				dx -= 4
			} else {
				dx += b.partSize()
			}
			op.GeoM.Translate(float64(dx), float64(dy))
			op.GeoM.Concat(*g)
			screen.DrawImage(img, op)
		}
	}
	if b.opened {
		x, y := b.position(sw, character)
		mx, my := b.margin()
		x = (x + mx + b.contentOffsetX) * consts.TileScale
		y = (y + my + b.contentOffsetY) * consts.TileScale
		x += dx
		y += dy
		font.DrawText(screen, b.content, x, y, consts.TextScale, data.TextAlignLeft, color.Black)
	}
}
