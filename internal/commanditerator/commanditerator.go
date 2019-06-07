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
	"fmt"

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

type CommandIterator struct {
	indices  []int // command index, branch index, command index, ...
	commands []*data.Command

	// Field that is not dumped
	labels map[string][]int

	terminating bool
}

func New(commands []*data.Command) *CommandIterator {
	c := &CommandIterator{
		indices:  []int{0},
		commands: commands,
		labels:   map[string][]int{},
	}
	c.unindentIfNeeded()
	c.recordLabel(c.commands, nil)
	return c
}

func (c *CommandIterator) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("indices")
	e.BeginArray()
	for _, i := range c.indices {
		e.EncodeInt(i)
	}
	e.EndArray()

	e.EncodeString("commands")
	e.BeginArray()
	for _, c := range c.commands {
		e.EncodeInterface(c)
	}
	e.EndArray()

	e.EndMap()
	return e.Flush()
}

func (c *CommandIterator) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "indices":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				c.indices = make([]int, n)
				for i := 0; i < n; i++ {
					c.indices[i] = d.DecodeInt()
				}
			}
		case "commands":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				c.commands = make([]*data.Command, n)
				for i := 0; i < n; i++ {
					if !d.SkipCodeIfNil() {
						c.commands[i] = &data.Command{}
						d.DecodeInterface(c.commands[i])
					}
				}
			}
		default:
			if err := d.Error(); err != nil {
				return fmt.Errorf("commanditerator: CommandIterator.DecodeMsgpack failed: %v", err)
			}
			return fmt.Errorf("commanditerator: CommandIterator.DecodeMsgpack failed: invalid key: %s", k)
		}
	}
	c.labels = map[string][]int{}
	c.recordLabel(c.commands, []int{})
	if err := d.Error(); err != nil {
		return fmt.Errorf("commanditerator: CommandIterator.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (c *CommandIterator) recordLabel(commands []*data.Command, pointer []int) {
	for ci, command := range commands {
		// Copy the p once so that other slices referring the underlying array should not be affected.
		p := make([]int, len(pointer))
		copy(p, pointer)
		p = append(p, ci)
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
			pp := make([]int, len(p))
			copy(pp, p)
			pp = append(pp, bi)
			c.recordLabel(b, pp)
		}
	}
}

func (c *CommandIterator) Rewind() {
	c.indices = []int{0}
	c.unindentIfNeeded()
}

func (c *CommandIterator) Terminate() {
	c.indices = []int{}
	c.terminating = false
}

func (c *CommandIterator) TerminateGracefully() {
	if c.IsTerminated() {
		return
	}
	c.terminating = true
}

func (c *CommandIterator) IsTerminated() bool {
	return len(c.indices) == 0
}

func (c *CommandIterator) unindentIfNeeded() {
	if len(c.indices) == 0 {
		return
	}
loop:
	cc := c.commands
	for i := 0; i < (len(c.indices)+1)/2; i++ {
		if len(cc) <= c.indices[i*2] {
			if 0 < i*2-1 {
				c.indices = c.indices[:i*2-1]
				c.indices[len(c.indices)-1]++
				goto loop
			}
			c.indices = []int{}
			return
		}
		if i < (len(c.indices)+1)/2-1 {
			cc = cc[c.indices[i*2]].Branches[c.indices[i*2+1]]
		}
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
	if c.terminating {
		c.Terminate()
		return
	}
	c.indices[len(c.indices)-1]++
	c.unindentIfNeeded()
}

func (c *CommandIterator) Choose(branchIndex int) {
	if c.terminating {
		c.Terminate()
		return
	}
	c.indices = append(c.indices, branchIndex, 0)
	c.unindentIfNeeded()
}

func (c *CommandIterator) Goto(label string) bool {
	if c.terminating {
		c.Terminate()
		return false
	}

	p, ok := c.labels[label]
	if !ok {
		// TODO: log error?
		return false
	}
	c.indices = make([]int, len(p))
	copy(c.indices, p)
	return true
}
