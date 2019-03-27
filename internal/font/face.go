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
	"github.com/hajimehoshi/bitmapfont"
	"golang.org/x/image/font"
	"golang.org/x/text/language"
)

var (
	bfFaces = map[int]font.Face{}
	scFaces = map[int]font.Face{}
	tcFaces = map[int]font.Face{}
)

func face(scale int, lang language.Tag) font.Face {
	switch lang {
	case language.SimplifiedChinese:
		f, ok := scFaces[scale]
		if !ok {
			f = scaleFont(gothic12r_sc, scale)
			scFaces[scale] = f
		}
		return f
	case language.TraditionalChinese:
		f, ok := tcFaces[scale]
		if !ok {
			f = scaleFont(gothic12r_tc, scale)
			tcFaces[scale] = f
		}
		return f
	default:
		f, ok := bfFaces[scale]
		if !ok {
			f = scaleFont(bitmapfont.Gothic12r, scale)
			bfFaces[scale] = f
		}
		return f
	}
}
