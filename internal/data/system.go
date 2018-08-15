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
	"golang.org/x/text/language"
)

type System struct {
	InitialPlayerState *InitialPlayerState `msgpack:"player"`
	DefaultLanguage    Language            `msgpack:"defaultLanguage"`
	TitleBGM           BGM                 `msgpack:"titleBgm"`
	GameName           UUID                `msgpack:"gameName"`
	TitleTextColor     string              `msgpack:"titleTextColor"`
}

func (s *System) UnmarshalJSON(data []uint8) error {
	type tmpSystem struct {
		InitialPlayerState *InitialPlayerState `json:"player"`
		DefualtLanguage    string              `json:"defaultLanguage"`
		TitleBGM           BGM                 `json:"titleBgm"`
		TitleTextColor     string              `json:"titleTextColor"`
	}
	var tmp *tmpSystem
	if err := unmarshalJSON(data, &tmp); err != nil {
		return err
	}
	s.InitialPlayerState = tmp.InitialPlayerState
	s.TitleBGM = tmp.TitleBGM
	s.TitleTextColor = tmp.TitleTextColor
	l, err := language.Parse(tmp.DefualtLanguage)
	if err != nil {
		return err
	}
	s.DefaultLanguage = Language(l)
	return nil
}

type InitialPlayerState struct {
	Image     string    `json:"image" msgpack:"image"`
	ImageType ImageType `json:"imageType" msgpack:"imageType"`
	MapID     int       `json:"mapId" msgpack:"mapId"`
	RoomID    int       `json:"roomId" msgpack:"roomId"`
	X         int       `json:"x" msgpack:"x"`
	Y         int       `json:"y" msgpack:"y"`
}
