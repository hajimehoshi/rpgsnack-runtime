// Copyright 2019 The RPGSnack Authors
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

package ui

import (
	"image"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
)

// includesInput reports whether the current input position is in the specified region.
//
// TODO: Add 'conforming' region for objects on a scrollable panel. Such objects can be moved out by scrolling.
func includesInput(offsetX, offsetY int, objectRegion image.Rectangle) bool {
	x, y := input.Position()
	x = int(float64(x) / consts.TileScale)
	y = int(float64(y) / consts.TileScale)
	x -= offsetX
	y -= offsetY
	return image.Pt(x, y).In(objectRegion)
}
