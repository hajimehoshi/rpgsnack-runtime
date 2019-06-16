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
	content     []rune
	soundEffect string
	delay       int

	// These members are not dumped.
	index                     int
	delayCount                int
	isSEPlayedInPreviousFrame bool
}

func newTypingEffect(content string, delay int, soundEffect string) *typingEffect {
	t := &typingEffect{
		soundEffect: soundEffect,
		delay:       delay,
	}
	t.SetContent(content, false)
	return t
}

func (t *typingEffect) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("content")
	e.EncodeString(string(t.content))

	e.EncodeString("soundEffect")
	e.EncodeString(t.soundEffect)

	e.EncodeString("delay")
	e.EncodeInt(t.delay)

	e.EndMap()
	return e.Flush()
}

func (t *typingEffect) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	content := ""
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "content":
			content = d.DecodeString()
		case "soundEffect":
			t.soundEffect = d.DecodeString()
		case "delay":
			t.delay = d.DecodeInt()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("window: typingEffect.DecodeMsgpack failed: %v", err)
	}
	t.SetContent(content, false)
	return nil
}

const (
	controlForceQuit = `\^`
	controlWaitShort = `\.`
	controlWaitLong  = `\|`
)

func visibleContent(content string) string {
	c := content
	if strings.Contains(c, controlForceQuit) {
		c = c[:strings.Index(c, controlForceQuit)]
	}
	c = strings.Replace(c, controlWaitShort, "", -1)
	c = strings.Replace(c, controlWaitLong, "", -1)
	return c
}

func (t *typingEffect) visibleContent() []rune {
	return []rune(visibleContent(string(t.content)))
}

func (t *typingEffect) visibleIndex() int {
	i := t.index
	i -= strings.Count(string(t.content[:t.index]), controlWaitShort) * len(controlWaitShort)
	i -= strings.Count(string(t.content[:t.index]), controlWaitLong) * len(controlWaitLong)

	max := len(t.visibleContent())
	if i > max {
		i = max
	}
	return i
}

func (t *typingEffect) forceQuit() bool {
	return strings.Contains(string(t.content), controlForceQuit)
}

func (t *typingEffect) isAnimating() bool {
	return t.delayCount > 0 || t.index < t.lastIndex()
}

func (t *typingEffect) lastIndex() int {
	i := len(t.content)
	c := string(t.content)
	if strings.Contains(c, controlForceQuit) {
		i = len([]rune(c[:strings.Index(c, controlForceQuit)]))
	}
	return i
}

func (t *typingEffect) shouldCloseWindow() bool {
	return t.index >= t.lastIndex() && t.forceQuit()
}

func (t *typingEffect) trySkipAnim() {
	if t.forceQuit() {
		return
	}
	// Give 1 delay so that isAnimation() still returns false in the frame when trySkipAnim() is called.
	t.delayCount = 1
	t.index = t.lastIndex()
}

func (t *typingEffect) SetContent(content string, overwrite bool) {
	t.content = []rune(content)
	t.delayCount = t.delay
	if t.index > 0 || t.delay == 0 || overwrite {
		t.index = t.lastIndex()
	}
}

func (t *typingEffect) update() {
	if t.delayCount > 0 {
		t.delayCount--
	}
	if t.delayCount == 0 {
		switch {
		case t.hasControl([]rune(controlWaitShort)):
			t.delayCount = 15
			t.index += len(controlWaitShort)
			return
		case t.hasControl([]rune(controlWaitLong)):
			t.delayCount = 60
			t.index += len(controlWaitLong)
			return
		}

		played := false
		if t.index < t.lastIndex() {
			t.index++
			if !t.isSEPlayedInPreviousFrame && !t.isLastRuneSpace() {
				played = t.playSE()
			}
		}
		t.isSEPlayedInPreviousFrame = played
		if t.index < t.lastIndex() {
			t.delayCount = t.delay
		}
	}
}

func (t *typingEffect) isLastRuneSpace() bool {
	if t.index == 0 {
		return false
	}
	return t.content[t.index-1] == ' '
}

func (t *typingEffect) hasControl(control []rune) bool {
	if t.index+len(control) > t.lastIndex() {
		return false
	}
	if string(t.content[t.index:t.index+len(control)]) == string(control) {
		return true
	}
	return false
}

func (t *typingEffect) playSE() bool {
	if t.soundEffect == "" {
		return false
	}
	audio.PlaySE(t.soundEffect, 0.2)
	return true
}

func (t *typingEffect) draw(screen *ebiten.Image, x, y int, textScale int, textAlign data.TextAlign, textColor color.Color, edgeColor color.Color, shadowColor color.Color) {
	i := t.visibleIndex()
	str := string(t.visibleContent())
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
