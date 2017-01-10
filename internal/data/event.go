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
	ID    int     `json:"id"`
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Pages []*Page `json:"pages"`
}

type Page struct {
	Conditions []*Condition         `json:"conditions"`
	Image      string               `json:"image"`
	ImageIndex int                  `json:"imageIndex"`
	Attitude   Attitude             `json:"attitude"`
	Dir        Dir                  `json:"dir"`
	DirFix     bool                 `json:"dirFix"`
	Walking    bool                 `json:"walking"`
	Stepping   bool                 `json:"stepping"`
	Through    bool                 `json:"through"`
	Priority   Priority             `json:"priority"`
	Trigger    Trigger              `json:"trigger"`
	Route      *CommandArgsSetRoute `json:"route"`
	Commands   []*Command           `json:"commands"`
}

type Dir int

const (
	DirUp Dir = iota
	DirRight
	DirDown
	DirLeft
)

type Attitude int

const (
	AttitudeLeft Attitude = iota
	AttitudeMiddle
	AttitudeRight
)

type Priority int

const (
	PriorityBelowCharacters Priority = iota
	PrioritySameAsCharacters
	PriorityAboveCharacters
)

type Trigger string

const (
	TriggerPlayer Trigger = "player"
	TriggerAuto   Trigger = "auto"
	TriggerDirect Trigger = "direct"
	TriggerNever  Trigger = "never"
)

type Speed int

const (
	Speed1 Speed = 1
	Speed2 Speed = 2
	Speed3 Speed = 3
	Speed4 Speed = 4
	Speed5 Speed = 5
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
	default:
		panic("not reach")
	}
}
