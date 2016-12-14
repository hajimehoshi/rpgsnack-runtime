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

type Event struct {
	ID    int     `json:"id"`
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Pages []*Page `json:"pages"`
}

type Page struct {
	Conditions []string   `json:"conditions"`
	Image      string     `json:"image"`
	ImageIndex int        `json:"imageIndex"`
	Attitude   Attitude   `json:"attitude"`
	Dir        Dir        `json:"dir"`
	DirFix     bool       `json:"dirFix"`
	Walking    bool       `json:"walking"`
	Stepping   bool       `json:"stepping"`
	Through    bool       `json:"through"`
	Priority   Priority   `json:"priority"`
	Trigger    Trigger    `json:"trigger"`
	Commands   []*Command `json:"commands"`
}

type Dir int

const (
	DirLeft Dir = iota
	DirRight
	DirUp
	DirDown
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
	TriggerTapped     Trigger = "tapped"
	TriggerEventTouch         = "event_touch"
	TriggerAuto               = "auto"
)

type Command struct {
	Name     CommandName            `json:"name"`
	Args     map[string]interface{} `json:"args"`
	Branches [][]*Command           `json:"branches"`
}

type CommandName string

const (
	CommandNameShowMessage CommandName = "show_message"
	CommandNameShowChoices             = "show_choices"
	CommandNameSetSwitch               = "set_switch"
	CommandNameMove                    = "move"
)

type ShowMessagePosition string

const (
	ShowMessagePositionSelf   ShowMessagePosition = "self"
	ShowMessagePositionPlayer                     = "player"
	ShowMessagePositionEvent                      = "event"
	ShowMessagePositionCenter                     = "center"
)
