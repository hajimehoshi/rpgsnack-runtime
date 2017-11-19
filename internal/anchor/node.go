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

package anchor

type rect struct {
	x0 float64
	y0 float64
	x1 float64
	y1 float64
}

type Node struct {
	rect     *rect
	parent   *Node
	children []*childNode
}

type childNode struct {
	node     *Node
	anchor   *rect
	toAnchor *rect
}

func (c *childNode) recalc() {
	w, h := c.node.parent.Size()
	a := c.anchor
	x0 := a.x0*w - c.toAnchor.x0
	y0 := a.y0*h - c.toAnchor.y0
	x1 := a.x1*w - c.toAnchor.x1
	y1 := a.y1*h - c.toAnchor.y1
	c.node.Resize(x0, y0, x1, y1)
}

func NewNode(x0, y0, x1, y1 float64) *Node {
	return &Node{
		rect: &rect{x0, y0, x1, y1},
	}
}

func (n *Node) AppendChild(child *Node, anchorX0, anchorY0, anchorX1, anchorY1 float64) {
	a := &rect{anchorX0, anchorY0, anchorX1, anchorY1}
	w := n.rect.x1 - n.rect.x0
	h := n.rect.y1 - n.rect.y0
	c := &childNode{
		node:   child,
		anchor: a,
		toAnchor: &rect{
			x0: a.x0*w - child.rect.x0,
			y0: a.y0*h - child.rect.y0,
			x1: a.x1*w - child.rect.x1,
			y1: a.y1*h - child.rect.y1,
		},
	}
	n.children = append(n.children, c)
	child.parent = n
}

func (n *Node) Resize(x0, y0, x1, y1 float64) {
	n.rect.x0 = x0
	n.rect.y0 = y0
	n.rect.x1 = x1
	n.rect.y1 = y1
	for _, c := range n.children {
		c.recalc()
	}
}

func (n *Node) Size() (w, h float64) {
	return n.rect.x1 - n.rect.x0, n.rect.y1 - n.rect.y0
}

func (n *Node) Abs() (x0, y0, x1, y1 float64) {
	px, py := 0.0, 0.0
	if n.parent != nil {
		px, py, _, _ = n.parent.Abs()
	}
	return px + n.rect.x0, py + n.rect.y0, px + n.rect.x1, py + n.rect.y1
}
