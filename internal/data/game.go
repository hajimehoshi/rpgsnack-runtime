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
	"fmt"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

type TableValueType string

const (
	TableValueTypeInt    TableValueType = "int"
	TableValueTypeString TableValueType = "string"
	TableValueTypeUUID   TableValueType = "uuid"
)

type ShopType string

const (
	ShopTypeMain ShopType = "main"
	ShopTypeHome ShopType = "home"
)

type CombineType string

const (
	CombineTypeUse     CombineType = "use"
	CombineTypeCombine CombineType = "combine"
)

type Game struct {
	Maps          []*Map          `json:"maps" msgpack:"maps"`
	Texts         *Texts          `json:"texts" msgpack:"texts"`
	Tables        []*Table        `json:"tables" msgpack:"tables"`
	TileSets      []*TileSet      `json:"tileSets" msgpack:"tileSets"`
	Achievements  []*Achievement  `json:"achievements" msgpack:"achievements"`
	Hints         []*Hint         `json:"hints" msgpack:"hints"`
	IAPProducts   []*IAPProduct   `json:"iapProducts" msgpack:"iapProducts"`
	Items         []*Item         `json:"items" msgpack:"items"`
	Combines      []*Combine      `json:"combines" msgpack:"combines"`
	CommonEvents  []*CommonEvent  `json:"commonEvents" msgpack:"commonEvents"`
	System        *System         `json:"system" msgpack:"system"`
	MessageStyles []*MessageStyle `json:"messageStyles" msgpack:"messageStyles"`
	Shops         []*Shop         `json:"shops" msgpack:"shops"`
}

type Table struct {
	Name    string                    `json:"name" msgpack:"name"`
	Schema  map[string]TableValueType `json:"schema" msgpack:"schema"`
	Records []*map[string]interface{} `json:"records" msgpack:"records"`
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
	ID      int    `json:"id" msgpack:"id"`
	Bundles []int  `json:"bundles" msgpack:"bundles"`
	Key     string `json:"key" msgpack:"key"`
	Name    UUID   `json:"name" msgpack:"name"`
	Desc    UUID   `json:"desc" msgpack:"desc"`
	Details UUID   `json:"details" msgpack:"details"`
	Type    string `json:"type" msgpack:"type"`
}

type Item struct {
	ID       int        `json:"id" msgpack:"id"`
	Group    int        `json:"group" msgpack:"group"`
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

type Shop struct {
	Name     ShopType `json:"name" msgpack:"name"`
	Products []int    `json:"products" msgpack:"products"`
}

type ShopProduct struct {
	ID       int    `json:"id" msgpack:"id"`
	Key      string `json:"key" msgpack:"key"`
	Name     string `json:"name" msgpack:"name"`
	Desc     string `json:"desc" msgpack:"desc"`
	Details  string `json:"details" msgpack:"details"`
	Type     string `json:"type" msgpack:"type"`
	Unlocked bool   `json:"unlocked" msgpack:"unlocked"`
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

func (g *Game) GetIAPProductByType(t string) *IAPProduct {
	for _, iapProduct := range g.IAPProducts {
		if iapProduct.Type == t {
			return iapProduct
		}
	}

	return nil
}

func (g *Game) IAPProductByID(id int) *IAPProduct {
	for _, iapProduct := range g.IAPProducts {
		if iapProduct.ID == id {
			return iapProduct
		}
	}

	return nil
}

func (g *Game) GetShopProducts(products []int) []*ShopProduct {
	shopProducts := []*ShopProduct{}
	for _, productID := range products {
		iapProduct := g.IAPProductByID(productID)
		if iapProduct != nil {
			shopProducts = append(shopProducts, &ShopProduct{
				ID:      iapProduct.ID,
				Key:     iapProduct.Key,
				Name:    g.Texts.Get(lang.Get(), iapProduct.Name),
				Desc:    g.Texts.Get(lang.Get(), iapProduct.Desc),
				Details: g.Texts.Get(lang.Get(), iapProduct.Details),
				Type:    iapProduct.Type,
			})
		}
	}
	return shopProducts
}

func (g *Game) GetShop(name ShopType) *Shop {
	for _, shop := range g.Shops {
		if shop.Name == name {
			return shop
		}
	}
	return nil
}

func (g *Game) IsShopAvailable(name ShopType) bool {
	return g.GetShop(name) != nil
}

func (g *Game) IsCombineAvailable() bool {
	return len(g.Combines) > 0
}

func (g *Game) GetTableValueType(tableName string, attrName string) TableValueType {
	for _, t := range g.Tables {
		if t.Name != tableName {
			continue
		}
		return t.Schema[attrName]
	}
	panic(fmt.Sprintf("GetTableValueType: could not find a schema type %s:%s", tableName, attrName))
}

func (g *Game) GetTableValue(tableName string, recordID int, attrName string) interface{} {
	id := float64(recordID)
	for _, t := range g.Tables {
		if t.Name != tableName {
			continue
		}
		for _, r := range t.Records {
			if (*r)["id"].(float64) == id {
				return (*r)[attrName]
			}
		}
	}

	panic(fmt.Sprintf("GetTableValue: could not find a value %s:%d:%s", tableName, recordID, attrName))
}
