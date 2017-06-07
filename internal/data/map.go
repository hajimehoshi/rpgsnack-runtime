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

type PassageOverrideType int

const (
	PassageOverrideTypeNone PassageOverrideType = iota
	PassageOverrideTypePassable
	PassageOverrideTypeBlock
)

type Map struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Rooms []*Room `json:"rooms"`
}

type Room struct {
	ID                   int                   `json:"id"`
	X                    int                   `json:"x"`
	Y                    int                   `json:"y"`
	Tiles                [][]int               `json:"tiles"`
	TileSets             []string              `json:"tileSets"`
	Events               []*Event              `json:"events"`
	Background           MapSprite             `json:"background"`
	Foreground           MapSprite             `json:"foreground"`
	PassageTypeOverrides []PassageOverrideType `json:"passageTypeOverrides"`
}

type MapSprite struct {
	Name    string `json:"name"`
	ScrollX int    `json:"scrollX"`
	ScrollY int    `json:"scrollY"`
}
