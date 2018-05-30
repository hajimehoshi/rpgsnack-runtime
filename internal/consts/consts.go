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

package consts

const (
	PaletteWidth           = 10
	TileSize               = 16
	MiniTileSize           = 8
	TileXNum               = 10
	TileYNum               = 22
	MapWidth               = TileXNum * TileSize
	MapHeight              = TileYNum * TileSize
	MapScaledWidth         = MapWidth * TileScale
	MapScaledHeight        = MapHeight * TileScale
	TileScale              = 3
	TextScale              = 2
	BigTextScale           = 3
	MaxFullscreenImageSize = 304
	SuperLargeScreenHeight = (TileYNum - 1) * TileSize * TileScale // For devices like iPhoneX
)
