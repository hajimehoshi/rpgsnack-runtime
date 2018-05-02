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
	InitialPosition *Position `msgpack:"player"`
	DefaultLanguage Language  `msgpack:"defaultLanguage"`
	TitleBGM        BGM       `msgpack:"titleBgm"`
	GameName        UUID      `msgpack:"gameName"`
}

func (s *System) UnmarshalJSON(data []uint8) error {
	type tmpSystem struct {
		InitialPosition *Position `json:"player"`
		DefualtLanguage string    `json:"defaultLanguage"`
		TitleBGM        BGM       `json:"titleBgm"`
	}
	var tmp *tmpSystem
	if err := unmarshalJSON(data, &tmp); err != nil {
		return err
	}
	s.InitialPosition = tmp.InitialPosition
	s.TitleBGM = tmp.TitleBGM
	l, err := language.Parse(tmp.DefualtLanguage)
	if err != nil {
		return err
	}
	s.DefaultLanguage = Language(l)
	return nil
}

type Position struct {
	MapID  int `json:"mapId" msgpack:"mapId"`
	RoomID int `json:"roomId" msgpack:"roomId"`
	X      int `json:"x" msgpack:"x"`
	Y      int `json:"y" msgpack:"y"`
}
