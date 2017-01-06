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

package mapscene

import (
	"fmt"

	"github.com/hajimehoshi/ebiten"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type event struct {
	data               *data.Event
	mapScene           *MapScene
	character          *character
	currentPageIndex   int
	commandIndex       *commandIndex
	steppingCount      int
	selfSwitches       [data.SelfSwitchNum]bool
	waitingCount       int
	waitingTint        bool
	waitingMessage     bool
	waitingChoosing    bool
	waitingTransfering bool
	dirBeforeRunning   data.Dir
	executingPage      *data.Page
}

func newEvent(eventData *data.Event, mapScene *MapScene) (*event, error) {
	c := &character{
		x: eventData.X,
		y: eventData.Y,
	}
	e := &event{
		data:             eventData,
		mapScene:         mapScene,
		character:        c,
		currentPageIndex: -1,
	}
	if err := e.updateCharacterIfNeeded(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *event) currentPage() *data.Page {
	if e.currentPageIndex == -1 {
		return nil
	}
	return e.data.Pages[e.currentPageIndex]
}

func (e *event) isPassable() bool {
	page := e.currentPage()
	if page == nil {
		return true
	}
	return page.Priority != data.PrioritySameAsCharacters
}

func (e *event) isRunnable() bool {
	page := e.currentPage()
	if page == nil {
		return true
	}
	return len(page.Commands) > 0
}

func (e *event) updateCharacterIfNeeded() error {
	i, err := e.calcPageIndex()
	if err != nil {
		return err
	}
	if e.currentPageIndex == i {
		return nil
	}
	e.currentPageIndex = i
	e.steppingCount = 0
	if i == -1 {
		c := e.character
		c.imageName = ""
		c.imageIndex = 0
		c.dirFix = false
		c.dir = data.Dir(0)
		c.attitude = data.AttitudeMiddle
		return nil
	}
	page := e.data.Pages[i]
	c := e.character
	c.imageName = page.Image
	c.imageIndex = page.ImageIndex
	c.dirFix = page.DirFix
	c.dir = page.Dir
	// page.Attitude is ignored so far.
	c.attitude = data.AttitudeMiddle
	return nil
}

func (e *event) meetsCondition(cond *data.Condition) (bool, error) {
	// TODO: Is it OK to allow null conditions?
	if cond == nil {
		return true, nil
	}
	switch cond.Type {
	case data.ConditionTypeSwitch:
		id := cond.ID
		v := e.mapScene.state().Variables().SwitchValue(id)
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeSelfSwitch:
		v := e.selfSwitches[cond.ID]
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeVariable:
		id := cond.ID
		v := e.mapScene.state().Variables().VariableValue(id)
		rhs := cond.Value.(int)
		switch cond.ValueType {
		case data.ConditionValueTypeConstant:
		case data.ConditionValueTypeVariable:
			rhs = e.mapScene.state().Variables().VariableValue(rhs)
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

func (e *event) meetsPageCondition(page *data.Page) (bool, error) {
	for _, cond := range page.Conditions {
		m, err := e.meetsCondition(cond)
		if err != nil {
			return false, err
		}
		if !m {
			return false, nil
		}
	}
	return true, nil
}

func (e *event) calcPageIndex() (int, error) {
	for i := len(e.data.Pages) - 1; i >= 0; i-- {
		page := e.data.Pages[i]
		m, err := e.meetsPageCondition(page)
		if err != nil {
			return 0, err
		}
		if m {
			return i, nil
		}
	}
	return -1, nil
}

func (e *event) tryRun(trigger data.Trigger) bool {
	if e.executingPage != nil {
		return false
	}
	if trigger == data.TriggerNever {
		return false
	}
	page := e.currentPage()
	if page == nil {
		return false
	}
	if page.Trigger != trigger {
		return false
	}
	e.executingPage = page
	return true
}

func (e *event) updateCommands() error {
	if e.executingPage == nil {
		return nil
	}
	if e.mapScene.player.isMovingByUserInput() {
		return nil
	}
	if e.commandIndex == nil {
		e.dirBeforeRunning = e.character.dir
		var dir data.Dir
		ex, ey := e.character.x, e.character.y
		px, py := e.mapScene.player.character.x, e.mapScene.player.character.y
		switch {
		case e.executingPage.Trigger == data.TriggerAuto:
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
		e.character.turn(dir)
		// page.Attitude is ignored so far.
		e.character.attitude = data.AttitudeMiddle
		e.steppingCount = 0
		e.commandIndex = newCommandIndex(e.currentPage())
	}
	if e.mapScene.gameState.Screen().IsFading() {
		return nil
	}
commandLoop:
	for !e.commandIndex.isTerminated() {
		c := e.commandIndex.command()
		if c.Name == data.CommandNameShowChoices {
			// Note: This just waits for message balloons
			if e.mapScene.balloons.isAnimating() {
				break commandLoop
			}
		} else {
			if e.mapScene.balloons.isBusy() {
				break commandLoop
			}
		}
		switch c.Name {
		case data.CommandNameIf:
			conditions := c.Args.(*data.CommandArgsIf).Conditions
			matches := true
			for _, c := range conditions {
				m, err := e.meetsCondition(c)
				if err != nil {
					return err
				}
				if !m {
					matches = false
					break commandLoop
				}
			}
			if matches {
				e.commandIndex.choose(0)
			} else if len(c.Branches) >= 2 {
				e.commandIndex.choose(1)
			} else {
				e.commandIndex.advance()
			}
		case data.CommandNameCallEvent:
			println(fmt.Sprintf("not implemented yet: %s", c.Name))
			e.commandIndex.advance()
		case data.CommandNameWait:
			if e.waitingCount == 0 {
				e.waitingCount = c.Args.(*data.CommandArgsWait).Time * 6
				break commandLoop
			}
			if e.waitingCount > 0 {
				e.waitingCount--
				if e.waitingCount == 0 {
					e.commandIndex.advance()
					continue commandLoop
				}
				break commandLoop
			}
			e.commandIndex.advance()
		case data.CommandNameShowMessage:
			if !e.waitingMessage {
				args := c.Args.(*data.CommandArgsShowMessage)
				content := data.Current().Texts.Get(language.Und, args.ContentID)
				e.mapScene.showMessage(content, e.mapScene.character(args.EventID, e))
				e.waitingMessage = true
				break commandLoop
			}
			e.commandIndex.advance()
			e.waitingMessage = false
		case data.CommandNameShowChoices:
			if !e.waitingChoosing {
				choices := []string{}
				for _, id := range c.Args.(*data.CommandArgsShowChoices).ChoiceIDs {
					choice := data.Current().Texts.Get(language.Und, id)
					choices = append(choices, choice)
				}
				e.mapScene.showChoices(choices)
				e.waitingChoosing = true
				break commandLoop
			}
			if e.mapScene.balloons.isOpened() {
				// Index is not determined yet: this is hacky!
				break commandLoop
			}
			e.commandIndex.choose(e.mapScene.balloons.ChosenIndex())
			e.waitingChoosing = false
		case data.CommandNameSetSwitch:
			args := c.Args.(*data.CommandArgsSetSwitch)
			e.mapScene.state().Variables().SetSwitchValue(args.ID, args.Value)
			e.commandIndex.advance()
		case data.CommandNameSetSelfSwitch:
			args := c.Args.(*data.CommandArgsSetSelfSwitch)
			e.selfSwitches[args.ID] = args.Value
			e.commandIndex.advance()
		case data.CommandNameSetVariable:
			args := c.Args.(*data.CommandArgsSetVariable)
			e.setVariable(args.ID, args.Op, args.ValueType, args.Value)
			e.commandIndex.advance()
		case data.CommandNameTransfer:
			args := c.Args.(*data.CommandArgsTransfer)
			if !e.waitingTransfering {
				e.waitingTransfering = true
				e.mapScene.fadeOut(30)
				break commandLoop
			}
			e.mapScene.transferPlayerImmediately(args.RoomID, args.X, args.Y)
			e.mapScene.fadeIn(30)
			e.waitingTransfering = false
			// TODO: advance is called too early.
			e.commandIndex.advance()
		case data.CommandNameSetRoute:
			println(fmt.Sprintf("not implemented yet: %s", c.Name))
			e.commandIndex.advance()
		case data.CommandNameTintScreen:
			if !e.waitingTint {
				args := c.Args.(*data.CommandArgsTintScreen)
				r := float64(args.Red) / 255
				g := float64(args.Green) / 255
				b := float64(args.Blue) / 255
				gray := float64(args.Gray) / 255
				e.mapScene.gameState.Screen().StartTint(r, g, b, gray, args.Time*6)
				e.waitingTint = args.Wait
			}
			if e.waitingTint {
				if e.mapScene.gameState.Screen().IsChangingTint() {
					break commandLoop
				}
				e.waitingTint = false
				e.commandIndex.advance()
				continue commandLoop
			}
			e.commandIndex.advance()
		case data.CommandNamePlaySE:
			println(fmt.Sprintf("not implemented yet: %s", c.Name))
			e.commandIndex.advance()
		case data.CommandNamePlayBGM:
			println(fmt.Sprintf("not implemented yet: %s", c.Name))
			e.commandIndex.advance()
		case data.CommandNameStopBGM:
			println(fmt.Sprintf("not implemented yet: %s", c.Name))
			e.commandIndex.advance()
		default:
			return fmt.Errorf("command not implemented: %s", c.Name)
		}
	}
	if e.commandIndex.isTerminated() {
		if e.mapScene.balloons.isBusy() {
			return nil
		}
		e.character.turn(e.dirBeforeRunning)
		e.executingPage = nil
		e.commandIndex = nil
		return nil
	}
	return nil
}

func (e *event) setVariable(id int, op data.SetVariableOp, valueType data.SetVariableValueType, value interface{}) {
	rhs := 0
	switch valueType {
	case data.SetVariableValueTypeConstant:
		rhs = value.(int)
	case data.SetVariableValueTypeVariable:
		rhs = e.mapScene.state().Variables().VariableValue(value.(int))
	case data.SetVariableValueTypeRandom:
		println(fmt.Sprintf("not implemented yet (set_variable): valueType %s", valueType))
		return
	case data.SetVariableValueTypeCharacter:
		args := value.(*data.SetVariableCharacterArgs)
		ch := e.mapScene.character(args.EventID, e)
		switch args.Type {
		case data.SetVariableCharacterTypeDirection:
			switch ch.dir {
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
		rhs += e.mapScene.state().Variables().VariableValue(id)
	case data.SetVariableOpSub:
		rhs -= e.mapScene.state().Variables().VariableValue(id)
	case data.SetVariableOpMul:
		rhs *= e.mapScene.state().Variables().VariableValue(id)
	case data.SetVariableOpDiv:
		rhs /= e.mapScene.state().Variables().VariableValue(id)
	case data.SetVariableOpMod:
		rhs %= e.mapScene.state().Variables().VariableValue(id)
	}
	e.mapScene.state().Variables().SetVariableValue(id, rhs)
}

func (e *event) update() error {
	if err := e.updateCharacterIfNeeded(); err != nil {
		return err
	}
	page := e.currentPage()
	if page == nil {
		return nil
	}
	if page.Stepping {
		switch {
		case e.steppingCount < 30:
			e.character.attitude = data.AttitudeMiddle
		case e.steppingCount < 60:
			e.character.attitude = data.AttitudeLeft
		case e.steppingCount < 90:
			e.character.attitude = data.AttitudeMiddle
		default:
			e.character.attitude = data.AttitudeRight
		}
		e.steppingCount++
		e.steppingCount %= 120
	}
	if err := e.updateCommands(); err != nil {
		return err
	}
	return nil
}

func (e *event) draw(screen *ebiten.Image) error {
	if err := e.character.draw(screen); err != nil {
		return err
	}
	return nil
}
