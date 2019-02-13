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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/tileset"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

const (
	markerAnimationInterval = 4
	markerSize              = 16
	itemPreviewPopupMargin  = 80
)

type MapScene struct {
	gameState            *gamestate.Game
	moveDstX             int
	moveDstY             int
	screenImage          *ebiten.Image
	tintScreenImage      *ebiten.Image
	triggeringFailed     bool
	initialState         bool
	cameraButton         *ui.Button
	cameraTaking         bool
	gameHeader           *ui.GameHeader
	screenShotImage      *ebiten.Image
	screenShotDialog     *ui.Dialog
	quitDialog           *ui.Dialog
	quitLabel            *ui.Label
	quitYesButton        *ui.Button
	quitNoButton         *ui.Button
	storeErrorDialog     *ui.Dialog
	storeErrorLabel      *ui.Label
	storeErrorOkButton   *ui.Button
	removeAdsButton      *ui.Button
	removeAdsDialog      *ui.Dialog
	removeAdsLabel       *ui.Label
	removeAdsYesButton   *ui.Button
	removeAdsNoButton    *ui.Button
	inventory            *ui.Inventory
	itemPreviewPopup     *ui.ItemPreviewPopup
	minigamePopup        *ui.MinigamePopup
	titleView            *ui.TitleView
	credits              *ui.Credits
	markerAnimationFrame int
	waitingRequestID     int
	isAdsRemoved         bool
	initialized          bool
	offsetY              int
	windowOffsetY        int
	inventoryHeight      int
	animation            animation
	onRewarded           func(*scene.Manager)

	err error
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

type sceneMaker struct{}

func (s *sceneMaker) NewMapScene() scene.Scene {
	return NewMapScene()
}

func (s *sceneMaker) NewMapSceneWithGame(game *gamestate.Game) scene.Scene {
	return NewMapSceneWithGame(game)
}

func (s *sceneMaker) NewSettingsScene() scene.Scene {
	return NewSettingsScene()
}

func NewTitleMapScene(savedGame *gamestate.Game) *MapScene {
	m := &MapScene{
		titleView: ui.NewTitleView(&sceneMaker{}),
	}
	m.gameState = gamestate.NewTitleGame(savedGame, m.shakeStartGameButton)
	return m
}

func (m *MapScene) updateOffsetY(sceneManager *scene.Manager) {
	_, sh := sceneManager.Size()

	// In case the device is super large (iPhoneX),
	// we do not do any of the layout work here
	// as we are always going to show the fullscreen
	if sh >= consts.SuperLargeScreenHeight {
		m.offsetY = 0
		m.windowOffsetY = sceneManager.BottomOffset()
		return
	}
	bottomOffset := consts.TileSize * consts.TileScale

	switch m.gameState.Map().CurrentRoom().LayoutMode {
	case data.RoomLayoutModeFixBottom:
		m.offsetY = sh - consts.MapScaledHeight + bottomOffset

	case data.RoomLayoutModeFixCenter:
		m.offsetY = (sh - consts.MapScaledHeight) / 2
		// Adjust the screen so that the bottom snaps to the grid
		m.offsetY -= m.offsetY % (consts.TileSize * consts.TileScale)

	case data.RoomLayoutModeScroll:
		character := m.gameState.Map().FocusingCharacter()
		// character can be nil for the very first Update() loop
		if character == nil {
			return
		}
		_, y := character.DrawPosition()
		t := -y*consts.TileScale + sh/2

		if t > 0 {
			t = 0
		}

		if t < sh-consts.MapScaledHeight+bottomOffset {
			t = sh - consts.MapScaledHeight + bottomOffset
		}
		m.offsetY = t

	default:
		panic(fmt.Sprintf("invalid layout mode: %s", m.gameState.Map().CurrentRoom().LayoutMode))
	}
	m.windowOffsetY = 0
}

func (m *MapScene) initUI(sceneManager *scene.Manager) {
	const (
		uiWidth = consts.MapWidth
	)

	if sceneManager.HasExtraBottomGrid() {
		m.inventoryHeight = 49 * consts.TileScale
	} else {
		m.inventoryHeight = (49 - consts.TileSize) * consts.TileScale
	}
	_, screenH := sceneManager.Size()
	m.offsetY = 0

	m.screenImage, _ = ebiten.NewImage(consts.MapWidth, consts.CeilDiv(screenH, consts.TileScale), ebiten.FilterNearest)
	m.tintScreenImage, _ = ebiten.NewImage(consts.MapWidth, consts.CeilDiv(screenH, consts.TileScale), ebiten.FilterDefault)

	screenShotImage, _ := ebiten.NewImage(480, 720, ebiten.FilterLinear)
	camera, _ := ebiten.NewImage(12, 12, ebiten.FilterNearest)
	camera.Fill(color.RGBA{0xff, 0, 0, 0xff})
	m.cameraButton = ui.NewImageButton(0, 0, camera, camera, "system/click")
	m.screenShotImage = screenShotImage
	m.screenShotDialog = ui.NewDialog((uiWidth-160)/2+4, 4, 152, 232)
	m.screenShotDialog.AddChild(ui.NewImageView(8, 8, 1.0/consts.TileScale/2, m.screenShotImage))

	if m.titleView == nil {
		m.gameHeader = ui.NewGameHeader()
	}

	// TODO: Implement the camera functionality later
	m.cameraButton.Hide()

	m.quitDialog = ui.NewDialog((uiWidth-160)/2+4, screenH/(2*consts.TileScale)-64, 152, 124)
	m.quitLabel = ui.NewLabel(16, 8)
	m.quitYesButton = ui.NewButton((152-120)/2, 72, 120, 20, "")
	m.quitNoButton = ui.NewButton((152-120)/2, 96, 120, 20, "system/cancel")

	m.quitDialog.AddChild(m.quitLabel)
	m.quitDialog.AddChild(m.quitYesButton)
	m.quitDialog.AddChild(m.quitNoButton)

	m.storeErrorDialog = ui.NewDialog((uiWidth-160)/2+4, 64, 152, 124)
	m.storeErrorLabel = ui.NewLabel(16, 8)
	m.storeErrorOkButton = ui.NewButton((152-120)/2, 96, 120, 20, "system/click")
	m.storeErrorDialog.AddChild(m.storeErrorLabel)
	m.storeErrorDialog.AddChild(m.storeErrorOkButton)

	m.removeAdsButton = ui.NewButton(104, 8, 52, 12, "system/click")
	m.removeAdsDialog = ui.NewDialog((uiWidth-160)/2+4, 64, 152, 124)
	m.removeAdsLabel = ui.NewLabel(16, 8)
	m.removeAdsYesButton = ui.NewButton((152-120)/2, 72, 120, 20, "system/click")
	m.removeAdsNoButton = ui.NewButton((152-120)/2, 96, 120, 20, "system/cancel")
	m.removeAdsDialog.AddChild(m.removeAdsLabel)
	m.removeAdsDialog.AddChild(m.removeAdsYesButton)
	m.removeAdsDialog.AddChild(m.removeAdsNoButton)

	m.inventory = ui.NewInventory(0, consts.CeilDiv(screenH-m.inventoryHeight, consts.TileScale), sceneManager.HasExtraBottomGrid())
	ty := consts.CeilDiv(screenH, consts.TileScale) - m.inventoryHeight - itemPreviewPopupMargin
	m.itemPreviewPopup = ui.NewItemPreviewPopup(ty)
	m.minigamePopup = ui.NewMinigamePopup(ty)
	m.quitDialog.AddChild(m.quitLabel)

	m.removeAdsButton.Hide() // TODO: Clock of Atonement does not need this feature, so turn it off for now

	m.credits = ui.NewCredits(false)

	m.quitYesButton.SetOnPressed(func(_ *ui.Button) {
		if m.gameState.IsAutoSaveEnabled() && !m.gameState.Map().IsBlockingEventExecuting() {
			m.gameState.RequestSave(0, sceneManager)
		}
		audio.Stop()
		g, err := savedGame(sceneManager)
		if err != nil {
			m.err = err
			return
		}
		sceneManager.GoToWithFading(NewTitleMapScene(g), FadingCount, FadingCount)
	})
	m.quitNoButton.SetOnPressed(func(_ *ui.Button) {
		m.quitDialog.Hide()
	})
	if m.gameHeader != nil {
		m.gameHeader.SetOnTitlePressed(func() {
			m.quitDialog.Show()
		})
	}
	m.storeErrorOkButton.SetOnPressed(func(_ *ui.Button) {
		m.storeErrorDialog.Hide()
	})
	m.removeAdsButton.SetOnPressed(func(_ *ui.Button) {
		m.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestGetIAPPrices(m.waitingRequestID)
	})
	m.removeAdsYesButton.SetOnPressed(func(_ *ui.Button) {
		m.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestPurchase(m.waitingRequestID, "ads_removal")
	})
	m.removeAdsNoButton.SetOnPressed(func(_ *ui.Button) {
		m.removeAdsDialog.Hide()
	})
	m.cameraButton.SetOnPressed(func(_ *ui.Button) {
		m.cameraTaking = true
		m.screenShotDialog.Show()
	})

	m.inventory.SetOnSlotPressed(func(_ *ui.Inventory, index int) {
		if index >= m.gameState.Items().ItemNum() {
			return
		}
		activeItemID := m.gameState.Items().ActiveItem()
		itemID := m.gameState.Items().ItemIDAt(index)
		switch mode := m.inventory.Mode(); mode {
		case ui.DefaultMode:
			if itemID == m.gameState.Items().ActiveItem() {
				m.gameState.Items().Deactivate()
			} else {
				m.gameState.Items().Activate(itemID)
			}
		case ui.PreviewMode:
			if sceneManager.Game().IsCombineAvailable() {
				combineItemID := 0
				if m.gameState.Items().ActiveItem() > 0 {
					if activeItemID != itemID {
						combineItemID = itemID
					}
					m.gameState.Items().SetCombineItem(combineItemID)
				}
				var item *data.Item
				for _, i := range sceneManager.Game().Items {
					if i.ID == itemID {
						item = i
						break
					}
				}
				if combineItemID != 0 {
					c := sceneManager.Game().CreateCombine(activeItemID, item.ID)
					m.itemPreviewPopup.SetCombineItem(item, c)
				} else {
					m.itemPreviewPopup.SetCombineItem(nil, nil)
				}
			} else {
				m.gameState.Items().Activate(itemID)
				m.gameState.Items().SetEventItem(itemID)
				m.inventory.SetActiveItemID(itemID)
			}
		default:
			panic(fmt.Sprintf("sceneimpl: invalid inventory mode: %d", mode))
		}
	})
	m.inventory.SetOnActiveItemPressed(func(_ *ui.Inventory) {
		m.gameState.Items().SetEventItem(m.gameState.Items().ActiveItem())
	})
	m.itemPreviewPopup.SetOnClosePressed(func(_ *ui.ItemPreviewPopup) {
		m.gameState.Items().SetEventItem(0)
		m.gameState.Items().SetCombineItem(0)
	})
	m.minigamePopup.SetOnClose(func() {
		mg := m.gameState.Minigame()
		m.gameState.RequestSavePermanentMinigame(0, sceneManager, mg.ID(), mg.Score(), mg.LastActiveAt())
		m.gameState.HideMinigame()
	})
	m.minigamePopup.SetOnProgress(func(progress int) {
		mg := m.gameState.Minigame()
		sceneManager.Requester().RequestSendAnalytics(fmt.Sprintf("minigame%d_progress_%d", mg.ID(), progress), "")
	})
	m.minigamePopup.SetOnSave(func() {
		mg := m.gameState.Minigame()
		m.gameState.RequestSavePermanentMinigame(0, sceneManager, mg.ID(), mg.Score(), mg.LastActiveAt())
	})
	m.minigamePopup.SetOnRequestRewardeAds(func() {
		m.waitingRequestID = sceneManager.GenerateRequestID()
		m.gameState.RequestRewardedAds(m.waitingRequestID, sceneManager, true)
		m.onRewarded = m.minigameRewarded
	})
	m.inventory.SetOnBackPressed(func(_ *ui.Inventory) {
		m.gameState.Items().SetEventItem(0)
		m.gameState.Items().SetCombineItem(0)
	})
	// TODO: ItemPreviewPopup is not standarized as the other Popups
	m.itemPreviewPopup.SetOnActionPressed(func(_ *ui.ItemPreviewPopup) {
		if m.gameState.Map().IsBlockingEventExecuting() {
			return
		}
		activeItemID := m.gameState.Items().ActiveItem()
		combineItemID := m.gameState.Items().CombineItem()
		if combineItemID != 0 {
			combine := sceneManager.Game().CreateCombine(activeItemID, combineItemID)
			m.gameState.StartCombineCommands(combine)
		} else {
			m.gameState.StartItemCommands(activeItemID)
		}
	})
}
func (m *MapScene) minigameRewarded(sceneManager *scene.Manager) {
	mg := m.gameState.Minigame()
	sceneManager.Requester().RequestSendAnalytics(fmt.Sprintf("minigame%d_reward", mg.ID()), "")
	m.minigamePopup.ActivateBoostMode()
}

func (m *MapScene) updateItemPreviewPopupVisibility(sceneManager *scene.Manager) {
	itemID := m.gameState.Items().EventItem()
	if itemID > 0 {
		m.inventory.SetActiveItemID(itemID)
		m.inventory.SetCombineItemID(m.gameState.Items().CombineItem())
		m.inventory.SetMode(ui.PreviewMode)
		var eventItem *data.Item
		for _, item := range sceneManager.Game().Items {
			if item.ID == itemID {
				eventItem = item
				break
			}
		}

		m.itemPreviewPopup.SetActiveItem(eventItem)
		m.itemPreviewPopup.Show()
	} else {
		m.itemPreviewPopup.SetActiveItem(nil)
		m.itemPreviewPopup.Hide()
		m.inventory.SetMode(ui.DefaultMode)
	}
}

func (m *MapScene) updatePurchasesState(sceneManager *scene.Manager) {
	m.isAdsRemoved = sceneManager.IsPurchased("ads_removal")
}

func (m *MapScene) runEventIfNeeded(sceneManager *scene.Manager) {
	if m.itemPreviewPopup.Visible() {
		m.triggeringFailed = false
		return
	}

	if m.titleView != nil {
		m.triggeringFailed = false
		return
	}

	x, y := input.Position()
	if y < ui.HeaderTouchAreaHeight {
		return
	}

	if _, sh := sceneManager.Size(); m.gameState.InventoryVisible() && y > sh-m.inventoryHeight {
		return
	}

	y -= m.offsetY
	if x < 0 || y < 0 {
		return
	}
	tx := x / consts.TileSize / consts.TileScale
	ty := y / consts.TileSize / consts.TileScale

	if input.Pressed() {
		m.gameState.Map().SetPressedPosition(tx, ty)
	}

	if m.gameState.Map().IsBlockingEventExecuting() {
		m.triggeringFailed = false
		return
	}

	if !input.Triggered() {
		return
	}

	m.markerAnimationFrame = 0
	// The bottom line of the map should not be tappable as that space is
	// reserved to avoid conflict with iPhoneX's HomeIndicator
	if tx < 0 || consts.TileXNum <= tx || ty < 0 || consts.TileYNum-1 <= ty {
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
	if m.waitingRequestID == 0 {
		return true
	}

	r := sceneManager.ReceiveResultIfExists(m.waitingRequestID)
	if r == nil {
		return false
	}
	m.waitingRequestID = 0
	switch r.Type {
	case scene.RequestTypeRewardedAds:
		if r.Succeeded {
			if m.onRewarded != nil {
				mg := m.gameState.Minigame()
				sceneManager.Requester().RequestSendAnalytics(fmt.Sprintf("minigame%d_reward", mg.ID()), "")
				m.onRewarded(sceneManager)
			}
		}

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
		text := texts.Text(lang.Get(), texts.TextIDRemoveAdsDesc)
		if _, ok := prices["ads_removal"]; ok {
			priceText = prices["ads_removal"]
		}
		m.removeAdsLabel.Text = fmt.Sprintf(text, priceText)
		m.removeAdsDialog.Show()
	case scene.RequestTypePurchase, scene.RequestTypeShowShop:
		// Note: Ideally we should show a notification toast to notify users about the result
		// For now, the notifications are handled on the native platform side
		if r.Succeeded {
			m.updatePurchasesState(sceneManager)
		}
		m.removeAdsDialog.Hide()
	}
	return false
}

func (m *MapScene) isUIBusy() bool {
	if m.quitDialog.Visible() {
		return true
	}
	if m.storeErrorDialog.Visible() {
		return true
	}
	if m.removeAdsDialog.Visible() {
		return true
	}
	if m.screenShotDialog.Visible() {
		return true
	}
	if m.credits.Visible() {
		return true
	}
	return false
}

func (m *MapScene) updateUI(sceneManager *scene.Manager) {
	l := lang.Get()
	m.quitLabel.Text = texts.Text(l, texts.TextIDBackToTitle)
	m.quitYesButton.SetText(texts.Text(l, texts.TextIDYes))
	m.quitNoButton.SetText(texts.Text(l, texts.TextIDNo))
	m.storeErrorLabel.Text = texts.Text(l, texts.TextIDStoreError)
	m.storeErrorOkButton.SetText(texts.Text(l, texts.TextIDOK))
	m.removeAdsYesButton.SetText(texts.Text(l, texts.TextIDYes))
	m.removeAdsNoButton.SetText(texts.Text(l, texts.TextIDNo))
	m.removeAdsButton.SetText(texts.Text(l, texts.TextIDRemoveAds))

	m.quitDialog.Update()
	m.screenShotDialog.Update()
	m.storeErrorDialog.Update()
	m.removeAdsDialog.Update()

	m.cameraButton.Update()
	if m.gameHeader != nil {
		m.gameHeader.Update(m.quitDialog.Visible())
	}

	if m.gameState.Map().IsBlockingEventExecuting() {
		m.removeAdsButton.Disable()
	} else {
		m.removeAdsButton.Enable()
	}
	m.removeAdsButton.Update()

	if m.gameState.InventoryVisible() {
		m.inventory.Show()
	} else {
		m.inventory.Hide()
	}
	m.inventory.SetDisabled(m.gameState.Map().IsBlockingEventExecuting())
	m.inventory.SetItems(m.gameState.Items().Items())
	m.inventory.SetActiveItemID(m.gameState.Items().ActiveItem())
	m.inventory.Update(sceneManager)

	m.itemPreviewPopup.Update(l)
	m.itemPreviewPopup.SetEnabled(!m.gameState.Map().IsBlockingEventExecuting())

	// Event handling
	m.updateItemPreviewPopupVisibility(sceneManager)

	if m.gameState.Minigame().Active() {
		m.minigamePopup.Show()
	} else {
		m.minigamePopup.Hide()
	}
	m.minigamePopup.Update(m.gameState.Minigame())
	m.minigamePopup.SetAdsLoaded(sceneManager.RewardedAdsLoaded())

	if m.titleView != nil {
		m.titleView.Update(sceneManager)
	}

	m.credits.Update()
}

func (m *MapScene) Update(sceneManager *scene.Manager) error {
	if m.err != nil {
		return m.err
	}

	if m.gameState.ShouldShowCredits() {
		m.credits.SetData(sceneManager.Credits())
		m.credits.Show()
		m.gameState.ShowedCredits()
	}

	m.animation.Update()

	if !m.initialized {
		m.initUI(sceneManager)
		m.initialized = true
	}

	m.updatePurchasesState(sceneManager)

	if ok := m.receiveRequest(sceneManager); !ok {
		return nil
	}

	if input.BackButtonPressed() {
		m.handleBackButton()
	}

	m.updateUI(sceneManager)
	if m.isUIBusy() {
		return nil
	}

	if m.initialState && m.gameState.IsAutoSaveEnabled() {
		m.gameState.RequestSave(0, sceneManager)
	}
	m.initialState = false
	if err := m.gameState.Update(sceneManager); err != nil {
		if err == gamestate.GoToTitle {
			m.goToTitle(sceneManager)
			return nil
		}
		return err
	}

	m.runEventIfNeeded(sceneManager)
	m.updateOffsetY(sceneManager)
	return nil
}

func (m *MapScene) goToTitle(sceneManager *scene.Manager) {
	audio.Stop()
	g, err := savedGame(sceneManager)
	if err != nil {
		m.err = err
		return
	}
	sceneManager.GoToWithFading(NewTitleMapScene(g), FadingCount, FadingCount)
}

func (m *MapScene) handleBackButton() {
	if m.gameState.Map().IsBlockingEventExecuting() {
		return
	}

	if m.storeErrorDialog.Visible() {
		audio.PlaySE("system/cancel", 1.0)
		m.storeErrorDialog.Hide()
		return
	}

	if m.quitDialog.Visible() {
		audio.PlaySE("system/cancel", 1.0)
		m.quitDialog.Hide()
		return
	}

	audio.PlaySE("system/click", 1.0)
	m.quitDialog.Show()
}

func (m *MapScene) drawTileLayer(layer int, priority data.Priority) {
	op := &ebiten.DrawImageOptions{}
	room := m.gameState.Map().CurrentRoom()

	for j := 0; j < consts.TileYNum; j++ {
		for i := 0; i < consts.TileXNum; i++ {
			tileIndex := tileset.TileIndex(i, j)
			tile := room.Tiles[layer][tileIndex]
			if tile == 0 {
				continue
			}
			imageID := tileset.ExtractImageID(tile)
			imageName := m.gameState.Map().FindImageName(imageID)
			index := 0
			if !tileset.IsAutoTile(imageName) {
				x, y := tileset.DecodeTile(tile)
				index = tileset.TileIndex(x, y)
			}
			passageType := tileset.PassageType(imageName, index)
			if layer == 2 || layer == 3 {
				if passageType == data.PassageTypeOver && priority != data.PriorityTop {
					continue
				}
				if passageType != data.PassageTypeOver && priority == data.PriorityTop {
					continue
				}
			}
			if tileset.IsAutoTile(imageName) {
				m.drawAutoTile(tile, op, i, j)
			} else {
				m.drawTile(tile, op, i, j)
			}
		}
	}
}

func (m *MapScene) drawTile(tile int, op *ebiten.DrawImageOptions, i int, j int) {
	imageID := tileset.ExtractImageID(tile)
	tileSetImg := m.gameState.Map().FindImage(imageID)
	if tileSetImg == nil {
		return
	}
	x, y := tileset.DecodeTile(tile)
	sx := x * consts.TileSize
	sy := y * consts.TileSize
	dx := i * consts.TileSize
	dy := j*consts.TileSize + m.offsetY/consts.TileScale
	// op is created outside of this function and other parameters than GeoM is not modified so far.
	op.GeoM.Reset()
	op.GeoM.Translate(float64(dx), float64(dy))
	m.screenImage.DrawImage(tileSetImg.SubImage(image.Rect(sx, sy, sx+consts.TileSize, sy+consts.TileSize)).(*ebiten.Image), op)
}

func (m *MapScene) drawAutoTile(tile int, op *ebiten.DrawImageOptions, i int, j int) {
	imageID := tileset.ExtractImageID(tile)
	tileSetImg := m.gameState.Map().FindImage(imageID)
	if tileSetImg == nil {
		return
	}
	autoTileSlice := tileset.DecodeAutoTile(tile)
	for index, value := range autoTileSlice {
		x, y := tileset.GetAutoTilePos(index, value)
		sx := x * consts.MiniTileSize
		sy := y * consts.MiniTileSize
		dx := i*consts.TileSize + index%2*consts.MiniTileSize
		dy := j*consts.TileSize + index/2*consts.MiniTileSize + m.offsetY/consts.TileScale
		// op is created outside of this function and other parameters is not modified so far.
		op.GeoM.Reset()
		op.GeoM.Translate(float64(dx), float64(dy))
		m.screenImage.DrawImage(tileSetImg.SubImage(image.Rect(sx, sy, sx+consts.MiniTileSize, sy+consts.MiniTileSize)).(*ebiten.Image), op)
	}

}

func (m *MapScene) drawTiles(priority data.Priority) {
	if priority == data.PriorityBottom {
		m.drawTileLayer(0, priority)
		m.drawTileLayer(1, priority)
	} else {
		m.drawTileLayer(2, priority)
		m.drawTileLayer(3, priority)
	}
}

func (m *MapScene) Draw(screen *ebiten.Image) {
	const mapWidth = consts.MapWidth

	if !m.initialized {
		return
	}

	if m.credits.Visible() {
		m.credits.Draw(screen)
		return
	}

	// Filling with black instead of clearing is necessary for tinting.
	// See the change 701eb1105ec126f09680f6a185ed1b1bf5235950.
	m.screenImage.Fill(color.Black)

	if background := m.gameState.Map().Background(m.gameState); background != "" {
		img := assets.GetImage("backgrounds/" + background + ".png")
		_, h := img.Size()
		diff := h - consts.MapHeight
		m.animation.Draw(m.screenImage, img, mapWidth, 0, m.offsetY/consts.TileScale-diff)
	}

	m.gameState.DrawPictures(m.screenImage, 0, m.offsetY/consts.TileScale, data.PicturePriorityBottom)
	for k := 0; k < 3; k++ {
		var p data.Priority
		switch k {
		case 0:
			p = data.PriorityBottom
		case 1:
			p = data.PriorityMiddle
		case 2:
			p = data.PriorityTop
		default:
			panic(fmt.Sprintf("sceneimpl: invalid priority: %d", k))
		}

		m.drawTiles(p)
		// Characters can be rendered in the upper black area.
		// That's why offset needs to be specified here.
		m.gameState.Map().DrawCharacters(m.screenImage, p, 0, m.offsetY/consts.TileScale)
	}

	m.gameState.DrawPictures(m.screenImage, 0, m.offsetY/consts.TileScale, data.PicturePriorityTop)

	if foreground := m.gameState.Map().Foreground(m.gameState); foreground != "" {
		img := assets.GetImage("foregrounds/" + foreground + ".png")
		_, h := img.Size()
		diff := h - consts.MapHeight
		m.animation.Draw(m.screenImage, img, mapWidth, 0, m.offsetY/consts.TileScale-diff)
	}

	m.gameState.DrawWeather(m.screenImage)
	m.gameState.DrawScreen(m.screenImage)

	tintScreenImage := m.screenImage
	if !m.gameState.ZeroTint() {
		tintScreenImage = m.tintScreenImage
		op := &ebiten.DrawImageOptions{}
		m.gameState.ApplyTintColor(&op.ColorM)
		op.CompositeMode = ebiten.CompositeModeCopy
		tintScreenImage.DrawImage(m.screenImage, op)
	}

	m.gameState.DrawPictures(tintScreenImage, 0, m.offsetY/consts.TileScale, data.PicturePriorityOverlay)

	op := &ebiten.DrawImageOptions{}
	m.gameState.ApplyShake(&op.GeoM)
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	// If the screen is shaking, there is a region in the screen that is not rendered. Clear first.
	if op.GeoM.Element(0, 2) != 0 || op.GeoM.Element(1, 2) != 0 {
		screen.Clear()
	}
	screen.DrawImage(tintScreenImage, op)

	if m.gameState.IsPlayerControlEnabled() && (m.gameState.Map().IsPlayerMovingByUserInput() || m.triggeringFailed) {
		x, y := m.moveDstX, m.moveDstY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*consts.TileSize), float64(y*consts.TileSize))
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(0, float64(m.offsetY))

		numFrames := m.markerAnimationFrame / markerAnimationInterval
		markerImage := assets.GetImage("system/game/marker.png")
		screen.DrawImage(markerImage.SubImage(image.Rect(markerSize*numFrames, 0, markerSize*(1+numFrames), markerSize)).(*ebiten.Image), op)

		w, _ := markerImage.Size()
		frameCount := w / markerSize
		if m.markerAnimationFrame < frameCount*markerAnimationInterval {
			m.markerAnimationFrame++
		} else {
			m.triggeringFailed = false
		}
	}

	m.itemPreviewPopup.Draw(screen)
	m.minigamePopup.Draw(screen)
	m.inventory.Draw(screen)

	m.cameraButton.Draw(screen)
	m.removeAdsButton.Draw(screen)

	m.gameState.DrawWindows(screen, 0, m.offsetY/consts.TileScale, m.windowOffsetY/consts.TileScale)
	if m.gameHeader != nil {
		m.gameHeader.Draw(screen)
	}

	m.screenShotDialog.Draw(screen)
	m.quitDialog.Draw(screen)
	m.storeErrorDialog.Draw(screen)
	m.removeAdsDialog.Draw(screen)

	if m.cameraTaking {
		m.cameraTaking = false
		m.screenShotImage.Clear()
		op := &ebiten.DrawImageOptions{}
		sw, _ := screen.Size()
		w, _ := m.screenShotImage.Size()
		op.GeoM.Translate((float64(w)-float64(sw))/2, 0)
		m.screenShotImage.DrawImage(screen, nil)
	}

	if m.titleView != nil {
		m.titleView.Draw(screen)
	}

	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	msg = ""
	font.DrawText(screen, msg, 160, 8, consts.TextScale, data.TextAlignLeft, color.White, len([]rune(msg)))
}

func (m *MapScene) Resize() {
	m.initialized = false
	if m.titleView != nil {
		m.titleView.Resize()
	}
}

func (m *MapScene) shakeStartGameButton() {
	m.titleView.ShakeStartGameButton()
}
