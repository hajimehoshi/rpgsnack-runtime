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

package data

import (
	"encoding/json"
	"fmt"
)

const (
	SelfSwitchNum = 4
)

type Event struct {
	ID    int     `json:"id"`
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Pages []*Page `json:"pages"`
}

type Page struct {
	Conditions []*Condition `json:"conditions"`
	Image      string       `json:"image"`
	ImageIndex int          `json:"imageIndex"`
	Attitude   Attitude     `json:"attitude"`
	Dir        Dir          `json:"dir"`
	DirFix     bool         `json:"dirFix"`
	Walking    bool         `json:"walking"`
	Stepping   bool         `json:"stepping"`
	Through    bool         `json:"through"`
	Priority   Priority     `json:"priority"`
	Trigger    Trigger      `json:"trigger"`
	Route      *Route       `json:"route"`
	Commands   []*Command   `json:"commands"`
}

type Dir int

const (
	DirUp Dir = iota
	DirRight
	DirDown
	DirLeft
)

type Attitude int

const (
	AttitudeLeft Attitude = iota
	AttitudeMiddle
	AttitudeRight
)

type Priority int

const (
	PriorityBelowCharacters Priority = iota
	PrioritySameAsCharacters
	PriorityAboveCharacters
)

type Trigger string

const (
	TriggerPlayer Trigger = "player"
	TriggerAuto           = "auto"
	TriggerDirect         = "direct"
	TriggerNever          = "never"
)

type Command struct {
	Name     CommandName
	Args     interface{}
	Branches [][]*Command
}

func (c *Command) UnmarshalJSON(data []uint8) error {
	type tmpCommand struct {
		Name     CommandName     `json:"name"`
		Branches [][]*Command    `json:"branches"`
		Args     json.RawMessage `json:"args"`
	}
	var tmp *tmpCommand
	if err := json.Unmarshal(data, &tmp); err != nil {
		return nil
	}
	c.Name = tmp.Name
	c.Branches = tmp.Branches
	switch c.Name {
	case CommandNameIf:
		var args *CommandArgsIf
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameCallEvent:
		return fmt.Errorf("not implemented yet: %s", c.Name)
	case CommandNameWait:
		var args *CommandArgsWait
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameShowMessage:
		var args *CommandArgsShowMessage
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameShowChoices:
		var args *CommandArgsShowChoices
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetSwitch:
		var args *CommandArgsSetSwitch
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetSelfSwitch:
		var args *CommandArgsSetSelfSwitch
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetVariable:
		var args *CommandArgsSetVariable
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameTransfer:
		var args *CommandArgsTransfer
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetRoute:
		return fmt.Errorf("not implemented yet: %s", c.Name)
	case CommandNameTintScreen:
		var args *CommandArgsTintScreen
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNamePlaySE:
		return fmt.Errorf("not implemented yet: %s", c.Name)
	case CommandNamePlayBGM:
		return fmt.Errorf("not implemented yet: %s", c.Name)
	case CommandNameStopBGM:
		return fmt.Errorf("not implemented yet: %s", c.Name)
	}
	return nil
}

type CommandName string

const (
	CommandNameIf            CommandName = "if"
	CommandNameCallEvent                 = "call_event"
	CommandNameWait                      = "wait"
	CommandNameShowMessage               = "show_message"
	CommandNameShowChoices               = "show_choices"
	CommandNameSetSwitch                 = "set_switch"
	CommandNameSetSelfSwitch             = "set_self_switch"
	CommandNameSetVariable               = "set_variable"
	CommandNameTransfer                  = "transfer"
	CommandNameSetRoute                  = "set_route"
	CommandNameTintScreen                = "tint_screen"
	CommandNamePlaySE                    = "play_se"
	CommandNamePlayBGM                   = "play_bgm"
	CommandNameStopBGM                   = "stop_bgm"
)

type CommandArgsIf struct {
	Conditions []*Condition `json:"conditions"`
}

type CommandArgsWait struct {
	Time int `json:"time"`
}

type CommandArgsShowMessage struct {
	EventID   int  `json:"eventId"`
	ContentID UUID `json:"content"`
}

type CommandArgsShowChoices struct {
	ChoiceIDs []UUID `json:"choices"`
}

type CommandArgsSetSwitch struct {
	ID    int  `json:"id"`
	Value bool `json:"value"`
}

type CommandArgsSetSelfSwitch struct {
	ID    int  `json:"id"`
	Value bool `json:"value"`
}

type CommandArgsSetVariable struct {
	ID        int                  `json:"id"`
	Op        SetVariableOp        `json:"op"`
	ValueType SetVariableValueType `json:"valueType"`
	Value     interface{}          `json:"value"`
}

type CommandArgsTransfer struct {
	RoomID int `json:"roomId"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

type CommandArgsTintScreen struct {
	Red   int  `json:"red"`
	Green int  `json:"green"`
	Blue  int  `json:"blue"`
	Gray  int  `json:"gray"`
	Time  int  `json:"time"`
	Wait  bool `json:"wait"`
}

type SetVariableOp string

const (
	SetVariableOpAssign SetVariableOp = "="
	SetVariableOpAdd                  = "+"
	SetVariableOpSub                  = "-"
	SetVariableOpMul                  = "*"
	SetVariableOpDiv                  = "/"
	SetVariableOpMod                  = "%"
)

type SetVariableValueType string

const (
	SetVariableValueTypeConstant  SetVariableValueType = "constant"
	SetVariableValueTypeVariable                       = "variable"
	SetVariableValueTypeRandom                         = "random"
	SetVariableValueTypeCharacter                      = "character"
)

type SetVariableCharacterType string

const (
	SetVariableCharacterTypeDirection SetVariableCharacterType = "direction"
)

type Condition struct {
	Type      ConditionType      `json:"type"`
	ID        int                `json:"id"`
	Comp      ConditionComp      `json:"comp"`
	ValueType ConditionValueType `json:"valueType"`
	Value     interface{}        `json:"value"`
}

/*func ConditionsFromMaps(m []interface{}) ([]*Condition, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var c []*Condition
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return c, nil
}*/

type ConditionType string

const (
	ConditionTypeSwitch     ConditionType = "switch"
	ConditionTypeSelfSwitch               = "self_switch"
	ConditionTypeVariable                 = "variable"
)

type ConditionComp string

const (
	ConditionCompEqualTo              ConditionComp = "=="
	ConditionCompNotEqualTo                         = "!="
	ConditionCompGreaterThanOrEqualTo               = ">="
	ConditionCompGreaterThan                        = ">"
	ConditionCompLessThanOrEqualTo                  = "<="
	ConditionCompLessThan                           = "<"
)

type ConditionValueType string

const (
	ConditionValueTypeConstant ConditionValueType = "constant"
	ConditionValueTypeVariable                    = "variable"
)

type Route struct {
	EventID int  `json:"eventId"`
	Repeat  bool `json:"repeat"`
	Skip    bool `json:"skip"`
	Wait    bool `json:"wait"`
	//Commands []*Command `json:"commands"`
}
