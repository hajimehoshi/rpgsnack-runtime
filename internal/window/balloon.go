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
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
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
	contentID      data.UUID
	content        string
	contentOffsetX int
	contentOffsetY int
	openingCount   int
	closingCount   int
	opened         bool
	balloonType    data.BalloonType
	messageStyle   *data.MessageStyle
	typingEffect   *typingEffect
	checked        bool

	offscreen *ebiten.Image
}

func (b *balloon) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("interpreterId")
	e.EncodeInt(b.interpreterID)

	e.EncodeString("x")
	e.EncodeInt(b.x)

	e.EncodeString("y")
	e.EncodeInt(b.y)

	e.EncodeString("width")
	e.EncodeInt(b.width)

	e.EncodeString("height")
	e.EncodeInt(b.height)

	e.EncodeString("hasArrow")
	e.EncodeBool(b.hasArrow)

	e.EncodeString("eventID")
	e.EncodeInt(b.eventID)

	e.EncodeString("contentID")
	e.EncodeInterface(&b.contentID)

	e.EncodeString("content")
	e.EncodeString(b.content)

	e.EncodeString("contentOffsetX")
	e.EncodeInt(b.contentOffsetX)

	e.EncodeString("contentOffsetY")
	e.EncodeInt(b.contentOffsetY)

	e.EncodeString("openingCount")
	e.EncodeInt(b.openingCount)

	e.EncodeString("closingCount")
	e.EncodeInt(b.closingCount)

	e.EncodeString("opened")
	e.EncodeBool(b.opened)

	e.EncodeString("balloonType")
	e.EncodeString(string(b.balloonType))

	e.EncodeString("messageStyle")
	e.EncodeAny(b.messageStyle)

	e.EncodeString("typingEffect")
	e.EncodeInterface(b.typingEffect)

	e.EncodeString("checked")
	e.EncodeBool(b.checked)

	e.EndMap()
	return e.Flush()
}

func (b *balloon) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "interpreterId":
			b.interpreterID = d.DecodeInt()
		case "x":
			b.x = d.DecodeInt()
		case "y":
			b.y = d.DecodeInt()
		case "width":
			b.width = d.DecodeInt()
		case "height":
			b.height = d.DecodeInt()
		case "hasArrow":
			b.hasArrow = d.DecodeBool()
		case "eventID":
			b.eventID = d.DecodeInt()
		case "contentID":
			d.DecodeInterface(&b.contentID)
		case "content":
			b.content = d.DecodeString()
		case "contentOffsetX":
			b.contentOffsetX = d.DecodeInt()
		case "contentOffsetY":
			b.contentOffsetY = d.DecodeInt()
		case "openingCount":
			b.openingCount = d.DecodeInt()
		case "closingCount":
			b.closingCount = d.DecodeInt()
		case "opened":
			b.opened = d.DecodeBool()
		case "balloonType":
			b.balloonType = data.BalloonType(d.DecodeString())
		case "messageStyle":
			if !d.SkipCodeIfNil() {
				b.messageStyle = &data.MessageStyle{}
				d.DecodeAny(b.messageStyle)
			}
		case "typingEffect":
			if !d.SkipCodeIfNil() {
				b.typingEffect = &typingEffect{}
				d.DecodeInterface(b.typingEffect)
			}
		case "checked":
			b.checked = d.DecodeBool()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("window: balloon.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func newBalloon(x, y, width, height int, contentID data.UUID, content string, balloonType data.BalloonType, interpreterID int, messageStyle *data.MessageStyle, checked bool) *balloon {
	font.DrawTextToScratchPad(content, consts.TextScale, lang.Get())

	b := &balloon{
		interpreterID: interpreterID,
		contentID:     contentID,
		content:       content,
		x:             x,
		y:             y,
		balloonType:   balloonType,
		messageStyle:  messageStyle,
		checked:       checked,
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

func newBalloonWithArrow(contentID data.UUID, content string, balloonType data.BalloonType, eventID int, interpreterID int, messageStyle *data.MessageStyle) *balloon {
	font.DrawTextToScratchPad(content, consts.TextScale, lang.Get())

	b := &balloon{
		interpreterID: interpreterID,
		contentID:     contentID,
		hasArrow:      true,
		eventID:       eventID,
		balloonType:   balloonType,
		messageStyle:  messageStyle,
	}
	b.setContent(content)
	return b
}

func (b *balloon) setContent(content string) {
	b.content = content
	if b.hasArrow {
		w, h, contentOffsetX, contentOffsetY := balloonSizeFromContent(visibleContent(b.content), b.balloonType)
		b.width = w
		b.height = h
		b.contentOffsetX = contentOffsetX
		b.contentOffsetY = contentOffsetY
	}
	if b.typingEffect != nil {
		b.typingEffect.SetContent(b.content)
	}
	b.offscreen = nil
}

func (b *balloon) trySkipTypingAnim() {
	b.typingEffect.trySkipAnim()
}

func (b *balloon) arrowPosition(screenWidth int, character *character.Character) (int, int) {
	if !b.hasArrow {
		panic("windows: hasArrow must be true at arrowPosition")
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
	if consts.MapWidth < x+b.width {
		x = consts.MapWidth - b.width
	}
	if x < 0 {
		x = 0
	}
	y := ay - b.height - 4
	return x, y
}

func (b *balloon) isClosed() bool {
	return !b.opened && b.openingCount == 0 && b.closingCount == 0
}

func (b *balloon) isOpened() bool {
	return b.opened
}

func (b *balloon) isAnimating() bool {
	return b.openingCount > 0 || b.closingCount > 0 || b.typingEffect.isAnimating()
}

func (b *balloon) open() {
	// TODO: This should be called only in the constructor?
	b.openingCount = balloonMaxCount
	b.typingEffect = newTypingEffect(b.content, b.messageStyle.TypingEffectDelay, b.messageStyle.SoundEffect)
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

func (b *balloon) characterAnimFinishTrigger() data.FinishTriggerType {
	if b.messageStyle.CharacterAnim == nil {
		return data.FinishTriggerTypeNone
	}
	return b.messageStyle.CharacterAnim.FinishTrigger
}

func (b *balloon) update(character *character.Character) {
	if b.closingCount > 0 {
		b.closingCount--
		b.opened = false
		if b.characterAnimFinishTrigger() == data.FinishTriggerTypeWindow {
			b.stopCharacterAnim(character)
		}
	}
	if b.openingCount > 0 {
		b.openingCount--
		if b.openingCount == 0 {
			b.opened = true
			b.playCharacterAnim(character)
		}
	}
	if b.opened && b.typingEffect.isAnimating() {
		b.typingEffect.update()
		if !b.typingEffect.isAnimating() && b.characterAnimFinishTrigger() == data.FinishTriggerTypeMessage {
			b.stopCharacterAnim(character)
		}
		if b.typingEffect.shouldCloseWindow() {
			b.close()
		}
	}
}

func (b *balloon) playCharacterAnim(character *character.Character) {
	if character == nil {
		return
	}
	a := b.messageStyle.CharacterAnim
	if a == nil {
		return
	}

	if !character.HasStoredState() {
		character.StoreState()
	}
	character.SetImage(a.ImageType, a.Image)
	character.SetStepping(true)
	character.SetSpeed(a.Speed)
}

func (b *balloon) stopCharacterAnim(character *character.Character) {
	if character == nil {
		return
	}
	character.RestoreStoredState()
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
	}
	g := ebiten.GeoM{}
	g.Translate(-cx, -cy)
	rate := b.openingRate()
	g.Scale(rate, rate)
	g.Translate(cx, cy)
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

func (b *balloon) assetImage() *ebiten.Image {
	if b.balloonType == data.BalloonTypeShout {
		return assets.GetImage("system/game/shout.png")
	} else {
		return assets.GetImage("system/game/balloon.png")
	}
}

func (b *balloon) ensureOffscreen() {
	if b.offscreen != nil {
		return
	}

	b.offscreen, _ = ebiten.NewImage(b.width, b.height, ebiten.FilterNearest)

	img := b.assetImage()

	op := &ebiten.DrawImageOptions{}
	pw, ph := b.width/b.partSize(), b.height/b.partSize()
	s := b.partSize()
	for j := 0; j < ph; j++ {
		for i := 0; i < pw; i++ {
			op.GeoM.Reset()
			sx, sy := 0, 0
			switch i {
			case 0:
			default:
				sx += s
			case pw - 1:
				sx += s * 2
			}
			switch j {
			case 0:
			default:
				sy += s
			case ph - 1:
				sy += s * 2
			}
			op.GeoM.Translate(float64(i*s), float64(j*s))
			b.offscreen.DrawImage(img.SubImage(image.Rect(sx, sy, sx+s, sy+s)).(*ebiten.Image), op)
		}
	}
}

func (b *balloon) draw(screen *ebiten.Image, character *character.Character, offsetX, offsetY int) {
	sw, _ := screen.Size()
	dx := math.Floor(float64(sw/consts.TileScale-consts.MapWidth)/2 + float64(offsetX))
	dy := math.Floor(float64(offsetY))
	if b.openingRate() > 0 {
		b.ensureOffscreen()

		img := b.assetImage()
		op := &ebiten.DrawImageOptions{}
		g := b.geoMForRate(screen, character)
		g.Translate(dx, dy)
		tx, ty := b.position(sw, character)
		op.GeoM.Translate(float64(tx), float64(ty))
		op.GeoM.Concat(*g)
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		if b.checked {
			op.ColorM.Scale(1.0, 1.0, 0.5, 1.0)
		}
		screen.DrawImage(b.offscreen, op)
		if b.checked {
			checkOp := &ebiten.DrawImageOptions{}
			checkOp.GeoM.Translate(float64(tx), float64(ty))
			checkOp.GeoM.Translate(float64(b.width-16), 4)
			checkOp.GeoM.Concat(*g)
			checkOp.GeoM.Scale(consts.TileScale, consts.TileScale)
			r := image.Rect(12, 12, 24, 24)
			screen.DrawImage(img.SubImage(r).(*ebiten.Image), checkOp)
		}

		if b.hasArrow && (b.balloonType == data.BalloonTypeNormal ||
			b.balloonType == data.BalloonTypeThink) {
			op := &ebiten.DrawImageOptions{}
			var r image.Rectangle
			switch t := b.balloonType; t {
			case data.BalloonTypeNormal:
				r = image.Rect(12, 0, 12+balloonArrowWidth, balloonArrowHeight)
			case data.BalloonTypeThink:
				r = image.Rect(18, 0, 18+balloonArrowWidth, balloonArrowHeight)
			default:
				panic(fmt.Sprintf("window: invalid balloon type: %d", t))
			}
			ax, ay := b.arrowPosition(sw, character)
			tx := ax
			ty := ay - balloonArrowHeight
			tx += b.partSize()

			const maxArrowX = consts.MapWidth - balloonArrowWidth - 8
			if tx > maxArrowX {
				op.GeoM.Scale(-1, 1)
				tx = ax - b.partSize()
			}
			op.GeoM.Translate(float64(tx), float64(ty))
			op.GeoM.Concat(*g)
			op.GeoM.Scale(consts.TileScale, consts.TileScale)
			screen.DrawImage(img.SubImage(r).(*ebiten.Image), op)
		}
	}
	if b.opened {
		x, y := b.position(sw, character)
		mx, my := b.margin()
		x = (x + mx + b.contentOffsetX) * consts.TileScale
		y = (y + my + b.contentOffsetY) * consts.TileScale
		x += int(dx * consts.TileScale)
		y += int(dy * consts.TileScale)
		b.typingEffect.draw(screen, x, y, consts.TextScale, data.TextAlignLeft, color.Black, nil, nil)
	}
}
