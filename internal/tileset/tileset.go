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

const (
	imgIDBits           = 15
	tilePosBits         = 8
	autoTileSectionBits = 3
)

// The algorithm is described in http://blog.rpgmakerweb.com/tutorials/anatomy-of-an-autotile
var miniTilePosMap = map[int]map[int][]int{
	// A
	0: {
		1: []int{2, 0},
		2: []int{0, 2},
		3: []int{2, 4},
		4: []int{2, 2},
		5: []int{0, 4},
	},
	// B
	1: {
		1: []int{3, 0},
		2: []int{3, 2},
		3: []int{1, 4},
		4: []int{1, 2},
		5: []int{3, 4},
	},
	// C
	2: {
		1: []int{2, 1},
		2: []int{0, 5},
		3: []int{2, 3},
		4: []int{2, 5},
		5: []int{0, 3},
	},
	// D
	3: {
		1: []int{3, 1},
		2: []int{3, 5},
		3: []int{1, 3},
		4: []int{1, 5},
		5: []int{3, 3},
	},
}

func TileIndex(x, y int) int {
	return y*consts.TileXNum + x
}

func ExtractImageID(tile int) int {
	const imgIDMask = (1 << imgIDBits) - 1
	return tile & imgIDMask
}

func DecodeTile(tile int) (int, int) {
	// tile hash is consist of three section
	// [8bit - x][8bit - y][15bit - image id]
	const (
		xOffset = imgIDBits + tilePosBits
		yOffset = imgIDBits
		yMask   = (1 << xOffset) - 1 - ((1 << yOffset) - 1)
	)
	x := tile >> xOffset
	y := (tile & yMask) >> yOffset
	return x, y
}

func DecodeAutoTile(tile int) []int {
	// auto tile hash is consist of five section
	// [3bit - a][3bit - b][3bit - c][3bit - d][15bit - image id]
	// a, b, c, d represent the four minitiles within a tile.
	const (
		dOffset = imgIDBits
		cOffset = imgIDBits + autoTileSectionBits
		bOffset = imgIDBits + autoTileSectionBits*2
		aOffset = imgIDBits + autoTileSectionBits*3
		dMask   = (1 << cOffset) - 1 - ((1 << dOffset) - 1)
		cMask   = (1 << bOffset) - 1 - ((1 << cOffset) - 1)
		bMask   = (1 << aOffset) - 1 - ((1 << bOffset) - 1)
	)
	a := tile >> aOffset
	b := (tile & bMask) >> bOffset
	c := (tile & cMask) >> cOffset
	d := (tile & dMask) >> dOffset
	return []int{a, b, c, d}
}

// PassageType gets passage type from the metadata attached to image.
// Returns passable when metadata doesn't exist or no passage type is set at the
// required position.
func PassageType(imageName string, index int) data.PassageType {
	metadata := assets.GetMetadata(imageName)
	if metadata == nil {
		return data.PassageTypePassable
	}
	p := metadata.PassageTypes
	if index >= len(p) {
		return data.PassageTypePassable
	}
	return p[index]
}

func IsAutoTile(imageName string) bool {
	metadata := assets.GetMetadata(imageName)
	if metadata == nil {
		return false
	}
	return metadata.IsAutoTile
}

func GetAutoTilePos(index int, value int) (int, int) {
	s := miniTilePosMap[index][value]
	return s[0], s[1]
}
