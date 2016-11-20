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
	"encoding/json"

	"github.com/hajimehoshi/tsugunai/internal/assets"
	"github.com/hajimehoshi/tsugunai/internal/data"
)

const (
	tileSize      = 16
	characterSize = 16
	tileXNum      = 10
	tileYNum      = 10
	textScale     = 2
	tileScale     = 3
)

const (
	gameMarginX = 0
	gameMarginY = 2.5 * tileSize * tileScale
)

func GameSize() (int, int) {
	return tileXNum*tileSize*tileScale + 2*gameMarginX, tileYNum*tileSize*tileScale + 2*gameMarginY
}

// TODO: This variable should belong to a struct.
var (
	tileSets []*data.TileSet
)

func init() {
	mapDataBytes := assets.MustAsset("data/tilesets.json")
	if err := json.Unmarshal(mapDataBytes, &tileSets); err != nil {
		panic(err)
	}
}
