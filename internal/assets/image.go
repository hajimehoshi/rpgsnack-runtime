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

	"golang.org/x/text/language"

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

func GetLocalizeImage(key string, lang language.Tag) *ebiten.Image {
	base, _ := lang.Base()
	img, ok := theAssets.images[path.Join("images", key+"_"+base.String()+".png")]
	if ok {
		return img
	}
	return GetImage(key + ".png")
}

func GetImage(key string) *ebiten.Image {
	img, ok := theAssets.images[path.Join("images", key)]
	if !ok {
		panic(fmt.Sprintf("assets: image %s not found", key))
	}
	return img
}

func GetIconImage(key string) *ebiten.Image {
	iconPath := path.Join("icons", key)
	return GetImage(iconPath)
}
