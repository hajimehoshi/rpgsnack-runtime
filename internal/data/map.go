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

type Dir int

const (
	DirLeft Dir = iota
	DirRight
	DirUp
	DirDown
)

type Priority int

const (
	PriorityBelowCharacters Priority = iota
	PrioritySameAsCharacters
	PriorityAboveCharacters
)

type Trigger int

const (
	TriggerActionButton Trigger = iota
	TriggerPlayerTouch
	TriggerEventTouch
	TriggerAuto
)

type Map struct {
	Name      string  `json:"name"`
	TileSetID int     `json:"tileSetId"`
	Rooms     []*Room `json:"rooms"`
}

type Room struct {
	Tiles  [][]int  `json:"tiles"`
	Events []*Event `json:"events"`
}

type Event struct {
	ID    int     `json:"id"`
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Pages []*Page `json:"pages"`
}

type Page struct {
	Condition  []string   `json:"condition"`
	Image      string     `json:"image"`
	ImageIndex int        `json:"imageIndex"`
	Dir        Dir        `json:"dir"`
	DirFix     bool       `json:"dirFix"`
	Walking    bool       `json:"walking"`
	Stepping   bool       `json:"stepping"`
	Through    bool       `json:"through"`
	Priority   Priority   `json:"priority"`
	Trigger    Trigger    `json:"trigger"`
	Commands   []*Command `json:"command"`
}

type Command struct {
	Command  string            `json:"command"`
	Args     map[string]string `json:"args"`
	Branches [][]*Command      `json:"branches"`
}
