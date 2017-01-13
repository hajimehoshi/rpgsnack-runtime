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
	"fmt"
	"log"
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
	if err := unmarshalJSON(data, &tmp); err != nil {
		return nil
	}
	c.Name = tmp.Name
	c.Branches = tmp.Branches
	switch c.Name {
	case CommandNameIf:
		var args *CommandArgsIf
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameLabel:
		var args *CommandArgsLabel
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameGoto:
		var args *CommandArgsGoto
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameCallEvent:
		var args *CommandArgsCallEvent
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameWait:
		var args *CommandArgsWait
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameShowMessage:
		var args *CommandArgsShowMessage
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameShowChoices:
		var args *CommandArgsShowChoices
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetSwitch:
		var args *CommandArgsSetSwitch
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetSelfSwitch:
		var args *CommandArgsSetSelfSwitch
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetVariable:
		var args *CommandArgsSetVariable
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameTransfer:
		var args *CommandArgsTransfer
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetRoute:
		var args *CommandArgsSetRoute
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameTintScreen:
		var args *CommandArgsTintScreen
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNamePlaySE:
		log.Printf("not implemented command: %s", c.Name)
	case CommandNamePlayBGM:
		log.Printf("not implemented command: %s", c.Name)
	case CommandNameStopBGM:
		log.Printf("not implemented command: %s", c.Name)
	case CommandNameGotoTitle:
	case CommandNameMoveCharacter:
		var args *CommandArgsMoveCharacter
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameTurnCharacter:
		var args *CommandArgsTurnCharacter
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameRotateCharacter:
		var args *CommandArgsRotateCharacter
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetCharacterProperty:
		var args *CommandArgsSetCharacterProperty
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetCharacterImage:
		var args *CommandArgsSetCharacterImage
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSetInnerVariable:
		fallthrough
	default:
		return fmt.Errorf("data: invalid command: %s", c.Name)
	}
	return nil
}

type CommandName string

const (
	CommandNameIf            CommandName = "if"
	CommandNameLabel         CommandName = "label"
	CommandNameGoto          CommandName = "goto"
	CommandNameCallEvent     CommandName = "call_event"
	CommandNameWait          CommandName = "wait"
	CommandNameShowMessage   CommandName = "show_message"
	CommandNameShowChoices   CommandName = "show_choices"
	CommandNameSetSwitch     CommandName = "set_switch"
	CommandNameSetSelfSwitch CommandName = "set_self_switch"
	CommandNameSetVariable   CommandName = "set_variable"
	CommandNameTransfer      CommandName = "transfer"
	CommandNameSetRoute      CommandName = "set_route"
	CommandNameTintScreen    CommandName = "tint_screen"
	CommandNamePlaySE        CommandName = "play_se"
	CommandNamePlayBGM       CommandName = "play_bgm"
	CommandNameStopBGM       CommandName = "stop_bgm"
	CommandNameGotoTitle     CommandName = "goto_title"

	// Route commands
	CommandNameMoveCharacter        CommandName = "move_character"
	CommandNameTurnCharacter        CommandName = "turn_character"
	CommandNameRotateCharacter      CommandName = "rotate_character"
	CommandNameSetCharacterProperty CommandName = "set_character_property"
	CommandNameSetCharacterImage    CommandName = "set_character_image"

	// Special commands
	CommandNameSetInnerVariable CommandName = "set_inner_variable"
)

type CommandArgsIf struct {
	Conditions []*Condition `json:"conditions"`
}

type CommandArgsLabel struct {
	Name string `json:"name"`
}

type CommandArgsGoto struct {
	Label string `json:"label"`
}

type CommandArgsCallEvent struct {
	EventID   int `json:"eventId"`
	PageIndex int `json:"pageIndex"`
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
	ID        int
	Op        SetVariableOp
	ValueType SetVariableValueType
	Value     interface{}
}

func (c *CommandArgsSetVariable) UnmarshalJSON(data []uint8) error {
	type tmpCommandArgsSetVariable struct {
		ID        int                  `json:"id"`
		Op        SetVariableOp        `json:"op"`
		ValueType SetVariableValueType `json:"valueType"`
		Value     json.RawMessage      `json:"value"`
	}
	var tmp *tmpCommandArgsSetVariable
	if err := unmarshalJSON(data, &tmp); err != nil {
		return err
	}
	c.ID = tmp.ID
	c.Op = tmp.Op
	c.ValueType = tmp.ValueType
	switch c.ValueType {
	case SetVariableValueTypeConstant:
		v := 0
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetVariableValueTypeVariable:
		v := 0
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetVariableValueTypeRandom:
		var v *SetVariableValueRandom
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetVariableValueTypeCharacter:
		var v *SetVariableCharacterArgs
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	default:
		return fmt.Errorf("data: invalid type: %s", c.ValueType)
	}
	return nil
}

type CommandArgsTransfer struct {
	RoomID int `json:"roomId"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

type CommandArgsSetRoute struct {
	EventID  int        `json:"eventId"`
	Repeat   bool       `json:"repeat"`
	Skip     bool       `json:"skip"`
	Wait     bool       `json:"wait"`
	Commands []*Command `json:"commands"`
}

type CommandArgsTintScreen struct {
	Red   int  `json:"red"`
	Green int  `json:"green"`
	Blue  int  `json:"blue"`
	Gray  int  `json:"gray"`
	Time  int  `json:"time"`
	Wait  bool `json:"wait"`
}

type CommandArgsMoveCharacter struct {
	Type     MoveCharacterType `json:type`
	Dir      Dir               `json:dir`
	Distance int               `json:distance`
	X        int               `json:x`
	Y        int               `json:y`
}

type CommandArgsTurnCharacter struct {
	Dir Dir `json:dir`
}

type CommandArgsRotateCharacter struct {
	Angle int `json:angle`
}

type CommandArgsSetCharacterProperty struct {
	Type  SetCharacterPropertyType
	Value interface{}
}

func (c *CommandArgsSetCharacterProperty) UnmarshalJSON(data []uint8) error {
	type tmpCommandArgsSetCharacterProperty struct {
		Type  SetCharacterPropertyType `json:"type"`
		Value json.RawMessage          `json:"value"`
	}
	var tmp *tmpCommandArgsSetCharacterProperty
	if err := unmarshalJSON(data, &tmp); err != nil {
		return err
	}
	c.Type = tmp.Type
	switch c.Type {
	case SetCharacterPropertyTypeVisibility:
		v := false
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetCharacterPropertyTypeDirFix:
		v := false
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetCharacterPropertyTypeStepping:
		v := false
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetCharacterPropertyTypeThrough:
		v := false
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetCharacterPropertyTypeWalking:
		v := false
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetCharacterPropertyTypeSpeed:
		var v Speed
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	default:
		return fmt.Errorf("data: invalid type: %s", c.Type)
	}
	return nil
}

type CommandArgsSetCharacterImage struct {
	Image          string `json:"image"`
	ImageIndex     int    `json:"imageIndex"`
	Frame          int    `json:"frame"`
	Dir            Dir    `json:"dir"`
	UseFrameAndDir bool   `json:"useFrameAndDir"`
}

type CommandArgsSetInnerVariable struct {
	Name  string `json:name`
	Value int    `json:value`
}

type SetVariableOp string

const (
	SetVariableOpAssign SetVariableOp = "="
	SetVariableOpAdd    SetVariableOp = "+"
	SetVariableOpSub    SetVariableOp = "-"
	SetVariableOpMul    SetVariableOp = "*"
	SetVariableOpDiv    SetVariableOp = "/"
	SetVariableOpMod    SetVariableOp = "%"
)

type SetVariableValueType string

const (
	SetVariableValueTypeConstant  SetVariableValueType = "constant"
	SetVariableValueTypeVariable  SetVariableValueType = "variable"
	SetVariableValueTypeRandom    SetVariableValueType = "random"
	SetVariableValueTypeCharacter SetVariableValueType = "character"
)

type SetVariableValueRandom struct {
	Begin int `json:"begin"`
	End   int `json:"end"`
}

type SetVariableCharacterArgs struct {
	Type    SetVariableCharacterType `json:"type"`
	EventID int                      `json:"eventId"`
}

type SetVariableCharacterType string

const (
	SetVariableCharacterTypeDirection SetVariableCharacterType = "direction"
)

type MoveCharacterType string

const (
	MoveCharacterTypeDirection MoveCharacterType = "direction"
	MoveCharacterTypeTarget    MoveCharacterType = "target"
	MoveCharacterTypeForward   MoveCharacterType = "forward"
	MoveCharacterTypeBackward  MoveCharacterType = "backward"
	MoveCharacterTypeToward    MoveCharacterType = "toward"
	MoveCharacterTypeRandom    MoveCharacterType = "random"
)

type SetCharacterPropertyType string

const (
	SetCharacterPropertyTypeVisibility SetCharacterPropertyType = "visibility"
	SetCharacterPropertyTypeDirFix     SetCharacterPropertyType = "dir_fix"
	SetCharacterPropertyTypeStepping   SetCharacterPropertyType = "stepping"
	SetCharacterPropertyTypeThrough    SetCharacterPropertyType = "through"
	SetCharacterPropertyTypeWalking    SetCharacterPropertyType = "walking"
	SetCharacterPropertyTypeSpeed      SetCharacterPropertyType = "speed"
)
