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

type Condition struct {
	Type      ConditionType      `json:"type" msgpack:"type"`
	ID        int                `json:"id" msgpack:"id"`
	Comp      ConditionComp      `json:"comp" msgpack:"comp"`
	ValueType ConditionValueType `json:"valueType" msgpack:"valueType"`
	Value     interface{}        `json:"value" msgpack:"value"`
}

type ConditionType string

const (
	ConditionTypeSwitch     ConditionType = "switch"
	ConditionTypeSelfSwitch ConditionType = "self_switch"
	ConditionTypeVariable   ConditionType = "variable"
	ConditionTypeItem       ConditionType = "item"
	ConditionTypeSpecial    ConditionType = "special" // This type is intended for inner only.
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

type ConditionItemValue string

const (
	ConditionItemOwn    ConditionItemValue = "own"
	ConditionItemNotOwn ConditionItemValue = "not_own"
	ConditionItemActive ConditionItemValue = "active"
)
