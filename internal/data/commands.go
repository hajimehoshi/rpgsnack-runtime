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
	"fmt"
	"runtime"

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
)

type Command struct {
	Name     CommandName
	Args     CommandArgs
	Branches [][]*Command
	IsFolded bool
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

	e.EncodeString("isFolded")
	e.EncodeBool(c.IsFolded)

	e.EndMap()
	return e.Flush()
}

var commandUnmarshalingCount = 0

func (c *Command) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	var args interface{}
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "name":
			c.Name = CommandName(d.DecodeString())
		case "args":
			d.DecodeAny(&args)
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
		case "isFolded":
			d.DecodeBool()

		default:
			if err := d.Error(); err != nil {
				return fmt.Errorf("data: Command.DecodeMsgpack failed: %v", err)
			}
			return fmt.Errorf("data: Command.DecodeMsgpack: invalid command structure: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: Command.DecodeMsgpack failed: %v", err)
	}

	// TODO: Avoid re-encoding the arg
	argsBin, err := msgpack.Marshal(args)
	if err != nil {
		return err
	}

	switch c.Name {
	case CommandNameNop:
	case CommandNameMemo:
		a := &CommandArgsMemo{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameIf:
		a := &CommandArgsIf{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameGroup:
		a := &CommandArgsGroup{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameLabel:
		a := &CommandArgsLabel{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameGoto:
		a := &CommandArgsGoto{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameGotoTitle:
		a := &CommandArgsGotoTitle{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameCallEvent:
		a := &CommandArgsCallEvent{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameCallCommonEvent:
		a := &CommandArgsCallCommonEvent{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameReturn:
	case CommandNameEraseEvent:
	case CommandNameWait:
		a := &CommandArgsWait{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameShowBalloon:
		a := &CommandArgsShowBalloon{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameShowMessage:
		a := &CommandArgsShowMessage{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		if a.TextAlign == "" {
			a.TextAlign = TextAlignLeft
		}
		c.Args = a
	case CommandNameShowHint:
	case CommandNameShowChoices:
		a := &CommandArgsShowChoices{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSetSwitch:
		a := &CommandArgsSetSwitch{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSetSelfSwitch:
		a := &CommandArgsSetSelfSwitch{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSetVariable:
		a := &CommandArgsSetVariable{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSavePermanent:
		a := &CommandArgsSavePermanent{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameLoadPermanent:
		a := &CommandArgsLoadPermanent{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameTransfer:
		a := &CommandArgsTransfer{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSetRoute:
		a := &CommandArgsSetRoute{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameShake:
		a := &CommandArgsShake{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameTintScreen:
		a := &CommandArgsTintScreen{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNamePlaySE:
		a := &CommandArgsPlaySE{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNamePlayBGM:
		a := &CommandArgsPlayBGM{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameStopBGM:
		a := &CommandArgsStopBGM{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSave:
	case CommandNameRequestReview:
	case CommandNameUnlockAchievement:
		a := &CommandArgsUnlockAchievement{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameAutoSave:
		a := &CommandArgsAutoSave{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNamePlayerControl:
		a := &CommandArgsPlayerControl{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNamePlayerSpeed:
		a := &CommandArgsPlayerSpeed{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameWeather:
		a := &CommandArgsWeather{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameControlHint:
		a := &CommandArgsControlHint{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNamePurchase:
		a := &CommandArgsPurchase{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameShowAds:
		a := &CommandArgsShowAds{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameOpenLink:
		a := &CommandArgsOpenLink{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameShare:
		a := &CommandArgsShare{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSendAnalytics:
		a := &CommandArgsSendAnalytics{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameShowShop:
		a := &CommandArgsShowShop{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameShowMinigame:
		a := &CommandArgsShowMinigame{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameVibrate:
		a := &CommandArgsVibrate{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameMoveCharacter:
		a := &CommandArgsMoveCharacter{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameTurnCharacter:
		a := &CommandArgsTurnCharacter{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameRotateCharacter:
		a := &CommandArgsRotateCharacter{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSetCharacterProperty:
		a := &CommandArgsSetCharacterProperty{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSetCharacterImage:
		a := &CommandArgsSetCharacterImage{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSetCharacterOpacity:
		a := &CommandArgsSetCharacterOpacity{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameAddItem:
		a := &CommandArgsAddItem{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameRemoveItem:
		a := &CommandArgsRemoveItem{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameShowInventory:
		a := &CommandArgsShowInventory{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameHideInventory:
	case CommandNameShowItem:
		a := &CommandArgsShowItem{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameHideItem:
	case CommandNameReplaceItem:
		a := &CommandArgsReplaceItem{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameShowPicture:
		a := &CommandArgsShowPicture{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		// TODO Implement Decoder
		if a.Priority == "" {
			a.Priority = PicturePriorityOverlay
		}
		c.Args = a
	case CommandNameErasePicture:
		a := &CommandArgsErasePicture{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameMovePicture:
		a := &CommandArgsMovePicture{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameScalePicture:
		a := &CommandArgsScalePicture{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameRotatePicture:
		a := &CommandArgsRotatePicture{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameFadePicture:
		a := &CommandArgsFadePicture{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameTintPicture:
		a := &CommandArgsTintPicture{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameChangePictureImage:
		a := &CommandArgsChangePictureImage{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameChangeBackground:
		a := &CommandArgsChangeBackground{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameChangeForeground:
		a := &CommandArgsChangeForeground{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameSpecial:
		a := &CommandArgsSpecial{}
		if err := msgpack.Unmarshal(argsBin, a); err != nil {
			return err
		}
		c.Args = a
	case CommandNameFinishPlayerMovingByUserInput:
	case CommandNameExecEventHere:
	default:
		return fmt.Errorf("data: Command.DecodeMsgpack: invalid command: %s", c.Name)
	}

	commandUnmarshalingCount++
	if commandUnmarshalingCount%8 == 0 {
		runtime.Gosched()
	}

	return nil
}

type CommandName string

const (
	CommandNameNop               CommandName = "nop"
	CommandNameMemo              CommandName = "memo"
	CommandNameIf                CommandName = "if"
	CommandNameGroup             CommandName = "group"
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
	CommandNameSavePermanent     CommandName = "save_permanent"
	CommandNameLoadPermanent     CommandName = "load_permanent"
	CommandNameTransfer          CommandName = "transfer"
	CommandNameSetRoute          CommandName = "set_route"
	CommandNameTintScreen        CommandName = "tint_screen"
	CommandNameShake             CommandName = "shake"
	CommandNamePlaySE            CommandName = "play_se"
	CommandNamePlayBGM           CommandName = "play_bgm"
	CommandNameStopBGM           CommandName = "stop_bgm"
	CommandNameSave              CommandName = "save"
	CommandNameGotoTitle         CommandName = "goto_title"
	CommandNameAutoSave          CommandName = "autosave"
	CommandNameGameClear         CommandName = "game_clear"
	CommandNamePlayerControl     CommandName = "player_control"
	CommandNamePlayerSpeed       CommandName = "player_speed"
	CommandNameWeather           CommandName = "weather"
	CommandNameUnlockAchievement CommandName = "unlock_achievement"
	CommandNameControlHint       CommandName = "control_hint"
	CommandNamePurchase          CommandName = "start_iap"
	CommandNameShowAds           CommandName = "show_ads"
	CommandNameOpenLink          CommandName = "open_link"
	CommandNameShare             CommandName = "share"
	CommandNameRequestReview     CommandName = "request_review"
	CommandNameSendAnalytics     CommandName = "send_analytics"
	CommandNameShowShop          CommandName = "show_shop"
	CommandNameShowMinigame      CommandName = "show_minigame"
	CommandNameVibrate           CommandName = "vibrate"

	CommandNameAddItem       CommandName = "add_item"
	CommandNameRemoveItem    CommandName = "remove_item"
	CommandNameReplaceItem   CommandName = "replace_item"
	CommandNameShowItem      CommandName = "show_item"
	CommandNameHideItem      CommandName = "hide_item"
	CommandNameShowInventory CommandName = "show_inventory"
	CommandNameHideInventory CommandName = "hide_inventory"

	CommandNameShowPicture        CommandName = "show_picture"
	CommandNameErasePicture       CommandName = "erase_picture"
	CommandNameMovePicture        CommandName = "move_picture"
	CommandNameScalePicture       CommandName = "scale_picture"
	CommandNameRotatePicture      CommandName = "rotate_picture"
	CommandNameFadePicture        CommandName = "fade_picture"
	CommandNameTintPicture        CommandName = "tint_picture"
	CommandNameChangePictureImage CommandName = "change_picture_image"
	CommandNameChangeBackground   CommandName = "change_background"
	CommandNameChangeForeground   CommandName = "change_foreground"

	// Route commands
	CommandNameMoveCharacter        CommandName = "move_character"
	CommandNameTurnCharacter        CommandName = "turn_character"
	CommandNameRotateCharacter      CommandName = "rotate_character"
	CommandNameSetCharacterProperty CommandName = "set_character_property"
	CommandNameSetCharacterImage    CommandName = "set_character_image"
	CommandNameSetCharacterOpacity  CommandName = "set_character_opacity"

	// Special commands
	CommandNameSpecial                       CommandName = "special"
	CommandNameFinishPlayerMovingByUserInput CommandName = "finish_player_moving_by_user_input"
	CommandNameExecEventHere                 CommandName = "exec_event_here"
)

type CommandArgsMemo struct {
	Content string `msgpack:"content"`
	Log     bool   `msgpack:"log"`
}

type CommandArgsIf struct {
	Conditions []*Condition `msgpack:"conditions"`
}

type CommandArgsGroup struct {
	Name string `msgpack:"name"`
}

type CommandArgsLabel struct {
	Name string `msgpack:"name"`
}

type CommandArgsGoto struct {
	Label string `msgpack:"label"`
}

type CommandArgsGotoTitle struct {
	Save bool `msgpack:"save"`
}

type CommandArgsCallEvent struct {
	EventID   int `msgpack:"eventId"`
	PageIndex int `msgpack:"pageIndex"`
}

type CommandArgsCallCommonEvent struct {
	EventID int `msgpack:"eventId"`
}

type CommandArgsWait struct {
	Time int `msgpack:"time"`
}

type CommandArgsShowBalloon struct {
	EventID        int         `msgpack:"eventId"`
	ContentID      UUID        `msgpack:"content"`
	BalloonType    BalloonType `msgpack:"balloonType"`
	MessageStyleID int         `msgpack:"messageStyleId"`
}

type CommandArgsShowMessage struct {
	EventID        int                 `msgpack:"eventId"`
	ContentID      UUID                `msgpack:"content"`
	Background     MessageBackground   `msgpack:"background"`
	PositionType   MessagePositionType `msgpack:"positionType"`
	TextAlign      TextAlign           `msgpack:"textAlign"`
	MessageStyleID int                 `msgpack:"messageStyleId"`
}

type ChoiceCondition struct {
	Visible *Condition `msgpack:"visible"`
	Checked *Condition `msgpack:"checked"`
}

type CommandArgsShowChoices struct {
	ChoiceIDs  []UUID             `msgpack:"choices"`
	Conditions []*ChoiceCondition `msgpack:"conditions"`
}

type CommandArgsSetSwitch struct {
	ID       int             `msgpack:"id"`
	IDType   SetSwitchIDType `msgpack:"idType"`
	Value    bool            `msgpack:"value"`
	Internal bool            `msgpack:"internal"`
}

type CommandArgsSetSelfSwitch struct {
	ID    int  `msgpack:"id"`
	Value bool `msgpack:"value"`
}

type CommandArgsSetVariable struct {
	ID        int                  `msgpack:"id"`
	IDType    SetVariableIDType    `msgpack:"idType"`
	Op        SetVariableOp        `msgpack:"op"`
	ValueType SetVariableValueType `msgpack:"valueType"`
	Value     interface{}          `msgpack:"value"`
	Internal  bool                 `msgpack:"internal"`
}

func (c *CommandArgsSetVariable) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("id")
	e.EncodeInt(c.ID)

	e.EncodeString("idType")
	e.EncodeString(string(c.IDType))

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
	case SetVariableValueTypeVariableRef:
		e.EncodeInt(c.Value.(int))
	case SetVariableValueTypeSwitch:
		e.EncodeInt(c.Value.(int))
	case SetVariableValueTypeSwitchRef:
		e.EncodeInt(c.Value.(int))
	case SetVariableValueTypeRandom:
		e.EncodeAny(c.Value)
	case SetVariableValueTypeCharacter:
		e.EncodeAny(c.Value)
	case SetVariableValueTypeIAPProduct:
		e.EncodeInt(c.Value.(int))
	case SetVariableValueTypeSystem:
		e.EncodeString(string(c.Value.(SystemVariableType)))
	case SetVariableValueTypeTable:
		e.EncodeAny(c.Value)

	default:
		return fmt.Errorf("data: CommandArgsSetVariable.EncodeMsgpack: invalid type: %s", c.ValueType)
	}

	e.EncodeString("internal")
	e.EncodeBool(c.Internal)

	e.EndMap()
	return e.Flush()
}

func (c *CommandArgsSetVariable) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	var value interface{}
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "id":
			c.ID = d.DecodeInt()
		case "idType":
			c.IDType = SetVariableIDType(d.DecodeString())
		case "op":
			c.Op = SetVariableOp(d.DecodeString())
		case "valueType":
			c.ValueType = SetVariableValueType(d.DecodeString())
		case "value":
			d.DecodeAny(&value)
		case "internal":
			c.Internal = d.DecodeBool()
		default:
			if err := d.Error(); err != nil {
				return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack failed: %v", err)
			}
			return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: invalid argument: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack failed: %v", err)
	}

	// TODO: Avoid re-encoding the arg
	valueBin, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}

	switch c.ValueType {
	case SetVariableValueTypeConstant:
		v, ok := InterfaceToInt(value)
		if !ok {
			return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: constant value must be an integer; got %v", value)
		}
		c.Value = v
	case SetVariableValueTypeVariable:
		v, ok := InterfaceToInt(value)
		if !ok {
			return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: variable value must be an integer; got %v", value)
		}
		c.Value = v
	case SetVariableValueTypeVariableRef:
		v, ok := InterfaceToInt(value)
		if !ok {
			return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: variable value must be an integer; got %v", value)
		}
		c.Value = v
	case SetVariableValueTypeSwitch:
		v, ok := InterfaceToInt(value)
		if !ok {
			return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: variable value must be an integer; got %v", value)
		}
		c.Value = v
	case SetVariableValueTypeSwitchRef:
		v, ok := InterfaceToInt(value)
		if !ok {
			return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: variable value must be an integer; got %v", value)
		}
		c.Value = v
	case SetVariableValueTypeRandom:
		v := &SetVariableValueRandom{}
		if err := msgpack.Unmarshal(valueBin, v); err != nil {
			return err
		}
		c.Value = v
	case SetVariableValueTypeCharacter:
		v := &SetVariableCharacterArgs{}
		if err := msgpack.Unmarshal(valueBin, v); err != nil {
			return err
		}
		c.Value = v
	case SetVariableValueTypeItemGroup:
		v := &SetVariableItemGroupArgs{}
		if err := msgpack.Unmarshal(valueBin, v); err != nil {
			return err
		}
		c.Value = v
	case SetVariableValueTypeIAPProduct:
		v, ok := InterfaceToInt(value)
		if !ok {
			return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: IAP product value must be an integer; got %v", value)
		}
		c.Value = v
	case SetVariableValueTypeSystem:
		c.Value = SystemVariableType(value.(string))
	case SetVariableValueTypeTable:
		v := &TableValueArgs{}
		if err := msgpack.Unmarshal(valueBin, v); err != nil {
			return err
		}
		c.Value = v
	default:
		return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: invalid type: %s", c.ValueType)
	}
	return nil
}

type CommandArgsSavePermanent struct {
	VariableID          int `msgpack:"variableId"`
	PermanentVariableID int `msgpack:"permanentVariableId"`
}

type CommandArgsLoadPermanent struct {
	VariableID          int `msgpack:"variableId"`
	PermanentVariableID int `msgpack:"permanentVariableId"`
}

type CommandArgsTransfer struct {
	ValueType  ValueType              `msgpack:"valueType"`
	RoomID     int                    `msgpack:"roomId"`
	X          int                    `msgpack:"x"`
	Y          int                    `msgpack:"y"`
	Dir        Dir                    `msgpack:"dir"`
	Transition TransferTransitionType `msgpack:"transition"`
}

type CommandArgsSetRoute struct {
	EventID  int        `msgpack:"eventId"`
	Repeat   bool       `msgpack:"repeat"`
	Skip     bool       `msgpack:"skip"`
	Wait     bool       `msgpack:"wait"`
	Internal bool       `msgpack:"internal"`
	Commands []*Command `msgpack:"commands"`
}

type CommandArgsShake struct {
	Power     int            `msgpack:"power"`
	Speed     int            `msgpack:"speed"`
	Time      int            `msgpack:"time"`
	Wait      bool           `msgpack:"wait"`
	Direction ShakeDirection `msgpack:"direction"`
}

type CommandArgsTintScreen struct {
	Red   int  `msgpack:"red"`
	Green int  `msgpack:"green"`
	Blue  int  `msgpack:"blue"`
	Gray  int  `msgpack:"gray"`
	Time  int  `msgpack:"time"`
	Wait  bool `msgpack:"wait"`
}

type CommandArgsPlaySE struct {
	Name   string `msgpack:"name"`
	Volume int    `msgpack:"volume"`
}

type CommandArgsPlayBGM struct {
	Name     string `msgpack:"name"`
	Volume   int    `msgpack:"volume"`
	FadeTime int    `msgpack:"fadeTime"`
}

type CommandArgsStopBGM struct {
	FadeTime int `msgpack:"fadeTime"`
}

type CommandArgsUnlockAchievement struct {
	ID int `msgpack:"id"`
}

type CommandArgsControlHint struct {
	ID   int             `msgpack:"id"`
	Type ControlHintType `msgpack:"type"`
}

type CommandArgsPurchase struct {
	ID int `msgpack:"id"`
}

type CommandArgsShowAds struct {
	Type     ShowAdsType `msgpack:"type"`
	ForceAds bool        `msgpack:"forceAds"`
}

type CommandArgsOpenLink struct {
	Type OpenLinkType `msgpack:"type"`
	Data string       `msgpack:"data"`
}

type CommandArgsShare struct {
	TextID UUID   `msgpack:"text"`
	Image  string `msgpack:"image"`
}

type CommandArgsSendAnalytics struct {
	EventName string `msgpack:"eventName"`
}

type CommandArgsShowShop struct {
	Products []int `msgpack:"products"`
}

type CommandArgsShowMinigame struct {
	ID       int `msgpack:"id"`
	ReqScore int `msgpack:"reqScore"`
}

type CommandArgsVibrate struct {
	Type string `msgpack:"type"`
}

type CommandArgsAutoSave struct {
	Enabled bool `msgpack:"enabled"`
}

type CommandArgsPlayerControl struct {
	Enabled bool `msgpack:"enabled"`
}

type CommandArgsPlayerSpeed struct {
	Value Speed `msgpack:"value"`
}

type CommandArgsWeather struct {
	Type WeatherType `msgpack:"type"`
}

type WeatherType string

const (
	WeatherTypeNone WeatherType = "none"
	WeatherTypeSnow WeatherType = "snow"
	WeatherTypeRain WeatherType = "rain"
)

type CommandArgsMoveCharacter struct {
	Type             MoveCharacterType `msgpack:"type"`
	Dir              Dir               `msgpack:"dir"`
	Distance         int               `msgpack:"distance"`
	X                int               `msgpack:"x"`
	Y                int               `msgpack:"y"`
	ValueType        ValueType         `msgpack:"valueType"`
	IgnoreCharacters bool              `msgpack:"ignoreCharacters"`
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

	e.EncodeString("valueType")
	e.EncodeString(string(c.ValueType))

	e.EncodeString("ignoreCharacters")
	e.EncodeBool(c.IgnoreCharacters)

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
		case "valueType":
			c.ValueType = ValueType(d.DecodeString())
		case "ignoreCharacters":
			c.IgnoreCharacters = d.DecodeBool()
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
	Dir Dir `msgpack:"dir"`
}

type CommandArgsRotateCharacter struct {
	Angle int `msgpack:"angle"`
}

type CommandArgsSetCharacterProperty struct {
	Type  SetCharacterPropertyType `msgpack:"type"`
	Value interface{}              `msgpack:"value"`
}

type CommandArgsSetCharacterOpacity struct {
	Opacity int  `msgpack:"opacity"`
	Time    int  `msgpack:"time"`
	Wait    bool `msgpack:"wait"`
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

func InterfaceToInt(v interface{}) (int, bool) {
	switch v := v.(type) {
	case int:
		return v, true
	case uint:
		return int(v), true
	case int8:
		return int(v), true
	case uint8:
		return int(v), true
	case int16:
		return int(v), true
	case uint16:
		return int(v), true
	case int32:
		return int(v), true
	case uint32:
		return int(v), true
	case int64:
		return int(v), true
	case uint64:
		return int(v), true
	case float32:
		// TODO: This should not happen?
		return int(v), true
	case float64:
		// This happens when the data is in JSON.
		return int(v), true
	}
	return 0, false
}

func (c *CommandArgsSetCharacterProperty) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	var value interface{}
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "type":
			c.Type = SetCharacterPropertyType(d.DecodeString())
		case "value":
			d.DecodeAny(&value)
		default:
			if err := d.Error(); err != nil {
				return fmt.Errorf("data: CommandArgsSetCharacterProperty.DecodeMsgpack failed: %v", err)
			}
			return fmt.Errorf("data: CommandArgsSetCharacterProperty.DecodeMsgpack: invalid argument: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: CommandArgsSetCharacterProperty.DecodeMsgpack failed: %v", err)
	}

	switch c.Type {
	case SetCharacterPropertyTypeVisibility:
		c.Value = value.(bool)
	case SetCharacterPropertyTypeDirFix:
		c.Value = value.(bool)
	case SetCharacterPropertyTypeStepping:
		c.Value = value.(bool)
	case SetCharacterPropertyTypeThrough:
		c.Value = value.(bool)
	case SetCharacterPropertyTypeWalking:
		c.Value = value.(bool)
	case SetCharacterPropertyTypeSpeed:
		v, ok := InterfaceToInt(value)
		if !ok {
			return fmt.Errorf("data: CommandArgsSetCharacterProperty.DecodeMsgpack: speed must be an integer; got %v", value)
		}
		c.Value = Speed(v)
	default:
		return fmt.Errorf("data: CommandArgsSetCharacterProperty.DecodeMsgpack: invalid type: %s", c.Type)
	}
	return nil
}

type CommandArgsSetCharacterImage struct {
	Image          interface{}
	ImageType      ImageType
	ImageValueType FileValueType
	Frame          int
	Dir            Dir
	UseFrameAndDir bool
}

func (c *CommandArgsSetCharacterImage) EncodeMsgpack(enc *msgpack.Encoder) error {
	// Default value
	if c.ImageValueType == "" {
		c.ImageValueType = FileValueTypeConstant
	}

	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("imageType")
	e.EncodeString(string(c.ImageType))

	e.EncodeString("imageValueType")
	e.EncodeString(string(c.ImageValueType))

	e.EncodeString("frame")
	e.EncodeInt(int(c.Frame))

	e.EncodeString("dir")
	e.EncodeInt(int(c.Dir))

	e.EncodeString("useFrameAndDir")
	e.EncodeBool(c.UseFrameAndDir)

	e.EncodeString("image")
	e.EncodeAny(c.Image)

	e.EndMap()
	return e.Flush()
}

func (c *CommandArgsSetCharacterImage) DecodeMsgpack(dec *msgpack.Decoder) error {
	// Default value
	if c.ImageValueType == "" {
		c.ImageValueType = FileValueTypeConstant
	}

	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	var imageValue interface{}
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "imageType":
			c.ImageType = ImageType(d.DecodeString())
		case "image":
			d.DecodeAny(&imageValue)
		case "imageValueType":
			c.ImageValueType = FileValueType(d.DecodeString())
		case "frame":
			c.Frame = d.DecodeInt()
		case "dir":
			c.Dir = Dir(d.DecodeInt())
		case "useFrameAndDir":
			c.UseFrameAndDir = d.DecodeBool()
		default:
			if err := d.Error(); err != nil {
				return fmt.Errorf("data: CommandArgsSetCharacterImage.DecodeMsgpack failed: %v", err)
			}
			return fmt.Errorf("data: CommandArgsSetCharacterImage.DecodeMsgpack: invalid argument: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: CommandArgsSetCharacterImage.DecodeMsgpack failed: %v", err)
	}

	// TODO: Avoid re-encoding the arg
	valueBin, err := msgpack.Marshal(imageValue)
	if err != nil {
		return err
	}

	switch c.ImageValueType {
	case FileValueTypeConstant:
		if imageValue != nil {
			c.Image = imageValue.(string)
		}
	case FileValueTypeTable:
		v := &TableValueArgs{}
		if err := msgpack.Unmarshal(valueBin, v); err != nil {
			return err
		}
		c.Image = v
	default:
		return fmt.Errorf("data: CommandArgsSetCharacterImage.DecodeMsgpack: invalid type: %s for image: %v", c.ImageValueType, imageValue)
	}
	return nil
}

type CommandArgsAddItem struct {
	ID int `msgpack:"id"`
}

type CommandArgsRemoveItem struct {
	ID int `msgpack:"id"`
}

type CommandArgsShowItem struct {
	ID int `msgpack:"id"`
}

type CommandArgsShowInventory struct {
	Group      int  `msgpack:"group"`
	Wait       bool `msgpack:"wait"`
	Cancelable bool `msgpack:"cancelable"`
}

type CommandArgsReplaceItem struct {
	ID         int   `msgpack:"id"`
	ReplaceIDs []int `msgpack:"replaceIds"`
}

type ValueType string

const (
	ValueTypeConstant ValueType = "constant"
	ValueTypeVariable ValueType = "variable"
)

type SelectType string

const (
	SelectTypeSingle SelectType = "single"
	SelectTypeMulti  SelectType = "multi"
)

type ShowPictureBlendType string

const (
	ShowPictureBlendTypeNormal ShowPictureBlendType = "normal"
	ShowPictureBlendTypeAdd    ShowPictureBlendType = "add"
)

type CommandArgsShowPicture struct {
	ID           int                  `msgpack:"id"`
	IDValueType  ValueType            `msgpack:"idValueType"`
	Image        string               `msgpack:"image"`
	OriginX      float64              `msgpack:"originX"`
	OriginY      float64              `msgpack:"originY"`
	X            int                  `msgpack:"x"`
	Y            int                  `msgpack:"y"`
	PosValueType ValueType            `msgpack:"posValueType"`
	ScaleX       int                  `msgpack:"scaleX"`
	ScaleY       int                  `msgpack:"scaleY"`
	Angle        int                  `msgpack:"angle"`
	Opacity      int                  `msgpack:"opacity"`
	Priority     PicturePriorityType  `msgpack:"priority"`
	BlendType    ShowPictureBlendType `msgpack:"blendType"`
}

type CommandArgsErasePicture struct {
	ID          interface{} `msgpack:"id"`
	IDValueType ValueType   `msgpack:"idValueType"`
	SelectType  SelectType  `msgpack:"selectType"`
}

type CommandArgsMovePicture struct {
	ID           int       `msgpack:"id"`
	IDValueType  ValueType `msgpack:"idValueType"`
	X            int       `msgpack:"x"`
	Y            int       `msgpack:"y"`
	PosValueType ValueType `msgpack:"posValueType"`
	Time         int       `msgpack:"time"`
	Wait         bool      `msgpack:"wait"`
}

type CommandArgsScalePicture struct {
	ID             int       `msgpack:"id"`
	IDValueType    ValueType `msgpack:"idValueType"`
	ScaleX         int       `msgpack:"scaleX"`
	ScaleY         int       `msgpack:"scaleY"`
	ScaleValueType ValueType `msgpack:"scaleValueType"`
	Time           int       `msgpack:"time"`
	Wait           bool      `msgpack:"wait"`
}

type CommandArgsRotatePicture struct {
	ID             int       `msgpack:"id"`
	IDValueType    ValueType `msgpack:"idValueType"`
	Angle          int       `msgpack:"angle"`
	AngleValueType ValueType `msgpack:"angleValueType"`
	Time           int       `msgpack:"time"`
	Wait           bool      `msgpack:"wait"`
}

type CommandArgsFadePicture struct {
	ID               int       `msgpack:"id"`
	IDValueType      ValueType `msgpack:"idValueType"`
	Opacity          int       `msgpack:"opacity"`
	OpacityValueType ValueType `msgpack:"opacityValueType"`
	Time             int       `msgpack:"time"`
	Wait             bool      `msgpack:"wait"`
}

type CommandArgsTintPicture struct {
	ID          int       `msgpack:"id"`
	IDValueType ValueType `msgpack:"idValueType"`
	Red         int       `msgpack:"red"`
	Green       int       `msgpack:"green"`
	Blue        int       `msgpack:"blue"`
	Gray        int       `msgpack:"gray"`
	Time        int       `msgpack:"time"`
	Wait        bool      `msgpack:"wait"`
}

type CommandArgsChangePictureImage struct {
	ID             int
	IDValueType    ValueType
	Image          interface{}
	ImageValueType FileValueType
}

func (c *CommandArgsChangePictureImage) EncodeMsgpack(enc *msgpack.Encoder) error {
	// Default value
	if c.ImageValueType == "" {
		c.ImageValueType = FileValueTypeConstant
	}

	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("id")
	e.EncodeInt(c.ID)

	e.EncodeString("idValueType")
	e.EncodeString(string(c.IDValueType))

	e.EncodeString("imageValueType")
	e.EncodeString(string(c.ImageValueType))

	e.EncodeString("image")
	e.EncodeAny(c.Image)

	e.EndMap()
	return e.Flush()
}

func (c *CommandArgsChangePictureImage) DecodeMsgpack(dec *msgpack.Decoder) error {
	// Default value
	if c.ImageValueType == "" {
		c.ImageValueType = FileValueTypeConstant
	}

	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	var imageValue interface{}
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "id":
			c.ID = d.DecodeInt()
		case "image":
			d.DecodeAny(&imageValue)
		case "idValueType":
			c.IDValueType = ValueType(d.DecodeString())
		case "imageValueType":
			c.ImageValueType = FileValueType(d.DecodeString())
		default:
			if err := d.Error(); err != nil {
				return fmt.Errorf("data: CommandArgsChangePictureImage.DecodeMsgpack failed: %v", err)
			}
			return fmt.Errorf("data: CommandArgsChangePictureImage.DecodeMsgpack: invalid argument: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: CommandArgsChangePictureImage.DecodeMsgpack failed: %v", err)
	}

	// TODO: Avoid re-encoding the arg
	valueBin, err := msgpack.Marshal(imageValue)
	if err != nil {
		return err
	}

	switch c.ImageValueType {
	case FileValueTypeConstant:
		if imageValue != nil {
			c.Image = imageValue.(string)
		}
	case FileValueTypeTable:
		v := &TableValueArgs{}
		if err := msgpack.Unmarshal(valueBin, v); err != nil {
			return err
		}
		c.Image = v
	default:
		return fmt.Errorf("data: CommandArgsChangePictureImage.DecodeMsgpack: invalid type: %s for image: %v", c.ImageValueType, imageValue)
	}
	return nil
}

type CommandArgsChangeBackground struct {
	Image          interface{}
	ImageValueType FileValueType
}

func (c *CommandArgsChangeBackground) EncodeMsgpack(enc *msgpack.Encoder) error {
	// Default value
	if c.ImageValueType == "" {
		c.ImageValueType = FileValueTypeConstant
	}

	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("imageValueType")
	e.EncodeString(string(c.ImageValueType))

	e.EncodeString("image")
	e.EncodeAny(c.Image)

	e.EndMap()
	return e.Flush()
}

func (c *CommandArgsChangeBackground) DecodeMsgpack(dec *msgpack.Decoder) error {
	// Default value
	if c.ImageValueType == "" {
		c.ImageValueType = FileValueTypeConstant
	}

	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	var imageValue interface{}
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "image":
			d.DecodeAny(&imageValue)
		case "imageValueType":
			c.ImageValueType = FileValueType(d.DecodeString())
		default:
			if err := d.Error(); err != nil {
				return fmt.Errorf("data: CommandArgsChangeBackground.DecodeMsgpack failed: %v", err)
			}
			return fmt.Errorf("data: CommandArgsChangeBackground.DecodeMsgpack: invalid argument: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: CommandArgsChangeBackground.DecodeMsgpack failed: %v", err)
	}

	// TODO: Avoid re-encoding the arg
	valueBin, err := msgpack.Marshal(imageValue)
	if err != nil {
		return err
	}

	switch c.ImageValueType {
	case FileValueTypeConstant:
		if imageValue != nil {
			c.Image = imageValue.(string)
		}
	case FileValueTypeTable:
		v := &TableValueArgs{}
		if err := msgpack.Unmarshal(valueBin, v); err != nil {
			return err
		}
		c.Image = v
	default:
		return fmt.Errorf("data: CommandArgsChangeBackground.DecodeMsgpack: invalid type: %s for image: %v", c.ImageValueType, imageValue)
	}
	return nil
}

type CommandArgsChangeForeground struct {
	Image          interface{}
	ImageValueType FileValueType
}

func (c *CommandArgsChangeForeground) EncodeMsgpack(enc *msgpack.Encoder) error {
	// Default value
	if c.ImageValueType == "" {
		c.ImageValueType = FileValueTypeConstant
	}

	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("imageValueType")
	e.EncodeString(string(c.ImageValueType))

	e.EncodeString("image")
	e.EncodeAny(c.Image)

	e.EndMap()
	return e.Flush()
}

func (c *CommandArgsChangeForeground) DecodeMsgpack(dec *msgpack.Decoder) error {
	// Default value
	if c.ImageValueType == "" {
		c.ImageValueType = FileValueTypeConstant
	}

	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	var imageValue interface{}
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "image":
			d.DecodeAny(&imageValue)
		case "imageValueType":
			c.ImageValueType = FileValueType(d.DecodeString())
		default:
			if err := d.Error(); err != nil {
				return fmt.Errorf("data: CommandArgsChangeForeground.DecodeMsgpack failed: %v", err)
			}
			return fmt.Errorf("data: CommandArgsChangeForeground.DecodeMsgpack: invalid argument: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("data: CommandArgsChangeForeground.DecodeMsgpack failed: %v", err)
	}

	// TODO: Avoid re-encoding the arg
	valueBin, err := msgpack.Marshal(imageValue)
	if err != nil {
		return err
	}

	switch c.ImageValueType {
	case FileValueTypeConstant:
		if imageValue != nil {
			c.Image = imageValue.(string)
		}
	case FileValueTypeTable:
		v := &TableValueArgs{}
		if err := msgpack.Unmarshal(valueBin, v); err != nil {
			return err
		}
		c.Image = v
	default:
		return fmt.Errorf("data: CommandArgsChangeForeground.DecodeMsgpack: invalid type: %s for image: %v", c.ImageValueType, imageValue)
	}
	return nil
}

type CommandArgsSpecial struct {
	Content string `msgpack:"content"`
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
	SetVariableValueTypeConstant    SetVariableValueType = "constant"
	SetVariableValueTypeVariable    SetVariableValueType = "variable"
	SetVariableValueTypeVariableRef SetVariableValueType = "variable_ref"
	SetVariableValueTypeSwitch      SetVariableValueType = "switch"
	SetVariableValueTypeSwitchRef   SetVariableValueType = "switch_ref"
	SetVariableValueTypeRandom      SetVariableValueType = "random"
	SetVariableValueTypeCharacter   SetVariableValueType = "character"
	SetVariableValueTypeItemGroup   SetVariableValueType = "item_group"
	SetVariableValueTypeIAPProduct  SetVariableValueType = "iap_product"
	SetVariableValueTypeSystem      SetVariableValueType = "system"
	SetVariableValueTypeTable       SetVariableValueType = "table"
)

type SetVariableIDType string

const (
	SetVariableIDTypeVal SetVariableIDType = "val"
	SetVariableIDTypeRef SetVariableIDType = "ref"
)

type SetSwitchIDType string

const (
	SetSwitchIDTypeVal SetSwitchIDType = "val"
	SetSwitchIDTypeRef SetSwitchIDType = "ref"
)

type TransferTransitionType string

const (
	TransferTransitionTypeNone  TransferTransitionType = "none"
	TransferTransitionTypeBlack TransferTransitionType = "black"
	TransferTransitionTypeWhite TransferTransitionType = "white"
)

type SetVariableValueRandom struct {
	Begin int `msgpack:"begin"`
	End   int `msgpack:"end"`
}

// TODO: Rename?
type SetVariableCharacterArgs struct {
	Type    SetVariableCharacterType `msgpack:"type"`
	EventID int                      `msgpack:"eventId"`
}

type SetVariableItemGroupArgs struct {
	Type  SetVariableItemGroupType `msgpack:"type"`
	Group int                      `msgpack:"group"`
}

type SetVariableSystem struct {
	Type    SetVariableCharacterType `msgpack:"type"`
	EventID int                      `msgpack:"eventId"`
}

type TableValueArgs struct {
	Type ValueType `json:"type" msgpack:"type"`
	Name string    `json:"name" msgpack:"name"`
	ID   int       `json:"id" msgpack:"id"`
	Attr string    `json:"attr" msgpack:"attr"`
}

type SetVariableCharacterType string

const (
	SetVariableCharacterTypeDirection SetVariableCharacterType = "direction"
	SetVariableCharacterTypeRoomX     SetVariableCharacterType = "room_x"
	SetVariableCharacterTypeRoomY     SetVariableCharacterType = "room_y"
	SetVariableCharacterTypeScreenX   SetVariableCharacterType = "screen_x"
	SetVariableCharacterTypeScreenY   SetVariableCharacterType = "screen_y"
	SetVariableCharacterTypeIsPressed SetVariableCharacterType = "pressed"
)

type SetVariableItemGroupType string

const (
	SetVariableItemGroupTypeOwned SetVariableItemGroupType = "owned"
	SetVariableItemGroupTypeTotal SetVariableItemGroupType = "total"
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
	SystemVariableCurrentTime           SystemVariableType = "current_time"
	SystemVariableActiveItemID          SystemVariableType = "active_item_id"
	SystemVariableEventItemID           SystemVariableType = "event_item_id"
)

type MessagePositionType string

const (
	MessagePositionBottom MessagePositionType = "bottom"
	MessagePositionMiddle MessagePositionType = "middle"
	MessagePositionTop    MessagePositionType = "top"
	MessagePositionAuto   MessagePositionType = "auto"
)

type MessageBackground string

const (
	MessageBackgroundDim         MessageBackground = "dim"
	MessageBackgroundTransparent MessageBackground = "transparent"
	MessageBackgroundBanner      MessageBackground = "banner"
)

type PicturePriorityType string

const (
	PicturePriorityBottom  PicturePriorityType = "bottom"
	PicturePriorityTop     PicturePriorityType = "top"
	PicturePriorityOverlay PicturePriorityType = "overlay"
)

type ShakeDirection string

const (
	ShakeDirectionHorizontal ShakeDirection = "horizontal"
	ShakeDirectionVertical   ShakeDirection = "vertical"
)

type OpenLinkType string

const (
	OpenLinkTypeApp        OpenLinkType = "app"
	OpenLinkTypeURL        OpenLinkType = "url"
	OpenLinkTypeReview     OpenLinkType = "review"
	OpenLinkTypeShowCredit OpenLinkType = "show_credit"
	OpenLinkTypePostCredit OpenLinkType = "post_credit"
	OpenLinkTypeMore       OpenLinkType = "more"
	OpenLinkTypeFacebook   OpenLinkType = "fb"
	OpenLinkTypeTwitter    OpenLinkType = "twitter"
)

type FileValueType string

const (
	FileValueTypeConstant FileValueType = "constant"
	FileValueTypeTable    FileValueType = "table"
)
