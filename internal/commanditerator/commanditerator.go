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
	commandIndices []int
	branchIndices  []int
}

func (p *pointer) appendCommand(i int) *pointer {
	ci := make([]int, len(p.commandIndices)+1)
	copy(ci, p.commandIndices)
	ci[len(ci)-1] = i
	return &pointer{
		commandIndices: ci,
		branchIndices:  p.branchIndices,
	}
}

func (p *pointer) appendBranch(i int) *pointer {
	bi := make([]int, len(p.branchIndices)+1)
	copy(bi, p.branchIndices)
	bi[len(bi)-1] = i
	return &pointer{
		commandIndices: p.commandIndices,
		branchIndices:  bi,
	}
}

type CommandIterator struct {
	commandIndices []int
	branchIndices  []int
	commands       []*data.Command
	labels         map[string]*pointer
}

func New(commands []*data.Command) *CommandIterator {
	c := &CommandIterator{
		commandIndices: []int{0},
		commands:       commands,
		labels:         map[string]*pointer{},
	}
	c.unindentIfNeeded()
	c.recordLabel(c.commands, &pointer{[]int{}, []int{}})
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
	c.commandIndices = []int{0}
	c.unindentIfNeeded()
}

func (c *CommandIterator) IsTerminated() bool {
	return len(c.commandIndices) == 0
}

func (c *CommandIterator) unindentIfNeeded() {
loop:
	for 0 < len(c.commandIndices) {
		branch := c.commands
		for i := 0; i < len(c.commandIndices); i++ {
			if len(branch) <= c.commandIndices[i] {
				c.commandIndices = c.commandIndices[:i]
				if len(c.commandIndices) > 0 {
					c.commandIndices[len(c.commandIndices)-1]++
				}
				if i > 0 {
					c.branchIndices = c.branchIndices[:i-1]
				}
				continue loop
			}
			if i < len(c.commandIndices)-1 {
				command := branch[c.commandIndices[i]]
				branch = command.Branches[c.branchIndices[i]]
				continue
			}
		}
		return
	}
}

func (c *CommandIterator) Command() *data.Command {
	branch := c.commands
	for i, bi := range c.branchIndices {
		command := branch[c.commandIndices[i]]
		branch = command.Branches[bi]
	}
	return branch[c.commandIndices[len(c.commandIndices)-1]]
}

func (c *CommandIterator) Advance() {
	c.commandIndices[len(c.commandIndices)-1]++
	c.unindentIfNeeded()
}

func (c *CommandIterator) Choose(branchIndex int) {
	c.branchIndices = append(c.branchIndices, branchIndex)
	c.commandIndices = append(c.commandIndices, 0)
	c.unindentIfNeeded()
}

func (c *CommandIterator) Goto(label string) bool {
	p, ok := c.labels[label]
	if !ok {
		// TODO: log error?
		return false
	}
	c.commandIndices = p.commandIndices
	c.branchIndices = p.branchIndices
	return true
}
