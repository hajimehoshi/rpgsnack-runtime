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

package data

import (
	"encoding/json"
)

type Condition struct {
	Type      ConditionType
	ID        int
	Comp      ConditionComp
	ValueType ConditionValueType
	Value     interface{}
}

func (c *Condition) UnmarshalJSON(data []uint8) error {
	type tmpCondition struct {
		Type      ConditionType      `json:"type"`
		ID        int                `json:"id"`
		Comp      ConditionComp      `json:"comp"`
		ValueType ConditionValueType `json:"valueType"`
		Value     interface{}        `json:"value"`
	}
	var tmp *tmpCondition
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	c.Type = tmp.Type
	c.ID = tmp.ID
	c.Comp = tmp.Comp
	c.ValueType = tmp.ValueType
	switch c.Type {
	case ConditionTypeSwitch:
		c.Value = tmp.Value.(bool)
	case ConditionTypeSelfSwitch:
		c.Value = tmp.Value.(bool)
	case ConditionTypeVariable:
		c.Value = int(tmp.Value.(float64))
	}
	return nil
}

type ConditionType string

const (
	ConditionTypeSwitch     ConditionType = "switch"
	ConditionTypeSelfSwitch ConditionType = "self_switch"
	ConditionTypeVariable   ConditionType = "variable"
)

type ConditionComp string

const (
	ConditionCompEqualTo              ConditionComp = "=="
	ConditionCompNotEqualTo           ConditionComp = "!="
	ConditionCompGreaterThanOrEqualTo ConditionComp = ">="
	ConditionCompGreaterThan          ConditionComp = ">"
	ConditionCompLessThanOrEqualTo    ConditionComp = "<="
	ConditionCompLessThan             ConditionComp = "<"
)

type ConditionValueType string

const (
	ConditionValueTypeConstant ConditionValueType = "constant"
	ConditionValueTypeVariable ConditionValueType = "variable"
)
