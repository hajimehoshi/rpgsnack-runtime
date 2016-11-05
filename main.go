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

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/font"
)

func update(screen *ebiten.Image) error {
	if err := font.DrawText(screen, "償いの時計 Clock of Atonement", 0, 0, textScale, color.White); err != nil {
		return err
	}
	return nil
}

const (
	tileSize  = 16
	tileXNum  = 10
	tileYNum  = 10
	textScale = 2
	mapScale  = 3
)

func main() {
	const title = "Clock of Atonement"
	if err := ebiten.Run(update, tileXNum*tileSize*mapScale, tileYNum*tileSize*mapScale, 1, title); err != nil {
		panic(err)
	}
}
