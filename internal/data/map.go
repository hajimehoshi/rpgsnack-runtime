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
	msgpack []byte
}

func (m *Map) ID() int {
	if err := m.ensureDecoded(); err != nil {
		panic(err)
	}
	return m.impl.ID
}

func (m *Map) Name() string {
	if err := m.ensureDecoded(); err != nil {
		panic(err)
	}
	return m.impl.Name
}

func (m *Map) Rooms() []*Room {
	if err := m.ensureDecoded(); err != nil {
		panic(err)
	}
	return m.impl.Rooms
}

func (m *Map) UnmarshalMsgpack(data []byte) error {
	m.msgpack = data
	return nil
}

func (m *Map) ensureDecoded() error {
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

	panic("data: the data format was not either Msgpack at (*Map).ensureDecoded")
}

type MapImpl struct {
	ID    int     `msgpack:"id"`
	Name  string  `msgpack:"name"`
	Rooms []*Room `msgpack:"rooms"`
}

type Room struct {
	ID                   int            `msgpack:"id"`
	X                    int            `msgpack:"x"`
	Y                    int            `msgpack:"y"`
	Tiles                [][]int        `msgpack:"tiles"`
	Events               []*Event       `msgpack:"events"`
	Background           MapSprite      `msgpack:"background"`
	Foreground           MapSprite      `msgpack:"foreground"`
	PassageTypeOverrides []PassageType  `msgpack:"passageTypeOverrides"`
	AutoBGM              bool           `msgpack:"autoBGM"`
	BGM                  BGM            `msgpack:"bgm"`
	LayoutMode           RoomLayoutMode `msgpack:"layoutMode"`
}

type MapSprite struct {
	Name    string `msgpack:"name"`
	ScrollX int    `msgpack:"scrollX"`
	ScrollY int    `msgpack:"scrollY"`
}
