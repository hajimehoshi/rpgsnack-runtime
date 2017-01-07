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

package character

import (
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type commandIndex struct {
	commandIndices []int
	branchIndices  []int
	commands       []*data.Command
}

func newCommandIndex(commands []*data.Command) *commandIndex {
	c := &commandIndex{
		commandIndices: []int{0},
		commands:       commands,
	}
	c.unindentIfNeeded()
	return c
}

func (c *commandIndex) isTerminated() bool {
	return len(c.commandIndices) == 0
}

func (c *commandIndex) unindentIfNeeded() {
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

func (c *commandIndex) command() *data.Command {
	branch := c.commands
	for i, bi := range c.branchIndices {
		command := branch[c.commandIndices[i]]
		branch = command.Branches[bi]
	}
	return branch[c.commandIndices[len(c.commandIndices)-1]]
}

func (c *commandIndex) advance() {
	c.commandIndices[len(c.commandIndices)-1]++
	c.unindentIfNeeded()
}

func (c *commandIndex) choose(branchIndex int) {
	c.branchIndices = append(c.branchIndices, branchIndex)
	c.commandIndices = append(c.commandIndices, 0)
	c.unindentIfNeeded()
}
