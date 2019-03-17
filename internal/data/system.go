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

type Title struct {
	MapID  int `msgpack:"mapId"`
	RoomID int `msgpack:"roomId"`
}

type System struct {
	Title              *Title              `msgpack:"title"`
	InitialPlayerState *InitialPlayerState `msgpack:"player"`
	DefaultLanguage    Language            `msgpack:"defaultLanguage"`
	TitleBGM           BGM                 `msgpack:"titleBgm"`
	GameName           UUID                `msgpack:"gameName"`
	TitleTextColor     string              `msgpack:"titleTextColor"`
	Switches           []*VariableData     `msgpack:"switches"`
	Variables          []*VariableData     `msgpack:"variables"`
}

type InitialPlayerState struct {
	Image     string    `msgpack:"image"`
	ImageType ImageType `msgpack:"imageType"`
	MapID     int       `msgpack:"mapId"`
	RoomID    int       `msgpack:"roomId"`
	X         int       `msgpack:"x"`
	Y         int       `msgpack:"y"`
}

type VariableData struct {
	ID       int             `msgpack:"id"`
	Name     string          `msgpack:"name"`
	Items    []*VariableItem `msgpack:"items"`
	IsFolded bool            `msgpack:"isFolded"`
}

type VariableItem struct {
	ID   int    `msgpack:"id"`
	Name string `msgpack:"name"`
}
