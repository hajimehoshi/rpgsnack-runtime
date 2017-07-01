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

package assets

import (
	"strings"

	"github.com/hajimehoshi/ebiten"
)

var theAssets = &assets{}

func Set(assets map[string][]uint8) error {
	theAssets.assets = assets
	theAssets.images = map[string]*ebiten.Image{}
	for file, bin := range assets {
		if strings.HasSuffix(file, ".png") {
			img, err := loadImage(file, bin, ebiten.FilterNearest)
			if err != nil {
				return err
			}
			theAssets.images[file] = img
		}
	}
	return nil
}

type assets struct {
	assets map[string][]uint8
	images map[string]*ebiten.Image
}

func Exists(path string) bool {
	_, ok := theAssets.assets[path]
	return ok
}

func GetResource(path string) []uint8 {
	return theAssets.assets[path]
}
