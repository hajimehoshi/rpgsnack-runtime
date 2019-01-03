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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
)

type typingEffect struct {
	animCount                 int
	content                   string
	soundEffect               string
	isSEPlayedInPreviousFrame bool
	delay                     int
}

func newTypingEffect(content string, delay int, soundEffect string) *typingEffect {
	t := &typingEffect{
		soundEffect: soundEffect,
		delay:       delay,
	}
	t.SetContent(content)
	return t
}

func (t *typingEffect) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("animCount")
	e.EncodeInt(t.animCount)

	e.EncodeString("content")
	e.EncodeString(t.content)

	e.EncodeString("soundEffect")
	e.EncodeString(t.soundEffect)

	e.EncodeString("playedSE")
	e.EncodeBool(t.isSEPlayedInPreviousFrame)

	e.EncodeString("delay")
	e.EncodeInt(t.delay)

	e.EndMap()
	return e.Flush()
}

func (t *typingEffect) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "animCount":
			t.animCount = d.DecodeInt()
		case "content":
			t.content = d.DecodeString()
		case "soundEffect":
			t.soundEffect = d.DecodeString()
		case "playedSE":
			t.isSEPlayedInPreviousFrame = d.DecodeBool()
		case "delay":
			t.delay = d.DecodeInt()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("window: typingEffect.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (t *typingEffect) animMaxCount() int {
	return len([]rune(t.content)) * t.delay
}

func (t *typingEffect) isAnimating() bool {
	return t.animCount < t.animMaxCount()
}

func (t *typingEffect) skipAnim() {
	t.animCount = t.animMaxCount()
}

func (t *typingEffect) SetContent(content string) {
	t.content = content
	// Finish animation forcely.
	if t.animCount > 0 {
		t.animCount = t.animMaxCount()
	}
}

func (t *typingEffect) update() {
	prevTextRuneCount := t.getCurrentTextRuneCount()
	if t.animCount < t.animMaxCount() {
		t.animCount++
	}
	currentTextRuneCount := t.getCurrentTextRuneCount()
	if currentTextRuneCount > 0 && currentTextRuneCount != prevTextRuneCount {
		t.playSE()
	}
}

func (t *typingEffect) playSE() {
	if t.soundEffect == "" {
		return
	}
	if !t.isSEPlayedInPreviousFrame && t.content[t.getCurrentTextRuneCount()-1] != ' ' {
		audio.PlaySE(t.soundEffect, 0.2)
		t.isSEPlayedInPreviousFrame = true
	} else {
		t.isSEPlayedInPreviousFrame = false
	}
}

func (t *typingEffect) draw(screen *ebiten.Image, x, y int, textScale int, textAlign data.TextAlign, textColor color.Color, edgeColor color.Color, shadowColor color.Color) {
	c := t.getCurrentTextRuneCount()
	if shadowColor != nil {
		// Shadow
		font.DrawText(screen, t.content, x+textScale*2, y, textScale, textAlign, shadowColor, c)
		font.DrawText(screen, t.content, x-textScale*2, y, textScale, textAlign, shadowColor, c)
		font.DrawText(screen, t.content, x, y+textScale*2, textScale, textAlign, shadowColor, c)
		font.DrawText(screen, t.content, x, y-textScale*2, textScale, textAlign, shadowColor, c)
		font.DrawText(screen, t.content, x+textScale, y+textScale, textScale, textAlign, shadowColor, c)
		font.DrawText(screen, t.content, x-textScale, y+textScale, textScale, textAlign, shadowColor, c)
		font.DrawText(screen, t.content, x+textScale, y-textScale, textScale, textAlign, shadowColor, c)
		font.DrawText(screen, t.content, x-textScale, y-textScale, textScale, textAlign, shadowColor, c)

		// Edge
		font.DrawText(screen, t.content, x+textScale, y, textScale, textAlign, edgeColor, c)
		font.DrawText(screen, t.content, x-textScale, y, textScale, textAlign, edgeColor, c)
		font.DrawText(screen, t.content, x, y+textScale, textScale, textAlign, edgeColor, c)
		font.DrawText(screen, t.content, x, y-textScale, textScale, textAlign, edgeColor, c)
	}
	font.DrawText(screen, t.content, x, y, textScale, textAlign, textColor, c)
}

func (t *typingEffect) getCurrentTextRuneCount() int {
	if t.animMaxCount() > 0 {
		return len([]rune(t.content)) * t.animCount / t.animMaxCount()
	}
	return len([]rune(t.content))
}
