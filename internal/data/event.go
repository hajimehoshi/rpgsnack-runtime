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
	Conditions []*Condition         `json:"conditions"`
	Image      string               `json:"image"`
	ImageIndex int                  `json:"imageIndex"`
	Attitude   Attitude             `json:"attitude"`
	Dir        Dir                  `json:"dir"`
	DirFix     bool                 `json:"dirFix"`
	Walking    bool                 `json:"walking"`
	Stepping   bool                 `json:"stepping"`
	Through    bool                 `json:"through"`
	Priority   Priority             `json:"priority"`
	Trigger    Trigger              `json:"trigger"`
	Route      *CommandArgsSetRoute `json:"route"`
	Commands   []*Command           `json:"commands"`
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
		var args *CommandArgsCallEvent
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
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
		var args *CommandArgsSetRoute
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameTintScreen:
		var args *CommandArgsTintScreen
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNamePlaySE:
		return fmt.Errorf("data: not implemented yet: %s", c.Name)
	case CommandNamePlayBGM:
		return fmt.Errorf("data: not implemented yet: %s", c.Name)
	case CommandNameStopBGM:
		return fmt.Errorf("data: not implemented yet: %s", c.Name)
	case CommandNameMoveCharacter:
		var args *CommandArgsMoveCharacter
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameTurnCharacter:
		var args *CommandArgsTurnCharacter
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameRotateCharacter:
		var args *CommandArgsRotateCharacter
		if err := json.Unmarshal(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	default:
		return fmt.Errorf("data: invalid command: %s", c.Name)
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

	// Route commands
	CommandNameMoveCharacter   = "move_character"
	CommandNameTurnCharacter   = "turn_character"
	CommandNameRotateCharacter = "rotate_character"
)

type CommandArgsIf struct {
	Conditions []*Condition `json:"conditions"`
}

type CommandArgsCallEvent struct {
	EventID   int `json:"event_id"`
	PageIndex int `json:"page_index"`
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
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	c.ID = tmp.ID
	c.Op = tmp.Op
	c.ValueType = tmp.ValueType
	switch c.ValueType {
	case SetVariableValueTypeConstant:
		v := 0
		if err := json.Unmarshal(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetVariableValueTypeVariable:
		v := 0
		if err := json.Unmarshal(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetVariableValueTypeRandom:
		return fmt.Errorf("data: not implemented yet (set_variable): valueType %s", c.ValueType)
	case SetVariableValueTypeCharacter:
		var v *SetVariableCharacterArgs
		if err := json.Unmarshal(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
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
	Dir      Dir `json:dir`
	Distance int `json:distance`
}

type CommandArgsTurnCharacter struct {
	Dir Dir `json:dir`
}

type CommandArgsRotateCharacter struct {
	Angle int `json:angle`
}

/*
move_character: dir: (int), distance: (int)
turn_character: dir: (int)
rotate_character: angle: (number: 90/180/270)
set_character_property: type:(string:"visibility"/"dir_fix"/"stepping"/"through"/"walking"/"speed") value: (bool or int)
wait: value: (int)
set_character_image: image: (string), imageIndex: (int)
play_se: // tbd
*/

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

type SetVariableCharacterArgs struct {
	Type    SetVariableCharacterType `json:"type"`
	EventID int                      `json:"eventId"`
}

type SetVariableCharacterType string

const (
	SetVariableCharacterTypeDirection SetVariableCharacterType = "direction"
)
