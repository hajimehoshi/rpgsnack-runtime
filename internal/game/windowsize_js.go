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

// +build js

package game

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/hajimehoshi/ebiten"
)

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func Scale() float64 {
	window := js.Global.Get("window")
	windowWidth := window.Get("innerWidth").Float()
	windowHeight := window.Get("innerHeight").Float()
	// Now window size is fixed. Adjust these values when necessary.
	width, height := 480, 720
	return min(windowWidth/float64(width), windowHeight/float64(height))
}

func adjustWindowSize() {
	ebiten.SetScreenScale(Scale())
}

func init() {
	js.Global.Get("window").Call("addEventListener", "resize", func() {
		adjustWindowSize()
	})
}
