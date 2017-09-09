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

package hints

import (
	"fmt"

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

type hintState int

const (
	hintStateInactive hintState = iota
	hintStateActiveUnread
	hintStateActiveRead
	hintStateCompleted
)

type Hints struct {
	states map[int]hintState
}

func (h *Hints) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()
	e.EncodeString("states")
	e.BeginMap()
	for k, v := range h.states {
		e.EncodeInt(k)
		e.EncodeInt(int(v))
	}
	e.EndMap()
	e.EndMap()
	return e.Flush()
}

func (h *Hints) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "states":
			if !d.SkipCodeIfNil() {
				n := d.DecodeMapLen()
				h.states = map[int]hintState{}
				for i := 0; i < n; i++ {
					k := d.DecodeInt()
					v := hintState(d.DecodeInt())
					h.states[k] = v
				}
			}
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("hints: Hints.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (h *Hints) Activate(id int) {
	if h.states == nil {
		h.states = map[int]hintState{}
	}
	if h.states[id] == hintStateInactive {
		h.states[id] = hintStateActiveUnread
	}
}

func (h *Hints) Pause(id int) {
	if h.states == nil {
		return
	}
	h.states[id] = hintStateInactive
	h.refreshActiveHints()
}

func (h *Hints) Complete(id int) {
	if h.states == nil {
		return
	}
	h.states[id] = hintStateCompleted
	h.refreshActiveHints()
}

func (h *Hints) ReadHint(id int) {
	if h.states == nil {
		return
	}
	h.states[id] = hintStateActiveRead
	h.refreshActiveHints()
}

func (h *Hints) refreshActiveHints() {
	// If all hints are marked as read, reset all to unread
	if h.activeUnreadHintCount() == 0 {
		for k := range h.states {
			if h.states[k] == hintStateActiveRead {
				h.states[k] = hintStateActiveUnread
			}
		}
	}
}

func (h *Hints) ActiveHintID() int {
	for k, v := range h.states {
		if v == hintStateActiveUnread {
			return k
		}
	}
	return -1
}

func (h *Hints) activeUnreadHintCount() int {
	count := 0
	for _, v := range h.states {
		if v == hintStateActiveUnread {
			count++
		}
	}
	return count
}

func (h *Hints) ActiveHintCount() int {
	count := 0
	for _, v := range h.states {
		if v == hintStateActiveUnread || v == hintStateActiveRead {
			count++
		}
	}
	return count
}
