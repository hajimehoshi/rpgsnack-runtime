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

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

func decodeImage(path string, bin []uint8) (*ebiten.Image, error) {
	img, err := png.Decode(bytes.NewReader(bin))
	if err != nil {
		return nil, err
	}
	eimg, err := ebiten.NewImageFromImage(img, ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}
	return eimg, nil
}

func GetLocalizedImagePngBytes(key string) []byte {
	s := lang.Normalize(lang.Get()).String()
	k := path.Join("images", key+"@"+s+".png")
	if bin, ok := theAssets.assets[k]; ok {
		return bin
	}
	k = path.Join("images", key+".png")
	return GetResource(path.Join("images", key+".png"))
}

func GetLocalizedImage(key string) *ebiten.Image {
	l := lang.Normalize(lang.Get())

	// Look for the exact localized image (ex: zh-Hant.png)
	k := path.Join("images", key+"@"+l.String()+".png")
	if img, ok := theAssets.images[k]; ok {
		return img
	}

	// If not fallback to the base (ex: zh.png)
	t, _ := l.Base()
	k = path.Join("images", key+"@"+t.String()+".png")
	if img, ok := theAssets.images[k]; ok {
		return img
	}
	if bin, ok := theAssets.assets[k]; ok {
		img, err := decodeImage(k, bin)
		if err != nil {
			panic(fmt.Sprintf("assets: image decode error: %s, %v", k, err))
		}
		theAssets.images[k] = img
		return img
	}

	// If no localized image was found, use the common one
	return GetImage(key + ".png")
}

func ImageExists(key string) bool {
	if Exists(path.Join("images", key+".png")) {
		return true
	}
	s := lang.Normalize(lang.Get()).String()
	if Exists(path.Join("images", key+"@"+s+".png")) {
		return true
	}
	return false
}

func GetImage(key string) *ebiten.Image {
	k := path.Join("images", key)
	img, ok := theAssets.images[k]
	if !ok {
		bin, ok := theAssets.assets[k]
		if !ok {
			panic(fmt.Sprintf("assets: image not found: %s", k))
		}
		var err error
		img, err = decodeImage(k, bin)
		if err != nil {
			panic(fmt.Sprintf("assets: image decode error: %s, %v", k, err))
		}
		theAssets.images[k] = img
	}
	return img
}

func GetIconImage(key string) *ebiten.Image {
	iconPath := path.Join("icons", key)
	return GetImage(iconPath)
}
