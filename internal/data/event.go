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
	"fmt"

	"github.com/vmihailenco/msgpack"
)

const (
	SelfSwitchNum = 10
)

type Event struct {
	impl    *EventImpl
	msgpack []byte
}

func (e *Event) ID() int {
	if err := e.ensureEncoded(); err != nil {
		panic(err)
	}
	return e.impl.ID
}

func (e *Event) Position() (int, int) {
	if err := e.ensureEncoded(); err != nil {
		panic(err)
	}
	return e.impl.X, e.impl.Y
}

func (e *Event) Pages() []*Page {
	if err := e.ensureEncoded(); err != nil {
		panic(err)
	}
	return e.impl.Pages
}

func (e *Event) UnmarshalMsgpack(data []byte) error {
	e.msgpack = data

	if !isEventLazilyParsed() {
		e.ensureEncoded()
	}
	return nil
}

func (e *Event) ensureEncoded() error {
	if e.impl != nil {
		return nil
	}

	var impl *EventImpl
	if e.msgpack != nil {
		if err := msgpack.Unmarshal(e.msgpack, &impl); err != nil {
			return err
		}
		e.impl = impl
		return nil
	}

	panic("data: the data format was not either Msgpack  at (*Event).ensureEncoded")
}

type EventImpl struct {
	ID    int     `msgpack:"id"`
	X     int     `msgpack:"x"`
	Y     int     `msgpack:"y"`
	Pages []*Page `msgpack:"pages"`
}

type CommonEvent struct {
	ID       int        `msgpack:"id"`
	Name     string     `msgpack:"name"`
	Commands []*Command `msgpack:"commands"`
}

type Page struct {
	Conditions []*Condition         `msgpack:"conditions"`
	Image      string               `msgpack:"image"`
	ImageType  ImageType            `msgpack:"imageType"`
	Frame      int                  `msgpack:"frame"`
	Dir        Dir                  `msgpack:"dir"`
	DirFix     bool                 `msgpack:"dirFix"`
	Walking    bool                 `msgpack:"walking"`
	Stepping   bool                 `msgpack:"stepping"`
	Through    bool                 `msgpack:"through"`
	Priority   Priority             `msgpack:"priority"`
	Speed      Speed                `msgpack:"speed"`
	Trigger    Trigger              `msgpack:"trigger"`
	Opacity    int                  `msgpack:"opacity"`
	Route      *CommandArgsSetRoute `msgpack:"route"`
	Commands   []*Command           `msgpack:"commands"`
}

type Dir int

const (
	DirNone  Dir = -1
	DirUp    Dir = 0
	DirRight Dir = 1
	DirDown  Dir = 2
	DirLeft  Dir = 3
)

type Priority string

const (
	PriorityBottom Priority = "bottom"
	PriorityMiddle Priority = "middle"
	PriorityTop    Priority = "top"
)

type ImageType string

const (
	ImageTypeCharacters ImageType = "character"
	ImageTypeIcons      ImageType = "icon"
)

type Trigger string

const (
	TriggerPlayer   Trigger = "player"
	TriggerAuto     Trigger = "auto"
	TriggerParallel Trigger = "parallel"
	TriggerDirect   Trigger = "direct"
	TriggerNever    Trigger = "never"
)

type Speed int

const (
	Speed1 Speed = 1
	Speed2 Speed = 2
	Speed3 Speed = 3
	Speed4 Speed = 4
	Speed5 Speed = 5
	Speed6 Speed = 6
)

func (s Speed) Frames() int {
	switch s {
	case Speed1:
		return 64
	case Speed2:
		return 32
	case Speed3:
		return 16
	case Speed4:
		return 8
	case Speed5:
		return 4
	case Speed6:
		return 2
	default:
		panic(fmt.Sprintf("data: invalid speed: %d", s))
	}
}

func (s Speed) SteppingIncrementFrames() int {
	switch s {
	case Speed1:
		return 1
	case Speed2:
		return 2
	case Speed3:
		return 3
	case Speed4:
		return 4
	case Speed5:
		return 6
	case Speed6:
		return 12
	}
	return 0
}

const MaxVolume = 100
