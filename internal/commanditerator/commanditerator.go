// Copyright 2016 Hajime Hoshi
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

package commanditerator

import (
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type pointer struct {
	indices []int // command index, branch index, command index, ...
}

func (p *pointer) appendCommand(i int) *pointer {
	return &pointer{
		indices: append(p.indices, i),
	}
}

func (p *pointer) appendBranch(i int) *pointer {
	return &pointer{
		indices: append(p.indices, i),
	}
}

type CommandIterator struct {
	indices  []int
	commands []*data.Command
	labels   map[string]*pointer
}

func New(commands []*data.Command) *CommandIterator {
	c := &CommandIterator{
		indices:  []int{0},
		commands: commands,
		labels:   map[string]*pointer{},
	}
	c.unindentIfNeeded()
	c.recordLabel(c.commands, &pointer{[]int{}})
	return c
}

func (c *CommandIterator) recordLabel(commands []*data.Command, pointer *pointer) {
	for ci, command := range commands {
		p := pointer.appendCommand(ci)
		if command.Name == data.CommandNameLabel {
			label := command.Args.(*data.CommandArgsLabel).Name
			if _, ok := c.labels[label]; !ok {
				c.labels[label] = p
			}
		}
		if command.Branches == nil {
			continue
		}
		for bi, b := range command.Branches {
			c.recordLabel(b, p.appendBranch(bi))
		}
	}
}

func (c *CommandIterator) Rewind() {
	c.indices = []int{0}
	c.unindentIfNeeded()
}

func (c *CommandIterator) IsTerminated() bool {
	return len(c.indices) == 0
}

func (c *CommandIterator) unindentIfNeeded() {
loop:
	for 0 < len(c.indices) {
		cc := c.commands
		for i := 0; i < (len(c.indices)+1)/2; i++ {
			if len(cc) <= c.indices[i*2] {
				if 0 < i*2-1 {
					c.indices = c.indices[:i*2-1]
				} else {
					c.indices = []int{}
				}
				if len(c.indices) > 0 {
					c.indices[len(c.indices)-1]++
				}
				continue loop
			}
			if i < (len(c.indices)+1)/2-1 {
				cc = cc[c.indices[i*2]].Branches[c.indices[i*2+1]]
				continue
			}
		}
		return
	}
}

func (c *CommandIterator) Command() *data.Command {
	cc := c.commands
	for i := 0; i < len(c.indices)/2; i++ {
		cc = cc[c.indices[i*2]].Branches[c.indices[i*2+1]]
	}
	return cc[c.indices[len(c.indices)-1]]
}

func (c *CommandIterator) Advance() {
	c.indices[len(c.indices)-1]++
	c.unindentIfNeeded()
}

func (c *CommandIterator) Choose(branchIndex int) {
	c.indices = append(c.indices, branchIndex, 0)
	c.unindentIfNeeded()
}

func (c *CommandIterator) Goto(label string) bool {
	p, ok := c.labels[label]
	if !ok {
		// TODO: log error?
		return false
	}
	// TODO: Copy?
	c.indices = p.indices
	return true
}
