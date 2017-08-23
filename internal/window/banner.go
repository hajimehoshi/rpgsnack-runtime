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

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
)

const (
	bannerMaxCount = 8
	bannerWidth    = 160
	bannerHeight   = 60
	bannerMarginX  = 4
	bannerMarginY  = 12
)

type banner struct {
	interpreterID int
	content       string
	openingCount  int
	closingCount  int
	opened        bool
	background    data.MessageBackground
	positionType  data.MessagePositionType
	textAlign     data.TextAlign
}

func (b *banner) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("interpreterId")
	e.EncodeInt(b.interpreterID)

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
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("window: banner.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func newBanner(content string, background data.MessageBackground, positionType data.MessagePositionType, textAlign data.TextAlign, interpreterID int) *banner {
	b := &banner{
		interpreterID: interpreterID,
		content:       content,
		background:    background,
		positionType:  positionType,
		textAlign:     textAlign,
	}
	return b
}

func (b *banner) isClosed() bool {
	return !b.opened && b.openingCount == 0 && b.closingCount == 0
}

func (b *banner) isOpened() bool {
	return b.opened
}

func (b *banner) isAnimating() bool {
	return b.openingCount > 0 || b.closingCount > 0
}

func (b *banner) open() {
	b.openingCount = bannerMaxCount
}

func (b *banner) close() {
	b.closingCount = bannerMaxCount
}

func (b *banner) update() error {
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

func (b *banner) position(screenHeight int) (int, int) {
	x := 0
	y := 0
	switch b.positionType {
	case data.MessagePositionBottom:
		y = screenHeight / consts.TileScale
	case data.MessagePositionMiddle:
		y = screenHeight / (consts.TileScale * 2)
	case data.MessagePositionTop:
		y = 0
	}
	return x, y
}

func (b *banner) draw(screen *ebiten.Image, character *character.Character) {
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
	sw, sh := screen.Size()
	dx := (sw - consts.TileXNum*consts.TileSize*consts.TileScale) / 2
	dy := 0

	switch b.background {
	case data.MessageBackgroundDim:
		// TODO
	case data.MessageBackgroundTransparent:
		textEdge = true
		textScale = consts.BigTextScale
	case data.MessageBackgroundBanner:
		if rate > 0 {
			img := assets.GetImage("system/banner.png")
			x, y := b.position(sh)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			op.GeoM.Scale(consts.TileScale, consts.TileScale)
			op.GeoM.Translate(float64(dx), float64(dy))
			op.ColorM.Scale(1, 1, 1, rate)
			screen.DrawImage(img, op)
		}
	}

	if b.opened {
		x, y := b.position(sh)
		x = (x + bannerMarginX) * consts.TileScale
		y = (y + bannerMarginY) * consts.TileScale
		switch b.textAlign {
		case data.TextAlignLeft:
		case data.TextAlignCenter:
			x += (consts.TileXNum*consts.TileSize - 2*bannerMarginX) * consts.TileScale / 2
		case data.TextAlignRight:
			x += (consts.TileXNum*consts.TileSize - 2*bannerMarginX) * consts.TileScale
		}
		x += dx
		y += dy

		if textEdge {
			font.DrawText(screen, b.content, x+textScale, y, textScale, b.textAlign, color.Black)
			font.DrawText(screen, b.content, x-textScale, y, textScale, b.textAlign, color.Black)
			font.DrawText(screen, b.content, x, y+textScale, textScale, b.textAlign, color.Black)
			font.DrawText(screen, b.content, x, y-textScale, textScale, b.textAlign, color.Black)
		}
		font.DrawText(screen, b.content, x, y, textScale, b.textAlign, color.White)
	}
}
