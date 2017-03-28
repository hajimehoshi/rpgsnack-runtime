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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/commanditerator"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type Interpreter struct {
	id                 int
	mapID              int // Note: This doesn't make sense when eventID == -1
	roomID             int // Note: This doesn't make sense when eventID == -1
	eventID            int
	commandIterator    *commanditerator.CommandIterator
	waitingCount       int
	waitingCommand     bool
	moveCharacterState *moveCharacterState
	repeat             bool
	sub                *Interpreter
	route              bool // True when used for event routing property.
	routeSkip          bool
	shouldGoToTitle    bool
	waitingRequestID   int // Note: When this is not 0, the game state can't be saved.

	// Fields that are not dumped
	gameState *Game
}

func NewInterpreter(gameState *Game, mapID, roomID, eventID int, commands []*data.Command) *Interpreter {
	return &Interpreter{
		id:              gameState.generateInterpreterID(),
		gameState:       gameState,
		mapID:           mapID,
		roomID:          roomID,
		eventID:         eventID,
		commandIterator: commanditerator.New(commands),
	}
}

type tmpInterpreter struct {
	ID                 int                              `json:"id"`
	MapID              int                              `json:"mapId"`
	RoomID             int                              `json:"roomId"`
	EventID            int                              `json:"eventId"`
	CommandIterator    *commanditerator.CommandIterator `json:"commandIterator"`
	WaitingCount       int                              `json:"waitingCount"`
	WaitingCommand     bool                             `json:"waitingCommand"`
	MoveCharacterState *moveCharacterState              `json:"moveCharacterState"`
	Repeat             bool                             `json:"repeat"`
	Sub                *Interpreter                     `json:"sub"`
	Route              bool                             `json:"route"`
	RouteSkip          bool                             `json:"routeSkip"`
	ShouldGoToTitle    bool                             `json:"shouldGoToTitle"`
	WaitingRequestID   int                              `json:"waitingRequestId"`
}

func (i *Interpreter) MarshalJSON() ([]uint8, error) {
	tmp := &tmpInterpreter{
		ID:                 i.id,
		MapID:              i.mapID,
		RoomID:             i.roomID,
		EventID:            i.eventID,
		CommandIterator:    i.commandIterator,
		WaitingCount:       i.waitingCount,
		WaitingCommand:     i.waitingCommand,
		MoveCharacterState: i.moveCharacterState,
		Repeat:             i.repeat,
		Sub:                i.sub,
		Route:              i.route,
		RouteSkip:          i.routeSkip,
		ShouldGoToTitle:    i.shouldGoToTitle,
		WaitingRequestID:   i.waitingRequestID,
	}
	return json.Marshal(tmp)
}

func (i *Interpreter) UnmarshalJSON(data []uint8) error {
	var tmp *tmpInterpreter
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	i.id = tmp.ID
	i.mapID = tmp.MapID
	i.roomID = tmp.RoomID
	i.eventID = tmp.EventID
	i.commandIterator = tmp.CommandIterator
	i.waitingCount = tmp.WaitingCount
	i.waitingCommand = tmp.WaitingCommand
	i.moveCharacterState = tmp.MoveCharacterState
	i.repeat = tmp.Repeat
	i.sub = tmp.Sub
	i.route = tmp.Route
	i.routeSkip = tmp.RouteSkip
	i.shouldGoToTitle = tmp.ShouldGoToTitle
	i.waitingRequestID = tmp.WaitingRequestID
	return nil
}

func (i *Interpreter) setGame(game *Game) {
	i.gameState = game
	if i.moveCharacterState != nil {
		i.moveCharacterState.setGame(game)
	}
}

func (i *Interpreter) waitingRequestResponse() bool {
	return i.waitingRequestID != 0
}

func (i *Interpreter) event() *character.Character {
	if i.eventID == character.PlayerEventID {
		return nil
	}
	if i.gameState.Map().mapID != i.mapID {
		return nil
	}
	if i.gameState.Map().roomID != i.roomID {
		return nil
	}
	for _, e := range i.gameState.Map().events {
		if i.eventID == e.EventID() {
			return e
		}
	}
	return nil
}

func (i *Interpreter) IsExecuting() bool {
	return i.commandIterator != nil
}

func (i *Interpreter) createChild(eventID int, commands []*data.Command) *Interpreter {
	child := NewInterpreter(i.gameState, i.mapID, i.roomID, eventID, commands)
	child.route = i.route
	return child
}

func (i *Interpreter) doOneCommand(sceneManager *scene.Manager) (bool, error) {
	if !i.gameState.windows.CanProceed(i.id) {
		return false, nil
	}
	if i.sub != nil {
		if err := i.sub.Update(sceneManager); err != nil {
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
		default:
			i.commandIterator.Advance()
		}
		return true, nil
	}
	c := i.commandIterator.Command()
	switch c.Name {
	case data.CommandNameIf:
		conditions := c.Args.(*data.CommandArgsIf).Conditions
		matches := true
		for _, c := range conditions {
			m, err := i.gameState.meetsCondition(c, i.eventID)
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
		room := i.gameState.Map().CurrentRoom()
		var event *data.Event
		for _, e := range room.Events {
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
		i.sub = i.createChild(eventID, commands)
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
	case data.CommandNameShowMessage:
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsShowMessage)
			content := data.Current().Texts.Get(sceneManager.Language(), args.ContentID)
			id := args.EventID
			if id == 0 {
				id = i.eventID
			}
			if ch := i.gameState.character(i.mapID, i.roomID, id); ch != nil {
				content = i.gameState.parseMessageSyntax(content)
				i.gameState.windows.ShowMessage(args.Type, content, args.BalloonType, args.PositionType, ch.EventID(), i.id)
				i.waitingCommand = true
				return false, nil
			}
		}
		// Advance command index first and check the next command.
		i.commandIterator.Advance()
		if !i.commandIterator.IsTerminated() {
			if i.commandIterator.Command().Name != data.CommandNameShowChoices {
				i.gameState.windows.CloseAll()
			}
		} else {
			i.gameState.windows.CloseAll()
		}
		i.waitingCommand = false
	case data.CommandNameShowHint:
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsShowHint)
			hintId := i.gameState.hints.ActiveHintId()
			// next time it shows next available hint
			i.gameState.hints.ReadHint(hintId)

			var hintText data.UUID
			for _, h := range data.Current().Hints {
				if h.ID == hintId {
					hintText = h.Text
					break
				}
			}

			var content string
			if hintText.String() == "" {
				content = "Undefined"
			} else {
				content = data.Current().Texts.Get(sceneManager.Language(), hintText)
			}

			id := args.EventID
			if id == 0 {
				id = i.eventID
			}
			if ch := i.gameState.character(i.mapID, i.roomID, id); ch != nil {
				content = i.gameState.parseMessageSyntax(content)
				i.gameState.windows.ShowMessage(args.Type, content, args.BalloonType, args.PositionType, ch.EventID(), i.id)
				i.waitingCommand = true
				return false, nil
			}
		}

		i.commandIterator.Advance()
		i.gameState.windows.CloseAll()
		i.waitingCommand = false
	case data.CommandNameShowChoices:
		if !i.waitingCommand {
			choices := []string{}
			for _, id := range c.Args.(*data.CommandArgsShowChoices).ChoiceIDs {
				choice := data.Current().Texts.Get(sceneManager.Language(), id)
				choice = i.gameState.parseMessageSyntax(choice)
				choices = append(choices, choice)
			}
			i.gameState.windows.ShowChoices(sceneManager, choices, i.id)
			i.waitingCommand = true
			return false, nil
		}
		if !i.gameState.windows.HasChosenIndex() {
			return false, nil
		}
		i.commandIterator.Choose(i.gameState.windows.ChosenIndex())
		i.waitingCommand = false
	case data.CommandNameSetSwitch:
		args := c.Args.(*data.CommandArgsSetSwitch)
		i.gameState.variables.SetSwitchValue(args.ID, args.Value)
		i.commandIterator.Advance()
		// Suspend executing to give other events chances to update their pages.
		return false, nil
	case data.CommandNameSetSelfSwitch:
		args := c.Args.(*data.CommandArgsSetSelfSwitch)
		m, r := i.gameState.Map().mapID, i.gameState.Map().roomID
		i.gameState.variables.SetSelfSwitchValue(m, r, i.eventID, args.ID, args.Value)
		i.commandIterator.Advance()
		// Suspend executing to give other events chances to update their pages.
		return false, nil
	case data.CommandNameSetVariable:
		args := c.Args.(*data.CommandArgsSetVariable)
		i.setVariable(sceneManager, args.ID, args.Op, args.ValueType, args.Value)
		i.commandIterator.Advance()
		// Suspend executing to give other events chances to update their pages.
		return false, nil
	case data.CommandNameTransfer:
		args := c.Args.(*data.CommandArgsTransfer)
		if !i.waitingCommand {
			i.gameState.screen.fadeOut(30)
			i.waitingCommand = true
			return false, nil
		}
		if i.gameState.screen.isFadedOut() {
			i.gameState.Map().transferPlayerImmediately(args.RoomID, args.X, args.Y, i)
			i.gameState.screen.fadeIn(30)
			return false, nil
		}
		if i.gameState.screen.isFading() {
			return false, nil
		}
		i.waitingCommand = false
		i.commandIterator.Advance()
	case data.CommandNameSetRoute:
		args := c.Args.(*data.CommandArgsSetRoute)
		id := args.EventID
		if id == 0 {
			id = i.eventID
		}
		sub := i.createChild(id, args.Commands)
		sub.repeat = args.Repeat
		sub.routeSkip = args.Skip
		if !args.Wait {
			// TODO: What if set_route w/o waiting already exists for this event?
			i.gameState.Map().addInterpreter(sub)
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
			i.gameState.screen.startTint(r, g, b, gray, args.Time*6)
			if !args.Wait {
				i.commandIterator.Advance()
				return true, nil
			}
			i.waitingCommand = args.Wait
		}
		if i.gameState.screen.isChangingTint() {
			return false, nil
		}
		i.waitingCommand = false
		i.commandIterator.Advance()
	case data.CommandNamePlaySE:
		args := c.Args.(*data.CommandArgsPlaySE)
		v := float64(args.Volume) / data.MaxVolume
		if err := audio.PlaySE(args.Name, v); err != nil {
			return false, err
		}
		i.commandIterator.Advance()
	case data.CommandNamePlayBGM:
		args := c.Args.(*data.CommandArgsPlayBGM)
		v := float64(args.Volume) / data.MaxVolume
		if err := audio.PlayBGM(args.Name, v); err != nil {
			return false, err
		}
		if args.FadeTime > 0 {
			println(fmt.Sprintf("fade time is not used so far: %d"), args.FadeTime)
		}
		i.commandIterator.Advance()
	case data.CommandNameStopBGM:
		if err := audio.StopBGM(); err != nil {
			return false, err
		}
		i.commandIterator.Advance()
	case data.CommandNameGotoTitle:
		i.shouldGoToTitle = true
		return false, GoToTitle
	case data.CommandUnlockAchievement:
		// TODO: Remove this command in the future.
		// Implement passive achievements instead.
		args := c.Args.(*data.CommandArgsUnlockAchievement)
		i.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestUnlockAchievement(i.waitingRequestID, args.ID)
		return false, nil
	case data.CommandControlHint:
		args := c.Args.(*data.CommandArgsControlHint)

		switch args.Type {
		case data.ControlHintPause:
			i.gameState.hints.Pause(args.ID)
		case data.ControlHintStart:
			i.gameState.hints.Activate(args.ID)
		case data.ControlHintComplete:
			i.gameState.hints.Complete(args.ID)
		}
		i.commandIterator.Advance()
	case data.CommandPurchase:
		// args := c.Args.(*data.CommandArgsPurchase)
		i.waitingRequestID = sceneManager.GenerateRequestID()

		// TODO: Retrieve productID from purchases database via args.ID
		sceneManager.Requester().RequestPurchase(i.waitingRequestID, "test_productId")
		return false, nil
	case data.CommandShowAds:
		args := c.Args.(*data.CommandArgsShowAds)
		i.waitingRequestID = sceneManager.GenerateRequestID()
		switch args.Type {
		case data.ShowAdsTypeRewarded:
			sceneManager.Requester().RequestRewardedAds(i.waitingRequestID)
		case data.ShowAdsTypeInterstitial:
			sceneManager.Requester().RequestInterstitialAds(i.waitingRequestID)
		}
		return false, nil
	case data.CommandOpenLink:
		args := c.Args.(*data.CommandArgsOpenLink)
		i.waitingRequestID = sceneManager.GenerateRequestID()
		// TODO: Define data.OpenLinkType
		sceneManager.Requester().RequestOpenLink(i.waitingRequestID, args.Type, args.Data)
		return false, nil
	case data.CommandNameMoveCharacter:
		ch := i.gameState.character(i.mapID, i.roomID, i.eventID)
		if ch == nil {
			i.commandIterator.Advance()
			return true, nil
		}
		if i.moveCharacterState == nil {
			m, err := newMoveCharacterState(
				i.gameState,
				i.mapID,
				i.roomID,
				i.eventID,
				c.Args.(*data.CommandArgsMoveCharacter),
				i.routeSkip)
			if err != nil {
				return false, err
			}
			i.moveCharacterState = m
		}
		if err := i.moveCharacterState.Update(); err != nil {
			return false, err
		}
		if !i.moveCharacterState.IsTerminated() {
			return false, nil
		}
		i.moveCharacterState = nil
		i.commandIterator.Advance()
	case data.CommandNameTurnCharacter:
		ch := i.gameState.character(i.mapID, i.roomID, i.eventID)
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
		ch := i.gameState.character(i.mapID, i.roomID, i.eventID)
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
		ch := i.gameState.character(i.mapID, i.roomID, i.eventID)
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
		ch := i.gameState.character(i.mapID, i.roomID, i.eventID)
		if ch == nil {
			i.commandIterator.Advance()
			return true, nil
		}
		ch.SetImage(args.Image, args.ImageIndex)
		if args.UseFrameAndDir {
			ch.SetFrame(args.Frame)
			ch.SetDir(args.Dir)
		}
		i.commandIterator.Advance()

	case data.CommandNameSetInnerVariable:
		args := c.Args.(*data.CommandArgsSetInnerVariable)
		i.gameState.variables.SetInnerVariableValue(args.Name, args.Value)
		i.commandIterator.Advance()
	default:
		return false, fmt.Errorf("invalid command: %s", c.Name)
	}
	return true, nil
}

func (i *Interpreter) Update(sceneManager *scene.Manager) error {
	if i.commandIterator == nil {
		return nil
	}
	for !i.commandIterator.IsTerminated() {
		cont, err := i.doOneCommand(sceneManager)
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
		if i.gameState.windows.IsBusy(i.id) {
			return nil
		}
		i.commandIterator = nil
		return nil
	}
	return nil
}

func (i *Interpreter) setVariable(sceneManager *scene.Manager, id int, op data.SetVariableOp, valueType data.SetVariableValueType, value interface{}) {
	rhs := 0
	switch valueType {
	case data.SetVariableValueTypeConstant:
		rhs = value.(int)
	case data.SetVariableValueTypeVariable:
		rhs = i.gameState.variables.VariableValue(value.(int))
	case data.SetVariableValueTypeRandom:
		v := value.(*data.SetVariableValueRandom)
		rhs = i.gameState.RandomValue(v.Begin, v.End+1)
	case data.SetVariableValueTypeCharacter:
		args := value.(*data.SetVariableCharacterArgs)
		id := args.EventID
		if id == 0 {
			id = i.eventID
		}
		ch := i.gameState.character(i.mapID, i.roomID, id)
		if ch == nil {
			// TODO: return error?
			return
		}
		dir := ch.Dir()
		switch args.Type {
		case data.SetVariableCharacterTypeDirection:
			switch dir {
			case data.DirUp:
				rhs = 0
			case data.DirRight:
				rhs = 1
			case data.DirDown:
				rhs = 2
			case data.DirLeft:
				rhs = 3
			default:
				panic("not reach")
			}
		default:
			println(fmt.Sprintf("not implemented yet (set_variable): type %s", args.Type))
		}
	case data.SetVariableValueTypeSystem:
		systemVariableType := value.(data.SystemVariableType)
		switch systemVariableType {
		case data.SystemVariableHintCount:
			rhs = i.gameState.hints.ActiveHintCount()
		case data.SystemVariableInterstitialAdsLoaded, data.SystemVariableRewardedAdsLoaded:
			if i.gameState.variables.InnerVariableValue(string(systemVariableType)) == 1 {
				rhs = 1
			}
		default:
			println(fmt.Sprintf("not implemented yet (set_variable): systemVariableType %s", systemVariableType))
		}
	}
	switch op {
	case data.SetVariableOpAssign:
	case data.SetVariableOpAdd:
		rhs = i.gameState.variables.VariableValue(id) + rhs
	case data.SetVariableOpSub:
		rhs = i.gameState.variables.VariableValue(id) - rhs
	case data.SetVariableOpMul:
		rhs = i.gameState.variables.VariableValue(id) * rhs
	case data.SetVariableOpDiv:
		rhs = i.gameState.variables.VariableValue(id) / rhs
	case data.SetVariableOpMod:
		rhs = i.gameState.variables.VariableValue(id) % rhs
	}
	i.gameState.variables.SetVariableValue(id, rhs)
}
