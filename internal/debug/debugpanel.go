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

package debug

import (
	"fmt"
	"image/color"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type DebugPanel struct {
	entities         []*data.VariableData
	game             *gamestate.Game
	touchStartY      int
	prevTouchY       int
	prevScrollY      float64
	scrollY          float64
	moved            bool
	velocity         float64
	entityType       DebugPanelType
	touchingGroupID  int
	touchingEntityID int
	touchingSlot     int
	editingEntityID  int
	numberInput      *NumberInput
	screenHeight     int
}
type DebugPanelType string

const (
	DebugPanelTypeSwitch   DebugPanelType = "switch"
	DebugPanelTypeVariable DebugPanelType = "variable"
)

const (
	rowHeight = 50
)

func NewDebugPanel(game *gamestate.Game, entityType DebugPanelType) *DebugPanel {
	d := &DebugPanel{
		game:        game,
		entityType:  entityType,
		numberInput: NewNumberInput(),
	}
	return d
}

func (d *DebugPanel) calcPosY(i int) int {
	return 10 + i*rowHeight + int(d.scrollY)
}

func (d *DebugPanel) updateScroll() {
	_, wy := input.Wheel()
	if math.Abs(wy) > 0.01 {
		d.velocity = wy * 4
		d.scrollY += d.velocity
	} else {
		_, iy := input.Position()
		if input.Triggered() {
			d.velocity = 0
			d.prevScrollY = d.scrollY
			d.touchStartY = iy
			d.prevTouchY = iy
			d.moved = false
		}
		if input.Pressed() {
			d.scrollY = d.prevScrollY + float64(iy-d.touchStartY)
			dy := iy - d.prevTouchY
			if math.Abs(float64(dy)) > 5 {
				d.moved = true
				d.velocity = float64(dy)
			}
			d.prevTouchY = iy
		} else {
			d.scrollY += d.velocity
		}
	}

	d.velocity *= 0.9
	if math.Abs(d.velocity) < 0.01 {
		d.velocity = 0
	}

	if d.scrollY > 0 {
		d.scrollY = 0
	}

	maxScrollY := (d.RowCount()+1)*rowHeight - d.screenHeight
	if int(d.scrollY) < -maxScrollY {
		d.scrollY = float64(-maxScrollY)
	}
}

func (d *DebugPanel) updateNumberInput() {
	if d.editingEntityID > 0 {
		d.game.SetVariableValue(d.editingEntityID, int64(d.numberInput.Value()))
	}
	if input.Triggered() {
		d.editingEntityID = 0
	}
	d.numberInput.Update()
}

func (d *DebugPanel) updateTouch() {
	d.touchingEntityID = 0
	d.touchingGroupID = 0
	if d.velocity != 0 || d.moved {
		return
	}
	i := 0
	for _, s := range d.entities {
		y := d.calcPosY(i)
		if d.isBoxInScreen(y) {
			if d.includesInput(20, y, 460, 40) {
				if input.Pressed() {
					d.touchingGroupID = s.ID
				}
				if input.Released() {
					audio.PlaySE("system/click", 0.1)
					s.IsFolded = !s.IsFolded
				}
			}
		}
		i++
		if !s.IsFolded {
			for _, si := range s.Items {
				y := d.calcPosY(i)
				if !d.isBoxInScreen(y) {
					i++
					continue
				}
				if d.includesInput(360, y, 100, 40) {
					d.updateItemTouch(si)
				}
				i++
			}
		}
	}
}

func (d *DebugPanel) RowCount() int {
	i := 0
	for _, s := range d.entities {
		i++
		if !s.IsFolded {
			i += len(s.Items)
		}
	}
	return i
}

func (d *DebugPanel) Update(sceneManager *scene.Manager) error {
	switch d.entityType {
	case DebugPanelTypeSwitch:
		d.entities = sceneManager.Game().System.Switches
	case DebugPanelTypeVariable:
		d.entities = sceneManager.Game().System.Variables
	default:
		panic(fmt.Sprintf("Update: invalid entityType %s", d.entityType))
	}
	d.updateScroll()
	d.updateTouch()
	d.updateNumberInput()

	return nil
}

func (d *DebugPanel) updateItemTouch(i *data.VariableItem) {
	ix, _ := input.Position()
	if input.Pressed() {
		d.touchingEntityID = i.ID
		d.touchingSlot = 1
		if ix < 380 {
			d.touchingSlot = 0
		}
		if ix > 440 {
			d.touchingSlot = 2
		}
	}
	switch d.entityType {
	case DebugPanelTypeSwitch:
		if input.Released() {
			audio.PlaySE("system/click", 0.1)
			d.game.SetSwitchValue(i.ID, !(d.game.SwitchValue(i.ID) > 0))
		}

	case DebugPanelTypeVariable:
		if input.Released() {
			if d.touchingSlot == 0 {
				audio.PlaySE("system/click", 0.1)
				d.game.SetVariableValue(i.ID, d.game.VariableValue(i.ID)-1)
			}
			if d.touchingSlot == 1 {
				d.editingEntityID = i.ID
				audio.PlaySE("system/click", 0.1)
				d.numberInput.SetValue(int(d.game.VariableValue(i.ID)))
			}
			if d.touchingSlot == 2 {
				audio.PlaySE("system/click", 0.1)
				d.game.SetVariableValue(i.ID, d.game.VariableValue(i.ID)+1)
			}
		}
	default:
		panic(fmt.Sprintf("touchItem: invalid entityType %s", d.entityType))
	}
}

func (d *DebugPanel) includesInput(x, y, w, h int) bool {
	ix, iy := input.Position()
	if x <= ix && ix < x+w && y <= iy && iy < y+h {
		return true
	}
	return false
}

func (d *DebugPanel) DrawBox(screen *ebiten.Image, padding, x, y, w, h int, text string, state int) {

	geoM := &ebiten.GeoM{}
	geoM.Translate(float64(x), float64(y))

	var s string

	colorM := &ebiten.ColorM{}
	switch state {
	case 0:
		s = "system/common/9patch_frame_off.png"
	case 1:
		s = "system/common/9patch_frame_on.png"
	case 2:
		s = "system/common/9patch_frame_on.png"
		colorM.Scale(1, 1, 0, 1)
	case 3:
		s = "system/common/9patch_frame_on.png"
		colorM.Scale(1, 0.7, 0, 1)
	}

	ui.DrawNinePatches(screen, assets.GetImage(s), w, h, geoM, colorM)

	font.DrawText(screen, text, int(padding+x), int(y+10), 1, data.TextAlignLeft, color.White, len([]rune(text)))
}

func (d *DebugPanel) renderValue(screen *ebiten.Image, i *data.VariableItem, y int) {
	var text string
	var state int
	switch d.entityType {
	case DebugPanelTypeSwitch:
		if d.game.SwitchValue(i.ID) > 0 {
			text = "ON"
			state = 1
		} else {
			text = "OFF"
			state = 0
		}
		if d.touchingEntityID == i.ID {
			state = 2
		}
		d.DrawBox(screen, 45, 360, y, 100, 40, text, state)
	case DebugPanelTypeVariable:
		touching := d.touchingEntityID == i.ID
		slot := d.touchingSlot
		state = 0
		text = strconv.Itoa(int(d.game.VariableValue(i.ID)))
		if slot == 1 {
			if d.editingEntityID == i.ID {
				text = d.numberInput.Text()
				state = 3
			}
			if touching {
				state = 2
			}
		}
		d.DrawBox(screen, 45, 360, y, 100, 40, text, state)
		state = 0
		if touching && slot == 0 {
			state = 2
		}
		d.DrawBox(screen, 4, 360, y, 20, 40, "◀", state)
		state = 0
		if touching && slot == 2 {
			state = 2
		}
		d.DrawBox(screen, 4, 440, y, 20, 40, "▶", state)
	default:
		panic(fmt.Sprintf("clickItem: invalid entityType %s", d.entityType))
	}

}

func (d *DebugPanel) isBoxInScreen(y int) bool {
	if y < -40 {
		return false
	}
	if y > d.screenHeight {
		return false
	}
	return true
}

func (d *DebugPanel) DrawItems(screen *ebiten.Image, startIndex int, items []*data.VariableItem) int {

	i := 0
	for _, s := range items {
		y := d.calcPosY(startIndex + i)
		if !d.isBoxInScreen(y) {
			i++
			continue
		}
		d.DrawBox(screen, 20, 40, y, 420, 40, s.Name, 0)
		d.renderValue(screen, s, y)
		i++
	}
	return i
}

func (d *DebugPanel) Draw(screen *ebiten.Image) {
	_, sh := screen.Size()
	d.screenHeight = sh
	i := 0
	for _, s := range d.entities {
		state := 1
		if d.touchingGroupID == s.ID {
			state = 2
		}
		y := d.calcPosY(i)
		if d.isBoxInScreen(y) {
			d.DrawBox(screen, 20, 20, y, 440, 40, s.Name, state)
		}
		i++
		if !s.IsFolded {
			i += d.DrawItems(screen, i, s.Items)
		}
	}
}
