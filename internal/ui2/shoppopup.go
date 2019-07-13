// Copyright 2019 The RPGSnack Authors
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

package ui2

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets/embedded"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
)

var (
	shopHeaderColor = color.RGBA{0xbb, 0xff, 0x66, 0xff}
	yellow          = color.RGBA{0xff, 0xff, 0, 0xff}
)

const ShopPopupTabInvalidID = -1

type ShopPopupTab struct {
	ID       int
	Name     string
	Products []*data.ShopProduct
}

type ShopPopupData struct {
	Tabs []*ShopPopupTab
}

type position struct {
	x int
	y int
}

type ShopPopup struct {
	screenHeight int

	popup *Popup

	mainPanels   map[int]*ScrollablePanel
	buyButtons   map[string]*Button
	detailPanels map[string]*ScrollablePanel
	tabs         map[int]*Button

	shopLabelBg *ImageView
	shopLabel   *Label
	closeButton *Button
	backButton  *Button
	border      *ImageView

	tabID     int
	detailKey string

	scrollPositions map[int]position

	loading      bool
	loadingCount int
}

type PurchaseRequester interface {
	RequestPurchase(key string)
}

type Purchased interface {
	IsPurchased(key string) bool
}

const (
	// TODO: When the screen height is changed, the position also should be updated.
	popupHeight      = 600
	shopHeaderHeight = 60
	shopFooterHeight = 60
)

const loadingSpeed = 5

var (
	shopLabelBgImage *ebiten.Image
	shopBorderImage  *ebiten.Image
)

func init() {
	shopLabelBgImage, _ = ebiten.NewImage(PopupWidth-2, shopHeaderHeight-1, ebiten.FilterNearest)
	shopLabelBgImage.Fill(color.RGBA{0x80, 0x80, 0x80, 0xff})
	shopBorderImage, _ = ebiten.NewImage(PopupWidth-2, 1, ebiten.FilterNearest)
	shopBorderImage.Fill(color.White)
}

func NewShopPopup(screenHeight int) *ShopPopup {
	s := &ShopPopup{
		screenHeight:    screenHeight,
		scrollPositions: map[int]position{},
	}

	s.shopLabelBg = NewImageView(1, 1, 1, shopLabelBgImage)
	s.shopLabel = NewLabel(PopupWidth/2, 12)
	s.shopLabel.SetTextAlign(data.TextAlignCenter)
	s.shopLabel.SetScale(2)

	img := embedded.Get("close")
	imgw, imgh := img.Size()
	s.closeButton = NewImageButton(
		PopupWidth-imgw-10,
		(shopHeaderHeight-imgh)/2,
		img,
		nil,
		"system/cancel",
	)
	s.closeButton.SetOnPressed(func(_ *Button) {
		s.popup.Hide()
	})

	img = embedded.Get("back")
	imgw, imgh = img.Size()
	s.backButton = NewImageButton(
		10,
		(shopHeaderHeight-imgh)/2,
		img,
		nil,
		"",
	)
	s.backButton.SetOnPressed(func(*Button) {
		s.goBack()
	})

	s.border = NewImageView(1, popupHeight-shopFooterHeight, 1, shopBorderImage)

	return s
}

func (s *ShopPopup) OnPurchased(succeeded bool) {
	s.loading = false
	s.loadingCount = 0
	if succeeded {
		s.Hide()
		return
	}
}

func (s *ShopPopup) SetShopData(shopData *ShopPopupData, req PurchaseRequester) {
	s.popup = NewPopup((s.screenHeight-popupHeight)/2, popupHeight)
	s.mainPanels = map[int]*ScrollablePanel{}
	s.buyButtons = map[string]*Button{}
	s.detailPanels = map[string]*ScrollablePanel{}
	s.tabs = map[int]*Button{}

	const (
		headerTextScale = 2
		descTextScale   = 1.5
		tabTextScale    = 1.5
		buttonTextScale = 1.5
		smallTextScale  = 1.5
	)

	const margin = 12

	for i, tab := range shopData.Tabs {
		const panelw = PopupWidth - margin*2
		h := popupHeight - shopHeaderHeight - 1
		if len(shopData.Tabs) > 1 {
			h = popupHeight - shopHeaderHeight - shopFooterHeight
		}
		panel := NewScrollablePanel(margin, shopHeaderHeight, panelw, h)
		panel.Hide()

		const lh = font.RenderingLineHeight

		w := (PopupWidth - margin*2) / len(shopData.Tabs)
		t := NewTextButton(i*w+margin, popupHeight-shopFooterHeight+(shopFooterHeight-40)/2, w, 40, "system/click")
		t.SetText(tab.Name)
		t.SetScale(tabTextScale)
		t.SetColor(color.White)
		t.SetDisabledColor(yellow)
		s.tabs[tab.ID] = t

		// Add all the ImageViews later for more efficient rendering. They use the linear filter.
		var imgs []*ImageView

		y := margin
		for i, p := range tab.Products {
			l := NewLabel(0, y)
			l.SetText(p.Name)
			l.SetScale(headerTextScale)
			if s := consts.SponsorTierType(p.Type); s.IsValid() {
				l.SetColor(s.Color())
			} else {
				l.SetColor(shopHeaderColor)
			}
			panel.AddChild(l)

			descX := panelw / 3
			descWidth := int(float64(panelw) * 2 / 3 * (1.0 / descTextScale))

			y += lh * headerTextScale
			l = NewLabel(descX, y)
			l.SetText(font.InsertNewLines(p.Desc, descWidth, lang.Get()))
			l.SetScale(descTextScale)
			panel.AddChild(l)

			img := assets.GetImage(fmt.Sprintf("shop/%s.png", shopImageNameFromKey(p.Key)))
			w, _ := img.Size()
			x := (descX - w/2) / 2
			imgv := NewImageView(x, y+margin, 1.0/2, img)
			imgv.SetFilter(ebiten.FilterLinear)
			imgs = append(imgs, imgv)

			y += 4 * lh * descTextScale

			key := p.Key
			if p.Details != "" {
				b := NewButton(panelw-100*2-12, y, 100, 40, "system/click")
				b.SetText(texts.Text(lang.Get(), texts.TextIDDetails))
				b.SetScale(buttonTextScale)
				b.SetOnPressed(func(*Button) {
					s.detailKey = key
				})
				panel.AddChild(b)
			}
			b := NewButton(panelw-100, y, 100, 40, "system/click")
			b.SetScale(buttonTextScale)
			b.SetOnPressed(func(*Button) {
				s.loading = true
				req.RequestPurchase(key)
			})
			s.buyButtons[p.Key] = b
			panel.AddChild(b)

			y += 40 // button height
			if i < len(tab.Products)-1 {
				y += lh * descTextScale
			}

			// Create a panel for details
			if p.Details != "" {
				panel := NewScrollablePanel(margin, shopHeaderHeight, panelw, popupHeight-shopHeaderHeight-40-margin)
				y := margin

				l := NewLabel(0, y)
				l.SetText(p.Name)
				l.SetScale(headerTextScale)
				if s := consts.SponsorTierType(p.Type); s.IsValid() {
					l.SetColor(s.Color())
				} else {
					l.SetColor(shopHeaderColor)
				}
				panel.AddChild(l)

				y += lh * headerTextScale
				w := int(float64(panelw) * (1.0 / descTextScale))
				l = NewLabel(0, y)
				l.SetText(font.InsertNewLines(p.Details, w, lang.Get()))
				l.SetScale(descTextScale)
				l.SetColor(color.White)
				panel.AddChild(l)
				s.detailPanels[p.Key] = panel
			}
		}

		for _, i := range imgs {
			panel.AddChild(i)
		}

		// Add a margin at the footer in the panel.
		panel.AddChild(NewPadding(0, y, 1, margin))

		s.mainPanels[tab.ID] = panel
	}

	s.popup.AddChild(s.shopLabelBg)
	s.popup.AddChild(s.shopLabel)
	s.popup.AddChild(s.closeButton)
	s.popup.AddChild(s.backButton)
	s.popup.AddChild(s.border)
	for _, p := range s.mainPanels {
		s.popup.AddChild(p)
	}
	for _, p := range s.detailPanels {
		s.popup.AddChild(p)
	}
	for _, t := range s.tabs {
		s.popup.AddChild(t)
	}

	for id, t := range s.tabs {
		id := id
		t.SetOnPressed(func(_ *Button) {
			s.tabID = id
		})
	}

	s.updateVisibilities()

	// Set the recorded scroll positions
	for id, p := range s.scrollPositions {
		if _, ok := s.mainPanels[id]; ok {
			s.mainPanels[id].SetScrollPosition(p.x, p.y)
		}
	}
}

func (s *ShopPopup) Visible() bool {
	return s.popup.Visible()
}

func (s *ShopPopup) Show() {
	s.popup.Show()
}

func (s *ShopPopup) Hide() {
	s.popup.Hide()
}

func (s *ShopPopup) HandleInput(offsetX, offsetY int) bool {
	if s.loading {
		return true
	}

	if input.BackButtonTriggered() {
		if s.goBack() {
			return true
		}
	}

	if s.popup.HandleInput(offsetX, offsetY) {
		return true
	}

	// If a popup is visible, do not propagate any input handling to parents.
	return true
}

func (s *ShopPopup) goBack() bool {
	if s.detailKey == "" {
		return false
	}

	audio.PlaySE("system/cancel", 1)
	s.detailKey = ""
	return true
}

func (s *ShopPopup) Update(prices map[string]string, purchased Purchased) {
	if s.loading {
		s.loadingCount++
		img := embedded.Get("loading")
		w, h := img.Size()
		s.loadingCount %= w / h * loadingSpeed
		return
	}

	s.shopLabel.SetText(texts.Text(lang.Get(), texts.TextIDShop))

	s.updatePrices(prices, purchased)
	s.updateVisibilities()

	s.popup.Update()

	for id, p := range s.mainPanels {
		if id == ShopPopupTabInvalidID {
			continue
		}
		x, y := p.ScrollPosition()
		s.scrollPositions[id] = position{x, y}
	}
}

func (s *ShopPopup) updatePrices(prices map[string]string, purchased Purchased) {
	for key, b := range s.buyButtons {
		if purchased.IsPurchased(key) {
			b.SetText(texts.Text(lang.Get(), texts.TextIDPurchased))
			b.Disable()
			b.SetColor(color.White)
			continue
		}
		b.Enable()
		b.SetColor(yellow)
		if p, ok := prices[key]; ok {
			b.SetText(p)
			continue
		}
		b.SetText(texts.Text(lang.Get(), texts.TextIDBuy))
	}
}

func (s *ShopPopup) updateVisibilities() {
	// If s.tabID is invalid e.g., by updating the shop data, use the first tab.
	found := false
	for id := range s.mainPanels {
		if id == s.tabID {
			found = true
			break
		}
	}
	if !found {
		var ids []int
		for id := range s.mainPanels {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		s.tabID = ids[0]
	}

	for id, p := range s.mainPanels {
		if s.detailKey != "" {
			p.Hide()
			s.tabs[id].Hide()
			s.border.Hide()
			continue
		}

		if len(s.mainPanels) == 1 {
			p.Show()
			s.tabs[id].Hide()
			s.border.Hide()
			continue
		}

		s.tabs[id].Show()
		s.border.Show()
		if id == s.tabID {
			p.Show()
			s.tabs[id].Disable()
		} else {
			p.Hide()
			s.tabs[id].Enable()
		}
	}

	for k, p := range s.detailPanels {
		if k == s.detailKey {
			p.Show()
		} else {
			p.Hide()
		}
	}

	if s.detailKey != "" {
		s.closeButton.Hide()
		s.backButton.Show()
	} else {
		s.closeButton.Show()
		s.backButton.Hide()
	}
}

func (s *ShopPopup) Draw(screen *ebiten.Image) {
	s.popup.Draw(screen)

	if !s.loading {
		return
	}
	op := &ebiten.DrawImageOptions{}
	w, h := shadowImage.Size()
	sw, sh := screen.Size()
	op.GeoM.Scale(float64(sw)/float64(w), float64(sh)/float64(h))
	screen.DrawImage(shadowImage, op)

	loadingImage := assets.GetImage("system/common/loading.png")
	op = &ebiten.DrawImageOptions{}
	w, h = loadingImage.Size()
	op.GeoM.Scale(3, 3)
	op.GeoM.Translate(float64(sw-h*3)/2, float64(sh-h*3)/2)
	sx, sy := (s.loadingCount/loadingSpeed)*h, 0
	screen.DrawImage(loadingImage.SubImage(image.Rect(sx, sy, sx+h, sy+h)).(*ebiten.Image), op)
}

func shopImageNameFromKey(key string) string {
	return strings.ReplaceAll(key, ".", "_")
}
