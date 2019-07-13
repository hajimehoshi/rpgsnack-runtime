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

package ui2

import (
	"image"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func diffAverage(xs []int) int {
	if len(xs) <= 1 {
		return 0
	}
	sum := 0
	for i := 0; i < len(xs)-1; i++ {
		sum += xs[i+1] - xs[i]
	}
	return sum / (len(xs) - 1)
}

type ScrollablePanel struct {
	x       int
	y       int
	width   int
	height  int
	visible bool

	nodes []Node

	pressing bool

	// offsetX and offsetY are the current scrolling offsets. This position represents the left-upper corner of
	// the content. The original positions are (0, 0), and usually take non-positive values.
	offsetX int
	offsetY int

	// origX and origY are the original offsetX/offsetY when dragging starts.
	origX int
	origY int

	// tmpDiffX and tmpDiffY are the difference between origX/origY and actual offset during dragging.
	tmpDiffX int
	tmpDiffY int

	// tmpOldDiffsX and tmpOldDiffsY are the tmpDiffX and tmpDiffY values in previous frames. These are used to
	// calculate velocities.
	tmpOldDiffsX []int
	tmpOldDiffsY []int

	// vx and vy are velocities of inertia.
	vx int
	vy int

	offscreen *ebiten.Image
}

func NewScrollablePanel(x, y, width, height int) *ScrollablePanel {
	return &ScrollablePanel{
		x:       x,
		y:       y,
		width:   width,
		height:  height,
		visible: true,
	}
}

func (s *ScrollablePanel) Show() {
	s.visible = true
}

func (s *ScrollablePanel) Hide() {
	s.visible = false
}

func (s *ScrollablePanel) AddChild(node Node) {
	s.nodes = append(s.nodes, node)
}

// childrenRegion returns a union region where child elements exist.
func (s *ScrollablePanel) childrenRegion() image.Rectangle {
	var r image.Rectangle
	for _, n := range s.nodes {
		r = r.Union(n.Region())
	}
	return r
}

// offsetRegion returns a region where the left-upper point can move.
// Usually, the region is (-width, -height) - (0, 0).
func (s *ScrollablePanel) offsetRegion() image.Rectangle {
	r := s.childrenRegion()
	x0 := min(-(r.Max.X - s.width), 0)
	y0 := min(-(r.Max.Y - s.height), 0)
	x1 := max(-r.Min.X, 0)
	y1 := max(-r.Min.Y, 0)
	return image.Rect(x0, y0, x1, y1)
}

func (s *ScrollablePanel) scrollable() (x, y bool) {
	r := s.offsetRegion()
	return r.Dx() > 0, r.Dy() > 0
}

func (s *ScrollablePanel) Region() image.Rectangle {
	return image.Rect(s.x, s.y, s.x+s.width, s.y+s.height)
}

// currentDiff returns differences between origX/origY and the actual cursor position.
func (s *ScrollablePanel) currentDiff() (int, int) {
	px, py := input.Position()
	r := s.offsetRegion()

	dx := px - s.origX
	dy := py - s.origY

	// ox and oy are the offsetX/offsetY assuming the current cursor position is used as the end of dragging.
	ox := s.offsetX + dx
	oy := s.offsetY + dy

	// If the scrolling reaches the edges, decrease the offset.
	const rate = 7.0 / 8.0
	if ox >= r.Max.X {
		dx -= int(float64(ox-r.Max.X) * rate)
	}
	if ox < r.Min.X {
		dx -= int(float64(ox-r.Min.X) * rate)
	}

	if oy >= r.Max.Y {
		dy -= int(float64(oy-r.Max.Y) * rate)
	}
	if oy < r.Min.Y {
		dy -= int(float64(oy-r.Min.Y) * rate)
	}

	return dx, dy
}

func (s *ScrollablePanel) inOffsetRegion() bool {
	r := s.offsetRegion()
	sx, sy := s.scrollable()
	x, y := s.offsetX, s.offsetY
	if sx && (x < r.Min.X || r.Max.X <= x) {
		return false
	}
	if sy && (y < r.Min.Y || r.Max.Y <= y) {
		return false
	}
	return true
}

func attenuate(v int, target int, rate float64) int {
	if v > target {
		return max(int(float64(v)*rate), target)
	}
	if v < target {
		return min(int(float64(v)*rate), target)
	}
	return target
}

func (s *ScrollablePanel) HandleInput(offsetX, offsetY int) bool {
	if !s.visible {
		return false
	}

	if x, y := input.Position(); image.Pt(x-offsetX, y-offsetY).In(s.Region()) {
		for _, n := range s.nodes {
			if n.HandleInput(s.x+s.offsetX+offsetX, s.y+s.offsetY+offsetY) {
				return true
			}
		}
	}

	if s.inOffsetRegion() {
		// Adjust positions by inertia.
		s.vx = attenuate(s.vx, 0, 31.0/32.0)
		s.vy = attenuate(s.vy, 0, 31.0/32.0)
		s.offsetX += s.vx
		s.offsetY += s.vy
	} else {
		r := s.offsetRegion()
		if s.offsetX < r.Min.X {
			s.offsetX += (r.Min.X - s.offsetX + 1) / 2
		}
		if s.offsetX >= r.Max.X {
			s.offsetX += (r.Max.X - s.offsetX - 1) / 2
		}
		if s.offsetY < r.Min.Y {
			s.offsetY += (r.Min.Y - s.offsetY + 1) / 2
		}
		if s.offsetY >= r.Max.Y {
			s.offsetY += (r.Max.Y - s.offsetY - 1) / 2
		}
		s.vx = 0
		s.vy = 0
	}

	if input.Released() {
		if !s.pressing {
			return false
		}
		s.pressing = false
		dx, dy := s.currentDiff()
		sx, sy := s.scrollable()
		if sx {
			s.offsetX += dx
			s.vx = diffAverage(append(s.tmpOldDiffsX, s.tmpDiffX)) * 2
		}
		if sy {
			s.offsetY += dy
			s.vy = diffAverage(append(s.tmpOldDiffsY, s.tmpDiffY)) * 2
		}

		s.origX = 0
		s.origY = 0
		s.tmpOldDiffsX = nil
		s.tmpOldDiffsY = nil
		s.tmpDiffX = 0
		s.tmpDiffY = 0
		return true
	}

	if !input.Pressed() {
		return false
	}

	if input.Triggered() {
		s.pressing = includesInput(offsetX, offsetY, s.Region())
		x, y := input.Position()
		sx, sy := s.scrollable()
		if sx {
			s.origX = x
		}
		if sy {
			s.origY = y
		}
		s.tmpOldDiffsX = nil
		s.tmpOldDiffsY = nil
		s.tmpDiffX = 0
		s.tmpDiffY = 0
	}

	if s.pressing {
		sx, sy := s.scrollable()
		dx, dy := s.currentDiff()
		const maxHistory = 10
		if sx {
			s.tmpOldDiffsX = append(s.tmpOldDiffsX, s.tmpDiffX)
			if len(s.tmpOldDiffsX) > maxHistory {
				s.tmpOldDiffsX = s.tmpOldDiffsX[len(s.tmpOldDiffsX)-maxHistory:]
			}
			s.tmpDiffX = dx
		}
		if sy {
			s.tmpOldDiffsY = append(s.tmpOldDiffsY, s.tmpDiffY)
			if len(s.tmpOldDiffsY) > maxHistory {
				s.tmpOldDiffsY = s.tmpOldDiffsY[len(s.tmpOldDiffsY)-maxHistory:]
			}
			s.tmpDiffY = dy
		}
	}
	return s.pressing
}

func (s *ScrollablePanel) Update() {
	if !s.visible {
		return
	}
	for _, n := range s.nodes {
		n.Update()
	}
}

func (s *ScrollablePanel) DrawAsChild(screen *ebiten.Image, offsetX, offsetY int) {
	if !s.visible {
		return
	}

	if s.offscreen == nil {
		s.offscreen, _ = ebiten.NewImage(s.width, s.height, ebiten.FilterDefault)
	}

	s.offscreen.Clear()
	for _, n := range s.nodes {
		n.DrawAsChild(s.offscreen, s.offsetX+s.tmpDiffX, s.offsetY+s.tmpDiffY)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.x+offsetX), float64(s.y+offsetY))
	screen.DrawImage(s.offscreen, op)
}

func (s *ScrollablePanel) SetScrollPosition(x, y int) {
	s.offsetX = x
	s.offsetY = y
	r := s.offsetRegion()
	s.offsetX = max(min(s.offsetX, r.Max.X), r.Min.X)
	s.offsetY = max(min(s.offsetY, r.Max.Y), r.Min.Y)
}

func (s *ScrollablePanel) ScrollPosition() (x, y int) {
	return s.offsetX, s.offsetY
}
