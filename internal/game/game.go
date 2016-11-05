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

package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/font"
)

const (
	tileSize  = 16
	tileXNum  = 10
	tileYNum  = 10
	textScale = 2
	mapScale  = 3
)

type Game struct {
}

func New() *Game {
	return &Game{}
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) error {
	if err := font.DrawText(screen, "償いの時計 Clock of Atonement", 0, 0, textScale, color.White); err != nil {
		return err
	}
	return nil
}

func (g *Game) Title() string {
	return "Clock of Atonement"
}

func (g *Game) Size() (int, int) {
	return tileXNum * tileSize * mapScale, tileYNum * tileSize * mapScale
}
