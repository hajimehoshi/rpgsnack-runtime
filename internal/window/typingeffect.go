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
	"unicode/utf8"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/vmihailenco/msgpack"
)

type typingEffect struct {
	animCount        int
	animMaxCount     int
	allTextDisplayed bool
	content          string
}

func NewTypingEffect(content string, delay int) *typingEffect {
	t := &typingEffect{
		content:          content,
		animMaxCount:     utf8.RuneCountInString(content) / 2 * delay,
		allTextDisplayed: false,
		animCount:        0,
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
	if t.animCount < t.animMaxCount {
		t.animCount++
	}
	if t.animCount == t.animMaxCount {
		t.allTextDisplayed = true
	}
}

func (t *typingEffect) getCurrentTextLength() int {
	if t.animMaxCount > 0 {
		return len(t.content) * t.animCount / t.animMaxCount
	}
	return len(t.content)
}
