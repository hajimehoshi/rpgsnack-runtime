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

func SetData(project []byte, assets []byte, progress []byte, purchases []byte, language string) {
	// Copy data here since the given data is just a reference and might be
	// broken in the mobile side.
	p := make([]byte, len(project))
	copy(p, project)

	a := make([]byte, len(assets))
	copy(a, assets)

	var p1 []byte
	if progress != nil {
		p1 = make([]byte, len(progress))
		copy(p1, progress)
	}

	var p2 []byte
	if purchases != nil {
		p2 = make([]byte, len(purchases))
		copy(p2, purchases)
	}

	data.SetData(p, a, p1, p2, language)
}

func IsRunning() bool {
	return running
}

func Start(widthInDP int, heightInDP int, requester Requester) error {
	const (
		minWidth  = 480
		minHeight = 672
	)

	running = true

	width := 0
	height := 0
	scale := 0.0
	if float64(heightInDP)/float64(widthInDP) > float64(minHeight)/minWidth {
		scale = float64(widthInDP) / minWidth
	} else {
		scale = float64(heightInDP) / minHeight
	}
	width = int(float64(widthInDP) / scale)
	height = int(float64(heightInDP) / scale)

	g := game.New(width, height, requester)
	if err := mobile.Start(g.Update, width, height, scale, game.Title()); err != nil {
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
