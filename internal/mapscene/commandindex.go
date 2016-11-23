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

package mapscene

import (
	"github.com/hajimehoshi/tsugunai/internal/data"
)

type commandIndex struct {
	commands []int
	branches []int
	page     *data.Page
}

func newCommandIndex(page *data.Page) *commandIndex {
	return &commandIndex{
		commands: []int{0},
		page:     page,
	}
}

func (c *commandIndex) isTerminated() bool {
	return len(c.commands) == 0
}

func (c *commandIndex) unindentIfNeeded() {
loop:
	for 0 < len(c.commands) {
		branch := c.page.Commands
		for i := 0; i < len(c.commands); i++ {
			if len(branch) <= c.commands[i] {
				c.commands = c.commands[:i]
				if len(c.commands) > 0 {
					c.commands[len(c.commands)-1]++
				}
				if i > 0 {
					c.branches = c.branches[:i-1]
				}
				continue loop
			}
			if i < len(c.commands)-1 {
				command := branch[c.commands[i]]
				branch = command.Branches[c.branches[i]]
				continue
			}
		}
		return
	}
}

func (c *commandIndex) command() *data.Command {
	branch := c.page.Commands
	for i, bi := range c.branches {
		command := branch[c.commands[i]]
		branch = command.Branches[bi]
	}
	return branch[c.commands[len(c.commands)-1]]
}

func (c *commandIndex) advance() {
	c.commands[len(c.commands)-1]++
	c.unindentIfNeeded()
}

func (c *commandIndex) choose(branchIndex int) {
	c.branches = append(c.branches, branchIndex)
	c.commands = append(c.commands, 0)
	c.unindentIfNeeded()
}
