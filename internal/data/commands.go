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
	"runtime"

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

var commandUnmarshalingCount = 0

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
	case CommandNameMemo:
		var args *CommandArgsMemo
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameIf:
		var args *CommandArgsIf
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameGroup:
		var args *CommandArgsGroup
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
	case CommandNameGotoTitle:
		var args *CommandArgsGotoTitle
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
	case CommandNameShake:
		var args *CommandArgsShake
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
		var args *CommandArgsStopBGM
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSave:
	case CommandNameRequestReview:
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
	case CommandNamePlayerSpeed:
		var args *CommandArgsPlayerSpeed
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameWeather:
		var args *CommandArgsWeather
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
	case CommandNameShare:
		var args *CommandArgsShare
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSendAnalytics:
		var args *CommandArgsSendAnalytics
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameShowShop:
		var args *CommandArgsShowShop
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameVibrate:
		var args *CommandArgsVibrate
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
	case CommandNameShowInventory:
		var args *CommandArgsShowInventory
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameHideInventory:
	case CommandNameShowItem:
		var args *CommandArgsShowItem
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameHideItem:
	case CommandNameReplaceItem:
		var args *CommandArgsReplaceItem
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameShowPicture:
		var args *CommandArgsShowPicture
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		// TODO Implement Decoder
		if args.Priority == "" {
			args.Priority = PicturePriorityOverlay
		}
		c.Args = args
	case CommandNameErasePicture:
		var args *CommandArgsErasePicture
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameMovePicture:
		var args *CommandArgsMovePicture
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameScalePicture:
		var args *CommandArgsScalePicture
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameRotatePicture:
		var args *CommandArgsRotatePicture
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameFadePicture:
		var args *CommandArgsFadePicture
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameTintPicture:
		var args *CommandArgsTintPicture
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameChangePictureImage:
		var args *CommandArgsChangePictureImage
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameChangeBackground:
		var args *CommandArgsChangeBackground
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameChangeForeground:
		var args *CommandArgsChangeForeground
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	case CommandNameSpecial:
		var args *CommandArgsSpecial
		if err := unmarshalJSON(tmp.Args, &args); err != nil {
			return err
		}
		c.Args = args
	default:
		return fmt.Errorf("data: invalid command: %s", c.Name)
	}

	// Force context switching to avoid freezing (#463)
	commandUnmarshalingCount++
	if commandUnmarshalingCount%8 == 0 {
		runtime.Gosched()
	}
	return nil
}

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
	Content string `json:"content" msgpack:"content"`
	Log     bool   `json:"log" msgpack:"log"`
}

type CommandArgsIf struct {
	Conditions []*Condition `json:"conditions" msgpack:"conditions"`
}

type CommandArgsGroup struct {
	Name string `json:"name" msgpack:"name"`
}

type CommandArgsLabel struct {
	Name string `json:"name" msgpack:"name"`
}

type CommandArgsGoto struct {
	Label string `json:"label" msgpack:"label"`
}

type CommandArgsGotoTitle struct {
	Save bool `json:"save" msgpack:"save"`
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
	EventID        int         `json:"eventId" msgpack:"eventId"`
	ContentID      UUID        `json:"content" msgpack:"content"`
	BalloonType    BalloonType `json:"balloonType" msgpack:"balloonType"`
	MessageStyleID int         `json:"messageStyleId" msgpack:"messageStyleId"`
}

type CommandArgsShowMessage struct {
	EventID        int                 `json:"eventId" msgpack:"eventId"`
	ContentID      UUID                `json:"content" msgpack:"content"`
	Background     MessageBackground   `json:"background" msgpack:"background"`
	PositionType   MessagePositionType `json:"positionType" msgpack:"positionType"`
	TextAlign      TextAlign           `json:"textAlign" msgpack:"textAlign"`
	MessageStyleID int                 `json:"messageStyleId" msgpack:"messageStyleId"`
}

type ChoiceCondition struct {
	Visible *Condition `json:"visible" msgpack:"visible"`
	Checked *Condition `json:"checked" msgpack:"checked"`
}

type CommandArgsShowChoices struct {
	ChoiceIDs  []UUID             `json:"choices" msgpack:"choices"`
	Conditions []*ChoiceCondition `json:"conditions" msgpack:"conditions"`
}

type CommandArgsSetSwitch struct {
	ID       int  `json:"id" msgpack:"id"`
	Value    bool `json:"value" msgpack:"value"`
	Internal bool `json:"internal" msgpack:"internal"`
}

type CommandArgsSetSelfSwitch struct {
	ID    int  `json:"id" msgpack:"id"`
	Value bool `json:"value" msgpack:"value"`
}

type CommandArgsSetVariable struct {
	ID        int                  `json:"id" msgpack:"id"`
	Op        SetVariableOp        `json:"op" msgpack:"op"`
	ValueType SetVariableValueType `json:"valueType" msgpack:"valueType"`
	Value     interface{}          `json:"value" msgpack:"value"`
	Internal  bool                 `json:"internal" msgpack:"internal"`
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

	e.EncodeString("internal")
	e.EncodeBool(c.Internal)

	e.EndMap()
	return e.Flush()
}

func (c *CommandArgsSetVariable) UnmarshalJSON(data []uint8) error {
	type tmpCommandArgsSetVariable struct {
		ID        int                  `json:"id"`
		Op        SetVariableOp        `json:"op"`
		ValueType SetVariableValueType `json:"valueType"`
		Value     json.RawMessage      `json:"value"`
		Internal  bool                 `json:"internal"`
	}
	var tmp *tmpCommandArgsSetVariable
	if err := unmarshalJSON(data, &tmp); err != nil {
		return err
	}
	c.ID = tmp.ID
	c.Op = tmp.Op
	c.ValueType = tmp.ValueType
	c.Internal = tmp.Internal
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
	var value interface{}
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "id":
			c.ID = d.DecodeInt()
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
	case SetVariableValueTypeIAPProduct:
		v, ok := InterfaceToInt(value)
		if !ok {
			return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: IAP product value must be an integer; got %v", value)
		}
		c.Value = v
	case SetVariableValueTypeSystem:
		c.Value = SystemVariableType(value.(string))
	default:
		return fmt.Errorf("data: CommandArgsSetVariable.DecodeMsgpack: invalid type: %s", c.ValueType)
	}
	return nil
}

type CommandArgsTransfer struct {
	ValueType  ValueType              `json:"valueType" msgpack:"valueType"`
	RoomID     int                    `json:"roomId" msgpack:"roomId"`
	X          int                    `json:"x" msgpack:"x"`
	Y          int                    `json:"y" msgpack:"y"`
	Dir        Dir                    `json:"dir" msgpack:"dir"`
	Transition TransferTransitionType `json:"transition" msgpack:"transition"`
}

type CommandArgsSetRoute struct {
	EventID  int        `json:"eventId" msgpack:"eventId"`
	Repeat   bool       `json:"repeat" msgpack:"repeat"`
	Skip     bool       `json:"skip" msgpack:"skip"`
	Wait     bool       `json:"wait" msgpack:"wait"`
	Internal bool       `json:"internal" msgpack:"internal"`
	Commands []*Command `json:"commands" msgpack:"commands"`
}

type CommandArgsShake struct {
	Power     int            `json:"power" msgpack:"power"`
	Speed     int            `json:"speed" msgpack:"speed"`
	Time      int            `json:"time" msgpack:"time"`
	Wait      bool           `json:"wait" msgpack:"wait"`
	Direction ShakeDirection `json:"direction" msgpack:"direction"`
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

type CommandArgsStopBGM struct {
	FadeTime int `json:"fadeTime" msgpack:"fadeTime"`
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
	Type     ShowAdsType `json:"type" msgpack:"type"`
	ForceAds bool        `json:"forceAds" msgpack:"forceAds"`
}

type CommandArgsOpenLink struct {
	Type string `json:"type" msgpack:"type"`
	Data string `json:"data" msgpack:"data"`
}

type CommandArgsShare struct {
	TextID UUID   `json:"text" msgpack:"text"`
	Image  string `json:"image" msgpack:"image"`
}

type CommandArgsSendAnalytics struct {
	EventName string `json:"eventName" msgpack:"eventName"`
}

type CommandArgsShowShop struct {
	Products []int `json:"products" msgpack:"products"`
}

type CommandArgsVibrate struct {
	Type string `json:"type" msgpack:"type"`
}

type CommandArgsAutoSave struct {
	Enabled bool `json:"enabled" msgpack:"enabled"`
}

type CommandArgsPlayerControl struct {
	Enabled bool `json:"enabled" msgpack:"enabled"`
}

type CommandArgsPlayerSpeed struct {
	Value Speed `json:"value" msgpack:"value"`
}

type CommandArgsWeather struct {
	Type WeatherType `json:"type" msgpack:"type"`
}

type WeatherType string

const (
	WeatherTypeNone WeatherType = "none"
	WeatherTypeSnow WeatherType = "snow"
	WeatherTypeRain WeatherType = "rain"
)

type CommandArgsMoveCharacter struct {
	Type             MoveCharacterType `json:"type" msgpack:"type"`
	Dir              Dir               `json:"dir" msgpack:"dir"`
	Distance         int               `json:"distance" msgpack:"distance"`
	X                int               `json:"x" msgpack:"x"`
	Y                int               `json:"y" msgpack:"y"`
	ValueType        ValueType         `json:"valueType" msgpack:"valueType"`
	IgnoreCharacters bool              `json:"ignoreCharacters" msgpack:"ignoreCharacters"`
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
	Dir Dir `json:"dir" msgpack:"dir"`
}

type CommandArgsRotateCharacter struct {
	Angle int `json:"angle" msgpack:"angle"`
}

type CommandArgsSetCharacterProperty struct {
	Type  SetCharacterPropertyType `json:"type" msgpack:"type"`
	Value interface{}              `json:"value" msgpack:"value"`
}

type CommandArgsSetCharacterOpacity struct {
	Opacity int  `json:"opacity" msgpack:"opacity"`
	Time    int  `json:"time" msgpack:"time"`
	Wait    bool `json:"wait" msgpack:"wait"`
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
	Image          string    `json:"image" msgpack:"image"`
	ImageType      ImageType `json:"imageType" msgpack:"imageType"`
	Frame          int       `json:"frame" msgpack:"frame"`
	Dir            Dir       `json:"dir" msgpack:"dir"`
	UseFrameAndDir bool      `json:"useFrameAndDir" msgpack:"useFrameAndDir"`
}

type CommandArgsAddItem struct {
	ID int `json:"id" msgpack:"id"`
}

type CommandArgsRemoveItem struct {
	ID int `json:"id" msgpack:"id"`
}

type CommandArgsShowItem struct {
	ID int `json:"id" msgpack:"id"`
}

type CommandArgsShowInventory struct {
	Group int `json:"group" msgpack:"group"`
}

type CommandArgsReplaceItem struct {
	ID         int   `json:"id" msgpack:"id"`
	ReplaceIDs []int `json:"replaceIds" msgpack:"replaceIds"`
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
	ID           int                  `json:"id" msgpack:"id"`
	IDValueType  ValueType            `json:"idValueType" msgpack:"idValueType"`
	Image        string               `json:"image" msgpack:"image"`
	OriginX      float64              `json:"originX" msgpack:"originX"`
	OriginY      float64              `json:"originY" msgpack:"originY"`
	X            int                  `json:"x" msgpack:"x"`
	Y            int                  `json:"y" msgpack:"y"`
	PosValueType ValueType            `json:"posValueType" msgpack:"posValueType"`
	ScaleX       int                  `json:"scaleX" msgpack:"scaleX"`
	ScaleY       int                  `json:"scaleY" msgpack:"scaleY"`
	Angle        int                  `json:"angle" msgpack:"angle"`
	Opacity      int                  `json:"opacity" msgpack:"opacity"`
	Priority     PicturePriorityType  `json:"priority" msgpack:"priority"`
	BlendType    ShowPictureBlendType `json:"blendType" msgpack:"blendType"`
}

type CommandArgsErasePicture struct {
	ID          interface{} `json:"id" msgpack:"id"`
	IDValueType ValueType   `json:"idValueType" msgpack:"idValueType"`
	SelectType  SelectType  `json:"selectType" msgpack:"selectType"`
}

type CommandArgsMovePicture struct {
	ID           int       `json:"id" msgpack:"id"`
	IDValueType  ValueType `json:"idValueType" msgpack:"idValueType"`
	X            int       `json:"x" msgpack:"x"`
	Y            int       `json:"y" msgpack:"y"`
	PosValueType ValueType `json:"posValueType" msgpack:"posValueType"`
	Time         int       `json:"time" msgpack:"time"`
	Wait         bool      `json:"wait" msgpack:"wait"`
}

type CommandArgsScalePicture struct {
	ID             int       `json:"id" msgpack:"id"`
	IDValueType    ValueType `json:"idValueType" msgpack:"idValueType"`
	ScaleX         int       `json:"scaleX" msgpack:"scaleX"`
	ScaleY         int       `json:"scaleY" msgpack:"scaleY"`
	ScaleValueType ValueType `json:"scaleValueType" msgpack:"scaleValueType"`
	Time           int       `json:"time" msgpack:"time"`
	Wait           bool      `json:"wait" msgpack:"wait"`
}

type CommandArgsRotatePicture struct {
	ID             int       `json:"id" msgpack:"id"`
	IDValueType    ValueType `json:"idValueType" msgpack:"idValueType"`
	Angle          int       `json:"angle" msgpack:"angle"`
	AngleValueType ValueType `json:"angleValueType" msgpack:"angleValueType"`
	Time           int       `json:"time" msgpack:"time"`
	Wait           bool      `json:"wait" msgpack:"wait"`
}

type CommandArgsFadePicture struct {
	ID               int       `json:"id" msgpack:"id"`
	IDValueType      ValueType `json:"idValueType" msgpack:"idValueType"`
	Opacity          int       `json:"opacity" msgpack:"opacity"`
	OpacityValueType ValueType `json:"opacityValueType" msgpack:"opacityValueType"`
	Time             int       `json:"time" msgpack:"time"`
	Wait             bool      `json:"wait" msgpack:"wait"`
}

type CommandArgsTintPicture struct {
	ID          int       `json:"id" msgpack:"id"`
	IDValueType ValueType `json:"idValueType" msgpack:"idValueType"`
	Red         int       `json:"red" msgpack:"red"`
	Green       int       `json:"green" msgpack:"green"`
	Blue        int       `json:"blue" msgpack:"blue"`
	Gray        int       `json:"gray" msgpack:"gray"`
	Time        int       `json:"time" msgpack:"time"`
	Wait        bool      `json:"wait" msgpack:"wait"`
}

type CommandArgsChangePictureImage struct {
	ID          int       `json:"id" msgpack:"id"`
	IDValueType ValueType `json:"idValueType" msgpack:"idValueType"`
	Image       string    `json:"image" msgpack:"image"`
}

type CommandArgsChangeBackground struct {
	Image string `json:"image" msgpack:"image"`
}

type CommandArgsChangeForeground struct {
	Image string `json:"image" msgpack:"image"`
}

type CommandArgsSpecial struct {
	Content string `json:"content" msgpack:"content"`
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
	SetVariableCharacterTypeIsPressed SetVariableCharacterType = "pressed"
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
