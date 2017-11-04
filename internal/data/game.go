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

type Game struct {
	Maps         []*Map         `json:"maps"`
	Texts        *Texts         `json:"texts"`
	TileSets     []*TileSet     `json:"tileSets"`
	Achievements []*Achievement `json:"achievements"`
	Hints        []*Hint        `json:"hints"`
	IAPProducts  []*IAPProduct  `json:"iapProducts"`
	Items        []*Item        `json:"items"`
	CommonEvents []*CommonEvent `json:"commonEvents"`
	System       *System        `json:"system"`
}

type BGM struct {
	Name   string `json:"name"`
	Volume int    `json:"volume"`
}

type Achievement struct {
	ID    int    `json:"id"`
	Name  UUID   `json:"name"`
	Desc  UUID   `json:"desc"`
	Image string `json:"image"`
}

type Hint struct {
	ID       int        `json:"id"`
	Commands []*Command `json:"commands"`
}

type IAPProduct struct {
	ID  int    `json:"id"`
	Key string `json:"key"`
}

type Item struct {
	ID       int        `json:"id"`
	Name     UUID       `json:"name"`
	Icon     string     `json:"icon"`
	Preview  string     `json:"preview"`
	Commands []*Command `json:"commands"`
}
