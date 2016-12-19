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

package titlescene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/data"
	"github.com/hajimehoshi/tsugunai/internal/font"
	"github.com/hajimehoshi/tsugunai/internal/input"
	"github.com/hajimehoshi/tsugunai/internal/mapscene"
	"github.com/hajimehoshi/tsugunai/internal/scene"
	"github.com/hajimehoshi/tsugunai/internal/task"
)

type TitleScene struct {
	gameData *data.Game
}

func New(gameData *data.Game) *TitleScene {
	return &TitleScene{gameData}
}

func (t *TitleScene) Update(subTasksUpdated bool, taskLine *task.TaskLine, sceneManager *scene.SceneManager) error {
	if subTasksUpdated {
		return nil
	}
	if input.Triggered() {
		mapScene, err := mapscene.New(t.gameData)
		if err != nil {
			return err
		}
		sceneManager.GoTo(mapScene)
	}
	return nil
}

func (t *TitleScene) Draw(screen *ebiten.Image) error {
	if err := font.DrawText(screen, "償いの時計\nClock of Atonement", 0, 0, scene.TextScale, color.White); err != nil {
		return err
	}
	return nil
}
