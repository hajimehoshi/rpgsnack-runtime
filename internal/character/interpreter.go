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

package character

import (
	"fmt"

	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

// TODO: Remove this
type MapScene interface {
	Player() *Player
	Character(eventID int, self *Event) interface{}
	TransferPlayerImmediately(roomID, x, y int, event *Event)
}

type interpreter struct {
	gameState      *gamestate.Game
	event          *Event
	page           *data.Page
	commandIndex   *commandIndex
	waitingCount   int
	waitingCommand bool
	mapScene       MapScene
}

func (i *interpreter) SetEvent(event *Event) {
	i.event = event
}

func (i *interpreter) IsExecuting() bool {
	return i.page != nil
}

func (i *interpreter) SetPage(page *data.Page) {
	i.page = page
}

func (i *interpreter) MeetsCondition(cond *data.Condition) (bool, error) {
	// TODO: Is it OK to allow null conditions?
	if cond == nil {
		return true, nil
	}
	switch cond.Type {
	case data.ConditionTypeSwitch:
		id := cond.ID
		v := i.gameState.Variables().SwitchValue(id)
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeSelfSwitch:
		v := i.event.SelfSwitch(cond.ID)
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeVariable:
		id := cond.ID
		v := i.gameState.Variables().VariableValue(id)
		rhs := cond.Value.(int)
		switch cond.ValueType {
		case data.ConditionValueTypeConstant:
		case data.ConditionValueTypeVariable:
			rhs = i.gameState.Variables().VariableValue(rhs)
		default:
			return false, fmt.Errorf("mapscene: invalid value type: %s", cond.ValueType)
		}
		switch cond.Comp {
		case data.ConditionCompEqualTo:
			return v == rhs, nil
		case data.ConditionCompNotEqualTo:
			return v != rhs, nil
		case data.ConditionCompGreaterThanOrEqualTo:
			return v >= rhs, nil
		case data.ConditionCompGreaterThan:
			return v > rhs, nil
		case data.ConditionCompLessThanOrEqualTo:
			return v <= rhs, nil
		case data.ConditionCompLessThan:
			return v < rhs, nil
		default:
			return false, fmt.Errorf("mapscene: invalid comp: %s", cond.Comp)
		}
	default:
		return false, fmt.Errorf("mapscene: invalid condition: %s", cond)
	}
	return false, nil
}

func (i *interpreter) Update() error {
	if i.page == nil {
		return nil
	}
	if i.commandIndex == nil {
		var dir data.Dir
		ex, ey := i.event.Position()
		px, py := i.mapScene.Player().character.x, i.mapScene.Player().character.y
		switch {
		case i.page.Trigger == data.TriggerAuto:
		case ex == px && ey == py:
			// The player and the event are at the same position.
		case ex > px && ey == py:
			dir = data.DirLeft
		case ex < px && ey == py:
			dir = data.DirRight
		case ex == px && ey > py:
			dir = data.DirUp
		case ex == px && ey < py:
			dir = data.DirDown
		default:
			panic("not reach")
		}
		i.event.StartEvent(dir)
		i.commandIndex = newCommandIndex(i.event.CurrentPage().Commands)
	}
commandLoop:
	for !i.commandIndex.isTerminated() {
		c := i.commandIndex.command()
		if !i.gameState.Windows().CanProceed() {
			break commandLoop
		}
		switch c.Name {
		case data.CommandNameIf:
			conditions := c.Args.(*data.CommandArgsIf).Conditions
			matches := true
			for _, c := range conditions {
				m, err := i.MeetsCondition(c)
				if err != nil {
					return err
				}
				if !m {
					matches = false
					break commandLoop
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
			println(fmt.Sprintf("not implemented yet: %s", c.Name))
			i.commandIndex.advance()
		case data.CommandNameWait:
			if i.waitingCount == 0 {
				i.waitingCount = c.Args.(*data.CommandArgsWait).Time * 6
				break commandLoop
			}
			if i.waitingCount > 0 {
				i.waitingCount--
				if i.waitingCount == 0 {
					i.commandIndex.advance()
					continue commandLoop
				}
				break commandLoop
			}
			i.commandIndex.advance()
		case data.CommandNameShowMessage:
			if !i.waitingCommand {
				args := c.Args.(*data.CommandArgsShowMessage)
				content := data.Current().Texts.Get(language.Und, args.ContentID)
				ch := i.mapScene.Character(args.EventID, i.event)
				content = i.gameState.ParseMessageSyntax(content)
				x := 0
				y := 0
				switch ch := ch.(type) {
				case *Player:
					x, y = ch.character.x, ch.character.y
				case *Event:
					x, y = ch.character.x, ch.character.y
				default:
					panic("not reach")
				}
				i.gameState.Windows().ShowMessage(content, x*scene.TileSize, y*scene.TileSize)
				i.waitingCommand = true
				break commandLoop
			}
			// Advance command index first and check the next command.
			i.commandIndex.advance()
			if !i.commandIndex.isTerminated() {
				if i.commandIndex.command().Name != data.CommandNameShowChoices {
					i.gameState.Windows().CloseAll()
				}
			} else {
				i.gameState.Windows().CloseAll()
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
				i.gameState.Windows().ShowChoices(choices)
				i.waitingCommand = true
				break commandLoop
			}
			if !i.gameState.Windows().HasChosenIndex() {
				break commandLoop
			}
			i.commandIndex.choose(i.gameState.Windows().ChosenIndex())
			i.waitingCommand = false
		case data.CommandNameSetSwitch:
			args := c.Args.(*data.CommandArgsSetSwitch)
			i.gameState.Variables().SetSwitchValue(args.ID, args.Value)
			i.commandIndex.advance()
		case data.CommandNameSetSelfSwitch:
			args := c.Args.(*data.CommandArgsSetSelfSwitch)
			i.event.SetSelfSwitch(args.ID, args.Value)
			i.commandIndex.advance()
		case data.CommandNameSetVariable:
			args := c.Args.(*data.CommandArgsSetVariable)
			i.setVariable(args.ID, args.Op, args.ValueType, args.Value)
			i.commandIndex.advance()
		case data.CommandNameTransfer:
			args := c.Args.(*data.CommandArgsTransfer)
			if !i.waitingCommand {
				i.gameState.Screen().FadeOut(30)
				i.waitingCommand = true
				break commandLoop
			}
			if i.gameState.Screen().IsFadedOut() {
				i.mapScene.TransferPlayerImmediately(args.RoomID, args.X, args.Y, i.event)
				i.gameState.Screen().FadeIn(30)
				break commandLoop
			}
			if i.gameState.Screen().IsFading() {
				break commandLoop
			}
			i.waitingCommand = false
			i.commandIndex.advance()
		case data.CommandNameSetRoute:
			println(fmt.Sprintf("not implemented yet: %s", c.Name))
			i.commandIndex.advance()
		case data.CommandNameTintScreen:
			if !i.waitingCommand {
				args := c.Args.(*data.CommandArgsTintScreen)
				r := float64(args.Red) / 255
				g := float64(args.Green) / 255
				b := float64(args.Blue) / 255
				gray := float64(args.Gray) / 255
				i.gameState.Screen().StartTint(r, g, b, gray, args.Time*6)
				if !args.Wait {
					i.commandIndex.advance()
					continue commandLoop
				}
				i.waitingCommand = args.Wait
			}
			if i.gameState.Screen().IsChangingTint() {
				break commandLoop
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
		default:
			return fmt.Errorf("command not implemented: %s", c.Name)
		}
	}
	if i.commandIndex.isTerminated() {
		if i.gameState.Windows().IsBusy() {
			return nil
		}
		i.gameState.Windows().CloseAll()
		i.event.EndEvent()
		i.page = nil
		i.commandIndex = nil
		return nil
	}
	return nil
}

func (i *interpreter) setVariable(id int, op data.SetVariableOp, valueType data.SetVariableValueType, value interface{}) {
	rhs := 0
	switch valueType {
	case data.SetVariableValueTypeConstant:
		rhs = value.(int)
	case data.SetVariableValueTypeVariable:
		rhs = i.gameState.Variables().VariableValue(value.(int))
	case data.SetVariableValueTypeRandom:
		println(fmt.Sprintf("not implemented yet (set_variable): valueType %s", valueType))
		return
	case data.SetVariableValueTypeCharacter:
		args := value.(*data.SetVariableCharacterArgs)
		ch := i.mapScene.Character(args.EventID, i.event)
		switch args.Type {
		case data.SetVariableCharacterTypeDirection:
			var dir data.Dir
			switch ch := ch.(type) {
			case *Player:
				dir = ch.character.dir
			case *Event:
				dir = ch.character.dir
			default:
				panic("not reach")
			}
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
		rhs += i.gameState.Variables().VariableValue(id)
	case data.SetVariableOpSub:
		rhs -= i.gameState.Variables().VariableValue(id)
	case data.SetVariableOpMul:
		rhs *= i.gameState.Variables().VariableValue(id)
	case data.SetVariableOpDiv:
		rhs /= i.gameState.Variables().VariableValue(id)
	case data.SetVariableOpMod:
		rhs %= i.gameState.Variables().VariableValue(id)
	}
	i.gameState.Variables().SetVariableValue(id, rhs)
}
