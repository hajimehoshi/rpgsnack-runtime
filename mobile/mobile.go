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

package mobile

import (
	"fmt"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/mobile"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/game"
)

var (
	theGame *game.Game

	startCalled = make(chan struct{})

	gameM sync.Mutex
)

func setData(project []byte, assets [][]byte, progress []byte, permanent []byte, purchases []byte, language string) {
	// Copy data here since the given data is just a reference and might be
	// broken in the mobile side.
	p := make([]byte, len(project))
	copy(p, project)

	var p1 []byte
	if progress != nil {
		p1 = make([]byte, len(progress))
		copy(p1, progress)
	}

	var p2 []byte
	if permanent != nil {
		p2 = make([]byte, len(permanent))
		copy(p2, permanent)
	}

	var p3 []byte
	if purchases != nil {
		p3 = make([]byte, len(purchases))
		copy(p3, purchases)
	}

	data.SetData(p, assets, p1, p2, p3, language)
}

func IsRunning() bool {
	select {
	case <-startCalled:
		return true
	default:
		return false
	}
}

func adjustScreenSize(widthInDP, heightInDP int) (width, height int, scale float64) {
	const (
		minWidth  = 480
		minHeight = 720
	)

	if float64(heightInDP)/float64(widthInDP) > float64(minHeight)/minWidth {
		scale = float64(widthInDP) / minWidth
	} else {
		scale = float64(heightInDP) / minHeight
	}
	width = int(math.Ceil(float64(widthInDP) / scale))
	height = int(math.Ceil(float64(heightInDP) / scale))
	return width, height, scale
}

var assetBytes [][]byte

func AppendAssetBytes(bytes []byte) {
	// Copy data here since the given data is just a reference and might be
	// broken in the mobile side.
	bs := make([]byte, len(bytes))
	copy(bs, bytes)
	assetBytes = append(assetBytes, bs)
}

func Start(widthInDP int, heightInDP int, requester Requester, project []byte, progress []byte, permanent []byte, purchases []byte, language string) (err error) {
	defer func() {
		close(startCalled)
	}()

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at Start: %v", err)
			}
		}
	}()

	setData(project, assetBytes, progress, permanent, purchases, language)

	width, height, scale := adjustScreenSize(widthInDP, heightInDP)
	g := game.New(width, height, requester)
	update := func(screen *ebiten.Image) error {
		gameM.Lock()
		defer gameM.Unlock()

		if err := g.Update(screen); err != nil {
			return err
		}
		return nil
	}
	if err := mobile.Start(update, width, height, scale, ""); err != nil {
		return err
	}
	theGame = g
	return nil
}

func SetScreenSize(widthInDP, heightInDP int) {
	gameM.Lock()
	defer gameM.Unlock()

	width, height, scale := adjustScreenSize(widthInDP, heightInDP)
	theGame.SetScreenSize(width, height, scale)
}

func Update() (err error) {
	<-startCalled

	defer func() {
		if r := recover(); r != nil {
			ok := false
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("error at Update: %v", err)
			}
		}
	}()
	return mobile.Update()
}

func UpdateTouchesOnAndroid(action int, id int, x, y int) {
	mobile.UpdateTouchesOnAndroid(action, id, x, y)
}

func UpdateTouchesOnIOS(phase int, ptr int64, x, y int) {
	mobile.UpdateTouchesOnIOS(phase, ptr, x, y)
}
