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
	X                   int
	Y                   int
	Visible             bool
	PressedSlotIndex    int
	items               []*data.Item
	activeItemID        int
	activeItemBoxButton *Button
}

const (
	frameXMargin = 54
	frameYMargin = 12
	itemXMargin  = 24
	itemYMargin  = 20
	itemSize     = 20
)

func NewInventory(x, y int) *Inventory {
	button := NewImageButton(0, y/consts.TileScale, assets.GetImage("system/active_item_box.png"), assets.GetImage("system/active_item_box_pressed.png"), "click")
	button.DisabledImage = assets.GetImage("system/active_item_box_pressed.png")

	return &Inventory{
		X:                   x,
		Y:                   y,
		Visible:             true,
		PressedSlotIndex:    -1,
		items:               []*data.Item{},
		activeItemID:        0,
		activeItemBoxButton: button,
	}
}

func (i *Inventory) pressedSlotIndex() int {
	if !input.Triggered() {
		return -1
	}

	x, y := input.Position()
	x -= frameXMargin
	y = (y - itemYMargin)

	if x >= frameXMargin/consts.TileScale && i.Y <= y && y < i.Y+itemSize*consts.TileScale {
		return (x - 4) / (itemSize * consts.TileScale)
	}

	return -1
}

func (i *Inventory) ActiveItemPressed() bool {
	return i.activeItemBoxButton.Pressed()
}

func (i *Inventory) Update() {
	i.PressedSlotIndex = i.pressedSlotIndex()
	i.activeItemBoxButton.Update()
	i.activeItemBoxButton.Disabled = i.activeItemID == 0
}

func (i *Inventory) Draw(screen *ebiten.Image) {
	if !i.Visible {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	op.GeoM.Translate(float64(i.X+frameXMargin), float64(i.Y+frameYMargin))
	screen.DrawImage(assets.GetImage("system/frame_inventory.png"), op)

	var activeItem *data.Item
	for index, item := range i.items {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(float64(frameXMargin+itemXMargin+i.X+index*itemSize*consts.TileScale), float64(i.Y+itemYMargin))
		if i.activeItemID == item.ID {
			op.ColorM.Translate(0.5, 0.5, 0.5, 0)
			activeItem = item
		}
		screen.DrawImage(assets.GetImage("items/icon/"+item.Icon+".png"), op)
	}

	i.activeItemBoxButton.Draw(screen)

	if activeItem != nil {
		dy := 0
		if i.activeItemBoxButton.Pressing() {
			dy = 3
		}
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(float64(i.X+14), float64(i.Y+14+dy))
		screen.DrawImage(assets.GetImage("items/icon/"+activeItem.Icon+".png"), op)
		if len(activeItem.Commands) > 0 {
			op = &ebiten.DrawImageOptions{}
			op.GeoM.Scale(consts.TileScale, consts.TileScale)
			op.GeoM.Translate(float64(i.X), float64(i.Y+dy))
			screen.DrawImage(assets.GetImage("system/item_box_info.png"), op)
		}
	}
}

func (i *Inventory) SetItems(items []*data.Item) {
	i.items = items
}

func (i *Inventory) SetActiveItemID(activeItemID int) {
	i.activeItemID = activeItemID
}
