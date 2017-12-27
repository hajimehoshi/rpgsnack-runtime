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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/vmihailenco/msgpack"
)

type typingEffect struct {
	animCount                 int
	animMaxCount              int
	allTextDisplayed          bool
	content                   string
	soundEffect               string
	isSEPlayedInPreviousFrame bool
}

func newTypingEffect(content string, delay int, soundEffect string) *typingEffect {
	t := &typingEffect{
		content:      content,
		animMaxCount: len([]rune(content)) * delay,
		soundEffect:  soundEffect,
	}
	return t
}

func (t *typingEffect) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("animCount")
	e.EncodeInt(t.animCount)

	e.EncodeString("animMaxCount")
	e.EncodeInt(t.animMaxCount)

	e.EncodeString("allTextDisplayed")
	e.EncodeBool(t.allTextDisplayed)

	e.EncodeString("content")
	e.EncodeString(t.content)

	e.EncodeString("soundEffect")
	e.EncodeString(t.soundEffect)

	e.EncodeString("playedSE")
	e.EncodeBool(t.isSEPlayedInPreviousFrame)

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
		case "animMaxCount":
			t.animMaxCount = d.DecodeInt()
		case "allTextDisplayed":
			t.allTextDisplayed = d.DecodeBool()
		case "content":
			t.content = d.DecodeString()
		case "soundEffect":
			t.soundEffect = d.DecodeString()
		case "playedSE":
			t.isSEPlayedInPreviousFrame = d.DecodeBool()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("window: typingEffect.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (t *typingEffect) isAnimating() bool {
	return !t.allTextDisplayed
}

func (t *typingEffect) skipAnim() {
	t.animCount = t.animMaxCount
}

func (t *typingEffect) update() {
	prevTextRuneCount := t.getCurrentTextRuneCount()
	if t.animCount < t.animMaxCount {
		t.animCount++
	}
	if t.animCount == t.animMaxCount {
		t.allTextDisplayed = true
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

func (t *typingEffect) getCurrentTextRuneCount() int {
	if t.animMaxCount > 0 {
		return len([]rune(t.content)) * t.animCount / t.animMaxCount
	}
	return len([]rune(t.content))
}
