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

package ui

const (
	partSize = 4
)

type ninePatchParts struct {
	width  int
	height int
}

func (n *ninePatchParts) Len() int {
	return (n.width / partSize) * (n.height / partSize)
}

func (n *ninePatchParts) Src(index int) (int, int, int, int) {
	xn := n.width / partSize
	yn := n.height / partSize
	sx, sy := 0, 0
	switch index % xn {
	case 0:
		sx = 0
	case xn - 1:
		sx = 2 * partSize
	default:
		sx = 1 * partSize
	}
	switch index / xn {
	case 0:
		sy = 0
	case yn - 1:
		sy = 2 * partSize
	default:
		sy = 1 * partSize
	}
	return sx, sy, sx + partSize, sy + partSize
}

func (n *ninePatchParts) Dst(index int) (int, int, int, int) {
	xn := n.width / partSize
	dx := (index % xn) * partSize
	dy := (index / xn) * partSize
	return dx, dy, dx + partSize, dy + partSize
}
