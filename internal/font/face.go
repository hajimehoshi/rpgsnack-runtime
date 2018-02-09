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
	"github.com/hajimehoshi/go-mplusbitmap"
	"golang.org/x/image/font"
)

var (
	faces = map[int]font.Face{}
)

func face(scale int) font.Face {
	// Use the same instance to use text cache efficiently.
	f, ok := faces[scale]
	if !ok {
		f = scaleFont(mplusbitmap.Gothic12r, scale)
		faces[scale] = f
	}
	return f
}
