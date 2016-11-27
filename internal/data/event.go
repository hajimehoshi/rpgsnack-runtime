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
	Dir        Dir        `json:"dir"`
	DirFix     bool       `json:"dirFix"`
	Walking    bool       `json:"walking"`
	Stepping   bool       `json:"stepping"`
	Through    bool       `json:"through"`
	Priority   Priority   `json:"priority"`
	Trigger    Trigger    `json:"trigger"`
	Commands   []*Command `json:"commands"`
}

type Command struct {
	Name     CommandName       `json:"name"`
	Args     map[string]string `json:"args"`
	Branches [][]*Command      `json:"branches"`
}

type CommandName string

const (
	CommandNameShowMessage CommandName = "show_message"
	CommandNameShowChoices             = "show_choices"
	CommandNameSetSwitch               = "set_switch"
)

type ShowMessagePosition string

const (
	ShowMessagePositionSelf   ShowMessagePosition = "self"
	ShowMessagePositionPlayer ShowMessagePosition = "player"
	ShowMessagePositionEvent  ShowMessagePosition = "event"
	ShowMessagePositionCenter                     = "center"
)

type SwitchValue string

const (
	SwitchValueFalse = "false"
	SwitchValueTrue  = "true"
)
