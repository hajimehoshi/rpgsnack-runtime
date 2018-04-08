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

type Map struct {
	ID    int     `json:"id" msgpack:"id"`
	Name  string  `json:"name" msgpack:"name"`
	Rooms []*Room `json:"rooms" msgpack:"rooms"`
}

type Room struct {
	ID                   int           `json:"id" msgpack:"id"`
	X                    int           `json:"x" msgpack:"x"`
	Y                    int           `json:"y" msgpack:"y"`
	Tiles                [][]int       `json:"tiles" msgpack:"tiles"`
	Events               []*Event      `json:"events" msgpack:"events"`
	Background           MapSprite     `json:"background" msgpack:"background"`
	Foreground           MapSprite     `json:"foreground" msgpack:"foreground"`
	PassageTypeOverrides []PassageType `json:"passageTypeOverrides" msgpack:"passageTypeOverrides"`
	AutoBGM              bool          `json:"autoBGM" msgpack:"autoBGM"`
	BGM                  BGM           `json:"bgm" msgpack:"bgm"`
}

type MapSprite struct {
	Name    string `json:"name" msgpack:"name"`
	ScrollX int    `json:"scrollX" msgpack:"scrollX"`
	ScrollY int    `json:"scrollY" msgpack:"scrollY"`
}
