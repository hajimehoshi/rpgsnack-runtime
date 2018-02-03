// Copyright 2018 Hajime Hoshi
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

package tileset

import (
	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

func TileIndex(x, y int) int {
	return y*consts.TileXNum + x
}

func DecodeTile(tile int) (int, int, int) {
	// tile hash is consist of three section
	// [8bit - x][8bit - y][15bit - image id]
	const yOffset = 15
	const xOffset = 23
	const imgIDMask = (1 << yOffset) - 1
	const yMask = (1 << xOffset) - 1 - imgIDMask
	imageID := tile & imgIDMask
	x := tile >> xOffset
	y := (tile & yMask) >> yOffset
	return x, y, imageID
}

// PassageType gets passage type from the metadata attached to image.
// Returns passable when metadata doesn't exist or no passage type is set at the
// required position.
func PassageType(imageName string, x, y int) data.PassageType {
	metadata := assets.GetMetadata(imageName)
	if metadata == nil {
		return data.PassageTypePassable
	}
	p := metadata.PassageTypes
	index := TileIndex(x, y)
	if index >= len(p) {
		return data.PassageTypePassable
	}
	return p[index]
}
