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

package assets

import (
	"bytes"
	"fmt"
	"image/png"
	"path"
	"strings"

	"github.com/hajimehoshi/ebiten"
)

func loadImage(path string, bin []uint8, filter ebiten.Filter) (*ebiten.Image, error) {
	img, err := png.Decode(bytes.NewReader(bin))
	if err != nil {
		return nil, err
	}
	eimg, err := ebiten.NewImageFromImage(img, filter)
	if err != nil {
		return nil, err
	}
	return eimg, nil
}

func SetResources(resources map[string][]uint8) error {
	theResources.resources = resources
	theResources.images = map[string]*ebiten.Image{}
	for file, bin := range resources {
		if strings.HasSuffix(file, ".png") {
			img, err := loadImage(file, bin, ebiten.FilterNearest)
			if err != nil {
				return err
			}
			theResources.images[file] = img
		}
	}
	return nil
}

var theResources = &resources{}

type resources struct {
	resources map[string][]uint8
	images    map[string]*ebiten.Image
}

func GetResource(path string) []uint8 {
	return theResources.resources[path]
}

func GetImage(key string) *ebiten.Image {
	img, ok := theResources.images[path.Join("images", key)]
	if !ok {
		panic(fmt.Sprintf("assets: image %s not found", key))
	}
	return img
}
