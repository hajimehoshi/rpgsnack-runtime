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
	"encoding/json"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
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
	positionType  data.MessagePositionType
}

type tmpBanner struct {
	InterpreterID int                      `json:"interpreterId"`
	Content       string                   `json:"content"`
	OpeningCount  int                      `json:"openingCount"`
	ClosingCount  int                      `json:"closingCount"`
	Opened        bool                     `json:"opened"`
	PositionType  data.MessagePositionType `json:"positionType"`
}

func (b *banner) MarshalJSON() ([]uint8, error) {
	tmp := &tmpBanner{
		InterpreterID: b.interpreterID,
		Content:       b.content,
		Opened:        b.opened,
		OpeningCount:  b.openingCount,
		ClosingCount:  b.closingCount,
		PositionType:  b.positionType,
	}
	return json.Marshal(tmp)
}

func (b *banner) UnmarshalJSON(data []uint8) error {
	var tmp *tmpBanner
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	b.interpreterID = tmp.InterpreterID
	b.content = tmp.Content
	b.opened = tmp.Opened
	b.openingCount = tmp.OpeningCount
	b.closingCount = tmp.ClosingCount
	b.positionType = tmp.PositionType
	return nil
}

func newBanner(content string, positionType data.MessagePositionType, interpreterID int) *banner {
	b := &banner{
		interpreterID: interpreterID,
		content:       content,
		positionType:  positionType,
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
		y = screenHeight / scene.TileScale
	case data.MessagePositionMiddle:
		y = screenHeight / (scene.TileScale * 2)
	case data.MessagePositionTop:
		y = 0
	}
	return x, y
}

func (b *banner) draw(screen *ebiten.Image, character *character.Character) {
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
	dx := (sw - scene.TileXNum*scene.TileSize*scene.TileScale) / 2
	dy := 0
	if rate > 0 {
		img := assets.GetImage("banner.png")
		x, y := b.position(sh)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		op.GeoM.Scale(scene.TileScale, scene.TileScale)
		op.GeoM.Translate(float64(dx), float64(dy))
		op.ColorM.Scale(1, 1, 1, rate)
		screen.DrawImage(img, op)
	}
	if b.opened {
		x, y := b.position(sh)
		x = (x + bannerMarginX) * scene.TileScale
		y = (y + bannerMarginY) * scene.TileScale
		x += dx
		y += dy
		font.DrawText(screen, b.content, x, y, scene.TextScale, color.White)
	}
}
