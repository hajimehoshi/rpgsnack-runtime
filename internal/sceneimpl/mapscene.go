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
}

func NewMapScene() *MapScene {
	tilesImage, _ := ebiten.NewImage(consts.TileXNum*consts.TileSize, (consts.TileYNum+consts.InventoryOffset)*consts.TileSize, ebiten.FilterNearest)
	m := &MapScene{
		tilesImage:   tilesImage,
		gameState:    gamestate.NewGame(),
		initialState: true,
	}
	m.initUI()
	return m
}

func NewMapSceneWithGame(game *gamestate.Game) *MapScene {
	tilesImage, _ := ebiten.NewImage(consts.TileXNum*consts.TileSize, (consts.TileYNum+consts.InventoryOffset)*consts.TileSize, ebiten.FilterNearest)
	m := &MapScene{
		tilesImage: tilesImage,
		gameState:  game,
	}
	m.initUI()
	return m
}

func (m *MapScene) offsetX(screenWidth int) float64 {
	return (float64(screenWidth) - consts.TileXNum*consts.TileSize*consts.TileScale) / 2
}

func (m *MapScene) initUI() {
	screenShotImage, _ := ebiten.NewImage(480, 720, ebiten.FilterLinear)
	camera, _ := ebiten.NewImage(12, 12, ebiten.FilterNearest)
	camera.Fill(color.RGBA{0xff, 0, 0, 0xff})
	m.cameraButton = ui.NewImageButton(0, 0, camera, "click")
	m.screenShotImage = screenShotImage
	m.screenShotDialog = ui.NewDialog(0, 4, 152, 232)
	m.screenShotDialog.AddChild(ui.NewImage(8, 8, 1.0/consts.TileScale/2, m.screenShotImage))
	m.titleButton = ui.NewButton(0, 2, 40, 12, "click")

	// TODO: Implement the camera functionality later
	m.cameraButton.Visible = false

	m.quitDialog = ui.NewDialog(0, 64, 152, 124)
	m.quitLabel = ui.NewLabel(16, 8)
	m.quitYesButton = ui.NewButton(0, 72, 120, 20, "click")
	m.quitNoButton = ui.NewButton(0, 96, 120, 20, "cancel")
	m.quitDialog.AddChild(m.quitLabel)
	m.quitDialog.AddChild(m.quitYesButton)
	m.quitDialog.AddChild(m.quitNoButton)

	m.storeErrorDialog = ui.NewDialog(0, 64, 152, 124)
	m.storeErrorLabel = ui.NewLabel(16, 8)
	m.storeErrorOkButton = ui.NewButton(0, 96, 120, 20, "click")
	m.storeErrorDialog.AddChild(m.storeErrorLabel)
	m.storeErrorDialog.AddChild(m.storeErrorOkButton)

	m.removeAdsButton = ui.NewButton(0, 8, 52, 12, "click")
	m.removeAdsDialog = ui.NewDialog(0, 64, 152, 124)
	m.removeAdsLabel = ui.NewLabel(16, 8)
	m.removeAdsYesButton = ui.NewButton(0, 72, 120, 20, "click")
	m.removeAdsNoButton = ui.NewButton(0, 96, 120, 20, "cancel")
	m.removeAdsDialog.AddChild(m.removeAdsLabel)
	m.removeAdsDialog.AddChild(m.removeAdsYesButton)
	m.removeAdsDialog.AddChild(m.removeAdsNoButton)
	m.inventory = ui.NewInventory(0, consts.TileYNum*consts.TileSize)
	m.itemPreviewPopup = ui.NewItemPreviewPopup(32, 32, 256, 256)
	m.quitDialog.AddChild(m.quitLabel)

	m.removeAdsButton.Visible = false // TODO: Clock of Atonement does not need this feature, so turn it off for now
}

func (m *MapScene) UpdatePurchasesState(sceneManager *scene.Manager) {
	m.isAdsRemoved = sceneManager.IsPurchased("ads_removal")
}

func (m *MapScene) runEventIfNeeded(sceneManager *scene.Manager) error {
	if m.gameState.Map().IsEventExecuting() {
		m.triggeringFailed = false
		return nil
	}
	if !input.Triggered() {
		return nil
	}
	x, y := input.Position()
	x -= sceneManager.MapOffsetX()
	y -= consts.GameMarginTop
	if x < 0 || y < 0 {
		return nil
	}
	tx := x / consts.TileSize / consts.TileScale
	ty := y / consts.TileSize / consts.TileScale
	if tx < 0 || consts.TileXNum <= tx || ty < 0 || consts.TileYNum <= ty {
		return nil
	}
	m.moveDstX = tx
	m.moveDstY = ty
	if m.gameState.Map().TryRunDirectEvent(tx, ty) {
		m.triggeringFailed = false
		return nil
	}
	if !m.gameState.Map().TryMovePlayerByUserInput(sceneManager, tx, ty) {
		m.triggeringFailed = true
		return nil
	}
	m.triggeringFailed = false
	return nil
}

func (m *MapScene) Update(sceneManager *scene.Manager) error {
	m.UpdatePurchasesState(sceneManager)
	if m.waitingRequestID != 0 {
		r := sceneManager.ReceiveResultIfExists(m.waitingRequestID)
		if r == nil {
			return nil
		}
		m.waitingRequestID = 0
		switch r.Type {
		case scene.RequestTypeIAPPrices:
			if !r.Succeeded {
				m.storeErrorDialog.Visible = true
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
			m.removeAdsDialog.Visible = true
		case scene.RequestTypePurchase:
			// Note: Ideally we should show a notification toast to notify users about the result
			// For now, the notifications are handled on the native platform side
			if r.Succeeded {
				m.UpdatePurchasesState(sceneManager)
			}
			m.removeAdsDialog.Visible = false
		}
		return nil
	}

	w, _ := sceneManager.Size()

	if input.BackButtonPressed() {
		m.handleBackButton()
	}

	m.quitLabel.Text = texts.Text(sceneManager.Language(), texts.TextIDBackToTitle)
	m.quitYesButton.Text = texts.Text(sceneManager.Language(), texts.TextIDYes)
	m.quitNoButton.Text = texts.Text(sceneManager.Language(), texts.TextIDNo)
	m.quitDialog.X = (w/consts.TileScale-160)/2 + 4
	m.quitYesButton.X = (m.quitDialog.Width - m.quitYesButton.Width) / 2
	m.quitNoButton.X = (m.quitDialog.Width - m.quitNoButton.Width) / 2
	if m.quitDialog.Visible {
		m.quitDialog.Update()
		if m.quitYesButton.Pressed() {
			if m.gameState.IsAutoSaveEnabled() {
				m.gameState.RequestSave(sceneManager)
			}
			if err := audio.Stop(); err != nil {
				return err
			}
			sceneManager.GoToWithFading(NewTitleScene(), 30)
			return nil
		}
		if m.quitNoButton.Pressed() {
			m.quitDialog.Visible = false
			return nil
		}
		return nil
	}

	if m.inventory.Visible {
		// TODO creating array for each loop does not seem to be the right thing
		items := []*data.Item{}
		for _, itemID := range m.gameState.Items().Items() {
			for _, item := range sceneManager.Game().Items {
				if itemID == item.ID {
					items = append(items, item)
				}
			}
		}
		m.inventory.SetItems(items)
		m.inventory.SetActiveItemID(m.gameState.Items().ActiveItem())
		m.inventory.Update()
		if m.inventory.PressedSlotIndex >= 0 && m.inventory.PressedSlotIndex < len(m.gameState.Items().Items()) {
			itemID := m.gameState.Items().Items()[m.inventory.PressedSlotIndex]
			if itemID == m.gameState.Items().ActiveItem() {
				if !m.itemPreviewPopup.Visible {
					for _, item := range sceneManager.Game().Items {
						if itemID == item.ID {
							m.itemPreviewPopup.SetItem(item)
							break
						}
					}
				}

				m.itemPreviewPopup.Visible = true
				m.gameState.Items().Deactivate()
			} else {
				m.gameState.Items().Activate(itemID)
			}
		}
	}

	m.itemPreviewPopup.X = (w/consts.TileScale-160)/2 + 16
	if m.itemPreviewPopup.Visible {
		m.itemPreviewPopup.Update()
		if m.itemPreviewPopup.PreviewPressed() {
			m.gameState.StartItemCommands()
		}
		if err := m.gameState.UpdateItemCommandsIfNeeded(sceneManager); err != nil {
			return err
		}

		// TODO: This is copied from above. Integrate this.
		if err := m.gameState.Screen().Update(); err != nil {
			return err
		}
		m.gameState.Windows().Update(sceneManager)
		return nil
	}

	m.storeErrorOkButton.X = (m.storeErrorDialog.Width - m.storeErrorOkButton.Width) / 2
	m.storeErrorLabel.Text = texts.Text(sceneManager.Language(), texts.TextIDStoreError)
	m.storeErrorOkButton.Text = texts.Text(sceneManager.Language(), texts.TextIDOK)
	m.storeErrorDialog.X = (w/consts.TileScale-160)/2 + 4
	if m.storeErrorDialog.Visible {
		m.storeErrorDialog.Update()
		if m.storeErrorOkButton.Pressed() {
			m.storeErrorDialog.Visible = false
			return nil
		}
		return nil
	}

	m.removeAdsYesButton.Text = texts.Text(sceneManager.Language(), texts.TextIDYes)
	m.removeAdsNoButton.Text = texts.Text(sceneManager.Language(), texts.TextIDNo)
	m.removeAdsDialog.X = (w/consts.TileScale-160)/2 + 4
	m.removeAdsYesButton.X = (m.removeAdsDialog.Width - m.removeAdsYesButton.Width) / 2
	m.removeAdsNoButton.X = (m.removeAdsDialog.Width - m.removeAdsNoButton.Width) / 2
	if m.removeAdsDialog.Visible {
		m.removeAdsDialog.Update()
		if m.removeAdsYesButton.Pressed() {
			m.waitingRequestID = sceneManager.GenerateRequestID()
			sceneManager.Requester().RequestPurchase(m.waitingRequestID, "ads_removal")
			return nil
		}
		if m.removeAdsNoButton.Pressed() {
			m.removeAdsDialog.Visible = false
			return nil
		}
		if m.removeAdsDialog.Visible {
			return nil
		}
	}
	// m.removeAdsButton.Visible = !m.isAdsRemoved

	// TODO: All UI parts' origin should be defined correctly
	// so that we don't need to adjust X positions here.
	m.titleButton.X = 4 + int(m.offsetX(w)/consts.TileScale)
	m.removeAdsButton.X = 104 + int(m.offsetX(w)/consts.TileScale)

	m.screenShotDialog.X = (w/consts.TileScale-160)/2 + 4
	if m.initialState && m.gameState.IsAutoSaveEnabled() {
		m.gameState.RequestSave(sceneManager)
	}
	m.initialState = false
	m.screenShotDialog.Update()
	if m.screenShotDialog.Visible {
		return nil
	}
	m.cameraButton.Update()

	if err := m.gameState.Update(sceneManager); err != nil {
		return err
	}
	if err := m.gameState.Screen().Update(); err != nil {
		return err
	}
	m.gameState.Windows().Update(sceneManager)
	if err := m.gameState.Map().Update(sceneManager); err != nil {
		if err == gamestate.GoToTitle {
			if err := audio.Stop(); err != nil {
				return err
			}
			sceneManager.GoToWithFading(NewTitleScene(), 60)
			return nil
		}
		return err
	}
	if err := m.runEventIfNeeded(sceneManager); err != nil {
		return err
	}
	if m.cameraButton.Pressed() {
		m.cameraTaking = true
		m.screenShotDialog.Visible = true
	}

	m.titleButton.Text = texts.Text(sceneManager.Language(), texts.TextIDTitle)
	m.titleButton.Disabled = m.gameState.Map().IsEventExecuting()
	m.titleButton.Update()
	if m.titleButton.Pressed() {
		m.quitDialog.Visible = true
	}

	m.removeAdsButton.Text = texts.Text(sceneManager.Language(), texts.TextIDRemoveAds)
	m.removeAdsButton.Disabled = m.gameState.Map().IsEventExecuting()
	m.removeAdsButton.Update()
	if m.removeAdsButton.Pressed() {
		m.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestGetIAPPrices(m.waitingRequestID)
	}
	return nil
}

func (m *MapScene) handleBackButton() {
	if m.storeErrorDialog.Visible {
		audio.PlaySE("cancel", 1.0)
		m.storeErrorDialog.Visible = false
		return
	}

	if m.quitDialog.Visible {
		audio.PlaySE("cancel", 1.0)
		m.quitDialog.Visible = false
		return
	}
	if m.quitDialog.Visible {
		audio.PlaySE("cancel", 1.0)
		m.quitDialog.Visible = false
		return
	}

	audio.PlaySE("click", 1.0)
	m.quitDialog.Visible = true
}

func (m *MapScene) Draw(screen *ebiten.Image) {
	m.tilesImage.Fill(color.Black)

	m.cameraButton.Draw(screen)
	m.titleButton.Draw(screen)
	m.removeAdsButton.Draw(screen)

	// TODO: This accesses *data.Game, but is it OK?
	room := m.gameState.Map().CurrentRoom()

	if room.Background.Name != "" {
		op := &ebiten.DrawImageOptions{}
		m.tilesImage.DrawImage(assets.GetImage("backgrounds/"+room.Background.Name+".png"), op)
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
					dy := j * consts.TileSize
					op.GeoM.Reset()
					op.GeoM.Translate(float64(dx), float64(dy))
					m.tilesImage.DrawImage(tileSetImg, op)
				}
			}
		}
		if k == 1 {
			m.gameState.Map().DrawCharacters(m.tilesImage)
		}
	}
	if room.Foreground.Name != "" {
		op := &ebiten.DrawImageOptions{}
		m.tilesImage.DrawImage(assets.GetImage("foregrounds/"+room.Foreground.Name+".png"), op)
	}

	m.inventory.Draw(m.tilesImage)

	sw, _ := screen.Size()
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	op.GeoM.Translate(m.offsetX(sw), consts.GameMarginTop)
	m.gameState.Screen().Draw(screen, m.tilesImage, op)

	if m.gameState.IsPlayerControlEnabled() && (m.gameState.Map().IsPlayerMovingByUserInput() || m.triggeringFailed) {
		x, y := m.moveDstX, m.moveDstY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*consts.TileSize), float64(y*consts.TileSize))
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(m.offsetX(sw), consts.GameMarginTop)
		screen.DrawImage(assets.GetImage("system/marker.png"), op)
	}

	if m.cameraTaking {
		m.cameraTaking = false
		m.screenShotImage.Clear()
		op := &ebiten.DrawImageOptions{}
		sw, _ := screen.Size()
		w, _ := m.screenShotImage.Size()
		op.GeoM.Translate((float64(w)-float64(sw))/2, 0)
		m.screenShotImage.DrawImage(screen, nil)
	}

	m.screenShotDialog.Draw(screen)
	m.quitDialog.Draw(screen)
	m.storeErrorDialog.Draw(screen)
	m.removeAdsDialog.Draw(screen)
	m.itemPreviewPopup.Draw(screen)

	m.gameState.DrawWindows(screen)

	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	font.DrawText(screen, msg, 160, 8, consts.TextScale, data.TextAlignLeft, color.White)
}
