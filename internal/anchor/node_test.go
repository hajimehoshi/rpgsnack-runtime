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

package anchor_test

import (
	"testing"

	. "github.com/hajimehoshi/rpgsnack-runtime/internal/anchor"
)

type Rect struct {
	X0 float64
	Y0 float64
	X1 float64
	Y1 float64
}

func TestRect(t *testing.T) {
	cases := []struct {
		Name     string
		Parent   *Rect
		Child    *Rect
		Anchor   *Rect
		Resize   *Rect
		ChildAbs *Rect
	}{
		{
			Name:     "no resize",
			Parent:   &Rect{10, 10, 20, 20},
			Child:    &Rect{1, 1, 9, 9},
			Anchor:   &Rect{0, 0, 1, 1},
			Resize:   &Rect{10, 10, 20, 20},
			ChildAbs: &Rect{11, 11, 19, 19},
		},
		{
			Name:     "enlarge",
			Parent:   &Rect{10, 10, 20, 20},
			Child:    &Rect{1, 1, 9, 9},
			Anchor:   &Rect{0, 0, 1, 1},
			Resize:   &Rect{10, 10, 30, 30},
			ChildAbs: &Rect{11, 11, 29, 29},
		},
		{
			Name:     "enlarge (anchor: left upper)",
			Parent:   &Rect{10, 10, 20, 20},
			Child:    &Rect{1, 1, 9, 9},
			Anchor:   &Rect{0, 0, 0, 0},
			Resize:   &Rect{10, 10, 30, 30},
			ChildAbs: &Rect{11, 11, 19, 19},
		},
		{
			Name:     "enlarge (anchor: upper)",
			Parent:   &Rect{10, 10, 20, 20},
			Child:    &Rect{1, 1, 9, 9},
			Anchor:   &Rect{0, 0, 1, 0},
			Resize:   &Rect{10, 10, 30, 30},
			ChildAbs: &Rect{11, 11, 29, 19},
		},
		{
			Name:     "enlarge (anchor: lower)",
			Parent:   &Rect{10, 10, 20, 20},
			Child:    &Rect{1, 1, 9, 9},
			Anchor:   &Rect{0, 1, 1, 1},
			Resize:   &Rect{10, 10, 30, 30},
			ChildAbs: &Rect{11, 21, 29, 29},
		},
		{
			Name:     "enlarge (anchor: center)",
			Parent:   &Rect{10, 10, 20, 20},
			Child:    &Rect{1, 1, 9, 9},
			Anchor:   &Rect{0.5, 0.5, 0.5, 0.5},
			Resize:   &Rect{10, 10, 30, 30},
			ChildAbs: &Rect{16, 16, 24, 24},
		},
		{
			Name:     "move",
			Parent:   &Rect{10, 10, 20, 20},
			Child:    &Rect{1, 1, 9, 9},
			Anchor:   &Rect{0, 0, 1, 1},
			Resize:   &Rect{20, 20, 30, 30},
			ChildAbs: &Rect{21, 21, 29, 29},
		},
	}
	for _, c := range cases {
		n1 := NewNode(10, 10, 20, 20)
		n2 := NewNode(1, 1, 9, 9)
		n1.AppendChild(n2, c.Anchor.X0, c.Anchor.Y0, c.Anchor.X1, c.Anchor.Y1)
		n1.Resize(c.Resize.X0, c.Resize.Y0, c.Resize.X1, c.Resize.Y1)
		x0, y0, x1, y1 := n2.Abs()
		r := c.ChildAbs
		if x0 != r.X0 || y0 != r.Y0 || x1 != r.X1 || y1 != r.Y1 {
			t.Errorf("case (%s): n2.Abs() = {%.2f, %.2f, %.2f, %.2f}, want: {%.2f, %.2f, %.2f, %.2f}",
				c.Name, x0, y0, x1, y1, r.X0, r.Y0, r.X1, r.Y1)
		}
	}
}
