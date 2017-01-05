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

package gamestate

type Variables struct {
	switches  []bool
	variables []int
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
