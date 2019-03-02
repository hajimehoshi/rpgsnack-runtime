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
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type InventoryMode int

const (
	DefaultMode InventoryMode = iota
	PreviewMode
)

type Inventory struct {
	x                   int
	y                   int
	showItemName        bool
	itemLabel           *Label
	visible             bool
	disabled            bool
	items               []*data.Item
	activeItemID        int
	combineItemID       int
	infoButton          *Button
	backButton          *Button
	pressStartIndex     int
	pressStartX         int
	pressStartY         int
	scrollX             int
	dragX               int
	dragging            bool
	scrolling           bool
	autoScrolling       bool
	prepareToCancel     bool
	pageIndex           int
	targetPageIndex     int
	stashedActiveItemID int
	stashedPageIndex    int
	bgPanel             *ImageView
	frameCover          *ImageView
	frameBase           *ImageView
	activeCardSlot      *ebiten.Image
	combineCardSlot     *ebiten.Image
	cardSlot            *ebiten.Image
	activeDot           *ebiten.Image
	dot                 *ebiten.Image
	mode                InventoryMode

	onSlotPressed    func(inventory *Inventory, index int)
	onOutsidePressed func(inventory *Inventory)
}

const (
	buttonOffsetX = 0
	buttonOffsetY = 0

	frameXMargin  = 34
	frameYMargin  = 2
	frameXPadding = 5
	frameYPadding = 6

	itemSize            = 19
	scrollDragThreshold = 5
	scrollBarWidth      = 120
	scrollBarHeight     = 24
	itemPerPageCount    = 6
	autoScrollSpeed     = 20
	snapDragX           = 60
	dotSpace            = 8
)

func NewInventory(x, y int, showItemName bool) *Inventory {
	backButton := NewImageButton(
		x,
		y,
		assets.GetImage("system/footer/back_button.png"),
		assets.GetImage("system/footer/back_button_on.png"),
		"system/cancel",
	)

	infoButton := NewImageButton(
		x+buttonOffsetX,
		y+buttonOffsetY,
		assets.GetImage("system/footer/info_button_off.png"),
		assets.GetImage("system/footer/info_button_on.png"),
		"system/click",
	)
	infoButton.DisabledImage = assets.GetImage("system/footer/info_button_disabled.png")
	infoButton.disabled = true

	bgPanel := NewImageView(x, y, 1.0, assets.GetImage("system/footer/panel.png"))
	frameCover := NewImageView(x+frameXMargin, y+4, 1.0, assets.GetImage("system/footer/inventory_mask.png"))
	frameBase := NewImageView(x+frameXMargin, y+4, 1.0, assets.GetImage("system/footer/inventory_bg.png"))

	itemLabel := NewLabel(x+80, y+28)
	itemLabel.Scale = 1.5
	itemLabel.Color = color.RGBA{0x3B, 0x58, 0x48, 0xff}
	itemLabel.TextAlign = data.TextAlignCenter

	return &Inventory{
		x:               x,
		y:               y,
		showItemName:    showItemName,
		itemLabel:       itemLabel,
		visible:         true,
		items:           []*data.Item{},
		activeItemID:    0,
		combineItemID:   0,
		infoButton:      infoButton,
		backButton:      backButton,
		bgPanel:         bgPanel,
		frameCover:      frameCover,
		frameBase:       frameBase,
		cardSlot:        assets.GetImage("system/footer/item_holder.png"),
		activeCardSlot:  assets.GetImage("system/footer/item_holder_active.png"),
		combineCardSlot: assets.GetImage("system/footer/item_holder_selected.png"),
		dot:             assets.GetImage("system/footer/dot_off.png"),
		activeDot:       assets.GetImage("system/footer/dot_on.png"),
		pageIndex:       0,
		targetPageIndex: 0,
		mode:            DefaultMode,
	}
}

func (i *Inventory) SetOnSlotPressed(f func(inventory *Inventory, index int)) {
	i.onSlotPressed = f
}

func (i *Inventory) SetOnOutsidePressed(f func(inventory *Inventory)) {
	i.onOutsidePressed = f
}

func (i *Inventory) Show() {
	i.visible = true
}

func (i *Inventory) Hide() {
	i.visible = false
}

func (i *Inventory) Visible() bool {
	return i.visible
}

func (i *Inventory) SetDisabled(disabled bool) {
	i.disabled = disabled
}

func (i *Inventory) slotIndexAt(x, y int) int {
	x -= (frameXMargin + frameXPadding) * consts.TileScale
	y = (y - (frameYPadding * consts.TileScale))

	if x >= 0 && i.y*consts.TileScale <= y && y < (i.y+itemSize)*consts.TileScale {
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

func (i *Inventory) SetOnActiveItemPressed(f func(inventory *Inventory)) {
	i.infoButton.SetOnPressed(func(_ *Button) {
		f(i)
	})
}

func (i *Inventory) SetOnBackPressed(f func(inventory *Inventory)) {
	i.backButton.SetOnPressed(func(_ *Button) {
		f(i)
	})
}

func (i *Inventory) calcScrollX(pageIndex int) int {
	return -(pageIndex * (scrollBarWidth - frameXPadding)) * consts.TileScale
}

func (i *Inventory) isTouchingScroll() bool {
	touchX, touchY := input.Position()
	sx := (i.x + frameXMargin + frameXPadding) * consts.TileScale
	sy := i.y * consts.TileScale
	return sx <= touchX && touchX < sx+scrollBarWidth*consts.TileScale && sy <= touchY && touchY < sy+scrollBarHeight*consts.TileScale
}

func (i *Inventory) Update(sceneManager *scene.Manager) {
	if !i.visible {
		return
	}

	activeItem := i.activeItem()
	if i.showItemName && activeItem != nil {
		i.itemLabel.Text = sceneManager.Game().Texts.Get(lang.Get(), activeItem.Name)
	} else {
		i.itemLabel.Text = ""
	}

	touchX, touchY := input.Position()

	if touchY < i.y*consts.TileScale {
		if input.Triggered() {
			i.prepareToCancel = true
			return
		}
		if i.prepareToCancel && input.Released() && i.onOutsidePressed != nil {
			i.onOutsidePressed(i)
		}
	}

	if input.Released() {
		i.prepareToCancel = false
	}

	if i.disabled {
		return
	}

	if input.Triggered() && i.isTouchingScroll() {
		i.pressStartX = touchX
		i.pressStartY = touchY
		i.pressStartIndex = i.slotIndexAt(touchX-(i.x*consts.TileScale+i.scrollX+i.dragX), touchY)
		i.autoScrolling = false
		i.dragging = true
	}
	if i.dragging {
		dx := touchX - i.pressStartX
		if math.Abs(float64(dx)) > scrollDragThreshold && i.isTouchingScroll() {
			i.scrolling = true
			i.dragX = dx
			i.pressStartIndex = -1
		}
	}
	if input.Released() {
		if !i.scrolling && i.isTouchingScroll() {
			index := i.slotIndexAt(touchX-(i.x*consts.TileScale+i.scrollX+i.dragX), touchY)
			if i.pressStartIndex == index && index >= 0 && index < len(i.items) {
				if i.onSlotPressed != nil {
					i.onSlotPressed(i, index)
				}
			}
		}
		i.pressStartIndex = -1
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
		i.dragging = false
	}

	i.infoButton.Update()
	i.infoButton.disabled = false
	if i.activeItemID == 0 {
		i.infoButton.disabled = true
	}
	i.backButton.Update()
	i.backButton.disabled = false
	if i.mode == DefaultMode {
		i.backButton.disabled = true
	}

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

	i.itemLabel.Update()
}

func (i *Inventory) activeItem() *data.Item {
	for index := 0; index < i.slotCount(); index++ {
		var item *data.Item
		if index < len(i.items) {
			item = i.items[index]
			if i.activeItemID == item.ID {
				return item
			}
		}
	}
	return nil
}

func (i *Inventory) Draw(screen *ebiten.Image) {
	if !i.visible {
		return
	}

	i.bgPanel.Draw(screen)
	i.frameBase.Draw(screen)

	activeItem := i.activeItem()
	for index := 0; index < i.slotCount(); index++ {
		var item *data.Item
		itemID := -2
		if index < len(i.items) {
			item = i.items[index]
			itemID = item.ID
		}

		tx := float64((i.x + frameXMargin + frameXPadding + index*itemSize) + (i.scrollX+i.dragX)/consts.TileScale)
		ty := float64(i.y+frameYPadding) + 1

		if tx < float64(i.x) || tx > float64(i.x+frameXMargin+scrollBarWidth) {
			continue
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(tx, ty)
		op.GeoM.Scale(consts.TileScale, consts.TileScale)

		if i.activeItemID == itemID {
			screen.DrawImage(i.activeCardSlot, op)
		} else {
			if i.mode == PreviewMode && item != nil && i.combineItemID == item.ID {
				screen.DrawImage(i.combineCardSlot, op)
			} else {
				screen.DrawImage(i.cardSlot, op)
			}
		}

		if item != nil {
			op.GeoM.Translate(3, 0)
			screen.DrawImage(assets.GetIconImage(item.Icon+".png"), op)
		}
	}

	i.frameCover.Draw(screen)

	if i.mode == DefaultMode {
		i.infoButton.Draw(screen)
		if activeItem != nil {
			dy := 2
			if i.infoButton.pressing {
				dy = 3
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(i.x*consts.TileScale+buttonOffsetX+10), float64(i.y+buttonOffsetY+5+dy))
			op.GeoM.Scale(consts.TileScale, consts.TileScale)
			screen.DrawImage(assets.GetIconImage(activeItem.Icon+".png"), op)
		}
	} else {
		i.backButton.Draw(screen)
	}

	centerX := frameXMargin + scrollBarWidth/2
	left := int(float64(centerX) - float64(i.pageCount())/2*dotSpace)

	// We only show dots UI if there are more than one page
	if i.pageCount() > 1 {
		op := &ebiten.DrawImageOptions{}
		for index := 0; index < i.pageCount(); index++ {
			var imagePart *ebiten.Image
			if index == i.pageIndex {
				imagePart = i.activeDot
			} else {
				imagePart = i.dot
			}
			op.GeoM.Reset()
			op.GeoM.Translate(float64(left+index*dotSpace), float64(i.y+26))
			op.GeoM.Scale(consts.TileScale, consts.TileScale)
			screen.DrawImage(imagePart, op)
		}
	}
	i.itemLabel.Draw(screen)
}

func (i *Inventory) SetMode(mode InventoryMode) {
	i.mode = mode
}

func (i *Inventory) SetItems(items []*data.Item) {
	i.items = items
}

func (i *Inventory) ActiveItemID() int {
	if i.mode == DefaultMode {
		return i.activeItemID
	}
	return 0
}

func (i *Inventory) SetActiveItemID(activeItemID int) {
	i.activeItemID = activeItemID
}

func (i *Inventory) SetCombineItemID(combineItemID int) {
	i.combineItemID = combineItemID
}

func (i *Inventory) Mode() InventoryMode {
	return i.mode
}
