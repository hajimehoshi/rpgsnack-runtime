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

var (
	current         *Game
	progress        []uint8
	purchases       []uint8
	prices          map[string]string // TODO: We want to use https://godoc.org/golang.org/x/text/currency
	defaultLanguage language.Tag
)

func Current() *Game {
	return current
}

func Progress() []uint8 {
	return progress
}

func Purchases() []uint8 {
	return purchases
}

func Price(productId string) string {
	if _, ok := prices[productId]; ok {
		return prices[productId]
	}
	return ""
}

// DefaultLanguage represents a default language in the environment the player is playing on.
func DefaultLanguage() language.Tag {
	return defaultLanguage
}

func UpdateProgress(p []uint8) {
	progress = p
}

func UpdatePurchases(p []uint8) {
	purchases = p
}

func UpdatePrices(p map[string]string) {
	prices = p
}

type jsonData struct {
	Game            []uint8
	Progress        []uint8
	Purchases       []uint8
	DefaultLanguage string
}

func Load() error {
	data, err := loadJSONData()
	if err != nil {
		return err
	}
	var gameData *Game
	if err := unmarshalJSON(data.Game, &gameData); err != nil {
		return err
	}
	current = gameData
	progress = data.Progress
	purchases = data.Purchases

	tag, err := language.Parse(data.DefaultLanguage)
	if err != nil {
		panic(err)
	}
	defaultLanguage = tag
	return nil
}
