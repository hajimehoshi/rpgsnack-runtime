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

type CombineType string

const (
	CombineTypeUse     CombineType = "use"
	CombineTypeCombine CombineType = "combine"
)

type Game struct {
	Maps          []*Map          `json:"maps" msgpack:"maps"`
	Texts         *Texts          `json:"texts" msgpack:"texts"`
	TileSets      []*TileSet      `json:"tileSets" msgpack:"tileSets"`
	Achievements  []*Achievement  `json:"achievements" msgpack:"achievements"`
	Hints         []*Hint         `json:"hints" msgpack:"hints"`
	IAPProducts   []*IAPProduct   `json:"iapProducts" msgpack:"iapProducts"`
	Items         []*Item         `json:"items" msgpack:"items"`
	Combines      []*Combine      `json:"combines" msgpack:"combines"`
	CommonEvents  []*CommonEvent  `json:"commonEvents" msgpack:"commonEvents"`
	System        *System         `json:"system" msgpack:"system"`
	MessageStyles []*MessageStyle `json:"messageStyles" msgpack:"messageStyles"`
}

type MessageStyle struct {
	ID                int            `json:"id" msgpack:"id"`
	Name              UUID           `json:"name" msgpack:"name"`
	TypingEffectDelay int            `json:"typingEffectDelay" msgpack:"typingEffectDelay"`
	SoundEffect       string         `json:"soundEffect" msgpack:"soundEffect"`
	CharacterAnim     *CharacterAnim `json:"characterAnim" msgpack:"characterAnim"`
}

type AssetMetadata struct {
	PassageTypes []PassageType `json:"passageTypes" msgpack:"passageTypes"`
	IsAutoTile   bool          `json:"isAutoTile" msgpack:"isAutoTile"`
}

type FinishTriggerType string

const (
	FinishTriggerTypeNone    FinishTriggerType = "none"
	FinishTriggerTypeMessage FinishTriggerType = "message"
	FinishTriggerTypeWindow  FinishTriggerType = "window"
)

type CharacterAnim struct {
	Image         string            `json:"image" msgpack:"image"`
	ImageType     ImageType         `json:"imageType" msgpack:"imageType"`
	Speed         Speed             `json:"speed" msgpack:"speed"`
	FinishTrigger FinishTriggerType `json:"finishTrigger" msgpack:"finishTrigger"`
}

type BGM struct {
	Name   string `json:"name" msgpack:"name"`
	Volume int    `json:"volume" msgpack:"volume"`
}

type Achievement struct {
	ID    int    `json:"id" msgpack:"id"`
	Name  UUID   `json:"name" msgpack:"name"`
	Desc  UUID   `json:"desc" msgpack:"desc"`
	Image string `json:"image" msgpack:"image"`
}

type Hint struct {
	ID       int        `json:"id" msgpack:"id"`
	Commands []*Command `json:"commands" msgpack:"commands"`
}

type IAPProduct struct {
	ID     int    `json:"id" msgpack:"id"`
	Key    string `json:"key" msgpack:"key"`
	Name   UUID   `json:"name" msgpack:"name"`
	Desc   UUID   `json:"desc" msgpack:"desc"`
	IsShop bool   `json:"is_shop" msgpack:"is_shop"`
	Tier   int    `json:"tier" msgpack:"tier"`
}

type Item struct {
	ID       int        `json:"id" msgpack:"id"`
	Name     UUID       `json:"name" msgpack:"name"`
	Icon     string     `json:"icon" msgpack:"icon"`
	Commands []*Command `json:"commands" msgpack:"commands"`
}

type Combine struct {
	ID       int         `json:"id" msgpack:"id"`
	Item1    int         `json:"item1" msgpack:"item1"`
	Item2    int         `json:"item2" msgpack:"item2"`
	Type     CombineType `json:"type" msgpack:"type"`
	Commands []*Command  `json:"commands" msgpack:"commands"`
}

func (g *Game) CreateCombine(itemID1, itemID2 int) *Combine {
	for _, combine := range g.Combines {
		if (combine.Item1 == itemID1 && combine.Item2 == itemID2) || (combine.Type == CombineTypeCombine && combine.Item1 == itemID2 && combine.Item2 == itemID1) {
			return combine
		}
	}
	return nil
}

func (g *Game) CreateDefaultMessageStyle() *MessageStyle {
	return &MessageStyle{TypingEffectDelay: 1}
}

func (g *Game) CreateChoicesMessageStyle() *MessageStyle {
	return &MessageStyle{TypingEffectDelay: 0}
}

func (g *Game) GetIAPProduct(key string) *IAPProduct {
	var iap *IAPProduct
	for _, p := range g.IAPProducts {
		if p.Key == key {
			iap = p
		}
	}
	return iap
}
