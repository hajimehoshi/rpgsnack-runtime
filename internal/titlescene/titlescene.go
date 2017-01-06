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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/mapscene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type TitleScene struct {
}

func New() *TitleScene {
	return &TitleScene{}
}

func (t *TitleScene) Update(sceneManager *scene.SceneManager) error {
	if input.Triggered() {
		mapScene, err := mapscene.New()
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
