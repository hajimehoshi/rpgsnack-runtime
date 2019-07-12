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

import (
	"fmt"
	"image/color"
)

// InterpreterID represents a unique identifier of interpreters
type InterpreterID int

const (
	PaletteWidth               = 10
	TileSize                   = 16
	MiniTileSize               = 8
	TileXNum                   = 10
	TileYNum                   = 22
	GuaranteedVisibleTileYNum  = 14
	MapWidth                   = TileXNum * TileSize
	MapHeight                  = TileYNum * TileSize
	GuaranteedVisibleMapHeight = GuaranteedVisibleTileYNum * TileSize
	MapScaledWidth             = MapWidth * TileScale
	MapScaledHeight            = MapHeight * TileScale
	TileScale                  = 3
	TextScale                  = 2
	BigTextScale               = 3
	MaxFullscreenImageSize     = 304
	HeaderHeight               = 16 * TileScale
)

func CeilDiv(x, y int) int {
	return (x-1)/y + 1
}

// HasExtraBottomGrid reports whether the screen has an extra bottom grid due to the screen is too large or not.
//
// HasExtraBottomGrid return true on devices like iPhone X.
func HasExtraBottomGrid(screenHeight int) bool {
	const superLargeScreenHeight = (TileYNum - 1) * TileSize * TileScale
	return screenHeight > superLargeScreenHeight
}

type SponsorTierType string

const (
	SponsorTierType_1_Donation SponsorTierType = "tier1_donation"
	SponsorTierType_2_Donation SponsorTierType = "tier2_donation"
	SponsorTierType_3_Donation SponsorTierType = "tier3_donation"
	SponsorTierType_4_Donation SponsorTierType = "tier4_donation"
)

func (t SponsorTierType) IsValid() bool {
	switch t {
	case SponsorTierType_1_Donation, SponsorTierType_2_Donation, SponsorTierType_3_Donation, SponsorTierType_4_Donation:
		return true
	}
	return false
}

func (t SponsorTierType) Level() int {
	switch t {
	case SponsorTierType_1_Donation:
		return 1
	case SponsorTierType_2_Donation:
		return 2
	case SponsorTierType_3_Donation:
		return 3
	case SponsorTierType_4_Donation:
		return 4
	default:
		panic(fmt.Sprintf("consts: invalid sponsor tier type: %s", t))
	}
}

func (t SponsorTierType) Color() color.Color {
	switch t {
	case SponsorTierType_1_Donation:
		return color.RGBA{0xcd, 0x7f, 0x32, 0xff} // bronze
	case SponsorTierType_2_Donation:
		return color.RGBA{0xc0, 0xc0, 0xc0, 0xff} // silver
	case SponsorTierType_3_Donation:
		return color.RGBA{0xff, 0xd7, 0x00, 0xff} // gold
	case SponsorTierType_4_Donation:
		return color.RGBA{0xee, 0xee, 0xee, 0xff} // platinum
	default:
		panic(fmt.Sprintf("consts: invalid sponsor tier type: %s", t))
	}
}
