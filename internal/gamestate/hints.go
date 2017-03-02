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

package gamestate

import (
	"encoding/json"
)

type hintState int

const (
	HintStateInactive hintState = iota
	HintStateActiveUnread
	HintStateActiveRead
	HintStateCompleted
)

type Hints struct {
	states map[int]hintState
}

type tmpHints struct {
	States map[int]hintState `json:"states"`
}

func (h *Hints) MarshalJSON() ([]uint8, error) {
	tmp := &tmpHints{
		States: h.states,
	}
	return json.Marshal(tmp)
}

func (h *Hints) UnmarshalJSON(data []uint8) error {
	var tmp *tmpHints
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	h.states = tmp.States
	return nil
}

func (h *Hints) Activate(id int) {
	if h.states == nil {
		h.states = map[int]hintState{}
	}
	if h.states[id] == HintStateInactive {
		h.states[id] = HintStateActiveUnread
	}
}

func (h *Hints) Pause(id int) {
	if h.states == nil {
		return
	}
	h.states[id] = HintStateInactive
	h.RefreshActiveHints()
}

func (h *Hints) Complete(id int) {
	if h.states == nil {
		return
	}
	h.states[id] = HintStateCompleted
	h.RefreshActiveHints()
}

func (h *Hints) ReadHint(id int) {
	if h.states == nil {
		return
	}
	h.states[id] = HintStateActiveRead
	h.RefreshActiveHints()
}

func (h *Hints) RefreshActiveHints() {
	// If all hints are marked as read, reset all to unread
	if h.ActiveUnreadHintCount() == 0 {
		for k := range h.states {
			if h.states[k] == HintStateActiveRead {
				h.states[k] = HintStateActiveUnread
			}
		}
	}
}

func (h *Hints) ActiveHintId() int {
	for k, v := range h.states {
		if v == HintStateActiveUnread {
			return k
		}
	}
	return -1
}

func (h *Hints) ActiveUnreadHintCount() int {
	count := 0
	for _, v := range h.states {
		if v == HintStateActiveUnread {
			count++
		}
	}
	return count
}

func (h *Hints) ActiveHintCount() int {
	count := 0
	for _, v := range h.states {
		if v == HintStateActiveUnread || v == HintStateActiveRead {
			count++
		}
	}
	return count
}
