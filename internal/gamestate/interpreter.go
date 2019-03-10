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
	"log"
	"math"
	"strconv"

	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
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
	mapID              int // Note: This doesn't make sense when eventID == PlayerEventID
	roomID             int // Note: This doesn't make sense when eventID == PlayerEventID
	eventID            int
	pageIndex          int
	commandIterator    *commanditerator.CommandIterator
	waitingCount       int
	waitingCommand     bool
	moveCharacterState *movecharacterstate.State
	repeat             bool
	sub                *Interpreter
	route              bool // True when used for event routing property.
	pageRoute          bool
	routeSkip          bool
	parallel           bool

	// Not dumped.
	waitingRequestID int
}

type InterpreterIDGenerator interface {
	GenerateInterpreterID() int
}

func NewInterpreter(idGen InterpreterIDGenerator, mapID, roomID, eventID, pageIndex int, commands []*data.Command) *Interpreter {
	return &Interpreter{
		id:              idGen.GenerateInterpreterID(),
		mapID:           mapID,
		roomID:          roomID,
		eventID:         eventID,
		pageIndex:       pageIndex,
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

	e.EncodeString("pageIndex")
	e.EncodeInt(i.pageIndex)

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

	e.EncodeString("pageRoute")
	e.EncodeBool(i.pageRoute)

	e.EncodeString("routeSkip")
	e.EncodeBool(i.routeSkip)

	e.EncodeString("parallel")
	e.EncodeBool(i.parallel)

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
		case "pageIndex":
			i.pageIndex = d.DecodeInt()
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
		case "pageRoute":
			i.pageRoute = d.DecodeBool()
		case "routeSkip":
			i.routeSkip = d.DecodeBool()
		case "parallel":
			i.parallel = d.DecodeBool()
		case "waitingRequestId":
			d.Skip()
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

func (i *Interpreter) createChild(gameState InterpreterIDGenerator, eventID int, pageIndex int, commands []*data.Command) *Interpreter {
	child := NewInterpreter(gameState, i.mapID, i.roomID, eventID, pageIndex, commands)
	child.route = i.route
	child.pageRoute = i.pageRoute
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
	// TODO: Instead of returnning boolean value, return enum value for code readability.

	// TODO: CanWindowProceed should always return true for route interpreters?
	if !i.route && !i.parallel && !gameState.CanWindowProceed(i.id) {
		return false, nil
	}
	if i.sub != nil {
		if err := i.sub.Update(sceneManager, gameState); err != nil {
			return false, err
		}
		if i.sub.IsExecuting() {
			return false, nil
		}
		i.sub = nil
		i.commandIterator.Advance()
		// Continue
		return true, nil
	}
	if i.waitingRequestID != 0 {
		r := sceneManager.ReceiveResultIfExists(i.waitingRequestID)
		if r == nil {
			return false, nil
		}
		i.waitingRequestID = 0
		switch r.Type {
		case scene.RequestTypePurchase, scene.RequestTypeRewardedAds, scene.RequestTypeShowShop:
			if r.Succeeded {
				i.commandIterator.Choose(0)
			} else {
				i.commandIterator.Choose(1)
			}
		case scene.RequestTypeIAPPrices:
			if r.Succeeded {
				var prices map[string]string
				if err := json.Unmarshal(r.Data, &prices); err != nil {
					return false, err
				}
				gameState.SetPrices(prices)
				i.commandIterator.Choose(0)
			} else {
				i.commandIterator.Choose(1)
			}
		case scene.RequestTypeSaveProgress:
			// The iterator is already proceeded.
		default:
			i.commandIterator.Advance()
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
	case data.CommandNameGroup:
		i.commandIterator.Choose(0)

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

		if i.mapID != gameState.currentMap.mapID {
			// TODO: warning?
			i.commandIterator.Advance()
			return true, nil
		}
		if i.roomID != gameState.currentMap.roomID {
			// TODO: warning?
			i.commandIterator.Advance()
			return true, nil
		}

		var event *data.Event
		for _, e := range gameState.CurrentEvents() {
			if e.ID() == eventID {
				event = e
				break
			}
		}
		if event == nil {
			// TODO: warning?
			i.commandIterator.Advance()
			return true, nil
		}
		page := event.Pages()[args.PageIndex]
		commands := page.Commands
		i.sub = i.createChild(gameState, eventID, args.PageIndex, commands)

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
		// TODO: Is this correct to the pass event id and the page index here?
		i.sub = i.createChild(gameState, i.eventID, i.pageIndex, c.Commands)

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
			id := args.EventID
			if id == 0 {
				id = i.eventID
			}
			messageStyle := i.findMessageStyle(sceneManager, args.MessageStyleID)
			if gameState.ShowBalloon(sceneManager, i.id, i.mapID, i.roomID, id, args.ContentID, args.BalloonType, messageStyle) {
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
			id := args.EventID
			if id == 0 {
				id = i.eventID
			}

			messageStyle := i.findMessageStyle(sceneManager, args.MessageStyleID)
			gameState.ShowMessage(sceneManager, i.id, id, args.ContentID, args.Background, args.PositionType, args.TextAlign, messageStyle)
			i.waitingCommand = true
			return false, nil
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

	case data.CommandNameShowHint:
		hintId := gameState.hints.ActiveHintID()
		// next time it shows next available hint
		gameState.hints.ReadHint(hintId)
		sceneManager.Requester().RequestSendAnalytics("show_hint", strconv.Itoa(hintId))
		hasHint := false
		for _, h := range sceneManager.Game().Hints {
			if h.ID == hintId {
				c := h.Commands
				i.sub = i.createChild(gameState, i.eventID, i.pageIndex, c)
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
			// Now there are other choice balloons. Let's wait.
			// TODO: I guess this never happens any longer.
			if gameState.windows.IsBusyWithChoosing() {
				return false, nil
			}
			gameState.ShowChoices(sceneManager, i.id, i.eventID, c.Args.(*data.CommandArgsShowChoices).ChoiceIDs, c.Args.(*data.CommandArgsShowChoices).Conditions)
			i.waitingCommand = true
			return false, nil
		}
		if !gameState.HasChosenWindowIndex() {
			return false, nil
		}
		if gameState.windows.IsBusy(i.id) {
			return false, nil
		}
		i.commandIterator.Choose(gameState.ChosenWindowIndex())
		i.waitingCommand = false

	case data.CommandNameSetSwitch:
		args := c.Args.(*data.CommandArgsSetSwitch)
		if args.ID >= variables.ReservedID && !args.Internal {
			return false, fmt.Errorf("gamestate: the switch ID (%d) must be < %d", args.ID, variables.ReservedID)
		}

		if args.IDType == data.SetSwitchIDTypeRef {
			gameState.SetSwitchRefValue(args.ID, args.Value)
		} else {
			gameState.SetSwitchValue(args.ID, args.Value)
		}
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

		if args.IDType == data.SetVariableIDTypeRef {
			if err := gameState.SetVariableRef(sceneManager, args.ID, args.Op, args.ValueType, args.Value, i.mapID, i.roomID, i.eventID); err != nil {
				return false, err
			}
		} else {
			if err := gameState.SetVariable(sceneManager, args.ID, args.Op, args.ValueType, args.Value, i.mapID, i.roomID, i.eventID); err != nil {
				return false, err
			}
		}

		i.commandIterator.Advance()

	case data.CommandNameSavePermanent:
		args := c.Args.(*data.CommandArgsSavePermanent)
		i.waitingRequestID = sceneManager.GenerateRequestID()
		gameState.RequestSavePermanentVariable(i.waitingRequestID, sceneManager, args.PermanentVariableID, args.VariableID)
		return false, nil

	case data.CommandNameLoadPermanent:
		args := c.Args.(*data.CommandArgsLoadPermanent)
		v := sceneManager.PermanentVariableValue(args.PermanentVariableID)
		if err := gameState.SetVariable(sceneManager, args.VariableID, data.SetVariableOpAssign, data.SetVariableValueTypeConstant, v, i.mapID, i.roomID, i.eventID); err != nil {
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
				roomID = int(gameState.VariableValue(roomID))
				x = int(gameState.VariableValue(x))
				y = int(gameState.VariableValue(y))
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
				roomID = int(gameState.VariableValue(roomID))
				x = int(gameState.VariableValue(x))
				y = int(gameState.VariableValue(y))
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
		sub := i.createChild(gameState, id, i.pageIndex, args.Commands)
		sub.repeat = args.Repeat
		sub.routeSkip = args.Skip

		if id != character.PlayerEventID && !args.Internal {
			gameState.Map().removeNonPageRoutes(id)
		}

		if !args.Wait {
			if id != character.PlayerEventID {
				// Set 'route' true so that the new route command does not
				// block the player's move (#380).
				sub.route = true
			}

			gameState.Map().addInterpreter(sub)
			i.commandIterator.Advance()
			return true, nil
		}

		i.sub = sub
	case data.CommandNameShake:
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsShake)
			if args.Time != 0 {
				gameState.StartShaking(args.Power, args.Speed, args.Time*6, args.Direction)
			} else {
				gameState.StopShaking()
			}
			forever := args.Time == -1
			if !args.Wait || forever {
				i.commandIterator.Advance()
				return true, nil
			}
			i.waitingCommand = args.Wait
		}
		if gameState.IsShaking() {
			return false, nil
		}
		i.waitingCommand = false
		i.commandIterator.Advance()
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
		// Proceed the command iterator before saving so that the game resumes from the next command.
		i.commandIterator.Advance()
		i.waitingRequestID = sceneManager.GenerateRequestID()
		gameState.RequestSave(i.waitingRequestID, sceneManager)
		return false, nil
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
	case data.CommandNamePlayerSpeed:
		args := c.Args.(*data.CommandArgsPlayerSpeed)
		gameState.SetPlayerSpeed(args.Value)
		i.commandIterator.Advance()
	case data.CommandNameWeather:
		args := c.Args.(*data.CommandArgsWeather)
		gameState.SetWeather(args.Type)
		i.commandIterator.Advance()
	case data.CommandNameGotoTitle:
		args := c.Args.(*data.CommandArgsGotoTitle)
		if args.Save {
			i.commandIterator.Advance()
			gameState.RequestSave(0, sceneManager)
		}
		return false, GoToTitle
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
			sceneManager.Requester().RequestRewardedAds(i.waitingRequestID, args.ForceAds)
		case data.ShowAdsTypeInterstitial:
			sceneManager.Requester().RequestInterstitialAds(i.waitingRequestID, args.ForceAds)
		}
		return false, nil

	case data.CommandNameOpenLink:
		args := c.Args.(*data.CommandArgsOpenLink)
		if args.Type == data.OpenLinkTypeShowCredit {
			gameState.ShowCredits()
			i.commandIterator.Advance()
			return false, nil
		}
		i.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(i.waitingRequestID, string(args.Type), args.Data)
		return false, nil

	case data.CommandNameShare:
		args := c.Args.(*data.CommandArgsShare)
		text := sceneManager.Game().Texts.Get(lang.Get(), args.TextID)

		var image []byte
		if args.Image != "" {
			i := assets.GetLocalizedImagePngBytes("pictures/" + args.Image)
			if i != nil {
				image = i
			}
		}
		i.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestShareImage(i.waitingRequestID, "", text, image)
		return false, nil

	case data.CommandNameSendAnalytics:
		args := c.Args.(*data.CommandArgsSendAnalytics)
		sceneManager.Requester().RequestSendAnalytics(args.EventName, "")
		// There is no need to wait for command. Proceed the command iterator.
		i.commandIterator.Advance()

	case data.CommandNameShowShop:
		i.waitingRequestID = sceneManager.GenerateRequestID()
		args := c.Args.(*data.CommandArgsShowShop)
		sceneManager.Requester().RequestShowShop(i.waitingRequestID, string(sceneManager.GetShopProductsData(args.Products)))
		return false, nil

	case data.CommandNameShowMinigame:
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsShowMinigame)

			score := 0
			var lastActiveAt int64
			if mg := sceneManager.PermanentMinigame(args.ID); mg != nil {
				score = mg.Score
				lastActiveAt = mg.LastActiveAt
			}
			// Resolve early in case minigame is already finished
			if score >= args.ReqScore {
				i.commandIterator.Choose(0)
				return false, nil
			}
			gameState.InitMinigame(args.ID, args.ReqScore, score)
			gameState.ShowMinigame(lastActiveAt)
			// In order to take care a case when ads are removed,
			// we have to notify the platform to initializing the ads here
			sceneManager.Requester().RequestOpenLink(0, "initialize_ads", "")
			i.waitingCommand = true
		}

		if gameState.Minigame().Active() {
			return false, nil
		}

		if gameState.Minigame().Success() {
			i.commandIterator.Choose(0)
		} else {
			i.commandIterator.Choose(1)
		}
		i.waitingCommand = false
		return false, nil

	case data.CommandNameVibrate:
		args := c.Args.(*data.CommandArgsVibrate)
		sceneManager.Requester().RequestVibration(args.Type)
		// There is no need to wait for command. Proceed the command iterator.
		i.commandIterator.Advance()
	case data.CommandNameRequestReview:
		sceneManager.Requester().RequestReview()
		// There is no need to wait for command. Proceed the command iterator.
		i.commandIterator.Advance()

	case data.CommandNameMoveCharacter:
		if ch := gameState.Character(i.mapID, i.roomID, i.eventID); ch == nil {
			i.commandIterator.Advance()
			return true, nil
		}
		if i.moveCharacterState == nil {
			args := c.Args.(*data.CommandArgsMoveCharacter)
			skip := i.routeSkip
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
				panic(fmt.Sprintf("gamestate: invalid character dir: %d at data.CommandNameRouteCharacter", ch.Dir()))
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
				panic(fmt.Sprintf("gamestate: invalid angle: %d at data.CommandNameRouteCharacter", args.Angle))
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
				panic(fmt.Sprintf("gamestate: invalid dir: %d at data.CommandNameRouteCharacter", dirI))
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
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsShowInventory)
			if !args.Wait {
				gameState.ShowInventory(args.Group, false, false)
				i.commandIterator.Advance()
				return true, nil
			}
			gameState.ShowInventory(args.Group, true, args.Cancelable)
			i.waitingCommand = true
		}

		if gameState.InventoryVisible() {
			return false, nil
		}

		i.commandIterator.Advance()
		i.waitingCommand = false

	case data.CommandNameHideInventory:
		gameState.HideInventory()
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
		id := args.ID
		if args.IDValueType == data.ValueTypeVariable {
			id = int(gameState.VariableValue(id))
		}
		if args.PosValueType == data.ValueTypeVariable {
			x = int(gameState.VariableValue(x))
			y = int(gameState.VariableValue(y))
		}
		scaleX := float64(args.ScaleX) / 100
		scaleY := float64(args.ScaleY) / 100
		angle := float64(args.Angle) * math.Pi / 180
		opacity := float64(args.Opacity) / 255
		gameState.pictures.Add(id, args.Image, x, y, scaleX, scaleY, angle, opacity, args.OriginX, args.OriginY, args.BlendType, args.Priority)
		i.commandIterator.Advance()

	case data.CommandNameErasePicture:
		args := c.Args.(*data.CommandArgsErasePicture)
		if args.SelectType == data.SelectTypeMulti {
			ids := make([]int, 2)
			interfaces := args.ID.([]interface{})
			for i := 0; i < 2; i++ {
				ok := false
				ids[i], ok = data.InterfaceToInt(interfaces[i])
				if !ok {
					return false, fmt.Errorf("gamestate: %v must be integer but not", interfaces[i])
				}
			}

			id1 := ids[0]
			id2 := ids[1]
			for i := id1; i <= id2; i++ {
				id := i
				if args.IDValueType == data.ValueTypeVariable {
					id = int(gameState.VariableValue(i))
				}
				gameState.pictures.Remove(id)
			}
		} else {
			id, ok := data.InterfaceToInt(args.ID)
			if !ok {
				return false, fmt.Errorf("gamestate: %v must be integer but not", args.ID)
			}
			if args.IDValueType == data.ValueTypeVariable {
				id = int(gameState.VariableValue(id))
			}
			gameState.pictures.Remove(id)
		}
		i.commandIterator.Advance()

	case data.CommandNameMovePicture:
		if i.waitingCount == 0 {
			args := c.Args.(*data.CommandArgsMovePicture)
			id := args.ID
			if args.IDValueType == data.ValueTypeVariable {
				id = int(gameState.VariableValue(id))
			}
			x := args.X
			y := args.Y
			if args.PosValueType == data.ValueTypeVariable {
				x = int(gameState.VariableValue(x))
				y = int(gameState.VariableValue(y))
			}
			gameState.pictures.MoveTo(id, x, y, args.Time*6)
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

			id := args.ID
			if args.IDValueType == data.ValueTypeVariable {
				id = int(gameState.VariableValue(id))
			}

			tx := args.ScaleX
			ty := args.ScaleY
			if args.ScaleValueType == data.ValueTypeVariable {
				tx = int(gameState.VariableValue(tx))
				ty = int(gameState.VariableValue(ty))
			}
			scaleX := float64(tx) / 100
			scaleY := float64(ty) / 100
			gameState.pictures.Scale(id, scaleX, scaleY, args.Time*6)
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

			id := args.ID
			if args.IDValueType == data.ValueTypeVariable {
				id = int(gameState.VariableValue(id))
			}

			t := args.Angle
			if args.AngleValueType == data.ValueTypeVariable {
				t = int(gameState.VariableValue(t))
			}
			angle := float64(t) * math.Pi / 180
			gameState.pictures.Rotate(id, angle, args.Time*6)
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
			id := args.ID
			if args.IDValueType == data.ValueTypeVariable {
				id = int(gameState.VariableValue(id))
			}

			opacity := args.Opacity
			if args.OpacityValueType == data.ValueTypeVariable {
				opacity = int(gameState.VariableValue(opacity))
			}
			o := float64(opacity) / 255
			gameState.pictures.Fade(id, o, args.Time*6)
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
			id := args.ID
			if args.IDValueType == data.ValueTypeVariable {
				id = int(gameState.VariableValue(id))
			}

			r := float64(args.Red) / 255
			g := float64(args.Green) / 255
			b := float64(args.Blue) / 255
			gray := float64(args.Gray) / 255
			gameState.pictures.Tint(id, r, g, b, gray, args.Time*6)
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
		id := args.ID
		if args.IDValueType == data.ValueTypeVariable {
			id = int(gameState.VariableValue(id))
		}

		gameState.pictures.ChangeImage(id, args.Image)
		i.commandIterator.Advance()

	case data.CommandNameChangeBackground:
		args := c.Args.(*data.CommandArgsChangeBackground)
		gameState.SetBackground(i.mapID, i.roomID, args.Image)
		i.commandIterator.Advance()

	case data.CommandNameChangeForeground:
		args := c.Args.(*data.CommandArgsChangeForeground)
		gameState.SetForeground(i.mapID, i.roomID, args.Image)
		i.commandIterator.Advance()

	case data.CommandNameSpecial:
		args := c.Args.(*data.CommandArgsSpecial)
		var content map[string]interface{}
		if err := json.Unmarshal([]byte(args.Content), &content); err != nil {
			return false, err
		}
		switch name := content["name"].(string); name {
		case "shake_start_game_button":
			gameState.ShakeStartGameButton()
		default:
			return false, fmt.Errorf("gamestate: invalid special command name: %q", name)
		}
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
		page, pageIndex := gameState.currentMap.currentPage(e)
		if page == nil {
			panic("gamestate: no page was found at data.CommandNameExecEventHere")
		}
		c := page.Commands
		i.sub = i.createChild(gameState, e.EventID(), pageIndex, c)

	case data.CommandNameMemo:
		args := c.Args.(*data.CommandArgsMemo)
		if args.Log {
			log.Print(gameState.ParseMessageSyntax(sceneManager, args.Content))
		}
		i.commandIterator.Advance()

	default:
		return false, fmt.Errorf("gamestate: invalid command: %s", c.Name)
	}

	// Continue
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

type ShopProduct struct {
	Key  string `json:"key" msgpack:"key"`
	Name string `json:"name" msgpack:"name"`
	Desc string `json:"desc" msgpack:"desc"`
}
