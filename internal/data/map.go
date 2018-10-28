// Copyright 2016 Hajime Hoshi
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

package data

import (
	"encoding/json"

	"github.com/vmihailenco/msgpack"
)

type RoomLayoutMode string

const (
	RoomLayoutModeFixBottom RoomLayoutMode = "fix_bottom"
	RoomLayoutModeFixCenter RoomLayoutMode = "fix_center"
	RoomLayoutModeScroll    RoomLayoutMode = "scroll"
)

type Map struct {
	impl    *MapImpl
	json    []byte
	msgpack []byte
}

func (m *Map) ID() int {
	if err := m.ensureEncoded(); err != nil {
		panic(err)
	}
	return m.impl.ID
}

func (m *Map) Name() string {
	if err := m.ensureEncoded(); err != nil {
		panic(err)
	}
	return m.impl.Name
}

func (m *Map) Rooms() []*Room {
	if err := m.ensureEncoded(); err != nil {
		panic(err)
	}
	return m.impl.Rooms
}

func (m *Map) ensureEncoded() error {
	if m.impl != nil {
		return nil
	}

	var impl *MapImpl
	if m.msgpack != nil {
		if err := msgpack.Unmarshal(m.msgpack, &impl); err != nil {
			return err
		}
		m.impl = impl
		return nil
	}

	if m.json != nil {
		if err := json.Unmarshal(m.json, &impl); err != nil {
			return err
		}
		m.impl = impl
		return nil
	}

	panic("not reached")
}

type MapImpl struct {
	ID    int     `json:"id" msgpack:"id"`
	Name  string  `json:"name" msgpack:"name"`
	Rooms []*Room `json:"rooms" msgpack:"rooms"`
}

func (m *Map) UnmarshalJSON(data []byte) error {
	m.json = data
	return nil
}

func (m *Map) UnmarshalMsgpack(data []byte) error {
	m.msgpack = data
	return nil
}

type Room struct {
	ID                   int            `json:"id" msgpack:"id"`
	X                    int            `json:"x" msgpack:"x"`
	Y                    int            `json:"y" msgpack:"y"`
	Tiles                [][]int        `json:"tiles" msgpack:"tiles"`
	Events               []*Event       `json:"events" msgpack:"events"`
	Background           MapSprite      `json:"background" msgpack:"background"`
	Foreground           MapSprite      `json:"foreground" msgpack:"foreground"`
	PassageTypeOverrides []PassageType  `json:"passageTypeOverrides" msgpack:"passageTypeOverrides"`
	AutoBGM              bool           `json:"autoBGM" msgpack:"autoBGM"`
	BGM                  BGM            `json:"bgm" msgpack:"bgm"`
	LayoutMode           RoomLayoutMode `json:"layoutMode" msgpack:"layoutMode"`
}

type MapSprite struct {
	Name    string `json:"name" msgpack:"name"`
	ScrollX int    `json:"scrollX" msgpack:"scrollX"`
	ScrollY int    `json:"scrollY" msgpack:"scrollY"`
}
