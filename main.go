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
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/game"
)

var theGame *game.Game

func update(screen *ebiten.Image) error {
	if err := theGame.Update(); err != nil {
		return err
	}
	if ebiten.IsRunningSlowly() {
		return nil
	}
	if err := theGame.Draw(screen); err != nil {
		return err
	}
	return nil
}

func main() {
	g, err := game.New()
	if err != nil {
		panic(err)
	}
	theGame = g
	w, h := theGame.Size()
	title := theGame.Title()
	if err := ebiten.Run(update, w, h, 1, title); err != nil {
		panic(err)
	}
}
