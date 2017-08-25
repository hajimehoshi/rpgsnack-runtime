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

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

type Command struct {
	Name     CommandName
	Args     CommandArgs
	Branches [][]*Command
}

type CommandArgs interface{}

func (c *Command) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("name")
	e.EncodeString(string(c.Name))

	e.EncodeString("args")
	e.EncodeAny(c.Args)

	e.EncodeString("branches")
	e.BeginArray()
	for _, b := range c.Branches {
		e.BeginArray()
		for _, command := range b {
			e.EncodeInterface(command)
		}
		e.EndArray()
	}
	e.EndArray()

	e.EndMap()
	return e.Flush()
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
	case CommandNameNop:
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
	case CommandNameCallCommonEvent:
		var args *CommandArgsCallCommonEvent
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameReturn:
	case CommandNameEraseEvent:
	case CommandNameWait:
		var args *CommandArgsWait
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameShowBalloon:
		var args *CommandArgsShowBalloon
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameShowMessage:
		var args *CommandArgsShowMessage
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		if args.TextAlign == "" {
			args.TextAlign = TextAlignLeft
		}
		c.Args = args
	case CommandNameShowHint:
		var args *CommandArgsShowHint
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
		var args *CommandArgsPlaySE
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNamePlayBGM:
		var args *CommandArgsPlayBGM
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameStopBGM:
	case CommandNameSave:
	case CommandNameGotoTitle:
	case CommandNameSyncIAP:
	case CommandNameUnlockAchievement:
		var args *CommandArgsUnlockAchievement
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameAutoSave:
		var args *CommandArgsAutoSave
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNamePlayerControl:
		var args *CommandArgsPlayerControl
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameControlHint:
		var args *CommandArgsControlHint
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNamePurchase:
		var args *CommandArgsPurchase
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameShowAds:
		var args *CommandArgsShowAds
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameOpenLink:
		var args *CommandArgsOpenLink
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
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
	case CommandNameSetCharacterOpacity:
		var args *CommandArgsSetCharacterOpacity
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameAddItem:
		var args *CommandArgsAddItem
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameRemoveItem:
		var args *CommandArgsRemoveItem
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameShowPicture:
		var args *CommandArgsShowPicture
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	default:
		return fmt.Errorf("data: invalid command: %s", c.Name)
	}
	return nil
}

func (c *Command) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "name":
			c.Name = CommandName(d.DecodeString())
		case "args":
			if c.Name == "" {
				return fmt.Errorf("data: 'name' should come before 'args'")
			}
			switch c.Name {
			case CommandNameNop:
				d.DecodeNil()
			case CommandNameIf:
				c.Args = &CommandArgsIf{}
				d.DecodeAny(c.Args)
			case CommandNameLabel:
				c.Args = &CommandArgsLabel{}
				d.DecodeAny(c.Args)
			case CommandNameGoto:
				c.Args = &CommandArgsGoto{}
				d.DecodeAny(c.Args)
			case CommandNameCallEvent:
				c.Args = &CommandArgsCallEvent{}
				d.DecodeAny(c.Args)
			case CommandNameCallCommonEvent:
				c.Args = &CommandArgsCallCommonEvent{}
				d.DecodeAny(c.Args)
			case CommandNameReturn:
				d.DecodeNil()
			case CommandNameEraseEvent:
				d.DecodeNil()
			case CommandNameWait:
				c.Args = &CommandArgsWait{}
				d.DecodeAny(c.Args)
			case CommandNameShowBalloon:
				args := &CommandArgsShowBalloon{}
				d.DecodeAny(c.Args)
				c.Args = args
			case CommandNameShowMessage:
				args := &CommandArgsShowMessage{}
				d.DecodeAny(c.Args)
				if args.TextAlign == "" {
					args.TextAlign = TextAlignLeft
				}
				c.Args = args
			case CommandNameShowHint:
				c.Args = &CommandArgsShowHint{}
				d.DecodeAny(c.Args)
			case CommandNameShowChoices:
				c.Args = &CommandArgsShowChoices{}
				d.DecodeAny(c.Args)
			case CommandNameSetSwitch:
				c.Args = &CommandArgsSetSwitch{}
				d.DecodeAny(c.Args)
			case CommandNameSetSelfSwitch:
				c.Args = &CommandArgsSetSelfSwitch{}
				d.DecodeAny(c.Args)
			case CommandNameSetVariable:
				c.Args = &CommandArgsSetVariable{}
				d.DecodeAny(c.Args)
			case CommandNameTransfer:
				c.Args = &CommandArgsTransfer{}
				d.DecodeAny(c.Args)
			case CommandNameSetRoute:
				c.Args = &CommandArgsSetRoute{}
				d.DecodeAny(c.Args)
			case CommandNameTintScreen:
				c.Args = &CommandArgsTintScreen{}
				d.DecodeAny(c.Args)
			case CommandNamePlaySE:
				c.Args = &CommandArgsPlaySE{}
				d.DecodeAny(c.Args)
			case CommandNamePlayBGM:
				c.Args = &CommandArgsPlayBGM{}
				d.DecodeAny(c.Args)
			case CommandNameStopBGM:
				d.DecodeNil()
			case CommandNameSave:
				d.DecodeNil()
			case CommandNameGotoTitle:
				d.DecodeNil()
			case CommandNameSyncIAP:
				d.DecodeNil()
			case CommandNameUnlockAchievement:
				c.Args = &CommandArgsUnlockAchievement{}
				d.DecodeAny(c.Args)
			case CommandNameAutoSave:
				c.Args = &CommandArgsAutoSave{}
				d.DecodeAny(c.Args)
			case CommandNamePlayerControl:
				c.Args = &CommandArgsPlayerControl{}
				d.DecodeAny(c.Args)
			case CommandNameControlHint:
				c.Args = &CommandArgsControlHint{}
				d.DecodeAny(c.Args)
			case CommandNamePurchase:
				c.Args = &CommandArgsPurchase{}
				d.DecodeAny(c.Args)
			case CommandNameShowAds:
				c.Args = &CommandArgsShowAds{}
				d.DecodeAny(c.Args)
			case CommandNameOpenLink:
				c.Args = &CommandArgsOpenLink{}
				d.DecodeAny(c.Args)
			case CommandNameMoveCharacter:
				c.Args = &CommandArgsMoveCharacter{}
				d.DecodeAny(c.Args)
			case CommandNameTurnCharacter:
				c.Args = &CommandArgsTurnCharacter{}
				d.DecodeAny(c.Args)
			case CommandNameRotateCharacter:
				c.Args = &CommandArgsRotateCharacter{}
				d.DecodeAny(c.Args)
			case CommandNameSetCharacterProperty:
				c.Args = &CommandArgsSetCharacterProperty{}
				d.DecodeAny(c.Args)
			case CommandNameSetCharacterImage:
				c.Args = &CommandArgsSetCharacterImage{}
				d.DecodeAny(c.Args)
			case CommandNameSetCharacterOpacity:
				c.Args = &CommandArgsSetCharacterOpacity{}
				d.DecodeAny(c.Args)
			case CommandNameAddItem:
				c.Args = &CommandArgsAddItem{}
				d.DecodeAny(c.Args)
			case CommandNameRemoveItem:
				c.Args = &CommandArgsRemoveItem{}
				d.DecodeAny(c.Args)
			case CommandNameShowPicture:
				c.Args = &CommandArgsShowPicture{}
				d.DecodeAny(c.Args)
			default:
				return fmt.Errorf("data: Command.DecodeMsgpack: invalid command: %s", c.Name)
			}
		case "branches":
			if d.SkipCodeIfNil() {
				continue
			}
			n := d.DecodeArrayLen()
			c.Branches = make([][]*Command, n)
			for i := 0; i < n; i++ {
				if d.SkipCodeIfNil() {
					continue
				}
				n := d.DecodeArrayLen()
				c.Branches[i] = make([]*Command, n)
				for j := 0; j < n; j++ {
					if d.SkipCodeIfNil() {
						continue
					}
					c.Branches[i][j] = &Command{}
					d.DecodeInterface(c.Branches[i][j])
				}
			}
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: Command.DecodeMsgpack failed: %v", err)
	}
	return nil
}

type CommandName string

const (
	CommandNameNop               CommandName = "nop"
	CommandNameIf                CommandName = "if"
	CommandNameLabel             CommandName = "label"
	CommandNameGoto              CommandName = "goto"
	CommandNameCallEvent         CommandName = "call_event"
	CommandNameCallCommonEvent   CommandName = "call_common_event"
	CommandNameReturn            CommandName = "return"
	CommandNameEraseEvent        CommandName = "erase_event"
	CommandNameWait              CommandName = "wait"
	CommandNameShowBalloon       CommandName = "show_balloon"
	CommandNameShowMessage       CommandName = "show_message"
	CommandNameShowHint          CommandName = "show_hint"
	CommandNameShowChoices       CommandName = "show_choices"
	CommandNameSetSwitch         CommandName = "set_switch"
	CommandNameSetSelfSwitch     CommandName = "set_self_switch"
	CommandNameSetVariable       CommandName = "set_variable"
	CommandNameTransfer          CommandName = "transfer"
	CommandNameSetRoute          CommandName = "set_route"
	CommandNameTintScreen        CommandName = "tint_screen"
	CommandNamePlaySE            CommandName = "play_se"
	CommandNamePlayBGM           CommandName = "play_bgm"
	CommandNameStopBGM           CommandName = "stop_bgm"
	CommandNameSave              CommandName = "save"
	CommandNameGotoTitle         CommandName = "goto_title"
	CommandNameAutoSave          CommandName = "autosave"
	CommandNameGameClear         CommandName = "game_clear"
	CommandNamePlayerControl     CommandName = "player_control"
	CommandNameUnlockAchievement CommandName = "unlock_achievement"
	CommandNameControlHint       CommandName = "control_hint"
	CommandNamePurchase          CommandName = "start_iap"
	CommandNameSyncIAP           CommandName = "sync_iap" // TODO: We might be able to remove this later
	CommandNameShowAds           CommandName = "show_ads"
	CommandNameOpenLink          CommandName = "open_link"

	CommandNameAddItem    CommandName = "add_item"
	CommandNameRemoveItem CommandName = "remove_item"

	CommandNameShowPicture CommandName = "show_picture"

	// Route commands
	CommandNameMoveCharacter        CommandName = "move_character"
	CommandNameTurnCharacter        CommandName = "turn_character"
	CommandNameRotateCharacter      CommandName = "rotate_character"
	CommandNameSetCharacterProperty CommandName = "set_character_property"
	CommandNameSetCharacterImage    CommandName = "set_character_image"
	CommandNameSetCharacterOpacity  CommandName = "set_character_opacity"

	// Special commands
	CommandNameFinishPlayerMovingByUserInput CommandName = "finish_player_moving_by_user_input"
	CommandNameExecEventHere                 CommandName = "exec_event_here"
)

type CommandArgsIf struct {
	Conditions []*Condition `json:"conditions" msgpack:"conditions"`
}

type CommandArgsLabel struct {
	Name string `json:"name" msgpack:"name"`
}

type CommandArgsGoto struct {
	Label string `json:"label" msgpack:"label"`
}

type CommandArgsCallEvent struct {
	EventID   int `json:"eventId" msgpack:"eventId"`
	PageIndex int `json:"pageIndex" msgpack:"pageIndex"`
}

type CommandArgsCallCommonEvent struct {
	EventID int `json:"eventId" msgpack:"eventId"`
}

type CommandArgsWait struct {
	Time int `json:"time" msgpack:"time"`
}

type CommandArgsShowBalloon struct {
	EventID     int         `json:"eventId" msgpack:"eventId"`
	ContentID   UUID        `json:"content" msgpack:"content"`
	BalloonType BalloonType `json:"balloonType" msgpack:"balloonType"`
}

type CommandArgsShowMessage struct {
	ContentID    UUID                `json:"content" msgpack:"content"`
	Background   MessageBackground   `json:"background" msgpack:"background"`
	PositionType MessagePositionType `json:"positionType" msgpack:"positionType"`
	TextAlign    TextAlign           `json:"textAlign" msgpack:"textAlign"`
}

type CommandArgsShowHint struct {
	ContentID    UUID                `json:"content" msgpack:"content"`
	Background   MessageBackground   `json:"textAlign" msgpack:"background"`
	PositionType MessagePositionType `json:"positionType" msgpack:"positionType"`
	TextAlign    TextAlign           `json:"textAlign" msgpack:"textAlign"`
}

type CommandArgsShowChoices struct {
	ChoiceIDs []UUID `json:"choices" msgpack:"choices"`
}

type CommandArgsSetSwitch struct {
	ID    int  `json:"id" msgpack:"id"`
	Value bool `json:"value" msgpack:"value"`
}

type CommandArgsSetSelfSwitch struct {
	ID    int  `json:"id" msgpack:"id"`
	Value bool `json:"value" msgpack:"value"`
}

type CommandArgsSetVariable struct {
	ID        int                  `json:"id"`
	Op        SetVariableOp        `json:"op"`
	ValueType SetVariableValueType `json:"valueType"`
	Value     interface{}          `json:"value"`
}

func (c *CommandArgsSetVariable) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("id")
	e.EncodeInt(c.ID)

	e.EncodeString("op")
	e.EncodeString(string(c.Op))

	e.EncodeString("valueType")
	e.EncodeString(string(c.ValueType))

	e.EncodeString("value")
	switch c.ValueType {
	case SetVariableValueTypeConstant:
		e.EncodeInt(c.Value.(int))
	case SetVariableValueTypeVariable:
		e.EncodeInt(c.Value.(int))
	case SetVariableValueTypeRandom:
		e.EncodeAny(c.Value)
	case SetVariableValueTypeCharacter:
		e.EncodeAny(c.Value)
	case SetVariableValueTypeIAPProduct:
		e.EncodeInt(c.Value.(int))
	case SetVariableValueTypeSystem:
		e.EncodeString(string(c.Value.(SystemVariableType)))
	default:
		return fmt.Errorf("data: CommandArgsSetVariable.EncodeMsgpack: invalid type: %s", c.ValueType)
	}

	e.EndMap()
	return e.Flush()
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
	case SetVariableValueTypeIAPProduct:
		v := 0
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	case SetVariableValueTypeSystem:
		var v SystemVariableType
		if err := unmarshalJSON(tmp.Value, &v); err != nil {
			return err
		}
		c.Value = v
	default:
		return fmt.Errorf("data: invalid type: %s", c.ValueType)
	}
	return nil
}

func (c *CommandArgsSetVariable) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "id":
			c.ID = d.DecodeInt()
		case "op":
			c.Op = SetVariableOp(d.DecodeString())
		case "valueType":
			c.ValueType = SetVariableValueType(d.DecodeString())
		case "value":
			switch c.ValueType {
			case SetVariableValueTypeConstant:
				c.Value = d.DecodeInt()
			case SetVariableValueTypeVariable:
				c.Value = d.DecodeInt()
			case SetVariableValueTypeRandom:
				if !d.SkipCodeIfNil() {
					v := &SetVariableValueRandom{}
					d.DecodeAny(v)
					c.Value = v
				}
			case SetVariableValueTypeCharacter:
				if !d.SkipCodeIfNil() {
					v := &SetVariableCharacterArgs{}
					d.DecodeAny(v)
					c.Value = v
				}
			case SetVariableValueTypeIAPProduct:
				c.Value = d.DecodeInt()
			case SetVariableValueTypeSystem:
				c.Value = SystemVariableType(d.DecodeString())
			default:
				return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: invalid type: %s", c.ValueType)
			}
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack failed: %v", err)
	}
	return nil
}

type CommandArgsTransfer struct {
	ValueType  TransferValueType      `json:"valueType" msgpack:"valueType"`
	RoomID     int                    `json:"roomId" msgpack:"roomId"`
	X          int                    `json:"x" msgpack:"x"`
	Y          int                    `json:"y" msgpack:"y"`
	Transition TransferTransitionType `json:"transition" msgpack:"transition"`
}

type CommandArgsSetRoute struct {
	EventID  int        `json:"eventId" msgpack:"eventId"`
	Repeat   bool       `json:"repeat" msgpack:"repeat"`
	Skip     bool       `json:"skip" msgpack:"skip"`
	Wait     bool       `json:"wait" msgpack:"wait"`
	Commands []*Command `json:"commands" msgpack:"commands"`
}

type CommandArgsTintScreen struct {
	Red   int  `json:"red" msgpack:"red"`
	Green int  `json:"green" msgpack:"green"`
	Blue  int  `json:"blue" msgpack:"blue"`
	Gray  int  `json:"gray" msgpack:"gray"`
	Time  int  `json:"time" msgpack:"time"`
	Wait  bool `json:"wait" msgpack:"wait"`
}

type CommandArgsPlaySE struct {
	Name   string `json:"name" msgpack:"name"`
	Volume int    `json:"volume" msgpack:"volume"`
}

type CommandArgsPlayBGM struct {
	Name     string `json:"name" msgpack:"name"`
	Volume   int    `json:"volume" msgpack:"volume"`
	FadeTime int    `json:"fadeTime" msgpack:"fadeTime"`
}

type CommandArgsUnlockAchievement struct {
	ID int `json:"id" msgpack:"id"`
}

type CommandArgsControlHint struct {
	ID   int             `json:"id" msgpack:"id"`
	Type ControlHintType `json:"type" msgpack:"type"`
}

type CommandArgsPurchase struct {
	ID int `json:"id" msgpack:"id"`
}

type CommandArgsShowAds struct {
	Type ShowAdsType `json:"type" msgpack:"type"`
}

type CommandArgsOpenLink struct {
	Type string `json:"type" msgpack:"type"`
	Data string `json:"data" msgpack:"data"`
}

type CommandArgsAutoSave struct {
	Enabled bool `json:"enabled" msgpack:"enabled"`
}

type CommandArgsPlayerControl struct {
	Enabled bool `json:"enabled" msgpack:"enabled"`
}

type CommandArgsMoveCharacter struct {
	Type             MoveCharacterType   `json:"type" msgpack:"type"`
	Dir              Dir                 `json:"dir" msgpack:"dir"`
	Distance         int                 `json:"distance" msgpack:"distance"`
	X                int                 `json:"x" msgpack:"x"`
	Y                int                 `json:"y" msgpack:"y"`
	ValueType        MoveTargetValueType `json:"valueType" msgpack:"valueType"`
	IgnoreCharacters bool                `json:"ignoreCharacters" msgpack:"ignoreCharacters"`
}

func (c *CommandArgsMoveCharacter) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("type")
	e.EncodeString(string(c.Type))

	e.EncodeString("dir")
	e.EncodeInt(int(c.Dir))

	e.EncodeString("distance")
	e.EncodeInt(c.Distance)

	e.EncodeString("x")
	e.EncodeInt(c.X)

	e.EncodeString("y")
	e.EncodeInt(c.Y)

	e.EndMap()
	return e.Flush()
}

func (c *CommandArgsMoveCharacter) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		k := d.DecodeString()
		switch k {
		case "type":
			c.Type = MoveCharacterType(d.DecodeString())
		case "dir":
			c.Dir = Dir(d.DecodeInt())
		case "distance":
			c.Distance = d.DecodeInt()
		case "x":
			c.X = d.DecodeInt()
		case "y":
			c.Y = d.DecodeInt()
		case "considerCharacters":
			d.Skip()
		default:
			return fmt.Errorf("data: CommandArgsMoveCharacter.DecodeMsgpack: invalid key: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: CommandArgsMoveCharacter.DecodeMsgpack failed: %v", err)
	}
	return nil
}

type CommandArgsTurnCharacter struct {
	Dir Dir `json:dir msgpack:"dir"`
}

type CommandArgsRotateCharacter struct {
	Angle int `json:angle msgpack:"angle"`
}

type CommandArgsSetCharacterProperty struct {
	Type  SetCharacterPropertyType `json:"type"`
	Value interface{}              `json:"value"`
}

type CommandArgsSetCharacterOpacity struct {
	Opacity int  `json:"opacity"`
	Time    int  `json:"time"`
	Wait    bool `json:"wait"`
}

func (c *CommandArgsSetCharacterProperty) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("type")
	e.EncodeString(string(c.Type))

	e.EncodeString("value")
	switch c.Type {
	case SetCharacterPropertyTypeVisibility:
		e.EncodeBool(c.Value.(bool))
	case SetCharacterPropertyTypeDirFix:
		e.EncodeBool(c.Value.(bool))
	case SetCharacterPropertyTypeStepping:
		e.EncodeBool(c.Value.(bool))
	case SetCharacterPropertyTypeThrough:
		e.EncodeBool(c.Value.(bool))
	case SetCharacterPropertyTypeWalking:
		e.EncodeBool(c.Value.(bool))
	case SetCharacterPropertyTypeSpeed:
		e.EncodeInt(int(c.Value.(Speed)))
	}

	e.EndMap()
	return e.Flush()
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

func (c *CommandArgsSetCharacterProperty) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch d.DecodeString() {
		case "type":
			c.Type = SetCharacterPropertyType(d.DecodeString())
		case "value":
			switch c.Type {
			case SetCharacterPropertyTypeVisibility:
				c.Value = d.DecodeBool()
			case SetCharacterPropertyTypeDirFix:
				c.Value = d.DecodeBool()
			case SetCharacterPropertyTypeStepping:
				c.Value = d.DecodeBool()
			case SetCharacterPropertyTypeThrough:
				c.Value = d.DecodeBool()
			case SetCharacterPropertyTypeWalking:
				c.Value = d.DecodeBool()
			case SetCharacterPropertyTypeSpeed:
				c.Value = Speed(d.DecodeInt())
			default:
				return fmt.Errorf("data: CommandArgsSetCharacterProperty.DecodeMsgpack: invalid type: %s", c.Type)
			}
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: CommandArgsSetCharacterProperty.DecodeMsgpack failed: %v", err)
	}
	return nil
}

type CommandArgsSetCharacterImage struct {
	Image          string `json:"image" msgpack:"image"`
	ImageIndex     int    `json:"imageIndex" msgpack:"imageIndex"`
	Frame          int    `json:"frame" msgpack:"frame"`
	Dir            Dir    `json:"dir" msgpack:"dir"`
	UseFrameAndDir bool   `json:"useFrameAndDir" msgpack:"useFrameAndDir"`
}

type CommandArgsAddItem struct {
	ID int `json:"id" msgpack:"id"`
}

type CommandArgsRemoveItem struct {
	ID int `json:"id" msgpack:"id"`
}

type ShowPicturePosValueType string

const (
	ShowPicturePosValueTypeConstant ShowPicturePosValueType = "constant"
	ShowPicturePosValueTypeVariable ShowPicturePosValueType = "variable"
)

type ShowPictureOrigin string

const (
	ShowPictureOriginUpperLeft ShowPictureOrigin = "upper_left"
	ShowPictureOriginCenter    ShowPictureOrigin = "center"
)

type ShowPictureBlendType string

const (
	ShowPictureBlendTypeNormal ShowPictureBlendType = "normal"
	ShowPictureBlendTypeAdd    ShowPictureBlendType = "add"
)

type CommandArgsShowPicture struct {
	ID           int                     `json:"id" msgpack:"id"`
	X            int                     `json:"x" msgpack:"x"`
	Y            int                     `json:"y" msgpack:"y"`
	PosValueType ShowPicturePosValueType `json:"posValueType" msgpack:"posValueType"`
	scaleX       int                     `json:"scaleX" msgpack:"scaleX"`
	scaleY       int                     `json:"scaleY" msgpack:"scaleY"`
	opacity      int                     `json:"opacity" msgpack:"opacity"`
	angle        int                     `json:"angle" msgpack:"angle"`
	origin       ShowPictureOrigin       `json:"origin" msgpack:"origin"`
	blendType    ShowPictureBlendType    `json:"blendType" msgpack:"blendType"`
	wait         bool                    `json:"wait" msgpack:"wait"`
	time         int                     `json:"time" msgpack:"time"`
	name         string                  `json:"name" msgpack:"name"`
}

type SetVariableOp string

const (
	SetVariableOpAssign SetVariableOp = "=" // TODO: Rename
	SetVariableOpAdd    SetVariableOp = "+"
	SetVariableOpSub    SetVariableOp = "-"
	SetVariableOpMul    SetVariableOp = "*"
	SetVariableOpDiv    SetVariableOp = "/"
	SetVariableOpMod    SetVariableOp = "%"
)

type SetVariableValueType string

const (
	SetVariableValueTypeConstant   SetVariableValueType = "constant"
	SetVariableValueTypeVariable   SetVariableValueType = "variable"
	SetVariableValueTypeRandom     SetVariableValueType = "random"
	SetVariableValueTypeCharacter  SetVariableValueType = "character"
	SetVariableValueTypeIAPProduct SetVariableValueType = "iap_product"
	SetVariableValueTypeSystem     SetVariableValueType = "system"
)

type TransferValueType string

const (
	TransferValueTypeConstant TransferValueType = "constant"
	TransferValueTypeVariable TransferValueType = "variable"
)

type MoveTargetValueType string

const (
	MoveTargetValueTypeConstant MoveTargetValueType = "constant"
	MoveTargetValueTypeVariable MoveTargetValueType = "variable"
)

type TransferTransitionType string

const (
	TransferTransitionTypeNone  TransferTransitionType = "none"
	TransferTransitionTypeBlack TransferTransitionType = "black"
	TransferTransitionTypeWhite TransferTransitionType = "white"
)

type SetVariableValueRandom struct {
	Begin int `json:"begin" msgpack:"begin"`
	End   int `json:"end" msgpack:"end"`
}

// TODO: Rename?
type SetVariableCharacterArgs struct {
	Type    SetVariableCharacterType `json:"type" msgpack:"type"`
	EventID int                      `json:"eventId" msgpack:"eventId"`
}

type SetVariableSystem struct {
	Type    SetVariableCharacterType `json:"type" msgpack:"type"`
	EventID int                      `json:"eventId" msgpack:"eventId"`
}

type SetVariableCharacterType string

const (
	SetVariableCharacterTypeDirection SetVariableCharacterType = "direction"
	SetVariableCharacterTypeRoomX     SetVariableCharacterType = "room_x"
	SetVariableCharacterTypeRoomY     SetVariableCharacterType = "room_y"
	SetVariableCharacterTypeScreenX   SetVariableCharacterType = "screen_x"
	SetVariableCharacterTypeScreenY   SetVariableCharacterType = "screen_y"
)

type ShowAdsType string

const (
	ShowAdsTypeRewarded     ShowAdsType = "rewarded"
	ShowAdsTypeInterstitial ShowAdsType = "interstitial"
)

type MoveCharacterType string

const (
	MoveCharacterTypeDirection MoveCharacterType = "direction"
	MoveCharacterTypeTarget    MoveCharacterType = "target"
	MoveCharacterTypeForward   MoveCharacterType = "forward"
	MoveCharacterTypeBackward  MoveCharacterType = "backward"
	MoveCharacterTypeToward    MoveCharacterType = "toward"
	MoveCharacterTypeAgainst   MoveCharacterType = "against"
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

type ControlHintType string

const (
	ControlHintPause    ControlHintType = "pause"
	ControlHintStart    ControlHintType = "start"
	ControlHintComplete ControlHintType = "complete"
)

type TextAlign string

const (
	TextAlignLeft   TextAlign = "left"
	TextAlignCenter TextAlign = "center"
	TextAlignRight  TextAlign = "right"
)

type BalloonType string

const (
	BalloonTypeNormal BalloonType = "normal"
	BalloonTypeThink  BalloonType = "think"
	BalloonTypeShout  BalloonType = "shout"
)

type SystemVariableType string

const (
	SystemVariableInterstitialAdsLoaded SystemVariableType = "interstitial_ads_loaded"
	SystemVariableRewardedAdsLoaded     SystemVariableType = "rewarded_ads_loaded"
	SystemVariableHintCount             SystemVariableType = "active_hint_count"
	SystemVariableRoomID                SystemVariableType = "room_id"
)

type MessagePositionType string

const (
	MessagePositionBottom MessagePositionType = "bottom"
	MessagePositionMiddle MessagePositionType = "middle"
	MessagePositionTop    MessagePositionType = "top"
)

type MessageBackground string

const (
	MessageBackgroundDim         MessageBackground = "dim"
	MessageBackgroundTransparent MessageBackground = "transparent"
	MessageBackgroundBanner      MessageBackground = "banner"
)
