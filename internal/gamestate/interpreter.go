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

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"strconv"

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/commanditerator"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/movecharacterstate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/variables"
)

type Interpreter struct {
	id                 int
	mapID              int // Note: This doesn't make sense when eventID == -1
	roomID             int // Note: This doesn't make sense when eventID == -1
	eventID            int
	commandIterator    *commanditerator.CommandIterator
	waitingCount       int
	waitingCommand     bool
	moveCharacterState *movecharacterstate.State
	repeat             bool
	sub                *Interpreter
	route              bool // True when used for event routing property.
	routeSkip          bool
	parallel           bool
	waitingRequestID   int // Note: When this is not 0, the game state can't be saved.
}

type InterpreterIDGenerator interface {
	GenerateInterpreterID() int
}

func NewInterpreter(idGen InterpreterIDGenerator, mapID, roomID, eventID int, commands []*data.Command) *Interpreter {
	return &Interpreter{
		id:              idGen.GenerateInterpreterID(),
		mapID:           mapID,
		roomID:          roomID,
		eventID:         eventID,
		commandIterator: commanditerator.New(commands),
	}
}

func (i *Interpreter) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("id")
	e.EncodeInt(i.id)

	e.EncodeString("mapId")
	e.EncodeInt(i.mapID)

	e.EncodeString("roomId")
	e.EncodeInt(i.roomID)

	e.EncodeString("eventId")
	e.EncodeInt(i.eventID)

	e.EncodeString("commandIterator")
	e.EncodeInterface(i.commandIterator)

	e.EncodeString("waitingCount")
	e.EncodeInt(i.waitingCount)

	e.EncodeString("waitingCommand")
	e.EncodeBool(i.waitingCommand)

	e.EncodeString("moveCharacterState")
	e.EncodeInterface(i.moveCharacterState)

	e.EncodeString("repeat")
	e.EncodeBool(i.repeat)

	e.EncodeString("sub")
	e.EncodeInterface(i.sub)

	e.EncodeString("route")
	e.EncodeBool(i.route)

	e.EncodeString("routeSkip")
	e.EncodeBool(i.routeSkip)

	e.EncodeString("parallel")
	e.EncodeBool(i.parallel)

	e.EncodeString("waitingRequestId")
	e.EncodeInt(i.waitingRequestID)

	e.EndMap()
	return e.Flush()
}

func (i *Interpreter) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for j := 0; j < n; j++ {
		switch k := d.DecodeString(); k {
		case "id":
			i.id = d.DecodeInt()
		case "mapId":
			i.mapID = d.DecodeInt()
		case "roomId":
			i.roomID = d.DecodeInt()
		case "eventId":
			i.eventID = d.DecodeInt()
		case "commandIterator":
			if !d.SkipCodeIfNil() {
				i.commandIterator = &commanditerator.CommandIterator{}
				d.DecodeInterface(i.commandIterator)
			}
		case "waitingCount":
			i.waitingCount = d.DecodeInt()
		case "waitingCommand":
			i.waitingCommand = d.DecodeBool()
		case "moveCharacterState":
			if !d.SkipCodeIfNil() {
				i.moveCharacterState = &movecharacterstate.State{}
				d.DecodeInterface(i.moveCharacterState)
			}
		case "repeat":
			i.repeat = d.DecodeBool()
		case "sub":
			if !d.SkipCodeIfNil() {
				i.sub = &Interpreter{}
				d.DecodeInterface(i.sub)
			}
		case "route":
			i.route = d.DecodeBool()
		case "routeSkip":
			i.routeSkip = d.DecodeBool()
		case "parallel":
			i.parallel = d.DecodeBool()
		case "waitingRequestId":
			i.waitingRequestID = d.DecodeInt()
		default:
			if err := d.Error(); err != nil {
				return err
			}
			return fmt.Errorf("gamestate: Interpreter.DecodeMsgpack failed: invalid key: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("gamestate: Interpreter.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (i *Interpreter) waitingRequestResponse() bool {
	return i.waitingRequestID != 0
}

func (i *Interpreter) IsExecuting() bool {
	return i.commandIterator != nil
}

func (i *Interpreter) createChild(gameState InterpreterIDGenerator, eventID int, commands []*data.Command) *Interpreter {
	child := NewInterpreter(gameState, i.mapID, i.roomID, eventID, commands)
	child.route = i.route
	return child
}

func (i *Interpreter) findMessageStyle(sceneManager *scene.Manager, messageStyleID int) *data.MessageStyle {
	messageStyles := sceneManager.Game().MessageStyles
	if messageStyleID > 0 {
		for index := range messageStyles {
			if messageStyles[index].ID == messageStyleID {
				return messageStyles[index]
			}
		}
	}
	return sceneManager.Game().CreateDefaultMessageStyle()
}

func (i *Interpreter) doOneCommand(sceneManager *scene.Manager, gameState *Game) (bool, error) {
	if !gameState.CanWindowProceed(i.id) {
		return false, nil
	}
	if i.sub != nil {
		if err := i.sub.Update(sceneManager, gameState); err != nil {
			return false, err
		}
		if !i.sub.IsExecuting() {
			i.sub = nil
			i.commandIterator.Advance()
		}
		return false, nil
	}
	if i.waitingRequestID != 0 {
		r := sceneManager.ReceiveResultIfExists(i.waitingRequestID)
		if r == nil {
			return false, nil
		}
		i.waitingRequestID = 0
		switch r.Type {
		case scene.RequestTypePurchase, scene.RequestTypeRewardedAds:
			if r.Succeeded {
				i.commandIterator.Choose(0)
			} else {
				i.commandIterator.Choose(1)
			}
		case scene.RequestTypeIAPPrices:
			if r.Succeeded {
				var prices map[string]string
				if err := json.Unmarshal(r.Data, &prices); err != nil {
					panic(err)
				}
				gameState.SetPrices(prices)
				i.commandIterator.Choose(0)
			} else {
				i.commandIterator.Choose(1)
			}
		default:
			i.commandIterator.Advance()
		}
		switch r.Type {
		case scene.RequestTypeRewardedAds, scene.RequestTypeInterstitialAds:
			audio.ResumeBGM()
		}
		return true, nil
	}
	c := i.commandIterator.Command()
	switch c.Name {
	case data.CommandNameNop:
		i.commandIterator.Advance()
	case data.CommandNameIf:
		conditions := c.Args.(*data.CommandArgsIf).Conditions
		matches := true
		for _, c := range conditions {
			m, err := gameState.MeetsCondition(c, i.eventID)
			if err != nil {
				return false, err
			}
			if !m {
				matches = false
				break
			}
		}
		if matches {
			i.commandIterator.Choose(0)
		} else if len(c.Branches) >= 2 {
			i.commandIterator.Choose(1)
		} else {
			i.commandIterator.Advance()
		}
	case data.CommandNameLabel:
		i.commandIterator.Advance()
	case data.CommandNameGoto:
		label := c.Args.(*data.CommandArgsGoto).Label
		if !i.commandIterator.Goto(label) {
			i.commandIterator.Advance()
		}
	case data.CommandNameCallEvent:
		args := c.Args.(*data.CommandArgsCallEvent)
		eventID := args.EventID
		if eventID == 0 {
			eventID = i.eventID
		}
		// TODO: Should i.mapID and i.roomID be considered here?
		var event *data.Event
		for _, e := range gameState.CurrentEvents() {
			if e.ID == eventID {
				event = e
				break
			}
		}
		if event == nil {
			// TODO: warning?
			i.commandIterator.Advance()
			return true, nil
		}
		page := event.Pages[args.PageIndex]
		commands := page.Commands
		i.sub = i.createChild(gameState, eventID, commands)

	case data.CommandNameCallCommonEvent:
		args := c.Args.(*data.CommandArgsCallCommonEvent)
		eventID := args.EventID
		var c *data.CommonEvent
		for _, e := range sceneManager.Game().CommonEvents {
			if e.ID == eventID {
				c = e
				break
			}
		}
		if c == nil {
			return false, fmt.Errorf("invalid common event ID: %d", eventID)
		}
		i.sub = i.createChild(gameState, i.eventID, c.Commands)

	case data.CommandNameReturn:
		i.commandIterator.Terminate()

	case data.CommandNameEraseEvent:
		i.commandIterator.Terminate()
		if ch := gameState.Character(i.mapID, i.roomID, i.eventID); ch != nil {
			ch.Erase()
		}

	case data.CommandNameWait:
		if i.waitingCount == 0 {
			time := c.Args.(*data.CommandArgsWait).Time
			// If Wait 0.0 is specified, treat is as one frame
			if time == 0 {
				i.waitingCount = 1
			} else {
				i.waitingCount = time * 6
			}
		}
		i.waitingCount--
		if i.waitingCount == 0 {
			i.commandIterator.Advance()
			return true, nil
		}
		return false, nil

	case data.CommandNameShowBalloon:
		args := c.Args.(*data.CommandArgsShowBalloon)
		if !i.waitingCommand {
			content := sceneManager.Game().Texts.Get(lang.Get(), args.ContentID)
			id := args.EventID
			if id == 0 {
				id = i.eventID
			}
			messageStyle := i.findMessageStyle(sceneManager, args.MessageStyleID)
			if gameState.ShowBalloon(i.id, i.mapID, i.roomID, id, content, args.BalloonType, messageStyle) {
				i.waitingCommand = true
				return false, nil
			}
		}
		if gameState.IsWindowAnimating(i.id) {
			return false, nil
		}
		// Advance command index first and check the next command.
		i.commandIterator.Advance()
		if !i.commandIterator.IsTerminated() {
			if i.commandIterator.Command().Name != data.CommandNameShowChoices {
				gameState.CloseAllWindows()
			}
		} else {
			gameState.CloseAllWindows()
		}
		i.waitingCommand = false

	case data.CommandNameShowMessage:
		args := c.Args.(*data.CommandArgsShowMessage)
		if !i.waitingCommand {
			content := sceneManager.Game().Texts.Get(lang.Get(), args.ContentID)
			id := args.EventID
			if id == 0 {
				id = i.eventID
			}

			messageStyle := i.findMessageStyle(sceneManager, args.MessageStyleID)
			gameState.ShowMessage(i.id, id, content, args.Background, args.PositionType, args.TextAlign, messageStyle)
			i.waitingCommand = true
			return false, nil
		}
		if gameState.IsWindowAnimating(i.id) {
			return false, nil
		}
		// Advance command index first and check the next command.
		i.commandIterator.Advance()
		gameState.CloseAllWindows()
		i.waitingCommand = false

	case data.CommandNameShowHint:
		hintId := gameState.hints.ActiveHintID()
		// next time it shows next available hint
		gameState.hints.ReadHint(hintId)
		sceneManager.Requester().RequestSendAnalytics("show_hint", strconv.Itoa(hintId))
		hasHint := false
		for _, h := range sceneManager.Game().Hints {
			if h.ID == hintId {
				c := h.Commands
				i.sub = i.createChild(gameState, i.eventID, c)
				hasHint = true
				break
			}
		}
		// Advance command index first and check the next command.
		if !hasHint {
			i.commandIterator.Advance()
		}

	case data.CommandNameShowChoices:
		if !i.waitingCommand {
			if gameState.windows.IsBusyWithChoosing() {
				return false, nil
			}
			gameState.ShowChoices(sceneManager, i.id, c.Args.(*data.CommandArgsShowChoices).ChoiceIDs)
			i.waitingCommand = true
			return false, nil
		}
		if !gameState.HasChosenWindowIndex() {
			return false, nil
		}
		i.commandIterator.Choose(gameState.ChosenWindowIndex())
		i.waitingCommand = false

	case data.CommandNameSetSwitch:
		args := c.Args.(*data.CommandArgsSetSwitch)
		if args.ID >= variables.ReservedID && !args.Internal {
			return false, fmt.Errorf("gamestate: the switch ID (%d) must be < %d", args.ID, variables.ReservedID)
		}
		gameState.SetSwitchValue(args.ID, args.Value)
		i.commandIterator.Advance()

	case data.CommandNameSetSelfSwitch:
		args := c.Args.(*data.CommandArgsSetSelfSwitch)
		gameState.SetSelfSwitchValue(i.eventID, args.ID, args.Value)
		i.commandIterator.Advance()

	case data.CommandNameSetVariable:
		args := c.Args.(*data.CommandArgsSetVariable)
		if args.ID >= variables.ReservedID && !args.Internal {
			return false, fmt.Errorf("gamestate: the variable ID (%d) must be < %d", args.ID, variables.ReservedID)
		}
		if err := gameState.SetVariable(sceneManager, args.ID, args.Op, args.ValueType, args.Value, i.mapID, i.roomID, i.eventID); err != nil {
			return false, err
		}
		i.commandIterator.Advance()

	case data.CommandNameTransfer:
		args := c.Args.(*data.CommandArgsTransfer)
		if args.Transition == data.TransferTransitionTypeNone {
			roomID := args.RoomID
			x := args.X
			y := args.Y
			if args.ValueType == data.ValueTypeVariable {
				roomID = gameState.VariableValue(roomID)
				x = gameState.VariableValue(x)
				y = gameState.VariableValue(y)
			}

			if args.Dir != data.DirNone {
				gameState.SetPlayerDir(args.Dir)
			}
			gameState.TransferPlayerImmediately(roomID, x, y, i)
			i.waitingCommand = false
			i.commandIterator.Advance()
			return true, nil
		}

		if !i.waitingCommand {
			if args.Transition == data.TransferTransitionTypeWhite {
				gameState.SetFadeColor(color.White)
			} else {
				gameState.SetFadeColor(color.Black)
			}
			gameState.FadeOut(30)
			i.waitingCommand = true
			return false, nil
		}
		if gameState.IsScreenFadedOut() {
			roomID := args.RoomID
			x := args.X
			y := args.Y
			if args.ValueType == data.ValueTypeVariable {
				roomID = gameState.VariableValue(roomID)
				x = gameState.VariableValue(x)
				y = gameState.VariableValue(y)
			}
			if args.Dir != data.DirNone {
				gameState.SetPlayerDir(args.Dir)
			}
			gameState.TransferPlayerImmediately(roomID, x, y, i)
			gameState.FadeIn(30)
			return false, nil
		}
		if gameState.IsScreenFading() {
			return false, nil
		}
		i.waitingCommand = false
		i.commandIterator.Advance()
	case data.CommandNameSetRoute:
		// Refresh events so that new event graphics can be seen before setting a route (#59)
		if err := gameState.RefreshEvents(); err != nil {
			return false, err
		}
		args := c.Args.(*data.CommandArgsSetRoute)
		id := args.EventID
		if id == 0 {
			id = i.eventID
		}
		sub := i.createChild(gameState, id, args.Commands)
		sub.repeat = args.Repeat
		sub.routeSkip = args.Skip
		if !args.Wait {
			// Set 'route' true so that the new route command does not
			// block the player's move (#380).
			if id != 0 {
				gameState.Map().removeRoutes(id)
				sub.route = true
			}
			gameState.Map().addInterpreter(sub)
			i.commandIterator.Advance()
			return true, nil
		}
		i.sub = sub
	case data.CommandNameTintScreen:
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsTintScreen)
			r := float64(args.Red) / 255
			g := float64(args.Green) / 255
			b := float64(args.Blue) / 255
			gray := float64(args.Gray) / 255
			gameState.StartTint(r, g, b, gray, args.Time*6)
			if !args.Wait {
				i.commandIterator.Advance()
				return true, nil
			}
			i.waitingCommand = args.Wait
		}
		if gameState.IsChangingTint() {
			return false, nil
		}
		i.waitingCommand = false
		i.commandIterator.Advance()
	case data.CommandNamePlaySE:
		args := c.Args.(*data.CommandArgsPlaySE)
		v := float64(args.Volume) / data.MaxVolume
		audio.PlaySE(args.Name, v)
		i.commandIterator.Advance()
	case data.CommandNamePlayBGM:
		args := c.Args.(*data.CommandArgsPlayBGM)
		v := float64(args.Volume) / data.MaxVolume
		audio.PlayBGM(args.Name, v, args.FadeTime*6)
		i.commandIterator.Advance()
	case data.CommandNameStopBGM:
		args := c.Args.(*data.CommandArgsStopBGM)
		audio.StopBGM(args.FadeTime * 6)
		i.commandIterator.Advance()
	case data.CommandNameSave:
		gameState.RequestSave(sceneManager)
		i.commandIterator.Advance()
	case data.CommandNameAutoSave:
		args := c.Args.(*data.CommandArgsAutoSave)
		gameState.SetAutoSaveEnabled(args.Enabled)
		i.commandIterator.Advance()
	case data.CommandNameGameClear:
		gameState.Clear()
		i.commandIterator.Advance()
	case data.CommandNamePlayerControl:
		args := c.Args.(*data.CommandArgsPlayerControl)
		gameState.SetPlayerControlEnabled(args.Enabled)
		i.commandIterator.Advance()
	case data.CommandNameWeather:
		// TODO: Implement this command in the future.
		i.commandIterator.Advance()
	case data.CommandNameGotoTitle:
		return false, GoToTitle
	case data.CommandNameSyncIAP:
		i.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestGetIAPPrices(i.waitingRequestID)
		return false, nil
	case data.CommandNameUnlockAchievement:
		// TODO: Remove this command in the future.
		// Implement passive achievements instead.
		args := c.Args.(*data.CommandArgsUnlockAchievement)
		i.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestUnlockAchievement(i.waitingRequestID, args.ID)
		return false, nil
	case data.CommandNameControlHint:
		args := c.Args.(*data.CommandArgsControlHint)
		switch args.Type {
		case data.ControlHintPause:
			gameState.PauseHint(args.ID)
		case data.ControlHintStart:
			gameState.ActivateHint(args.ID)
		case data.ControlHintComplete:
			gameState.CompleteHint(args.ID)
		}
		i.commandIterator.Advance()
	case data.CommandNamePurchase:
		args := c.Args.(*data.CommandArgsPurchase)
		i.waitingRequestID = sceneManager.GenerateRequestID()

		var key string
		for _, i := range sceneManager.Game().IAPProducts {
			if i.ID == args.ID {
				key = i.Key
				break
			}
		}

		sceneManager.Requester().RequestPurchase(i.waitingRequestID, key)
		return false, nil
	case data.CommandNameShowAds:
		args := c.Args.(*data.CommandArgsShowAds)
		i.waitingRequestID = sceneManager.GenerateRequestID()
		switch args.Type {
		case data.ShowAdsTypeRewarded:
			sceneManager.Requester().RequestRewardedAds(i.waitingRequestID)
		case data.ShowAdsTypeInterstitial:
			sceneManager.Requester().RequestInterstitialAds(i.waitingRequestID)
		}
		audio.PauseBGM()
		return false, nil
	case data.CommandNameOpenLink:
		args := c.Args.(*data.CommandArgsOpenLink)
		i.waitingRequestID = sceneManager.GenerateRequestID()
		// TODO: Define data.OpenLinkType
		sceneManager.Requester().RequestOpenLink(i.waitingRequestID, args.Type, args.Data)
		return false, nil
	case data.CommandNameSendAnalytics:
		args := c.Args.(*data.CommandArgsSendAnalytics)
		sceneManager.Requester().RequestSendAnalytics(args.EventName, "")
		i.commandIterator.Advance()
	case data.CommandNameRequestReview:
		sceneManager.Requester().RequestReview()
		i.commandIterator.Advance()
	case data.CommandNameMoveCharacter:
		if ch := gameState.Character(i.mapID, i.roomID, i.eventID); ch == nil {
			i.commandIterator.Advance()
			return true, nil
		}
		if i.moveCharacterState == nil {
			args := c.Args.(*data.CommandArgsMoveCharacter)
			skip := i.routeSkip
			if args.Type == data.MoveCharacterTypeTarget {
				skip = false
			}
			m := movecharacterstate.New(
				gameState,
				i.mapID,
				i.roomID,
				i.eventID,
				args,
				skip)
			if m == nil {
				if i.routeSkip {
					i.commandIterator.Advance()
					return true, nil
				}
				return false, nil
			}
			i.moveCharacterState = m
		}
		i.moveCharacterState.Update(gameState)
		if !i.moveCharacterState.IsTerminated(gameState) {
			return false, nil
		}
		i.moveCharacterState = nil
		i.commandIterator.Advance()

	case data.CommandNameTurnCharacter:
		ch := gameState.Character(i.mapID, i.roomID, i.eventID)
		if ch == nil {
			i.commandIterator.Advance()
			return true, nil
		}
		// Check IsMoving() first since the character might be moving at this time.
		if ch.IsMoving() {
			return false, nil
		}
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsTurnCharacter)
			ch.SetDir(args.Dir)
			i.waitingCommand = true
			return false, nil
		}
		i.waitingCommand = false
		i.commandIterator.Advance()

	case data.CommandNameRotateCharacter:
		ch := gameState.Character(i.mapID, i.roomID, i.eventID)
		if ch == nil {
			i.commandIterator.Advance()
			return true, nil
		}
		// Check IsMoving() first since the character might be moving at this time.
		if ch.IsMoving() {
			return false, nil
		}
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsRotateCharacter)
			dirI := 0
			switch ch.Dir() {
			case data.DirUp:
				dirI = 0
			case data.DirRight:
				dirI = 1
			case data.DirDown:
				dirI = 2
			case data.DirLeft:
				dirI = 3
			default:
				panic("not reach")
			}
			switch args.Angle {
			case 0:
			case 90:
				dirI += 1
			case 180:
				dirI += 2
			case 270:
				dirI += 3
			default:
				panic("not reach")
			}
			dirI %= 4
			var dir data.Dir
			switch dirI {
			case 0:
				dir = data.DirUp
			case 1:
				dir = data.DirRight
			case 2:
				dir = data.DirDown
			case 3:
				dir = data.DirLeft
			default:
				panic("not reach")
			}
			ch.Turn(dir)
			i.waitingCommand = true
			return false, nil
		}
		i.waitingCommand = false
		i.commandIterator.Advance()
	case data.CommandNameSetCharacterProperty:
		args := c.Args.(*data.CommandArgsSetCharacterProperty)
		ch := gameState.Character(i.mapID, i.roomID, i.eventID)
		if ch == nil {
			i.commandIterator.Advance()
			return true, nil
		}
		switch args.Type {
		case data.SetCharacterPropertyTypeVisibility:
			ch.SetVisibility(args.Value.(bool))
		case data.SetCharacterPropertyTypeDirFix:
			ch.SetDirFix(args.Value.(bool))
		case data.SetCharacterPropertyTypeStepping:
			ch.SetStepping(args.Value.(bool))
		case data.SetCharacterPropertyTypeThrough:
			ch.SetThrough(args.Value.(bool))
		case data.SetCharacterPropertyTypeWalking:
			ch.SetWalking(args.Value.(bool))
		case data.SetCharacterPropertyTypeSpeed:
			ch.SetSpeed(args.Value.(data.Speed))
		default:
			return false, fmt.Errorf("invaid set_character_property type: %s", args.Type)
		}
		i.commandIterator.Advance()
	case data.CommandNameSetCharacterImage:
		args := c.Args.(*data.CommandArgsSetCharacterImage)
		ch := gameState.Character(i.mapID, i.roomID, i.eventID)
		if ch == nil {
			i.commandIterator.Advance()
			return true, nil
		}
		ch.SetImage(args.ImageType, args.Image)
		if args.UseFrameAndDir {
			ch.SetFrame(args.Frame)
			ch.SetDir(args.Dir)
		}
		i.commandIterator.Advance()

	case data.CommandNameSetCharacterOpacity:
		ch := gameState.Character(i.mapID, i.roomID, i.eventID)
		if ch == nil {
			i.commandIterator.Advance()
			return true, nil
		}
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsSetCharacterOpacity)
			ch.ChangeOpacity(args.Opacity, args.Time*6)
			if !args.Wait {
				i.commandIterator.Advance()
				return true, nil
			}
			i.waitingCommand = args.Wait
		}
		if ch.IsChangingOpacity() {
			return false, nil
		}
		i.waitingCommand = false
		i.commandIterator.Advance()

	case data.CommandNameAddItem:
		args := c.Args.(*data.CommandArgsAddItem)
		gameState.AddItem(args.ID)
		i.commandIterator.Advance()

	case data.CommandNameRemoveItem:
		args := c.Args.(*data.CommandArgsRemoveItem)
		gameState.RemoveItem(args.ID)
		i.commandIterator.Advance()

	case data.CommandNameShowInventory:
		gameState.SetInventoryVisible(true)
		i.commandIterator.Advance()

	case data.CommandNameHideInventory:
		gameState.SetInventoryVisible(false)
		i.commandIterator.Advance()

	case data.CommandNameShowItem:
		args := c.Args.(*data.CommandArgsShowItem)
		gameState.SetEventItem(args.ID)
		if gameState.Items().Includes(args.ID) {
			gameState.Items().Activate(args.ID)
			gameState.Items().SetCombineItem(0)
		}
		i.commandIterator.Advance()

	case data.CommandNameHideItem:
		gameState.SetEventItem(0)
		i.commandIterator.Advance()

	case data.CommandNameReplaceItem:
		args := c.Args.(*data.CommandArgsReplaceItem)
		for _, id := range args.ReplaceIDs {
			gameState.InsertItemBefore(args.ID, id)
		}
		gameState.RemoveItem(args.ID)
		i.commandIterator.Advance()

	case data.CommandNameShowPicture:
		args := c.Args.(*data.CommandArgsShowPicture)
		x := args.X
		y := args.Y
		if args.PosValueType == data.ValueTypeVariable {
			x = gameState.VariableValue(x)
			y = gameState.VariableValue(y)
		}
		scaleX := float64(args.ScaleX) / 100
		scaleY := float64(args.ScaleY) / 100
		angle := float64(args.Angle) * math.Pi / 180
		opacity := float64(args.Opacity) / 255
		gameState.pictures.Add(args.ID, args.Image, x, y, scaleX, scaleY, angle, opacity, args.OriginX, args.OriginY, args.BlendType)
		i.commandIterator.Advance()

	case data.CommandNameErasePicture:
		args := c.Args.(*data.CommandArgsErasePicture)
		gameState.pictures.Remove(args.ID)
		i.commandIterator.Advance()

	case data.CommandNameMovePicture:
		if i.waitingCount == 0 {
			args := c.Args.(*data.CommandArgsMovePicture)
			x := args.X
			y := args.Y
			if args.PosValueType == data.ValueTypeVariable {
				x = gameState.VariableValue(x)
				y = gameState.VariableValue(y)
			}
			gameState.pictures.MoveTo(args.ID, x, y, args.Time*6)
			if !args.Wait {
				i.commandIterator.Advance()
				return true, nil
			}
			i.waitingCount = args.Time * 6
		}
		if i.waitingCount > 0 {
			i.waitingCount--
		}
		if i.waitingCount > 0 {
			return false, nil
		}
		i.commandIterator.Advance()

	case data.CommandNameScalePicture:
		if i.waitingCount == 0 {
			args := c.Args.(*data.CommandArgsScalePicture)
			scaleX := float64(args.ScaleX) / 100
			scaleY := float64(args.ScaleY) / 100
			gameState.pictures.Scale(args.ID, scaleX, scaleY, args.Time*6)
			if !args.Wait {
				i.commandIterator.Advance()
				return true, nil
			}
			i.waitingCount = args.Time * 6
		}
		if i.waitingCount > 0 {
			i.waitingCount--
		}
		if i.waitingCount > 0 {
			return false, nil
		}
		i.commandIterator.Advance()

	case data.CommandNameRotatePicture:
		if i.waitingCount == 0 {
			args := c.Args.(*data.CommandArgsRotatePicture)
			angle := float64(args.Angle) * math.Pi / 180
			gameState.pictures.Rotate(args.ID, angle, args.Time*6)
			if !args.Wait {
				i.commandIterator.Advance()
				return true, nil
			}
			i.waitingCount = args.Time * 6
		}
		if i.waitingCount > 0 {
			i.waitingCount--
		}
		if i.waitingCount > 0 {
			return false, nil
		}
		i.commandIterator.Advance()

	case data.CommandNameFadePicture:
		if i.waitingCount == 0 {
			args := c.Args.(*data.CommandArgsFadePicture)
			opacity := float64(args.Opacity) / 255
			gameState.pictures.Fade(args.ID, opacity, args.Time*6)
			if !args.Wait {
				i.commandIterator.Advance()
				return true, nil
			}
			i.waitingCount = args.Time * 6
		}
		if i.waitingCount > 0 {
			i.waitingCount--
		}
		if i.waitingCount > 0 {
			return false, nil
		}
		i.commandIterator.Advance()

	case data.CommandNameTintPicture:
		if i.waitingCount == 0 {
			args := c.Args.(*data.CommandArgsTintPicture)
			r := float64(args.Red) / 255
			g := float64(args.Green) / 255
			b := float64(args.Blue) / 255
			gray := float64(args.Gray) / 255
			gameState.pictures.Tint(args.ID, r, g, b, gray, args.Time*6)
			if !args.Wait {
				i.commandIterator.Advance()
				return true, nil
			}
			i.waitingCount = args.Time * 6
		}
		if i.waitingCount > 0 {
			i.waitingCount--
		}
		if i.waitingCount > 0 {
			return false, nil
		}
		i.commandIterator.Advance()

	case data.CommandNameChangePictureImage:
		args := c.Args.(*data.CommandArgsChangePictureImage)
		gameState.pictures.ChangeImage(args.ID, args.Image)
		i.commandIterator.Advance()

	case data.CommandNameChangeBackground:
		args := c.Args.(*data.CommandArgsChangeBackground)
		gameState.Map().CurrentRoom().Background.Name = args.Image
		i.commandIterator.Advance()

	case data.CommandNameChangeForeground:
		args := c.Args.(*data.CommandArgsChangeForeground)
		gameState.Map().CurrentRoom().Foreground.Name = args.Image
		i.commandIterator.Advance()

	case data.CommandNameFinishPlayerMovingByUserInput:
		gameState.currentMap.FinishPlayerMovingByUserInput()
		i.commandIterator.Advance()

	case data.CommandNameExecEventHere:
		e := gameState.ExecutableEventAtPlayer()
		if e == nil {
			i.commandIterator.Advance()
			break
		}
		page := gameState.currentMap.currentPage(e)
		if page == nil {
			panic("not reached")
		}
		c := page.Commands
		i.sub = i.createChild(gameState, e.EventID(), c)

	case data.CommandNameMemo:
		i.commandIterator.Advance()

	default:
		return false, fmt.Errorf("interpreter: invalid command: %s", c.Name)
	}
	return true, nil
}

func (i *Interpreter) Update(sceneManager *scene.Manager, gameState *Game) error {
	if i.commandIterator == nil {
		return nil
	}
	for !i.commandIterator.IsTerminated() {
		cont, err := i.doOneCommand(sceneManager, gameState)
		if err != nil {
			return err
		}
		if !cont {
			break
		}
	}
	if i.commandIterator.IsTerminated() {
		if i.repeat {
			i.commandIterator.Rewind()
			return nil
		}
		if gameState.windows.IsBusy(i.id) {
			return nil
		}
		i.commandIterator = nil
		return nil
	}
	return nil
}
