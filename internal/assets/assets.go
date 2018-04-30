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
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

var theAssets = &assets{}

func Set(assets map[string][]byte, metadata map[string]*data.AssetMetadata) error {
	theAssets.assets = assets
	theAssets.metadata = metadata
	theAssets.images = map[string]*ebiten.Image{}
	return nil
}

type assets struct {
	assets   map[string][]byte
	metadata map[string]*data.AssetMetadata
	images   map[string]*ebiten.Image
}

func Exists(path string) bool {
	_, ok := theAssets.assets[path]
	return ok
}

func GetResource(path string) []byte {
	return theAssets.assets[path]
}

func GetMetadata(imageName string) *data.AssetMetadata {
	return theAssets.metadata["images/"+imageName+"_metadata.json"]
}
