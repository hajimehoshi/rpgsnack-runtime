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

	"github.com/vmihailenco/msgpack"
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

type rawData struct {
	Project   []uint8
	Assets    []uint8
	Progress  []uint8
	Purchases []uint8
	Language  []uint8
}

type Project struct {
	Data *Game `json:data`
}

type LoadedData struct {
	Game      *Game
	Assets    map[string][]uint8
	Progress  []uint8
	Purchases []string
	Language  language.Tag
}

func Load(projectPath string) (*LoadedData, error) {
	data, err := loadRawData(projectPath)
	if err != nil {
		return nil, err
	}
	var project *Project
	if err := unmarshalJSON(data.Project, &project); err != nil {
		return nil, err
	}
	gameData := project.Data
	var assets map[string][]uint8
	if err := msgpack.Unmarshal(data.Assets, &assets); err != nil {
		return nil, fmt.Errorf("data: msgpack.Unmarshal error: %s", err.Error())
	}
	var purchases []string
	if data.Purchases != nil {
		if err := unmarshalJSON(data.Purchases, &purchases); err != nil {
			return nil, err
		}
	} else {
		purchases = []string{}
	}
	var tag language.Tag
	if data.Language != nil {
		var langId string
		if err := unmarshalJSON(data.Language, &langId); err != nil {
			return nil, err
		}
		tag, err = language.Parse(langId)
		if err != nil {
			return nil, err
		}
	} else {
		tag = gameData.System.DefaultLanguage
	}

	// Keep only the language base so far.
	// TODO: If the language is Chinese, the region code might be needed to keep.
	tag = tag.Parent()

	return &LoadedData{
		Game:      gameData,
		Assets:    assets,
		Purchases: purchases,
		Progress:  data.Progress,
		Language:  tag,
	}, nil
}
