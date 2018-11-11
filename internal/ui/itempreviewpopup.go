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
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
)

type ItemPreviewPopup struct {
	y               int
	item            *data.Item
	combineItem     *data.Item
	combine         *data.Combine
	visible         bool
	fadeImage       *ebiten.Image
	frameImage      *ebiten.Image
	previewBoxImage *ebiten.Image
	nodes           []Node
	closeButton     *Button
	previewButton   *Button
	actionButton    *Button
}

func NewItemPreviewPopup(y int) *ItemPreviewPopup {
	closeButton := NewImageButton(
		135,
		5,
		assets.GetImage("system/itempreview/cancel_off.png"),
		assets.GetImage("system/itempreview/cancel_on.png"),
		"system/cancel",
	)

	actionButton := NewImageButton(
		54,
		93,
		assets.GetImage("system/itempreview/action_button_off.png"),
		assets.GetImage("system/itempreview/action_button_on.png"),
		"system/click",
	)
	frameImage := assets.GetImage("system/itempreview/details.png")
	previewBoxImage := assets.GetImage("system/itempreview/preview_box.png")

	nodes := []Node{closeButton, actionButton}

	fadeImage, err := ebiten.NewImage(16, 16, ebiten.FilterNearest)
	if err != nil {
		panic(err)
	}

	return &ItemPreviewPopup{
		y:               y,
		fadeImage:       fadeImage,
		frameImage:      frameImage,
		previewBoxImage: previewBoxImage,
		nodes:           nodes,
		closeButton:     closeButton,
		actionButton:    actionButton,
	}
}

func (i *ItemPreviewPopup) Update(lang language.Tag) {
	if !i.visible {
		return
	}
	for _, n := range i.nodes {
		n.UpdateAsChild(i.visible, 0, i.y)
	}
	i.actionButton.Text = texts.Text(lang, texts.TextIDItemCheck)
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

func (i *ItemPreviewPopup) SetOnClosePressed(f func(itemPreviewPopup *ItemPreviewPopup)) {
	i.closeButton.SetOnPressed(func(_ *Button) {
		f(i)
	})
}

func (i *ItemPreviewPopup) SetOnActionPressed(f func(itemPreviewPopup *ItemPreviewPopup)) {
	i.actionButton.SetOnPressed(func(_ *Button) {
		f(i)
	})
}

func (i *ItemPreviewPopup) SetEnabled(enabled bool) {
	i.actionButton.Disabled = !enabled
	i.closeButton.Disabled = !enabled
}

func (i *ItemPreviewPopup) SetActiveItem(item *data.Item) {
	if i.item != item {
		i.item = item
		i.combineItem = nil
		i.combine = nil
	}
}

func (i *ItemPreviewPopup) SetCombineItem(item *data.Item, combine *data.Combine) {
	i.combineItem = item
	i.combine = combine
}

func (i *ItemPreviewPopup) drawItem(screen *ebiten.Image, x, y float64, icon string) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	screen.DrawImage(i.previewBoxImage, op)

	if i.item.Icon != "" {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate((x+3)/consts.TileScale, (y+3)/consts.TileScale)
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
	op.GeoM.Translate(6, float64(i.y))
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	screen.DrawImage(i.frameImage, op)

	op = &ebiten.DrawImageOptions{}
	if i.combineItem != nil && i.combineItem.ID != i.item.ID {
		i.drawItem(screen, 16, float64(i.y)+28, i.item.Icon)
		i.drawItem(screen, 88, float64(i.y)+28, i.combineItem.Icon)
	} else {
		i.drawItem(screen, 54, float64(i.y)+28, i.item.Icon)
	}

	for _, n := range i.nodes {
		n.DrawAsChild(screen, 0, i.y)
	}
}
