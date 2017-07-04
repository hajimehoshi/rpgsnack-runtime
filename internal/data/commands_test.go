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

package data_test

import (
	"reflect"
	"testing"

	"github.com/vmihailenco/msgpack"

	. "github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

func TestSetVariable(t *testing.T) {
	c := &Command{
		Name: CommandNameSetVariable,
	}
	tests := []struct {
		args *CommandArgsSetVariable
	}{
		{
			args: &CommandArgsSetVariable{
				ID:        1,
				Op:        SetVariableOpAssign,
				ValueType: SetVariableValueTypeConstant,
				Value:     1,
			},
		},
		{
			args: &CommandArgsSetVariable{
				ID:        1,
				Op:        SetVariableOpAdd,
				ValueType: SetVariableValueTypeRandom,
				Value: &SetVariableValueRandom{
					Begin: 1,
					End:   2,
				},
			},
		},
		{
			args: &CommandArgsSetVariable{
				ID:        1,
				Op:        SetVariableOpSub,
				ValueType: SetVariableValueTypeCharacter,
				Value: &SetVariableCharacterArgs{
					Type:    SetVariableCharacterTypeDirection,
					EventID: 3,
				},
			},
		},
	}
	for _, test := range tests {
		c.Args = test.args
		b, err := msgpack.Marshal(c)
		if err != nil {
			t.Error(err)
		}
		var c2 *Command
		if err := msgpack.Unmarshal(b, &c2); err != nil {
			t.Error(err)
		}
		if c2.Name != c.Name {
			t.Errorf("got: %s, want: %s", c2.Name, c.Name)
		}
		args := c.Args.(*CommandArgsSetVariable)
		args2 := c2.Args.(*CommandArgsSetVariable)
		if args2.ID != args.ID {
			t.Errorf("got: %s, want: %s", args2.ID, args.ID)
		}
		if args2.Op != args.Op {
			t.Errorf("got: %s, want: %s", args2.Op, args.Op)
		}
		if args2.ValueType != args.ValueType {
			t.Errorf("got: %s, want: %s", args2.ValueType, args.ValueType)
		}
		if !reflect.DeepEqual(args2.Value, args.Value) {
			t.Errorf("got: %v, want: %v", args2.Value, args.Value)
		}
	}
}

func TestSetCharacterProperty(t *testing.T) {
	c := &Command{
		Name: CommandNameSetCharacterProperty,
	}
	tests := []struct {
		args *CommandArgsSetCharacterProperty
	}{
		{
			args: &CommandArgsSetCharacterProperty{
				Type:  SetCharacterPropertyTypeVisibility,
				Value: false,
			},
		},
		{
			args: &CommandArgsSetCharacterProperty{
				Type:  SetCharacterPropertyTypeSpeed,
				Value: Speed1,
			},
		},
	}
	for _, test := range tests {
		c.Args = test.args
		b, err := msgpack.Marshal(c)
		if err != nil {
			t.Error(err)
		}
		var c2 *Command
		if err := msgpack.Unmarshal(b, &c2); err != nil {
			t.Error(err)
		}
		if c2.Name != c.Name {
			t.Errorf("got: %s, want: %s", c2.Name, c.Name)
		}
		args := c.Args.(*CommandArgsSetCharacterProperty)
		args2 := c2.Args.(*CommandArgsSetCharacterProperty)
		if args2.Type != args.Type {
			t.Errorf("got: %s, want: %s", args2.Type, args.Type)
		}
		if !reflect.DeepEqual(args2.Value, args.Value) {
			t.Errorf("got: %v, want: %v", args2.Value, args.Value)
		}
	}
}
