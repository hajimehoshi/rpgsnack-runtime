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
	"encoding/json"
	"fmt"

	"golang.org/x/text/language"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func unmarshalJSON(data []uint8, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		switch err := err.(type) {
		case *json.UnmarshalTypeError:
			begin := max(int(err.Offset)-20, 0)
			end := min(int(err.Offset)+40, len(data))
			part := string(data[begin:end])
			return fmt.Errorf("data JSON type error: %s:\n%s", err.Error(), part)
		case *json.SyntaxError:
			begin := max(int(err.Offset)-20, 0)
			end := min(int(err.Offset)+40, len(data))
			part := string(data[begin:end])
			return fmt.Errorf("data: JSON syntax error: %s:\n%s", err.Error(), part)
		}
		return err
	}
	return nil
}

type jsonData struct {
	Game      []uint8
	Progress  []uint8
	Purchases []uint8
	Language  string
}

type LoadedData struct {
	Game      *Game
	Progress  []uint8
	Purchases []string
	Language  language.Tag
}

func Load() (*LoadedData, error) {
	data, err := loadJSONData()
	if err != nil {
		return nil, err
	}
	var gameData *Game
	if err := unmarshalJSON(data.Game, &gameData); err != nil {
		return nil, err
	}
	var purchases []string
	if data.Purchases != nil {
		if err := unmarshalJSON(data.Purchases, &purchases); err != nil {
			return nil, err
		}
	} else {
		purchases = []string{}
	}
	tag, err := language.Parse(data.Language)
	if err != nil {
		return nil, err
	}
	return &LoadedData{
		Game:      gameData,
		Purchases: purchases,
		Progress:  data.Progress,
		Language:  tag,
	}, nil
}
