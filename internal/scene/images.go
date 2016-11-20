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

package scene

import (
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/assets"
)

var theImageCache *imageCache

func initImageCache() error {
	theImageCache = &imageCache{
		cache: map[string]*ebiten.Image{},
	}
	files, err := assets.AssetDir("images")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		img, err := assets.LoadImage("images/"+file, ebiten.FilterNearest)
		if err != nil {
			return err
		}
		theImageCache.cache[file] = img
	}
	return nil
}

func init() {
	// TODO: The image should be loaded asyncly.
	if err := initImageCache(); err != nil {
		panic(err)
	}
}

type imageCache struct {
	cache map[string]*ebiten.Image
}

func (i *imageCache) Get(path string) *ebiten.Image {
	return i.cache[path]
}
