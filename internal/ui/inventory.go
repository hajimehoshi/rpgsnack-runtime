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
	autoScrolling       bool
	pageIndex           int
	targetPageIndex     int
	bgPanel             *ImageView
	frameCover          *ImageView
	frameBase           *ImageView
	activeCardSlot      *ImagePart
	cardSlot            *ImagePart
	activeDot           *ImagePart
	dot                 *ImagePart
}

const (
	buttonOffsetX       = 0
	buttonOffsetY       = 0
	frameXMargin        = 34
	frameYMargin        = 2
	frameXPadding       = 6
	frameYPadding       = 6
	itemSize            = 19
	scrollDragThreshold = 5
	scrollBarWidth      = 120
	itemPerPageCount    = 6
	autoScrollSpeed     = 20
	snapDragX           = 60
	dotSpace            = 8
)

func NewInventory(x, y int) *Inventory {
	button := NewImageButton(
		x+buttonOffsetX/consts.TileScale,
		y+buttonOffsetY/consts.TileScale,
		NewImagePartWithRect(assets.GetImage("system/ui_footer.png"), 40, 0, 36, 32),
		NewImagePartWithRect(assets.GetImage("system/ui_footer.png"), 0, 0, 36, 32),
		"click",
	)
	button.DisabledImage = NewImagePartWithRect(assets.GetImage("system/ui_footer.png"), 80, 0, 36, 32)

	bgPanel := NewImageView(x, y, 1.0, NewImagePartWithRect(assets.GetImage("system/ui_footer.png"), 0, 32, 160, 40))
	frameCover := NewImageView(x+34, y+4, 1.0, NewImagePartWithRect(assets.GetImage("system/ui_footer.png"), 0, 144, 128, 24))
	frameBase := NewImageView(x+34, y+4, 1.0, NewImagePartWithRect(assets.GetImage("system/ui_footer.png"), 0, 168, 128, 24))

	// TODO: Bug fix: x,y should be in the tile-scaled world
	return &Inventory{
		X:                   x * consts.TileScale,
		Y:                   y * consts.TileScale,
		Visible:             true,
		PressedSlotIndex:    -1,
		items:               []*data.Item{},
		activeItemID:        0,
		activeItemBoxButton: button,
		bgPanel:             bgPanel,
		frameCover:          frameCover,
		frameBase:           frameBase,
		cardSlot:            NewImagePartWithRect(assets.GetImage("system/ui_footer.png"), 120, 0, 18, 18),
		activeCardSlot:      NewImagePartWithRect(assets.GetImage("system/ui_footer.png"), 138, 0, 18, 18),
		dot:                 NewImagePartWithRect(assets.GetImage("system/ui_footer.png"), 120, 24, 8, 8),
		activeDot:           NewImagePartWithRect(assets.GetImage("system/ui_footer.png"), 128, 24, 8, 8),
		pageIndex:           0,
		targetPageIndex:     0,
	}
}

func (i *Inventory) slotIndexAt(x, y int) int {
	x -= (frameXMargin + frameXPadding) * consts.TileScale
	y = (y - (frameYPadding * consts.TileScale))

	if x >= 0 && i.Y <= y && y < i.Y+itemSize*consts.TileScale {
		return x / (itemSize * consts.TileScale)
	}

	return -1
}

func (i *Inventory) pageCount() int {
	return int(math.Max(1, math.Ceil(float64(len(i.items))/float64(itemPerPageCount))))
}

func (i *Inventory) slotCount() int {
	return i.pageCount() * itemPerPageCount
}

func (i *Inventory) ActiveItemPressed() bool {
	return i.activeItemBoxButton.Pressed()
}

func (i *Inventory) calcScrollX(pageIndex int) int {
	return -(pageIndex * (scrollBarWidth - frameXPadding)) * consts.TileScale
}

func (i *Inventory) Update() {
	touchX, touchY := input.Position()
	i.PressedSlotIndex = -1
	if input.Triggered() {
		i.pressStartX = touchX
		i.pressStartY = touchY
		i.pressStartIndex = i.slotIndexAt(touchX-(i.X+i.scrollX+i.dragX), touchY)
		i.autoScrolling = false
	}
	if input.Pressed() {
		dx := touchX - i.pressStartX
		if math.Abs(float64(dx)) > scrollDragThreshold && touchY > i.Y {
			i.scrolling = true
			i.dragX = dx
		}
	}
	if input.Released() {
		if !i.scrolling && touchX > (i.X+(frameXMargin+frameXPadding)*consts.TileScale) {
			index := i.slotIndexAt(touchX-(i.X+i.scrollX+i.dragX), touchY)
			if i.pressStartIndex == index {
				i.PressedSlotIndex = index
				i.pressStartIndex = -1
			}
		}

		i.targetPageIndex = i.pageIndex
		if i.dragX > snapDragX && i.pageIndex > 0 {
			i.targetPageIndex = i.pageIndex - 1
		}
		if i.dragX < -snapDragX && i.pageIndex < i.pageCount()-1 {
			i.targetPageIndex = i.pageIndex + 1
		}
		i.autoScrolling = true
		i.scrollX += i.dragX
		i.dragX = 0
		i.scrolling = false
	}

	i.activeItemBoxButton.Update()
	i.activeItemBoxButton.Disabled = i.activeItemID == 0

	if i.autoScrolling {
		targetX := i.calcScrollX(i.targetPageIndex)
		dx := float64(targetX - i.scrollX)
		if dx > 0 {
			i.scrollX += autoScrollSpeed
		} else {
			i.scrollX -= autoScrollSpeed
		}
		if math.Abs(dx) < autoScrollSpeed {
			i.scrollX = targetX
			i.pageIndex = i.targetPageIndex
			i.autoScrolling = false
		}
	}
}

func (i *Inventory) Draw(screen *ebiten.Image) {
	if !i.Visible {
		return
	}

	i.bgPanel.Draw(screen)
	i.frameBase.Draw(screen)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	op.GeoM.Translate(float64(i.X+frameXMargin*consts.TileScale), float64(i.Y+frameYMargin*consts.TileScale))

	var activeItem *data.Item
	for index := 0; index < i.slotCount(); index++ {
		var item *data.Item
		itemID := -2
		if index < len(i.items) {
			item = i.items[index]
			itemID = item.ID
			if i.activeItemID == item.ID {
				activeItem = item
			}
		}

		tx := float64(i.X + (frameXMargin+frameXPadding+index*itemSize)*consts.TileScale + i.scrollX + i.dragX)
		ty := float64(i.Y+frameYPadding*consts.TileScale) + 2

		if tx < float64(i.X+frameXMargin) || tx > float64(i.X+(frameXMargin+scrollBarWidth)*consts.TileScale) {
			continue
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(tx, ty)
		if i.activeItemID == itemID {
			i.activeCardSlot.Draw(screen, &op.GeoM, &ebiten.ColorM{})
		} else {
			i.cardSlot.Draw(screen, &op.GeoM, &ebiten.ColorM{})
		}

		if item != nil {
			op.GeoM.Translate(3, 0)
			screen.DrawImage(assets.GetIconImage(item.Icon+".png"), op)
		}
	}

	i.frameCover.Draw(screen)

	i.activeItemBoxButton.Draw(screen)

	if activeItem != nil {
		dy := 0
		if i.activeItemBoxButton.Pressing() {
			dy = 6
		}
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(float64(i.X+buttonOffsetX+30), float64(i.Y+buttonOffsetY+14+dy))
		screen.DrawImage(assets.GetIconImage(activeItem.Icon+".png"), op)
	}

	centerX := frameXMargin + scrollBarWidth/2
	left := int(float64(centerX) - float64(i.pageCount())/2*dotSpace)

	// We only show dots UI if there are more than one page
	if i.pageCount() > 1 {
		for index := 0; index < i.pageCount(); index++ {
			var imagePart *ImagePart
			if index == i.pageIndex {
				imagePart = i.activeDot
			} else {
				imagePart = i.dot
			}
			geoM := &ebiten.GeoM{}
			geoM.Translate(float64(left+index*dotSpace), float64(i.Y/consts.TileScale+26))
			geoM.Scale(consts.TileScale, consts.TileScale)
			imagePart.Draw(screen, geoM, &ebiten.ColorM{})
		}
	}
}

func (i *Inventory) SetItems(items []*data.Item) {
	i.items = items
}

func (i *Inventory) SetActiveItemID(activeItemID int) {
	i.activeItemID = activeItemID
}
