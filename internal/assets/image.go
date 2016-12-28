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
	"image/png"

	"github.com/hajimehoshi/ebiten"
)

func loadImage(path string, filter ebiten.Filter) (*ebiten.Image, error) {
	bin := MustAsset(path)
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

func initImageCache(imageCache *imageCache) error {
	files, err := AssetDir("images")
	if err != nil {
		return err
	}
	for _, file := range files {
		img, err := loadImage("images/"+file, ebiten.FilterNearest)
		if err != nil {
			return err
		}
		imageCache.cache[file] = img
	}
	return nil
}

var theImageCache *imageCache

func init() {
	theImageCache = &imageCache{
		cache: map[string]*ebiten.Image{},
	}
	ch := make(chan error)
	go func() {
		defer close(ch)
		if err := initImageCache(theImageCache); err != nil {
			ch <- err
			return
		}
	}()
	theImageCache.loadingCh = ch
}

type imageCache struct {
	cache     map[string]*ebiten.Image
	loadingCh chan error
}

func (i *imageCache) Get(path string) *ebiten.Image {
	return i.cache[path]
}

func (i *imageCache) IsLoading() bool {
	if i.loadingCh == nil {
		return false
	}
	select {
	case err, ok := <-i.loadingCh:
		if err != nil {
			panic(err)
		}
		if !ok {
			i.loadingCh = nil
			return true
		}
	default:
	}
	return true
}

func IsLoading() bool {
	return theImageCache.IsLoading()
}

func GetImage(path string) *ebiten.Image {
	return theImageCache.Get(path)
}
