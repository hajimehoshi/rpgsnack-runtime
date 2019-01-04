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
	"strings"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
)

type typingEffect struct {
	index                     int
	delayCount                int
	content                   []rune
	soundEffect               string
	isSEPlayedInPreviousFrame bool
	delay                     int
	forceQuit                 bool
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

	e.EncodeString("index")
	e.EncodeInt(t.index)

	e.EncodeString("delayCount")
	e.EncodeInt(t.delayCount)

	e.EncodeString("content")
	e.EncodeString(string(t.content))

	e.EncodeString("soundEffect")
	e.EncodeString(t.soundEffect)

	e.EncodeString("playedSE")
	e.EncodeBool(t.isSEPlayedInPreviousFrame)

	e.EncodeString("delay")
	e.EncodeInt(t.delay)

	e.EncodeString("forceQuit")
	e.EncodeBool(t.forceQuit)

	e.EndMap()
	return e.Flush()
}

func (t *typingEffect) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "index":
			t.index = d.DecodeInt()
		case "delayCount":
			t.delayCount = d.DecodeInt()
		case "content":
			t.content = []rune(d.DecodeString())
		case "soundEffect":
			t.soundEffect = d.DecodeString()
		case "playedSE":
			t.isSEPlayedInPreviousFrame = d.DecodeBool()
		case "delay":
			t.delay = d.DecodeInt()
		case "forceQuit":
			t.forceQuit = d.DecodeBool()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("window: typingEffect.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (t *typingEffect) isAnimating() bool {
	return t.delayCount > 0
}

func (t *typingEffect) shouldCloseWindow() bool {
	return !t.isAnimating() && t.forceQuit
}

func (t *typingEffect) trySkipAnim() {
	if t.forceQuit {
		return
	}
	t.delayCount = 0
}

func (t *typingEffect) SetContent(content string) {
	t.content = []rune(content)
	if strings.Contains(content, `\^`) {
		t.content = []rune(content[:strings.Index(content, `\^`)])
		t.forceQuit = true
	}
	t.delayCount = t.delay
	if t.delay == 0 {
		t.index = len(t.content)
	}
}

func (t *typingEffect) update() {
	if t.delayCount > 0 {
		t.delayCount--
	}
	if t.delayCount == 0 {
		if t.index < len(t.content) {
			t.index++
			t.playSE()
		}
		if t.index < len(t.content) {
			t.delayCount = t.delay
		}
	}
}

func (t *typingEffect) playSE() {
	if t.soundEffect == "" {
		return
	}
	if !t.isSEPlayedInPreviousFrame && t.content[t.index-1] != ' ' {
		audio.PlaySE(t.soundEffect, 0.2)
		t.isSEPlayedInPreviousFrame = true
	} else {
		t.isSEPlayedInPreviousFrame = false
	}
}

func (t *typingEffect) draw(screen *ebiten.Image, x, y int, textScale int, textAlign data.TextAlign, textColor color.Color, edgeColor color.Color, shadowColor color.Color) {
	i := t.index
	str := string(t.content)
	if shadowColor != nil {
		// Shadow
		font.DrawText(screen, str, x+textScale*2, y, textScale, textAlign, shadowColor, i)
		font.DrawText(screen, str, x-textScale*2, y, textScale, textAlign, shadowColor, i)
		font.DrawText(screen, str, x, y+textScale*2, textScale, textAlign, shadowColor, i)
		font.DrawText(screen, str, x, y-textScale*2, textScale, textAlign, shadowColor, i)
		font.DrawText(screen, str, x+textScale, y+textScale, textScale, textAlign, shadowColor, i)
		font.DrawText(screen, str, x-textScale, y+textScale, textScale, textAlign, shadowColor, i)
		font.DrawText(screen, str, x+textScale, y-textScale, textScale, textAlign, shadowColor, i)
		font.DrawText(screen, str, x-textScale, y-textScale, textScale, textAlign, shadowColor, i)

		// Edge
		font.DrawText(screen, str, x+textScale, y, textScale, textAlign, edgeColor, i)
		font.DrawText(screen, str, x-textScale, y, textScale, textAlign, edgeColor, i)
		font.DrawText(screen, str, x, y+textScale, textScale, textAlign, edgeColor, i)
		font.DrawText(screen, str, x, y-textScale, textScale, textAlign, edgeColor, i)
	}
	font.DrawText(screen, str, x, y, textScale, textAlign, textColor, i)
}
