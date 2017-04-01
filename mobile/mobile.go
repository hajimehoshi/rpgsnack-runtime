// Copyright 2017 Hajime Hoshi
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

// +build android ios

package mobile

import (
	"github.com/hajimehoshi/ebiten/mobile"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/game"
)

var (
	running bool
	theGame *game.Game
)

func SetData(game []uint8, progress []uint8, purchases []uint8) {
	// Copy data here since the given data is just a reference and might be
	// broken in the mobile side.
	g := make([]uint8, len(game))
	copy(g, game)
	var p1 []uint8
	if progress != nil {
		p1 = make([]uint8, len(progress))
		copy(p1, progress)
	}
	var p2 []uint8
	if purchases != nil {
		p2 = make([]uint8, len(purchases))
		copy(p2, purchases)
	}
	data.SetData(g, p1, p2)
}

func IsRunning() bool {
	return running
}

func Start(screenWidth int, screenHeight int, scale float64, requester Requester) error {
	running = true
	g, err := game.New(screenWidth, screenHeight, requester)
	if err != nil {
		return err
	}
	if err := mobile.Start(g.Update, screenWidth, screenHeight, scale, game.Title()); err != nil {
		return err
	}
	theGame = g
	return nil
}

func Update() error {
	return mobile.Update()
}

func UpdateTouchesOnAndroid(action int, id int, x, y int) {
	mobile.UpdateTouchesOnAndroid(action, id, x, y)
}

func UpdateTouchesOnIOS(phase int, ptr int64, x, y int) {
	mobile.UpdateTouchesOnIOS(phase, ptr, x, y)
}
