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

package commanditerator_test

import (
	"encoding/json"
	"testing"

	. "github.com/hajimehoshi/rpgsnack-runtime/internal/commanditerator"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

func makeLabelCommand(name string) *data.Command {
	return &data.Command{
		Name: data.CommandNameLabel,
		Args: &data.CommandArgsLabel{
			Name: name,
		},
	}
}

func makeBranches(commands ...[]*data.Command) *data.Command {
	return &data.Command{
		Name:     data.CommandNameNop,
		Args:     nil,
		Branches: commands,
	}
}

func TestGoto(t *testing.T) {
	commands := []*data.Command{
		makeLabelCommand("foo"),
		makeLabelCommand("bar"),
		makeBranches(
			[]*data.Command{
				makeLabelCommand("baz"),
				makeLabelCommand("qux"),
			},
			[]*data.Command{
				makeBranches(
					[]*data.Command{
						makeLabelCommand("quux"),
					},
				),
			},
			[]*data.Command{
				makeLabelCommand("foo"), // should be ignored
				makeLabelCommand("corge"),
			}),
	}
	it := New(commands)
	cases := []struct {
		In  string
		Out string
	}{
		{
			In:  "foo",
			Out: "foo",
		},
		{
			In:  "bar",
			Out: "bar",
		},
		{
			In:  "baz",
			Out: "baz",
		},
		{
			In:  "qux",
			Out: "qux",
		},
		{
			In:  "quux",
			Out: "quux",
		},
		{
			In:  "corge",
			Out: "corge",
		},
	}
	for _, c := range cases {
		if !it.Goto(c.In) {
			t.Errorf("goto failed")
			continue
		}
		command := it.Command()
		if command.Name != data.CommandNameLabel {
			t.Errorf("command is not '%v' for %v", data.CommandNameLabel, c.In)
			continue
		}
		it.Command()
		got := command.Args.(*data.CommandArgsLabel).Name
		want := c.Out
		if got != want {
			t.Errorf("it.Command().Args.Name == %v want: %v", got, want)
		}

		// JSON marshaling
		j, err := json.Marshal(it)
		if err != nil {
			t.Fatal(err)
		}
		var it2 *CommandIterator
		if err := json.Unmarshal(j, &it2); err != nil {
			t.Fatal(err)
		}
		if !it2.Goto(c.In) {
			t.Errorf("goto failed")
			continue
		}
		command = it2.Command()
		if command.Name != data.CommandNameLabel {
			t.Errorf("command is not '%v' for %v", data.CommandNameLabel, c.In)
			continue
		}
	}
}
