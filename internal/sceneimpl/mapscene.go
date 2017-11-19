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

package sceneimpl

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type MapScene struct {
	gameState          *gamestate.Game
	moveDstX           int
	moveDstY           int
	tilesImage         *ebiten.Image
	triggeringFailed   bool
	initialState       bool
	cameraButton       *ui.Button
	cameraTaking       bool
	titleButton        *ui.Button
	screenShotImage    *ebiten.Image
	screenShotDialog   *ui.Dialog
	quitDialog         *ui.Dialog
	quitLabel          *ui.Label
	quitYesButton      *ui.Button
	quitNoButton       *ui.Button
	storeErrorDialog   *ui.Dialog
	storeErrorLabel    *ui.Label
	storeErrorOkButton *ui.Button
	removeAdsButton    *ui.Button
	removeAdsDialog    *ui.Dialog
	removeAdsLabel     *ui.Label
	removeAdsYesButton *ui.Button
	removeAdsNoButton  *ui.Button
	inventory          *ui.Inventory
	itemPreviewPopup   *ui.ItemPreviewPopup
	waitingRequestID   int
	isAdsRemoved       bool
	initialized        bool
	offsetX            float64
	offsetY            float64
}

func NewMapScene() *MapScene {
	m := &MapScene{
		gameState:    gamestate.NewGame(),
		initialState: true,
	}
	return m
}

func NewMapSceneWithGame(game *gamestate.Game) *MapScene {
	m := &MapScene{
		gameState: game,
	}
	return m
}

func (m *MapScene) initUI(sceneManager *scene.Manager) {
	const footerHeight = 33 * consts.TileScale

	screenW, screenH := sceneManager.Size()
	tilesImage, _ := ebiten.NewImage(screenW/consts.TileScale, screenH/consts.TileScale, ebiten.FilterNearest)
	m.tilesImage = tilesImage

	m.offsetX = (float64(screenW) - consts.TileXNum*consts.TileSize*consts.TileScale) / 2
	m.offsetY = float64(screenH) - consts.TileYNum*consts.TileSize*consts.TileScale - footerHeight

	screenShotImage, _ := ebiten.NewImage(480, 720, ebiten.FilterLinear)
	camera, _ := ebiten.NewImage(12, 12, ebiten.FilterNearest)
	camera.Fill(color.RGBA{0xff, 0, 0, 0xff})
	cameraImagePart := ui.NewImagePart(camera)
	m.cameraButton = ui.NewImageButton(0, 0, cameraImagePart, cameraImagePart, "click")
	m.screenShotImage = screenShotImage
	m.screenShotDialog = ui.NewDialog((screenW/consts.TileScale-160)/2+4, 4, 152, 232)
	m.screenShotDialog.AddChild(ui.NewImageView(8, 8, 1.0/consts.TileScale/2, ui.NewImagePart(m.screenShotImage)))
	m.titleButton = ui.NewButton(4+int(m.offsetX/consts.TileScale), 2, 40, 12, "click")

	// TODO: Implement the camera functionality later
	m.cameraButton.Visible = false

	m.quitDialog = ui.NewDialog((screenW/consts.TileScale-160)/2+4, 64, 152, 124)
	m.quitLabel = ui.NewLabel(16, 8)
	m.quitYesButton = ui.NewButton((152-120)/2, 72, 120, 20, "click")
	m.quitNoButton = ui.NewButton((152-120)/2, 96, 120, 20, "cancel")

	m.quitDialog.AddChild(m.quitLabel)
	m.quitDialog.AddChild(m.quitYesButton)
	m.quitDialog.AddChild(m.quitNoButton)

	m.storeErrorDialog = ui.NewDialog((screenW/consts.TileScale-160)/2+4, 64, 152, 124)
	m.storeErrorLabel = ui.NewLabel(16, 8)
	m.storeErrorOkButton = ui.NewButton((152-120)/2, 96, 120, 20, "click")
	m.storeErrorDialog.AddChild(m.storeErrorLabel)
	m.storeErrorDialog.AddChild(m.storeErrorOkButton)

	m.removeAdsButton = ui.NewButton(104+int(m.offsetX/consts.TileScale), 8, 52, 12, "click")
	m.removeAdsDialog = ui.NewDialog((screenW/consts.TileScale-160)/2+4, 64, 152, 124)
	m.removeAdsLabel = ui.NewLabel(16, 8)
	m.removeAdsYesButton = ui.NewButton((152-120)/2, 72, 120, 20, "click")
	m.removeAdsNoButton = ui.NewButton((152-120)/2, 96, 120, 20, "cancel")
	m.removeAdsDialog.AddChild(m.removeAdsLabel)
	m.removeAdsDialog.AddChild(m.removeAdsYesButton)
	m.removeAdsDialog.AddChild(m.removeAdsNoButton)

	m.inventory = ui.NewInventory(int(m.offsetX/consts.TileScale), (screenH-footerHeight)/consts.TileScale)
	m.itemPreviewPopup = ui.NewItemPreviewPopup((screenW/consts.TileScale-160)/2+16, int(m.offsetY/consts.TileScale))
	m.quitDialog.AddChild(m.quitLabel)

	m.removeAdsButton.Visible = false // TODO: Clock of Atonement does not need this feature, so turn it off for now
	m.initialized = true
}

func (m *MapScene) updatePurchasesState(sceneManager *scene.Manager) {
	m.isAdsRemoved = sceneManager.IsPurchased("ads_removal")
}

func (m *MapScene) runEventIfNeeded(sceneManager *scene.Manager) {
	if m.itemPreviewPopup.Visible() {
		m.triggeringFailed = false
		return
	}
	if m.gameState.Map().IsBlockingEventExecuting() {
		m.triggeringFailed = false
		return
	}
	if !input.Triggered() {
		return
	}
	x, y := input.Position()
	x -= int(m.offsetX)
	y -= int(m.offsetY)
	if x < 0 || y < 0 {
		return
	}
	tx := x / consts.TileSize / consts.TileScale
	ty := y / consts.TileSize / consts.TileScale
	if tx < 0 || consts.TileXNum <= tx || ty < 0 || consts.TileYNum <= ty {
		return
	}
	m.moveDstX = tx
	m.moveDstY = ty
	if m.gameState.Map().TryRunDirectEvent(m.gameState, tx, ty) {
		m.triggeringFailed = false
		return
	}
	if !m.gameState.Map().TryMovePlayerByUserInput(sceneManager, m.gameState, tx, ty) {
		m.triggeringFailed = true
		return
	}
	m.triggeringFailed = false
}

func (m *MapScene) receiveRequest(sceneManager *scene.Manager) bool {
	if m.waitingRequestID != 0 {
		r := sceneManager.ReceiveResultIfExists(m.waitingRequestID)
		if r == nil {
			return false
		}
		m.waitingRequestID = 0
		switch r.Type {
		case scene.RequestTypeIAPPrices:
			if !r.Succeeded {
				m.storeErrorDialog.Show()
				break
			}
			priceText := "???"
			var prices map[string]string
			if err := json.Unmarshal(r.Data, &prices); err != nil {
				panic(err)
			}
			text := texts.Text(sceneManager.Language(), texts.TextIDRemoveAdsDesc)
			if _, ok := prices["ads_removal"]; ok {
				priceText = prices["ads_removal"]
			}
			m.removeAdsLabel.Text = fmt.Sprintf(text, priceText)
			m.removeAdsDialog.Show()
		case scene.RequestTypePurchase:
			// Note: Ideally we should show a notification toast to notify users about the result
			// For now, the notifications are handled on the native platform side
			if r.Succeeded {
				m.updatePurchasesState(sceneManager)
			}
			m.removeAdsDialog.Hide()
		}
		return false
	}

	return true
}

func (m *MapScene) updateQuitDialog(sceneManager *scene.Manager) (bool, error) {
	m.quitLabel.Text = texts.Text(sceneManager.Language(), texts.TextIDBackToTitle)
	m.quitYesButton.Text = texts.Text(sceneManager.Language(), texts.TextIDYes)
	m.quitNoButton.Text = texts.Text(sceneManager.Language(), texts.TextIDNo)
	if m.quitDialog.Visible() {
		m.quitDialog.Update()
		if m.quitYesButton.Pressed() {
			if m.gameState.IsAutoSaveEnabled() {
				m.gameState.RequestSave(sceneManager)
			}
			if err := audio.Stop(); err != nil {
				return false, err
			}
			sceneManager.GoToWithFading(NewTitleScene(), 30)
			return false, nil
		}
		if m.quitNoButton.Pressed() {
			m.quitDialog.Hide()
			return false, nil
		}
		return false, nil
	}

	return true, nil
}

func (m *MapScene) updateInventory(sceneManager *scene.Manager) {
	if m.gameState.InventoryVisible() {
		m.inventory.Show()
	} else {
		m.inventory.Hide()
	}
	activeItemID := m.gameState.Items().ActiveItem()
	if m.inventory.Visible() {
		// TODO creating array for each loop does not seem to be the right thing
		items := []*data.Item{}
		for _, itemID := range m.gameState.Items().Items() {
			for _, item := range sceneManager.Game().Items {
				if itemID == item.ID {
					items = append(items, item)
					break
				}
			}
		}

		m.inventory.SetItems(items)
		if m.inventory.Mode() == ui.DefaultMode {
			m.inventory.SetActiveItemID(activeItemID)
		}
		if !m.gameState.Map().IsBlockingEventExecuting() {
			m.inventory.Update()
		}

		if m.inventory.PressedSlotIndex() >= 0 && m.inventory.PressedSlotIndex() < len(m.gameState.Items().Items()) {
			itemID := m.gameState.Items().Items()[m.inventory.PressedSlotIndex()]
			if m.inventory.Mode() == ui.DefaultMode {
				if itemID == activeItemID {
					m.gameState.Items().Deactivate()
				} else {
					m.gameState.Items().Activate(itemID)
				}
			} else {
				var combineItem *data.Item
				for _, item := range sceneManager.Game().Items {
					if m.inventory.CombineItemID() == item.ID {
						combineItem = item
						break
					}
				}

				m.itemPreviewPopup.SetCombineItem(combineItem, sceneManager.Game().CreateCombine(activeItemID, m.inventory.CombineItemID()))
			}
		}

		if eventItemID := m.gameState.Items().EventItem(); eventItemID > 0 {
			m.inventory.SetActiveItemID(eventItemID)
			m.inventory.SetMode(ui.PreviewMode)

			var eventItem *data.Item
			for _, item := range sceneManager.Game().Items {
				if item.ID == eventItemID {
					eventItem = item
					break
				}
			}

			m.itemPreviewPopup.SetActiveItem(eventItem, sceneManager.Game().Texts.Get(sceneManager.Language(), eventItem.Desc))
			m.itemPreviewPopup.Show()
		}

		if m.inventory.ActiveItemPressed() {
			m.gameState.Items().SetEventItem(activeItemID)
		}
	}

	if m.itemPreviewPopup.ClosePressed() || m.inventory.BackPressed() {
		m.gameState.Items().SetEventItem(0)
	}

	// TODO: using Pressed() in if statement requires us to call extra Update()... this is ugly
	if m.itemPreviewPopup.Visible() && m.gameState.Items().EventItem() == 0 {
		m.itemPreviewPopup.SetActiveItem(nil, "")
		m.itemPreviewPopup.Hide()
		m.itemPreviewPopup.Update(sceneManager)
		m.inventory.Update()
		m.inventory.SetMode(ui.DefaultMode)
	}

	if m.itemPreviewPopup.Visible() {
		if !m.gameState.ExecutingItemCommands() && !m.gameState.Map().IsBlockingEventExecuting() {
			// TODO: ItemPreviewPopup is not standarized as the other Popups
			m.itemPreviewPopup.Update(sceneManager)
			if m.itemPreviewPopup.ActionPressed() {
				if m.inventory.CombineItemID() != 0 {
					combine := sceneManager.Game().CreateCombine(activeItemID, m.inventory.CombineItemID())
					m.gameState.StartCombineCommands(combine)
				} else {
					m.gameState.StartItemCommands(activeItemID)
				}
			}
		}
	}
}

func (m *MapScene) updateStoreDialog(sceneManager *scene.Manager) bool {
	m.storeErrorLabel.Text = texts.Text(sceneManager.Language(), texts.TextIDStoreError)
	m.storeErrorOkButton.Text = texts.Text(sceneManager.Language(), texts.TextIDOK)
	if m.storeErrorDialog.Visible() {
		m.storeErrorDialog.Update()
		if m.storeErrorOkButton.Pressed() {
			m.storeErrorDialog.Hide()
			return false
		}
		return false
	}

	return true
}

func (m *MapScene) updateRemoveAdsDialog(sceneManager *scene.Manager) bool {
	m.removeAdsYesButton.Text = texts.Text(sceneManager.Language(), texts.TextIDYes)
	m.removeAdsNoButton.Text = texts.Text(sceneManager.Language(), texts.TextIDNo)
	if m.removeAdsDialog.Visible() {
		m.removeAdsDialog.Update()
		if m.removeAdsYesButton.Pressed() {
			m.waitingRequestID = sceneManager.GenerateRequestID()
			sceneManager.Requester().RequestPurchase(m.waitingRequestID, "ads_removal")
			return false
		}
		if m.removeAdsNoButton.Pressed() {
			m.removeAdsDialog.Hide()
			return false
		}
		if m.removeAdsDialog.Visible() {
			return false
		}
	}

	return true
}

func (m *MapScene) Update(sceneManager *scene.Manager) error {
	if !m.initialized {
		m.initUI(sceneManager)
	}
	m.updatePurchasesState(sceneManager)

	if ok := m.receiveRequest(sceneManager); !ok {
		return nil
	}

	if input.BackButtonPressed() {
		m.handleBackButton()
	}

	if ok, err := m.updateQuitDialog(sceneManager); !ok {
		return err
	}

	m.updateInventory(sceneManager)

	if ok := m.updateStoreDialog(sceneManager); !ok {
		return nil
	}

	if ok := m.updateRemoveAdsDialog(sceneManager); !ok {
		return nil
	}

	if m.initialState && m.gameState.IsAutoSaveEnabled() {
		m.gameState.RequestSave(sceneManager)
	}
	m.initialState = false
	m.screenShotDialog.Update()
	if m.screenShotDialog.Visible() {
		return nil
	}
	m.cameraButton.Update()

	if err := m.gameState.Update(sceneManager); err != nil {
		if err == gamestate.GoToTitle {
			return m.goToTitle(sceneManager)
		}
		return err
	}

	m.runEventIfNeeded(sceneManager)
	if m.cameraButton.Pressed() {
		m.cameraTaking = true
		m.screenShotDialog.Show()
	}

	m.titleButton.Text = texts.Text(sceneManager.Language(), texts.TextIDTitle)
	m.titleButton.Disabled = m.gameState.Map().IsBlockingEventExecuting()
	m.titleButton.Update()
	if m.titleButton.Pressed() {
		m.quitDialog.Show()
	}

	m.removeAdsButton.Text = texts.Text(sceneManager.Language(), texts.TextIDRemoveAds)
	m.removeAdsButton.Disabled = m.gameState.Map().IsBlockingEventExecuting()
	m.removeAdsButton.Update()
	if m.removeAdsButton.Pressed() {
		m.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestGetIAPPrices(m.waitingRequestID)
	}
	return nil
}

func (m *MapScene) goToTitle(sceneManager *scene.Manager) error {
	if err := audio.Stop(); err != nil {
		return err
	}
	sceneManager.GoToWithFading(NewTitleScene(), 60)
	return nil
}

func (m *MapScene) handleBackButton() {
	if m.storeErrorDialog.Visible() {
		audio.PlaySE("cancel", 1.0)
		m.storeErrorDialog.Hide()
		return
	}

	if m.quitDialog.Visible() {
		audio.PlaySE("cancel", 1.0)
		m.quitDialog.Hide()
		return
	}
	if m.quitDialog.Visible() {
		audio.PlaySE("cancel", 1.0)
		m.quitDialog.Hide()
		return
	}

	audio.PlaySE("click", 1.0)
	m.quitDialog.Show()
}

func (m *MapScene) Draw(screen *ebiten.Image) {
	if !m.initialized {
		return
	}
	m.tilesImage.Fill(color.Black)

	// TODO: This accesses *data.Game, but is it OK?
	room := m.gameState.Map().CurrentRoom()

	if room.Background.Name != "" {
		m.gameState.Map().DrawFullscreenImage(m.tilesImage, assets.GetImage("backgrounds/"+room.Background.Name+".png"), 0, m.offsetY/consts.TileScale)
	}
	op := &ebiten.DrawImageOptions{}
	for k := 0; k < 3; k++ {
		layer := 0
		if k >= 1 {
			layer = 1
		}
		tileSet := m.gameState.Map().TileSet(layer)
		if tileSet != nil {
			tileSetImg := assets.GetImage("tilesets/" + tileSet.Name + ".png")
			for j := 0; j < consts.TileYNum; j++ {
				for i := 0; i < consts.TileXNum; i++ {
					tile := room.Tiles[layer][j*consts.TileXNum+i]
					if layer == 1 {
						p := tileSet.PassageTypes[tile]
						if k == 1 && p == data.PassageTypeOver {
							continue
						}
						if k == 2 && p != data.PassageTypeOver {
							continue
						}
					}
					sx := tile % consts.PaletteWidth * consts.TileSize
					sy := tile / consts.PaletteWidth * consts.TileSize
					r := image.Rect(sx, sy, sx+consts.TileSize, sy+consts.TileSize)
					op.SourceRect = &r
					dx := i * consts.TileSize
					dy := j*consts.TileSize + int(m.offsetY/consts.TileScale)
					op.GeoM.Reset()
					op.GeoM.Translate(float64(dx), float64(dy))
					m.tilesImage.DrawImage(tileSetImg, op)
				}
			}
		}
		var p data.Priority
		switch k {
		case 0:
			p = data.PriorityBottom
		case 1:
			p = data.PriorityMiddle
		case 2:
			p = data.PriorityTop
		default:
			panic("not reached")
		}
		m.gameState.Map().DrawCharacters(m.tilesImage, p, 0, m.offsetY/consts.TileScale)
	}
	if room.Foreground.Name != "" {
		m.gameState.Map().DrawFullscreenImage(m.tilesImage, assets.GetImage("foregrounds/"+room.Foreground.Name+".png"), 0, m.offsetY/consts.TileScale)
	}

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	op.GeoM.Translate(m.offsetX, 0)
	m.gameState.DrawScreen(screen, m.tilesImage, op)
	m.gameState.DrawPictures(screen, m.offsetX, m.offsetY)

	if m.gameState.IsPlayerControlEnabled() && (m.gameState.Map().IsPlayerMovingByUserInput() || m.triggeringFailed) {
		x, y := m.moveDstX, m.moveDstY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*consts.TileSize), float64(y*consts.TileSize))
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(m.offsetX, m.offsetY)
		screen.DrawImage(assets.GetImage("system/marker.png"), op)
	}

	m.itemPreviewPopup.Draw(screen)
	m.inventory.Draw(screen)
	m.gameState.DrawWindows(screen, 0, m.offsetY)

	if m.cameraTaking {
		m.cameraTaking = false
		m.screenShotImage.Clear()
		op := &ebiten.DrawImageOptions{}
		sw, _ := screen.Size()
		w, _ := m.screenShotImage.Size()
		op.GeoM.Translate((float64(w)-float64(sw))/2, 0)
		m.screenShotImage.DrawImage(screen, nil)
	}

	m.cameraButton.Draw(screen)
	m.titleButton.Draw(screen)
	m.removeAdsButton.Draw(screen)

	m.screenShotDialog.Draw(screen)
	m.quitDialog.Draw(screen)
	m.storeErrorDialog.Draw(screen)
	m.removeAdsDialog.Draw(screen)

	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	font.DrawText(screen, msg, 160, 8, consts.TextScale, data.TextAlignLeft, color.White)
}
