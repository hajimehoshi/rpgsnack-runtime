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
	Maps          []*Map          `msgpack:"maps"`
	Texts         *Texts          `msgpack:"texts"`
	Tables        []*Table        `msgpack:"tables"`
	TileSets      []*TileSet      `msgpack:"tileSets"`
	Achievements  []*Achievement  `msgpack:"achievements"`
	Hints         []*Hint         `msgpack:"hints"`
	IAPProducts   []*IAPProduct   `msgpack:"iapProducts"`
	Items         []*Item         `msgpack:"items"`
	Combines      []*Combine      `msgpack:"combines"`
	CommonEvents  []*CommonEvent  `msgpack:"commonEvents"`
	System        *System         `msgpack:"system"`
	MessageStyles []*MessageStyle `msgpack:"messageStyles"`
	Shops         []*Shop         `msgpack:"shops"`
}

type Table struct {
	Name    string                    `msgpack:"name"`
	Schema  map[string]TableValueType `msgpack:"schema"`
	Records []*map[string]interface{} `msgpack:"records"`
}

type MessageStyle struct {
	ID                int            `msgpack:"id"`
	Name              UUID           `msgpack:"name"`
	TypingEffectDelay int            `msgpack:"typingEffectDelay"`
	SoundEffect       string         `msgpack:"soundEffect"`
	CharacterAnim     *CharacterAnim `msgpack:"characterAnim"`
}

type AssetMetadata struct {
	PassageTypes []PassageType `msgpack:"passageTypes"`
	IsAutoTile   bool          `msgpack:"isAutoTile"`
}

type FinishTriggerType string

const (
	FinishTriggerTypeNone    FinishTriggerType = "none"
	FinishTriggerTypeMessage FinishTriggerType = "message"
	FinishTriggerTypeWindow  FinishTriggerType = "window"
)

type CharacterAnim struct {
	Image         string            `msgpack:"image"`
	ImageType     ImageType         `msgpack:"imageType"`
	Speed         Speed             `msgpack:"speed"`
	FinishTrigger FinishTriggerType `msgpack:"finishTrigger"`
}

type BGM struct {
	Name   string `msgpack:"name"`
	Volume int    `msgpack:"volume"`
}

type Achievement struct {
	ID    int    `msgpack:"id"`
	Name  UUID   `msgpack:"name"`
	Desc  UUID   `msgpack:"desc"`
	Image string `msgpack:"image"`
}

type Hint struct {
	ID       int        `msgpack:"id"`
	Commands []*Command `msgpack:"commands"`
}

type IAPProduct struct {
	ID      int    `msgpack:"id"`
	Bundles []int  `msgpack:"bundles"`
	Key     string `msgpack:"key"`
	Name    UUID   `msgpack:"name"`
	Desc    UUID   `msgpack:"desc"`
	Details UUID   `msgpack:"details"`
	Type    string `msgpack:"type"`
}

type Item struct {
	ID       int        `msgpack:"id"`
	Group    int        `msgpack:"group"`
	Name     UUID       `msgpack:"name"`
	Icon     string     `msgpack:"icon"`
	Commands []*Command `msgpack:"commands"`
}

type Combine struct {
	ID       int         `msgpack:"id"`
	Item1    int         `msgpack:"item1"`
	Item2    int         `msgpack:"item2"`
	Type     CombineType `msgpack:"type"`
	Commands []*Command  `msgpack:"commands"`
}

type Shop struct {
	Name     ShopType `msgpack:"name"`
	Products []int    `msgpack:"products"`
}

type ShopProduct struct {
	ID       int    `msgpack:"id"`
	Key      string `msgpack:"key"`
	Name     string `msgpack:"name"`
	Desc     string `msgpack:"desc"`
	Details  string `msgpack:"details"`
	Type     string `msgpack:"type"`
	Unlocked bool   `msgpack:"unlocked"`
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
	id := recordID
	for _, t := range g.Tables {
		if t.Name != tableName {
			continue
		}
		for _, r := range t.Records {
			i, ok := InterfaceToInt((*r)["id"])
			if !ok {
				panic(fmt.Sprintf("GetTableValue: failed to convert ID %v", (*r)["id"]))
			}
			if i == id {
				return (*r)[attrName]
			}
		}
	}

	panic(fmt.Sprintf("GetTableValue: could not find a value %s:%d:%s", tableName, recordID, attrName))
}
