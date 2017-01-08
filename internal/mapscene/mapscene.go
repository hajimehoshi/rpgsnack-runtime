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

package mapscene

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
)

type MapScene struct {
	gameState  *gamestate.Game
	moveDstX   int
	moveDstY   int
	tilesImage *ebiten.Image
}

func New() (*MapScene, error) {
	tilesImage, err := ebiten.NewImage(scene.TileXNum*scene.TileSize, scene.TileYNum*scene.TileSize, ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	state, err := gamestate.NewGame()
	if err != nil {
		return nil, err
	}
	mapScene := &MapScene{
		tilesImage: tilesImage,
		gameState:  state,
	}
	return mapScene, nil
}

func (m *MapScene) movePlayerIfNeeded() error {
	if !input.Triggered() {
		return nil
	}
	x, y := input.Position()
	tx := (x - scene.GameMarginX) / scene.TileSize / scene.TileScale
	ty := (y - scene.GameMarginTop) / scene.TileSize / scene.TileScale
	result, err := m.gameState.Map().TryMovePlayerByUserInput(tx, ty)
	if err != nil {
		return err
	}
	if !result {
		return nil
	}
	m.moveDstX = tx
	m.moveDstY = ty
	return nil
}

func (m *MapScene) Update(sceneManager *scene.SceneManager) error {
	if err := m.gameState.Screen().Update(); err != nil {
		return err
	}
	if err := m.gameState.Windows().Update(); err != nil {
		return err
	}
	if err := m.gameState.Map().Update(); err != nil {
		return err
	}
	m.gameState.Map().TryRunAutoEvent()
	if err := m.movePlayerIfNeeded(); err != nil {
		return err
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

func (m *MapScene) Draw(screen *ebiten.Image) error {
	m.tilesImage.Clear()
	tileset, err := m.gameState.Map().TileSet()
	if err != nil {
		return err
	}
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = &tilesImageParts{
		room:    m.gameState.Map().CurrentRoom(),
		tileSet: tileset,
		layer:   0,
	}
	if err := m.tilesImage.DrawImage(assets.GetImage(tileset.Images[0]), op); err != nil {
		return err
	}
	op.ImageParts = &tilesImageParts{
		room:     m.gameState.Map().CurrentRoom(),
		tileSet:  tileset,
		layer:    1,
		overOnly: false,
	}
	if err := m.tilesImage.DrawImage(assets.GetImage(tileset.Images[1]), op); err != nil {
		return err
	}
	if err := m.gameState.Map().DrawPlayer(m.tilesImage); err != nil {
		return err
	}
	if err := m.gameState.Map().DrawEvents(m.tilesImage); err != nil {
		return err
	}
	op = &ebiten.DrawImageOptions{}
	tileSet, err := m.gameState.Map().TileSet()
	if err != nil {
		return err
	}
	op.ImageParts = &tilesImageParts{
		room:     m.gameState.Map().CurrentRoom(),
		tileSet:  tileSet,
		layer:    1,
		overOnly: true,
	}
	if err := m.tilesImage.DrawImage(assets.GetImage(tileset.Images[1]), op); err != nil {
		return err
	}
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scene.TileScale, scene.TileScale)
	op.GeoM.Translate(scene.GameMarginX, scene.GameMarginTop)
	m.gameState.Screen().Apply(&op.ColorM)
	if err := screen.DrawImage(m.tilesImage, op); err != nil {
		return err
	}
	if m.gameState.Map().IsPlayerMovingByUserInput() {
		x, y := m.moveDstX, m.moveDstY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*scene.TileSize), float64(y*scene.TileSize))
		op.GeoM.Scale(scene.TileScale, scene.TileScale)
		op.GeoM.Translate(scene.GameMarginX, scene.GameMarginTop)
		if err := screen.DrawImage(assets.GetImage("marker.png"), op); err != nil {
			return err
		}
	}
	if err := m.gameState.Windows().Draw(screen); err != nil {
		return nil
	}
	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	if err := font.DrawText(screen, msg, 0, 0, scene.TextScale, color.White); err != nil {
		return err
	}
	return nil
}
