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

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
)

type ItemPreviewPopup struct {
	x               int
	y               int
	item            *data.Item
	combineItem     *data.Item
	combine         *data.Combine
	visible         bool
	fadeImage       *ebiten.Image
	frameImage      *ebiten.Image
	bgBoxImage      *ebiten.Image
	previewBoxImage *ebiten.Image
	nodes           []Node
	closeButton     *Button
	previewButton   *Button
	actionButton    *Button
	desc            string
}

func NewItemPreviewPopup(x, y int) *ItemPreviewPopup {
	closeButton := NewImageButton(
		120,
		25,
		NewImagePart(assets.GetImage("system/item_cancel_off.png")),
		NewImagePart(assets.GetImage("system/item_cancel_on.png")),
		"cancel",
	)

	actionButton := NewImageButton(
		40,
		114,
		NewImagePart(assets.GetImage("system/item_action_button_off.png")),
		NewImagePart(assets.GetImage("system/item_action_button_on.png")),
		"click",
	)
	frameImage := assets.GetImage("system/item_details.png")
	bgBoxImage := assets.GetImage("system/item_bg_box.png")
	previewBoxImage := assets.GetImage("system/item_preview_box.png")

	nodes := []Node{closeButton, actionButton}

	fadeImage, err := ebiten.NewImage(16, 16, ebiten.FilterNearest)
	if err != nil {
		panic(err)
	}

	return &ItemPreviewPopup{
		x:               x,
		y:               y,
		fadeImage:       fadeImage,
		frameImage:      frameImage,
		bgBoxImage:      bgBoxImage,
		previewBoxImage: previewBoxImage,
		nodes:           nodes,
		closeButton:     closeButton,
		actionButton:    actionButton,
	}
}

func (i *ItemPreviewPopup) Update(sceneManager *scene.Manager) {
	for _, n := range i.nodes {
		n.UpdateAsChild(i.visible, i.x, i.y)
	}
	i.actionButton.Text = texts.Text(sceneManager.Language(), texts.TextIDItemCheck)
}

func (i *ItemPreviewPopup) Show() {
	i.visible = true
}

func (i *ItemPreviewPopup) Hide() {
	i.visible = false
}

func (i *ItemPreviewPopup) Visible() bool {
	return i.visible
}

func (i *ItemPreviewPopup) ClosePressed() bool {
	return i.closeButton.Pressed()
}

func (i *ItemPreviewPopup) ActionPressed() bool {
	return i.actionButton.Pressed()
}

func (i *ItemPreviewPopup) SetActiveItem(item *data.Item, desc string) {
	if i.item != item {
		i.item = item
		i.desc = desc
		i.combineItem = nil
		i.combine = nil
	}
}

func (i *ItemPreviewPopup) SetCombineItem(item *data.Item, combine *data.Combine) {
	i.combineItem = item
	i.combine = combine
}

func (i *ItemPreviewPopup) DrawItem(screen *ebiten.Image, x, y float64, icon string) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	screen.DrawImage(i.previewBoxImage, op)

	if i.item.Icon != "" {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x/consts.TileScale, y/consts.TileScale)
		op.GeoM.Scale(consts.TileScale*3, consts.TileScale*3)
		screen.DrawImage(assets.GetIconImage(icon+".png"), op)
	}
}

func (i *ItemPreviewPopup) Draw(screen *ebiten.Image) {
	if !i.visible {
		return
	}

	w, h := i.fadeImage.Size()
	sw, sh := screen.Size()
	sx := float64(sw) / float64(w)
	sy := float64(sh) / float64(h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sx, sy)
	op.ColorM.Translate(0, 0, 0, 1)
	op.ColorM.Scale(1, 1, 1, 0.5)
	screen.DrawImage(i.fadeImage, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(6, 35)
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	screen.DrawImage(i.frameImage, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(12, 65)
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	screen.DrawImage(i.bgBoxImage, op)

	i.DrawItem(screen, 16, 70, i.item.Icon)
	if i.combineItem != nil && i.combineItem.ID != i.item.ID {
		i.DrawItem(screen, 92, 70, i.combineItem.Icon)
	} else {
		font.DrawText(screen, i.desc, 68*consts.TileScale+consts.TextScale, 72*consts.TileScale+consts.TextScale, consts.TextScale, data.TextAlignLeft, color.Black)
		font.DrawText(screen, i.desc, 68*consts.TileScale, 72*consts.TileScale, consts.TextScale, data.TextAlignLeft, color.White)
	}

	for _, n := range i.nodes {
		n.DrawAsChild(screen, i.x, i.y)
	}
}
