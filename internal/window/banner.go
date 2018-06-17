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

package window

import (
	"fmt"
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
	bannerMaxCount = 8
	bannerWidth    = 160
	bannerHeight   = 75
	bannerPaddingX = 4
)

type banner struct {
	interpreterID int
	contentID     data.UUID
	content       string
	openingCount  int
	closingCount  int
	opened        bool
	background    data.MessageBackground
	positionType  data.MessagePositionType
	textAlign     data.TextAlign
	playerY       int
	messageStyle  *data.MessageStyle
	typingEffect  *typingEffect
	eventID       int
}

func (b *banner) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("interpreterId")
	e.EncodeInt(b.interpreterID)

	e.EncodeString("contentID")
	e.EncodeInterface(b.contentID)

	e.EncodeString("content")
	e.EncodeString(b.content)

	e.EncodeString("openingCount")
	e.EncodeInt(b.openingCount)

	e.EncodeString("closingCount")
	e.EncodeInt(b.closingCount)

	e.EncodeString("opened")
	e.EncodeBool(b.opened)

	e.EncodeString("positionType")
	e.EncodeString(string(b.positionType))

	e.EncodeString("textAlign")
	e.EncodeString(string(b.textAlign))

	e.EncodeString("background")
	e.EncodeString(string(b.background))

	e.EncodeString("messageStyle")
	e.EncodeAny(b.messageStyle)

	e.EncodeString("typingEffect")
	e.EncodeInterface(b.typingEffect)

	e.EncodeString("eventID")
	e.EncodeInt(b.eventID)

	e.EndMap()
	return e.Flush()
}

func (b *banner) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "interpreterId":
			b.interpreterID = d.DecodeInt()
		case "contentID":
			d.DecodeInterface(&b.contentID)
		case "content":
			b.content = d.DecodeString()
		case "openingCount":
			b.openingCount = d.DecodeInt()
		case "closingCount":
			b.closingCount = d.DecodeInt()
		case "opened":
			b.opened = d.DecodeBool()
		case "positionType":
			b.positionType = data.MessagePositionType(d.DecodeString())
		case "textAlign":
			b.textAlign = data.TextAlign(d.DecodeString())
		case "background":
			b.background = data.MessageBackground(d.DecodeString())
		case "messageStyle":
			// TODO: This should not be nil?
			if !d.SkipCodeIfNil() {
				b.messageStyle = &data.MessageStyle{}
				d.DecodeAny(b.messageStyle)
			}
		case "typingEffect":
			// TODO: This should not be nil?
			if !d.SkipCodeIfNil() {
				b.typingEffect = &typingEffect{}
				d.DecodeInterface(b.typingEffect)
			}
		case "eventID":
			b.eventID = d.DecodeInt()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("window: banner.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func newBanner(contentID data.UUID, content string, eventID int, background data.MessageBackground, positionType data.MessagePositionType, textAlign data.TextAlign, interpreterID int, messageStyle *data.MessageStyle) *banner {
	font.DrawTextToScratchPad(content, consts.TextScale, lang.Get())

	b := &banner{
		interpreterID: interpreterID,
		contentID:     contentID,
		content:       content,
		background:    background,
		positionType:  positionType,
		textAlign:     textAlign,
		messageStyle:  messageStyle,
		eventID:       eventID,
	}
	return b
}

func (b *banner) SetContent(content string) {
	b.content = content
	if b.typingEffect != nil {
		b.typingEffect.SetContent(b.content)
	}
}

func (b *banner) isClosed() bool {
	return !b.opened && b.openingCount == 0 && b.closingCount == 0
}

func (b *banner) isOpened() bool {
	return b.opened
}

func (b *banner) isAnimating() bool {
	return b.openingCount > 0 || b.closingCount > 0 || b.typingEffect.isAnimating()
}

func (b *banner) skipTypingAnim() {
	b.typingEffect.skipAnim()
}

func (b *banner) open() {
	b.openingCount = bannerMaxCount
	b.typingEffect = newTypingEffect(b.content, b.messageStyle.TypingEffectDelay, b.messageStyle.SoundEffect)
}

func (b *banner) close() {
	b.closingCount = bannerMaxCount
}

func (b *banner) characterAnimFinishTrigger() data.FinishTriggerType {
	if b.messageStyle.CharacterAnim == nil {
		return data.FinishTriggerTypeNone
	}
	return b.messageStyle.CharacterAnim.FinishTrigger
}

func (b *banner) update(playerY int, character *character.Character) error {
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
	}
	b.playerY = playerY
	return nil
}

func (b *banner) playCharacterAnim(character *character.Character) {
	if character == nil {
		return
	}
	CharacterAnim := b.messageStyle.CharacterAnim
	if CharacterAnim == nil {
		return
	}

	if !character.HasStoredState() {
		character.StoreState()
	}
	character.SetImage(CharacterAnim.ImageType, CharacterAnim.Image)
	character.SetStepping(true)
	character.SetSpeed(CharacterAnim.Speed)
}

func (b *banner) stopCharacterAnim(character *character.Character) {
	if character == nil {
		return
	}
	character.RestoreStoredState()
}

func (b *banner) position(screen *ebiten.Image) (int, int) {
	x := 0
	y := 0
	positionType := b.positionType
	if positionType == data.MessagePositionAuto {
		if b.playerY < consts.TileYNum/2 {
			positionType = data.MessagePositionBottom
		} else {
			positionType = data.MessagePositionTop
		}
	}
	_, sh := screen.Size()
	cy := (sh/consts.TileScale - bannerHeight) / 2
	ty := (consts.GuaranteedVisibleMapHeight - bannerHeight) / 2

	switch positionType {
	case data.MessagePositionBottom:
		y = cy + ty
	case data.MessagePositionMiddle:
		y = cy
	case data.MessagePositionTop:
		y = cy - ty
	}
	return x, y
}

func (b *banner) draw(screen *ebiten.Image, offsetX, offsetY int) {
	textScale := consts.TextScale
	textEdge := false
	rate := 0.0
	switch {
	case b.opened:
		rate = 1
	case b.openingCount > 0:
		rate = 1 - float64(b.openingCount)/float64(bannerMaxCount)
	case b.closingCount > 0:
		rate = float64(b.closingCount) / float64(bannerMaxCount)
	}
	sw, _ := screen.Size()
	dx := math.Floor(float64(sw/consts.TileScale-consts.MapWidth)/2 + float64(offsetX))
	dy := math.Floor(float64(offsetY))

	switch b.background {
	case data.MessageBackgroundDim:
		// TODO
	case data.MessageBackgroundTransparent:
		textEdge = true
		textScale = consts.BigTextScale
	case data.MessageBackgroundBanner:
		if rate > 0 {
			img := assets.GetImage("system/game/banner.png")
			x, y := b.position(screen)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			op.GeoM.Translate(float64(dx), float64(dy))
			op.GeoM.Scale(consts.TileScale, consts.TileScale)
			op.ColorM.Scale(1, 1, 1, rate)
			screen.DrawImage(img, op)
		}
	}

	if b.opened {
		displayTextLength := b.typingEffect.getCurrentTextRuneCount()
		_, th := font.MeasureSize(b.content)
		x, y := b.position(screen)
		x = (x + bannerPaddingX) * consts.TileScale
		y = y*consts.TileScale + (bannerHeight*consts.TileScale-th*textScale)/2
		switch b.textAlign {
		case data.TextAlignLeft:
		case data.TextAlignCenter:
			x += (consts.MapWidth - 2*bannerPaddingX) * consts.TileScale / 2
		case data.TextAlignRight:
			x += (consts.MapWidth - 2*bannerPaddingX) * consts.TileScale
		}
		x += int(dx * consts.TileScale)
		y += int(dy * consts.TileScale)

		if textEdge {
			shadowColor := color.RGBA{0, 0, 0, 64}
			font.DrawText(screen, b.content, x+textScale*2, y, textScale, b.textAlign, shadowColor, displayTextLength)
			font.DrawText(screen, b.content, x-textScale*2, y, textScale, b.textAlign, shadowColor, displayTextLength)
			font.DrawText(screen, b.content, x, y+textScale*2, textScale, b.textAlign, shadowColor, displayTextLength)
			font.DrawText(screen, b.content, x, y-textScale*2, textScale, b.textAlign, shadowColor, displayTextLength)
			font.DrawText(screen, b.content, x+textScale, y+textScale, textScale, b.textAlign, shadowColor, displayTextLength)
			font.DrawText(screen, b.content, x-textScale, y+textScale, textScale, b.textAlign, shadowColor, displayTextLength)
			font.DrawText(screen, b.content, x+textScale, y-textScale, textScale, b.textAlign, shadowColor, displayTextLength)
			font.DrawText(screen, b.content, x-textScale, y-textScale, textScale, b.textAlign, shadowColor, displayTextLength)

			edgeColor := color.Black
			font.DrawText(screen, b.content, x+textScale, y, textScale, b.textAlign, edgeColor, displayTextLength)
			font.DrawText(screen, b.content, x-textScale, y, textScale, b.textAlign, edgeColor, displayTextLength)
			font.DrawText(screen, b.content, x, y+textScale, textScale, b.textAlign, edgeColor, displayTextLength)
			font.DrawText(screen, b.content, x, y-textScale, textScale, b.textAlign, edgeColor, displayTextLength)
		}
		font.DrawText(screen, b.content, x, y, textScale, b.textAlign, color.White, displayTextLength)
	}
}
