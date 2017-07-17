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

package variables

import (
	"encoding/json"
	"fmt"

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

type Variables struct {
	switches     []bool
	selfSwitches map[string][]bool
	variables    []int
}

type tmpVariables struct {
	Switches     []bool            `json:"switches"`
	SelfSwitches map[string][]bool `json:"selfSwitches"`
	Variables    []int             `json:"variables"`
}

func (v *Variables) MarshalJSON() ([]uint8, error) {
	tmp := &tmpVariables{
		Switches:     v.switches,
		SelfSwitches: v.selfSwitches,
		Variables:    v.variables,
	}
	return json.Marshal(tmp)
}

func (v *Variables) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("switches")
	e.BeginArray()
	for _, val := range v.switches {
		e.EncodeBool(val)
	}
	e.EndArray()

	e.EncodeString("selfSwitches")
	e.BeginMap()
	for k, val := range v.selfSwitches {
		e.EncodeString(k)
		e.BeginArray()
		for _, s := range val {
			e.EncodeBool(s)
		}
		e.EndArray()
	}
	e.EndMap()

	e.EncodeString("variables")
	e.BeginArray()
	for _, val := range v.variables {
		e.EncodeInt(val)
	}
	e.EndArray()

	e.EndMap()
	return e.Flush()
}

func (v *Variables) UnmarshalJSON(data []uint8) error {
	var tmp *tmpVariables
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	v.switches = tmp.Switches
	v.selfSwitches = tmp.SelfSwitches
	v.variables = tmp.Variables
	return nil
}

func (v *Variables) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "switches":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				v.switches = make([]bool, n)
				for i := 0; i < n; i++ {
					v.switches[i] = d.DecodeBool()
				}
			}
		case "selfSwitches":
			if !d.SkipCodeIfNil() {
				n := d.DecodeMapLen()
				v.selfSwitches = map[string][]bool{}
				for i := 0; i < n; i++ {
					k := d.DecodeString()
					v.selfSwitches[k] = nil
					if !d.SkipCodeIfNil() {
						n := d.DecodeArrayLen()
						a := make([]bool, n)
						for i := 0; i < n; i++ {
							a[i] = d.DecodeBool()
						}
						v.selfSwitches[k] = a
					}
				}
			}
		case "variables":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				v.variables = make([]int, n)
				for i := 0; i < n; i++ {
					v.variables[i] = d.DecodeInt()
				}
			}
		case "innerVariables":
			d.Skip()
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("variables: Variables.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (v *Variables) SwitchValue(id int) bool {
	if len(v.switches) < id+1 {
		zeros := make([]bool, id+1-len(v.switches))
		v.switches = append(v.switches, zeros...)
	}
	return v.switches[id]
}

func (v *Variables) SetSwitchValue(id int, value bool) {
	if len(v.switches) < id+1 {
		zeros := make([]bool, id+1-len(v.switches))
		v.switches = append(v.switches, zeros...)
	}
	v.switches[id] = value
}

func (v *Variables) SelfSwitchValue(mapID, roomID, eventID int, id int) bool {
	key := fmt.Sprintf("%d_%d_%d", mapID, roomID, eventID)
	if v.selfSwitches == nil {
		v.selfSwitches = map[string][]bool{}
	}
	values, ok := v.selfSwitches[key]
	if !ok {
		v.selfSwitches[key] = make([]bool, data.SelfSwitchNum)
		return false
	}
	return values[id]
}

func (v *Variables) SetSelfSwitchValue(mapID, roomID, eventID int, id int, value bool) {
	key := fmt.Sprintf("%d_%d_%d", mapID, roomID, eventID)
	if v.selfSwitches == nil {
		v.selfSwitches = map[string][]bool{}
	}
	values, ok := v.selfSwitches[key]
	if !ok {
		v.selfSwitches[key] = make([]bool, data.SelfSwitchNum)
		v.selfSwitches[key][id] = value
		return
	}
	values[id] = value
}

func (v *Variables) VariableValue(id int) int {
	if len(v.variables) < id+1 {
		zeros := make([]int, id+1-len(v.variables))
		v.variables = append(v.variables, zeros...)
	}
	return v.variables[id]
}

func (v *Variables) SetVariableValue(id int, value int) {
	if len(v.variables) < id+1 {
		zeros := make([]int, id+1-len(v.variables))
		v.variables = append(v.variables, zeros...)
	}
	v.variables[id] = value
}
