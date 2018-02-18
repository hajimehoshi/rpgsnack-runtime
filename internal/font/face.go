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

package font

import (
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/go-mplusbitmap"
	"golang.org/x/image/font"
	"golang.org/x/text/language"
)

var (
	scTTF *truetype.Font
	tcTTF *truetype.Font

	scFaces = map[int]font.Face{}
	tcFaces = map[int]font.Face{}
)

func ensureSCTTF() *truetype.Font {
	if scTTF != nil {
		return scTTF
	}

	bs, err := getSCTTF()
	if err != nil {
		panic(err)
	}
	scTTF, err = truetype.Parse(bs)
	if err != nil {
		panic(err)
	}
	return scTTF
}

func ensureTCTTF() *truetype.Font {
	if tcTTF != nil {
		return tcTTF
	}

	bs, err := getTCTTF()
	if err != nil {
		panic(err)
	}
	tcTTF, err = truetype.Parse(bs)
	if err != nil {
		panic(err)
	}
	return tcTTF
}

func face(scale int, lang language.Tag) font.Face {
	const dpi = 72

	switch lang {
	case language.SimplifiedChinese:
		f, ok := scFaces[scale]
		if !ok {
			f = truetype.NewFace(ensureSCTTF(), &truetype.Options{
				Size:    12 * float64(scale),
				DPI:     dpi,
				Hinting: font.HintingFull,
			})
			scFaces[scale] = f
		}
		return f
	case language.TraditionalChinese:
		f, ok := tcFaces[scale]
		if !ok {
			f = truetype.NewFace(ensureTCTTF(), &truetype.Options{
				Size:    12 * float64(scale),
				DPI:     dpi,
				Hinting: font.HintingFull,
			})
			tcFaces[scale] = f
		}
		return f
	default:
		return scaleFont(mplusbitmap.Gothic12r, scale)
	}
}
