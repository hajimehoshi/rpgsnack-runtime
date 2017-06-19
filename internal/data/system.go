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
	InitialPosition *Position
	DefaultLanguage language.Tag
}

func (s *System) UnmarshalJSON(data []uint8) error {
	type tmpSystem struct {
		InitialPosition *Position `json:"player"`
		DefualtLanguage string    `json:"defaultLanguage"`
	}
	var tmp *tmpSystem
	if err := unmarshalJSON(data, &tmp); err != nil {
		return err
	}
	s.InitialPosition = tmp.InitialPosition
	l, err := language.Parse(tmp.DefualtLanguage)
	if err != nil {
		return err
	}
	s.DefaultLanguage = l
	return nil
}

type Position struct {
	MapID  int `json:"mapId"`
	RoomID int `json:"roomId"`
	X      int `json:"x"`
	Y      int `json:"y"`
}
