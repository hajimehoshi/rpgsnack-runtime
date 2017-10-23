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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type ItemPreviewPopup struct {
	X         int
	Y         int
	Width     int
	Height    int
	item      *data.Item
	Visible   bool
	fadeImage *ebiten.Image

	nodes         []Node
	closeButton   *Button
	previewButton *Button
}

func NewItemPreviewPopup(x, y, width, height int) *ItemPreviewPopup {
	previewButton := NewButton(0, 40, 120, 120, "ok")
	closeButton := NewButton(0, 160, 100, 20, "cancel")
	closeButton.Text = "Close"

	nodes := []Node{previewButton, closeButton}

	fadeImage, err := ebiten.NewImage(16, 16, ebiten.FilterNearest)
	if err != nil {
		panic(err)
	}

	return &ItemPreviewPopup{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		fadeImage: fadeImage,

		nodes:         nodes,
		closeButton:   closeButton,
		previewButton: previewButton,
	}
}

func (i *ItemPreviewPopup) Update() {
	i.previewButton.X = (i.Width - i.previewButton.Width*i.previewButton.ScaleX) / 2
	i.closeButton.X = (i.Width - i.closeButton.Width) / 2

	for _, n := range i.nodes {
		n.UpdateAsChild(i.Visible, i.X, i.Y)
	}

	if i.closeButton.Pressed() {
		i.item = nil
	}
}

func (i *ItemPreviewPopup) PreviewPressed() bool {
	return i.previewButton.Pressed()
}

func (i *ItemPreviewPopup) Item() *data.Item {
	return i.item
}

func (i *ItemPreviewPopup) SetItem(item *data.Item) {
	i.item = item
	if i.item == nil || (i.item.Preview == "" && i.item.Icon == "") {
		i.previewButton.Visible = false
		return
	}
	i.previewButton.Visible = true
	if i.item.Preview != "" {
		i.previewButton.Image = NewImagePart(assets.GetImage("items/preview/" + i.item.Preview + ".png"))
		i.previewButton.ScaleX = 1
		i.previewButton.ScaleY = 1
		i.previewButton.Width = 120
		i.previewButton.Height = 120

	} else {
		i.previewButton.Image = NewImagePart(assets.GetIconImage(i.item.Icon + ".png"))
		i.previewButton.ScaleX = 6
		i.previewButton.ScaleY = 6
		i.previewButton.Width = 16
		i.previewButton.Height = 16
	}
}

func (i *ItemPreviewPopup) Draw(screen *ebiten.Image) {
	if !i.Visible {
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

	for _, n := range i.nodes {
		n.DrawAsChild(screen, i.X, i.Y)
	}
}
