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
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type MapScene struct {
	gameState        *gamestate.Game
	moveDstX         int
	moveDstY         int
	tilesImage       *ebiten.Image
	triggeringFailed bool
	initialState     bool
	cameraButton     *ui.Button
	cameraTaking     bool
	screenShotImage  *ebiten.Image
	screenShotDialog *ui.Dialog
}

func NewMapScene() (*MapScene, error) {
	tilesImage, _ := ebiten.NewImage(scene.TileXNum*scene.TileSize, scene.TileYNum*scene.TileSize, ebiten.FilterNearest)
	state, err := gamestate.NewGame()
	if err != nil {
		return nil, err
	}
	screenShotImage, _ := ebiten.NewImage(480, 720, ebiten.FilterLinear)
	camera, _ := ebiten.NewImage(12, 12, ebiten.FilterNearest)
	camera.Fill(color.RGBA{0xff, 0, 0, 0xff})
	m := &MapScene{
		tilesImage:       tilesImage,
		gameState:        state,
		initialState:     true,
		cameraButton:     ui.NewImageButton(0, 0, camera),
		screenShotImage:  screenShotImage,
		screenShotDialog: ui.NewDialog(0, 4, 152, 232),
	}
	m.screenShotDialog.AddChild(ui.NewImage(8, 8, 1.0/scene.TileScale/2, m.screenShotImage))
	return m, nil
}

func NewMapSceneWithGame(game *gamestate.Game) (*MapScene, error) {
	tilesImage, err := ebiten.NewImage(scene.TileXNum*scene.TileSize, scene.TileYNum*scene.TileSize, ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	mapScene := &MapScene{
		tilesImage: tilesImage,
		gameState:  game,
	}
	return mapScene, nil
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
	y -= scene.GameMarginTop
	if x < 0 || y < 0 {
		return nil
	}
	tx := x / scene.TileSize / scene.TileScale
	ty := y / scene.TileSize / scene.TileScale
	if tx < 0 || scene.TileXNum <= tx || ty < 0 || scene.TileYNum <= ty {
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
	w, _ := sceneManager.Size()
	m.screenShotDialog.X = (w/scene.TileScale-160)/2 + 4
	if m.initialState {
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
			sceneManager.GoTo(NewTitleScene())
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
	return nil
}

type tilesImageParts struct {
	room     *data.Room
	tileSet  *data.TileSet
	layer    int
	overOnly bool
}

func (t *tilesImageParts) Len() int {
	return scene.TileXNum * scene.TileYNum
}

func (t *tilesImageParts) Src(index int) (int, int, int, int) {
	tile := t.room.Tiles[t.layer][index]
	if t.layer == 1 {
		p := t.tileSet.PassageTypes[t.layer][tile]
		if !t.overOnly && p == data.PassageTypeOver {
			return 0, 0, 0, 0
		}
		if t.overOnly && p != data.PassageTypeOver {
			return 0, 0, 0, 0
		}
	}
	// TODO: 8 is a magic number and should be replaced.
	x := tile % 8 * scene.TileSize
	y := tile / 8 * scene.TileSize
	return x, y, x + scene.TileSize, y + scene.TileSize
}

func (t *tilesImageParts) Dst(index int) (int, int, int, int) {
	x := index % scene.TileXNum * scene.TileSize
	y := index / scene.TileXNum * scene.TileSize
	return x, y, x + scene.TileSize, y + scene.TileSize
}

func (m *MapScene) Draw(screen *ebiten.Image) {
	m.tilesImage.Fill(color.Black)
	m.cameraButton.Draw(screen)
	tileSet := m.gameState.Map().TileSet()
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &tilesImageParts{
		room:    m.gameState.Map().CurrentRoom(),
		tileSet: tileSet,
		layer:   0,
	}
	m.tilesImage.DrawImage(assets.GetImage(tileSet.Images[0]), op)
	op.ImageParts = &tilesImageParts{
		room:     m.gameState.Map().CurrentRoom(),
		tileSet:  tileSet,
		layer:    1,
		overOnly: false,
	}
	m.tilesImage.DrawImage(assets.GetImage(tileSet.Images[1]), op)
	m.gameState.Map().DrawCharacters(m.tilesImage)
	op = &ebiten.DrawImageOptions{}
	op.ImageParts = &tilesImageParts{
		room:     m.gameState.Map().CurrentRoom(),
		tileSet:  tileSet,
		layer:    1,
		overOnly: true,
	}
	m.tilesImage.DrawImage(assets.GetImage(tileSet.Images[1]), op)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scene.TileScale, scene.TileScale)
	sw, _ := screen.Size()
	tx := (float64(sw) - scene.TileXNum*scene.TileSize*scene.TileScale) / 2
	op.GeoM.Translate(tx, scene.GameMarginTop)
	m.gameState.Screen().Draw(screen, m.tilesImage, op)
	if m.gameState.Map().IsPlayerMovingByUserInput() || m.triggeringFailed {
		x, y := m.moveDstX, m.moveDstY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*scene.TileSize), float64(y*scene.TileSize))
		op.GeoM.Scale(scene.TileScale, scene.TileScale)
		op.GeoM.Translate(tx, scene.GameMarginTop)
		screen.DrawImage(assets.GetImage("marker.png"), op)
	}
	m.gameState.DrawWindows(screen)
	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	font.DrawText(screen, msg, 0, 0, scene.TextScale, color.White)
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
}
