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
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type DebugPanel struct {
	game       *gamestate.Game
	entityType DebugPanelType
}
type DebugPanelType string

const (
	DebugPanelTypeSwitch   DebugPanelType = "switch"
	DebugPanelTypeVariable DebugPanelType = "variable"
)

func NewDebugPanel(game *gamestate.Game, entityType DebugPanelType) *DebugPanel {
	d := &DebugPanel{
		game:       game,
		entityType: entityType,
	}
	return d
}

func (d *DebugPanel) Update(sceneManager *scene.Manager) error {
	return nil
}

func (d *DebugPanel) Draw(screen *ebiten.Image) {
}
