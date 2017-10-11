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
	"math"

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
	pressStartIndex     int
	pressStartX         int
	pressStartY         int
	scrollX             int
	dragX               int
	scrolling           bool
}

const (
	buttonOffsetX       = 9
	buttonOffsetY       = 6
	frameXMargin        = 24
	frameYMargin        = 4
	itemXMargin         = 6
	itemYMargin         = 6
	itemSize            = 21
	scrollDragThreshold = 5
	inventoryWidth      = 160
)

func NewInventory(x, y int) *Inventory {
	button := NewImageButton(buttonOffsetX/consts.TileScale, (y+buttonOffsetY)/consts.TileScale, NewImagePart(assets.GetImage("system/active_item_box.png")), NewImagePart(assets.GetImage("system/active_item_box_pressed.png")), "click")
	button.DisabledImage = NewImagePart(assets.GetImage("system/active_item_box_disabled.png"))
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

func (i *Inventory) slotIndexAt(x, y int) int {
	x -= (frameXMargin + itemXMargin) * consts.TileScale
	y = (y - (itemYMargin * consts.TileScale))

	if x >= 0 && i.Y <= y && y < i.Y+itemSize*consts.TileScale {
		return x / (itemSize * consts.TileScale)
	}

	return -1
}

func (i *Inventory) ActiveItemPressed() bool {
	return i.activeItemBoxButton.Pressed()
}

func (i *Inventory) Update() {
	touchX, touchY := input.Position()
	i.PressedSlotIndex = -1
	if input.Triggered() {
		i.pressStartX = touchX
		i.pressStartY = touchY
		i.pressStartIndex = i.slotIndexAt(touchX-(i.scrollX+i.dragX), touchY)
	}
	if input.Pressed() {
		dx := touchX - i.pressStartX
		if math.Abs(float64(dx)) > scrollDragThreshold && touchY > i.Y {
			scrollBarWidth := inventoryWidth - frameXMargin
			maxX := (itemXMargin + len(i.items)*itemSize - scrollBarWidth) * consts.TileScale

			if maxX > 0 {
				i.scrolling = true
				i.dragX = dx
				if i.scrollX+i.dragX > 0 {
					i.dragX = -i.scrollX
				}

				if (i.scrollX + i.dragX) < -maxX {
					i.dragX = -maxX - i.scrollX
				}
			}
		}
	}
	if input.Released() {
		if !i.scrolling && touchX > frameXMargin*consts.TileScale {
			index := i.slotIndexAt(touchX-(i.scrollX+i.dragX), touchY)
			if i.pressStartIndex == index {
				i.PressedSlotIndex = index
				i.pressStartIndex = -1
			}
		}
		i.scrollX += i.dragX
		i.dragX = 0
		i.scrolling = false
	}

	i.activeItemBoxButton.Update()
	i.activeItemBoxButton.Disabled = i.activeItemID == 0
}

func (i *Inventory) Draw(screen *ebiten.Image) {
	if !i.Visible || len(i.items) == 0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	op.GeoM.Translate(float64(i.X+frameXMargin*consts.TileScale), float64(i.Y+frameYMargin*consts.TileScale))
	screen.DrawImage(assets.GetImage("system/frame_inventory.png"), op)

	var activeItem *data.Item
	for index, item := range i.items {
		if i.activeItemID == item.ID {
			activeItem = item
		}

		tx := float64((frameXMargin+itemXMargin+i.X+index*itemSize)*consts.TileScale + i.scrollX + i.dragX)
		ty := float64(i.Y + itemYMargin*consts.TileScale)

		if tx < frameXMargin || tx > inventoryWidth*consts.TileScale {
			continue
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(tx, ty)
		if i.activeItemID == item.ID {
			screen.DrawImage(assets.GetImage("system/card_frame_selected.png"), op)
		} else {
			screen.DrawImage(assets.GetImage("system/card_frame.png"), op)
		}
		op.GeoM.Translate(3, 3)
		screen.DrawImage(assets.GetIconImage(item.Icon+".png"), op)
	}

	i.activeItemBoxButton.Draw(screen)

	if activeItem != nil {
		dy := 0
		if i.activeItemBoxButton.Pressing() {
			dy = 3
		}
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(float64(i.X+buttonOffsetX+14), float64(i.Y+buttonOffsetY+14+dy))
		screen.DrawImage(assets.GetIconImage(activeItem.Icon+".png"), op)
		if len(activeItem.Commands) > 0 {
			op = &ebiten.DrawImageOptions{}
			op.GeoM.Scale(consts.TileScale, consts.TileScale)
			op.GeoM.Translate(float64(i.X+buttonOffsetX), float64(i.Y+buttonOffsetY+dy))
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
