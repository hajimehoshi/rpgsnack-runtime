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

package embedded

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten"
)

var images = map[string]*ebiten.Image{}

func Get(key string) *ebiten.Image {
	img, ok := images[key]
	if ok {
		return img
	}
	var orig []byte
	switch key {
	case "back":
		orig = back_png
	case "close":
		orig = close_png
	case "loading":
		orig = loading_png
	default:
		panic(fmt.Sprintf("embedded: key not found: %s", key))
	}

	i, _, err := image.Decode(bytes.NewReader(orig))
	if err != nil {
		panic(fmt.Sprintf("embedded: decode error: %s: %v", key, err))
	}

	images[key], _ = ebiten.NewImageFromImage(i, ebiten.FilterDefault)
	return images[key]
}
