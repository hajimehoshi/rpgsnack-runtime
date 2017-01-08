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
	"fmt"

	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type char interface {
	Position() (int, int)
	Dir() data.Dir
	IsMoving() bool
	Move(dir data.Dir)
	Turn(dir data.Dir)
}

type Interpreter struct {
	gameState      *Game
	mapID          int // Note: This doesn't make sense when eventID == -1
	roomID         int // Note: This doesn't make sense when eventID == -1
	eventID        int
	commandIndex   *commandIndex
	waitingCount   int
	waitingCommand bool
	sub            *Interpreter
}

func NewInterpreter(gameState *Game, mapID, roomID, eventID int, commands []*data.Command) *Interpreter {
	return &Interpreter{
		gameState:    gameState,
		mapID:        mapID,
		roomID:       roomID,
		eventID:      eventID,
		commandIndex: newCommandIndex(commands),
	}
}

func (i *Interpreter) event() *character.Event {
	if i.eventID == -1 {
		return nil
	}
	if i.gameState.mapID != i.mapID {
		return nil
	}
	if i.gameState.roomID != i.roomID {
		return nil
	}
	for _, e := range i.gameState.events {
		if i.eventID == e.ID() {
			return e
		}
	}
	return nil
}

func (i *Interpreter) IsExecuting() bool {
	return i.commandIndex != nil
}

func (i *Interpreter) character(id int) char {
	if id == -1 {
		return i.gameState.player
	}
	if i.gameState.mapID != i.mapID {
		return nil
	}
	if i.gameState.roomID != i.roomID {
		return nil
	}
	if id == 0 {
		id = i.eventID
	}
	for _, e := range i.gameState.events {
		if id == e.ID() {
			return e
		}
	}
	return nil
}

func (i *Interpreter) doOneCommand() (bool, error) {
	c := i.commandIndex.command()
	if !i.gameState.windows.CanProceed() {
		return false, nil
	}
	if i.sub != nil {
		if err := i.sub.Update(); err != nil {
			return false, err
		}
		if !i.sub.IsExecuting() {
			i.sub = nil
			i.commandIndex.advance()
		}
		return false, nil
	}
	switch c.Name {
	case data.CommandNameIf:
		conditions := c.Args.(*data.CommandArgsIf).Conditions
		matches := true
		for _, c := range conditions {
			m, err := i.gameState.MeetsCondition(c, i.eventID)
			if err != nil {
				return false, err
			}
			if !m {
				matches = false
				break
			}
		}
		if matches {
			i.commandIndex.choose(0)
		} else if len(c.Branches) >= 2 {
			i.commandIndex.choose(1)
		} else {
			i.commandIndex.advance()
		}
	case data.CommandNameCallEvent:
		args := c.Args.(*data.CommandArgsCallEvent)
		eventID := args.EventID
		if eventID == 0 {
			eventID = i.eventID
		}
		// TODO: Should i.mapID and i.roomID be considered here?
		room := i.gameState.CurrentRoom()
		var event *data.Event
		for _, e := range room.Events {
			if e.ID == eventID {
				event = e
				break
			}
		}
		if event == nil {
			// TODO: warning?
			i.commandIndex.advance()
			return true, nil
		}
		page := event.Pages[args.PageIndex]
		commands := page.Commands
		i.sub = NewInterpreter(i.gameState, i.mapID, i.roomID, eventID, commands)
	case data.CommandNameWait:
		if i.waitingCount == 0 {
			i.waitingCount = c.Args.(*data.CommandArgsWait).Time * 6
		}
		if i.waitingCount == 0 {
			// Time 0.0[s] is specified.
			i.commandIndex.advance()
			return true, nil
		}
		i.waitingCount--
		if i.waitingCount == 0 {
			i.commandIndex.advance()
			return true, nil
		}
		return false, nil
	case data.CommandNameShowMessage:
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsShowMessage)
			content := data.Current().Texts.Get(language.Und, args.ContentID)
			if ch := i.character(args.EventID); ch != nil {
				x, y := ch.Position()
				content = i.gameState.ParseMessageSyntax(content)
				i.gameState.windows.ShowMessage(content, x*scene.TileSize, y*scene.TileSize)
				i.waitingCommand = true
				return false, nil
			}
		}
		// Advance command index first and check the next command.
		i.commandIndex.advance()
		if !i.commandIndex.isTerminated() {
			if i.commandIndex.command().Name != data.CommandNameShowChoices {
				i.gameState.windows.CloseAll()
			}
		} else {
			i.gameState.windows.CloseAll()
		}
		i.waitingCommand = false
	case data.CommandNameShowChoices:
		if !i.waitingCommand {
			choices := []string{}
			for _, id := range c.Args.(*data.CommandArgsShowChoices).ChoiceIDs {
				choice := data.Current().Texts.Get(language.Und, id)
				choice = i.gameState.ParseMessageSyntax(choice)
				choices = append(choices, choice)
			}
			i.gameState.windows.ShowChoices(choices)
			i.waitingCommand = true
			return false, nil
		}
		if !i.gameState.windows.HasChosenIndex() {
			return false, nil
		}
		i.commandIndex.choose(i.gameState.windows.ChosenIndex())
		i.waitingCommand = false
	case data.CommandNameSetSwitch:
		args := c.Args.(*data.CommandArgsSetSwitch)
		i.gameState.variables.SetSwitchValue(args.ID, args.Value)
		i.commandIndex.advance()
	case data.CommandNameSetSelfSwitch:
		args := c.Args.(*data.CommandArgsSetSelfSwitch)
		m, r := i.gameState.mapID, i.gameState.roomID
		i.gameState.variables.SetSelfSwitchValue(m, r, i.eventID, args.ID, args.Value)
		i.commandIndex.advance()
	case data.CommandNameSetVariable:
		args := c.Args.(*data.CommandArgsSetVariable)
		i.setVariable(args.ID, args.Op, args.ValueType, args.Value)
		i.commandIndex.advance()
	case data.CommandNameTransfer:
		args := c.Args.(*data.CommandArgsTransfer)
		if !i.waitingCommand {
			i.gameState.screen.fadeOut(30)
			i.waitingCommand = true
			return false, nil
		}
		if i.gameState.screen.isFadedOut() {
			i.gameState.transferPlayerImmediately(args.RoomID, args.X, args.Y, i)
			i.gameState.screen.fadeIn(30)
			return false, nil
		}
		if i.gameState.screen.isFading() {
			return false, nil
		}
		i.waitingCommand = false
		i.commandIndex.advance()
	case data.CommandNameSetRoute:
		args := c.Args.(*data.CommandArgsSetRoute)
		id := args.EventID
		if id == 0 {
			id = i.eventID
		}
		i.sub = NewInterpreter(i.gameState, i.mapID, i.roomID, id, args.Commands)
	case data.CommandNameTintScreen:
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsTintScreen)
			r := float64(args.Red) / 255
			g := float64(args.Green) / 255
			b := float64(args.Blue) / 255
			gray := float64(args.Gray) / 255
			i.gameState.screen.startTint(r, g, b, gray, args.Time*6)
			if !args.Wait {
				i.commandIndex.advance()
				return true, nil
			}
			i.waitingCommand = args.Wait
		}
		if i.gameState.screen.isChangingTint() {
			return false, nil
		}
		i.waitingCommand = false
		i.commandIndex.advance()
	case data.CommandNamePlaySE:
		println(fmt.Sprintf("not implemented yet: %s", c.Name))
		i.commandIndex.advance()
	case data.CommandNamePlayBGM:
		println(fmt.Sprintf("not implemented yet: %s", c.Name))
		i.commandIndex.advance()
	case data.CommandNameStopBGM:
		println(fmt.Sprintf("not implemented yet: %s", c.Name))
		i.commandIndex.advance()
	case data.CommandNameMoveCharacter:
		ch := i.character(i.eventID)
		if ch == nil {
			i.commandIndex.advance()
			return true, nil
		}
		// Check IsMoving() first since the character might be moving at this time.
		if ch.IsMoving() {
			return false, nil
		}
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsMoveCharacter)
			ch.Move(args.Dir)
			i.waitingCommand = true
			return false, nil
		}
		i.waitingCommand = false
		i.commandIndex.advance()
	case data.CommandNameTurnCharacter:
		ch := i.character(i.eventID)
		if ch == nil {
			i.commandIndex.advance()
			return true, nil
		}
		// Check IsMoving() first since the character might be moving at this time.
		if ch.IsMoving() {
			return false, nil
		}
		if !i.waitingCommand {
			args := c.Args.(*data.CommandArgsTurnCharacter)
			ch.Turn(args.Dir)
			i.waitingCommand = true
			return false, nil
		}
		i.waitingCommand = false
		i.commandIndex.advance()
	case data.CommandNameRotateCharacter:
		ch := i.character(i.eventID)
		if ch == nil {
			i.commandIndex.advance()
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
			}
			switch args.Angle {
			case 0:
			case 90:
				dirI += 1
			case 180:
				dirI += 2
			case 270:
				dirI += 3
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
			}
			ch.Turn(dir)
			i.waitingCommand = true
			return false, nil
		}
		i.waitingCommand = false
		i.commandIndex.advance()
	case data.CommandNameSetInnerVariable:
		args := c.Args.(*data.CommandArgsSetInnerVariable)
		i.gameState.variables.SetInnerVariableValue(args.Name, args.Value)
		i.commandIndex.advance()
	default:
		return false, fmt.Errorf("invaid command: %s", c.Name)
	}
	return true, nil
}

func (i *Interpreter) Update() error {
	if i.commandIndex == nil {
		return nil
	}
	for !i.commandIndex.isTerminated() {
		cont, err := i.doOneCommand()
		if err != nil {
			return err
		}
		if !cont {
			break
		}
	}
	if i.commandIndex.isTerminated() {
		if i.gameState.windows.IsBusy() {
			return nil
		}
		i.gameState.windows.CloseAll()
		i.commandIndex = nil
		return nil
	}
	return nil
}

func (i *Interpreter) setVariable(id int, op data.SetVariableOp, valueType data.SetVariableValueType, value interface{}) {
	rhs := 0
	switch valueType {
	case data.SetVariableValueTypeConstant:
		rhs = value.(int)
	case data.SetVariableValueTypeVariable:
		rhs = i.gameState.variables.VariableValue(value.(int))
	case data.SetVariableValueTypeRandom:
		println(fmt.Sprintf("not implemented yet (set_variable): valueType %s", valueType))
		return
	case data.SetVariableValueTypeCharacter:
		args := value.(*data.SetVariableCharacterArgs)
		ch := i.character(args.EventID)
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
			}
		default:
			println(fmt.Sprintf("not implemented yet (set_variable): type %s", args.Type))
		}
	}
	switch op {
	case data.SetVariableOpAssign:
	case data.SetVariableOpAdd:
		rhs += i.gameState.variables.VariableValue(id)
	case data.SetVariableOpSub:
		rhs -= i.gameState.variables.VariableValue(id)
	case data.SetVariableOpMul:
		rhs *= i.gameState.variables.VariableValue(id)
	case data.SetVariableOpDiv:
		rhs /= i.gameState.variables.VariableValue(id)
	case data.SetVariableOpMod:
		rhs %= i.gameState.variables.VariableValue(id)
	}
	i.gameState.variables.SetVariableValue(id, rhs)
}
