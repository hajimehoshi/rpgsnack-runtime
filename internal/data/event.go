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

const (
	SelfSwitchNum = 4
)

type Event struct {
	ID    int     `json:"id" msgpack:"id"`
	X     int     `json:"x" msgpack:"x"`
	Y     int     `json:"y" msgpack:"y"`
	Pages []*Page `json:"pages" msgpack:"pages"`
}

type CommonEvent struct {
	ID       int        `json:"id" msgpack:"id"`
	Name     string     `json:"name" msgpack:"name"`
	Commands []*Command `json:"commands" msgpack:"commands"`
}

type Page struct {
	Conditions []*Condition         `json:"conditions" msgpack:"conditions"`
	Image      string               `json:"image" msgpack:"image"`
	ImageType  ImageType            `json:"imageType" msgpack:"imageType"`
	Frame      int                  `json:"frame" msgpack:"frame"`
	Dir        Dir                  `json:"dir" msgpack:"dir"`
	DirFix     bool                 `json:"dirFix" msgpack:"dirFix"`
	Walking    bool                 `json:"walking" msgpack:"walking"`
	Stepping   bool                 `json:"stepping" msgpack:"stepping"`
	Through    bool                 `json:"through" msgpack:"through"`
	Priority   Priority             `json:"priority" msgpack:"priority"`
	Speed      Speed                `json:"speed" msgpack:"speed"`
	Trigger    Trigger              `json:"trigger" msgpack:"trigger"`
	Opacity    int                  `json:"opacity" msgpack:"opacity"`
	Route      *CommandArgsSetRoute `json:"route" msgpack:"route"`
	Commands   []*Command           `json:"commands" msgpack:"commands"`
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
		panic("not reach")
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
		return 5
	case Speed6:
		return 6
	}
	return 0
}

const MaxVolume = 100
