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

//go:generate go-bindata -nocompress -pkg=assets -ignore=/\. images
//go:generate gofmt -s -w .

package assets

import (
	"bytes"
	"image/png"
	"sync/atomic"

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
	atomic.StoreInt32(&theImageCache.loading, 1)
	go func() {
		if err := initImageCache(theImageCache); err != nil {
			panic(err)
		}
		atomic.StoreInt32(&theImageCache.loading, 0)
	}()
}

type imageCache struct {
	cache   map[string]*ebiten.Image
	loading int32
}

func (i *imageCache) Get(path string) *ebiten.Image {
	return i.cache[path]
}

func (i *imageCache) IsLoading() bool {
	return atomic.LoadInt32(&i.loading) != 0
}

func IsLoading() bool {
	return theImageCache.IsLoading()
}

func GetImage(path string) *ebiten.Image {
	return theImageCache.Get(path)
}
