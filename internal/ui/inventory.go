// Copyright 2017 Hajime Hoshi
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

package ui

import (
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
)

type Inventory struct {
	X                int
	Y                int
	Visible          bool
	PressedSlotIndex int
	items            []*data.Item
	activeItemId     int
}

const itemYMargin = 2

func NewInventory(x, y int) *Inventory {
	return &Inventory{
		X:                x,
		Y:                y,
		Visible:          true,
		PressedSlotIndex: -1,
		items:            []*data.Item{},
		activeItemId:     0,
	}
}

func (i *Inventory) pressedSlotIndex() int {
	if !input.Triggered() {
		return -1
	}

	x, y := input.Position()
	x /= consts.TileScale
	y = (y - consts.GameMarginTop - itemYMargin) / consts.TileScale

	if i.Y <= y && y < i.Y+20 {
		return (x - 4) / 20
	}

	return -1
}

func (i *Inventory) Update() {
	i.PressedSlotIndex = i.pressedSlotIndex()
}

func (i *Inventory) Draw(screen *ebiten.Image) {
	if !i.Visible {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(i.X), float64(i.Y))
	screen.DrawImage(assets.GetImage("system/frame_inventory.png"), op)

	for index, item := range i.items {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(4+i.X+index*20), float64(i.Y+itemYMargin))
		if i.activeItemId == item.ID {
			op.ColorM.Translate(0.5, 0.5, 0.5, 0)
		}
		screen.DrawImage(assets.GetImage("items/icon/"+item.Icon+".png"), op)
	}
}

func (i *Inventory) SetItems(items []*data.Item) {
	i.items = items
}

func (i *Inventory) SetActiveItemID(activeItemId int) {
	i.activeItemId = activeItemId
}
