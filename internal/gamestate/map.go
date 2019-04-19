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
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	pathpkg "github.com/hajimehoshi/rpgsnack-runtime/internal/path"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/sort"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/tileset"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/variables"
)

type passableOnMap struct {
	through          bool
	ignoreCharacters bool
	m                *Map
}

func (p *passableOnMap) At(x, y int) bool {
	return p.m.Passable(p.through, x, y, p.ignoreCharacters)
}

type Map struct {
	player                      *character.Character
	mapID                       int
	roomID                      int
	events                      []*character.Character
	eventPageIndices            map[int]int
	executingEventIDByUserInput int
	interpreters                map[int]*Interpreter
	playerInterpreterID         int
	itemInterpreter             *Interpreter

	// Fields that are not dumped
	isTitle                   bool
	gameData                  *data.Game
	isPlayerMovingByUserInput bool
	origSpeed                 data.Speed
	pressedMapX               int
	pressedMapY               int
}

func NewMap() *Map {
	return &Map{
		mapID:        1,
		interpreters: map[int]*Interpreter{},
		pressedMapX:  -1,
		pressedMapY:  -1,
	}
}

func NewTitleMap() *Map {
	m := NewMap()
	m.isTitle = true
	return m
}

func (m *Map) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("player")
	e.EncodeInterface(m.player)

	e.EncodeString("mapId")
	e.EncodeInt(m.mapID)

	e.EncodeString("roomId")
	e.EncodeInt(m.roomID)

	e.EncodeString("events")
	e.BeginArray()
	for _, v := range m.events {
		e.EncodeInterface(v)
	}
	e.EndArray()

	e.EncodeString("eventPageIndices")
	e.BeginMap()
	for k, v := range m.eventPageIndices {
		e.EncodeInt(k)
		e.EncodeInt(v)
	}
	e.EndMap()

	e.EncodeString("executingEventIdByUserInput")
	e.EncodeInt(m.executingEventIDByUserInput)

	e.EncodeString("interpreters")
	e.BeginMap()
	for k, v := range m.interpreters {
		e.EncodeInt(k)
		e.EncodeInterface(v)
	}
	e.EndMap()

	e.EncodeString("playerInterpreterId")
	e.EncodeInt(m.playerInterpreterID)

	e.EncodeString("itemInterpreter")
	e.EncodeInterface(m.itemInterpreter)

	e.EndMap()
	return e.Flush()
}

func (m *Map) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		k := d.DecodeString()
		switch k {
		case "player":
			if !d.SkipCodeIfNil() {
				m.player = &character.Character{}
				d.DecodeInterface(m.player)
			}
		case "mapId":
			m.mapID = d.DecodeInt()
		case "roomId":
			m.roomID = d.DecodeInt()
		case "events":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				m.events = make([]*character.Character, n)
				for i := 0; i < n; i++ {
					if !d.SkipCodeIfNil() {
						m.events[i] = &character.Character{}
						d.DecodeInterface(m.events[i])
					}
				}
			}
		case "eventPageIndices":
			if !d.SkipCodeIfNil() {
				n := d.DecodeMapLen()
				m.eventPageIndices = map[int]int{}
				for i := 0; i < n; i++ {
					k := d.DecodeInt()
					v := d.DecodeInt()
					m.eventPageIndices[k] = v
				}
			}
		case "executingEventIdByUserInput":
			m.executingEventIDByUserInput = d.DecodeInt()
		case "interpreters":
			if !d.SkipCodeIfNil() {
				n := d.DecodeMapLen()
				m.interpreters = map[int]*Interpreter{}
				for i := 0; i < n; i++ {
					k := d.DecodeInt()
					m.interpreters[k] = nil
					if !d.SkipCodeIfNil() {
						m.interpreters[k] = &Interpreter{}
						d.DecodeInterface(m.interpreters[k])
					}
				}
			}
		case "playerInterpreterId":
			m.playerInterpreterID = d.DecodeInt()
		case "itemInterpreter":
			if !d.SkipCodeIfNil() {
				m.itemInterpreter = &Interpreter{}
				d.DecodeInterface(m.itemInterpreter)
			}
		default:
			if err := d.Error(); err != nil {
				return err
			}
			return fmt.Errorf("gamestate: Map.DecodeMsgpack failed: unknown key: %s", k)
		}
	}
	if err := d.Error(); err != nil {
		return fmt.Errorf("gamestate: Map.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (m *Map) addInterpreter(interpreter *Interpreter) {
	m.interpreters[interpreter.id] = interpreter
}

func (m *Map) setRoomID(gameState *Game, id int, interpreter *Interpreter) error {
	m.roomID = id
	m.executingEventIDByUserInput = 0
	m.events = nil
	m.eventPageIndices = map[int]int{}

	if m.CurrentRoom().AutoBGM {
		gameState.SetBGM(m.CurrentRoom().BGM)
	}

	for _, e := range m.CurrentRoom().Events {
		x, y := e.Position()
		event := character.NewEvent(e.ID(), x, y)
		m.events = append(m.events, event)
		m.eventPageIndices[event.EventID()] = character.PlayerEventID
	}
	sort.Slice(m.events, func(i, j int) bool {
		return m.events[i].EventID() < m.events[j].EventID()
	})
	m.resetInterpreters(gameState)
	if interpreter != nil {
		m.addInterpreter(interpreter)
	}
	return nil
}

func (m *Map) resetInterpreters(gameState *Game) {
	m.abortPlayerInterpreter(gameState)
	m.interpreters = map[int]*Interpreter{}
}

func (m *Map) IsBlockingEventExecuting() bool {
	for _, i := range m.interpreters {
		if i.id == m.playerInterpreterID {
			if m.IsPlayerMovingByUserInput() {
				continue
			} else if i.IsExecuting() {
				return true
			}
		}
		if i.route {
			continue
		}
		if i.parallel {
			continue
		}
		if i.IsExecuting() {
			return true
		}
	}
	if m.itemInterpreter != nil && m.itemInterpreter.IsExecuting() {
		return true
	}
	return false
}

func (m *Map) meetsPageCondition(gameState *Game, page *data.Page, eventID int) (bool, error) {
	for _, cond := range page.Conditions {
		m, err := gameState.MeetsCondition(cond, eventID)
		if err != nil {
			return false, err
		}
		if !m {
			return false, nil
		}
	}
	return true, nil
}

func (m *Map) calcPageIndex(gameState *Game, ch *character.Character) (int, error) {
	if ch.Erased() {
		return -1, nil
	}
	var event *data.Event
	for _, e := range m.CurrentRoom().Events {
		if e.ID() == ch.EventID() {
			event = e
			break
		}
	}
	if event == nil {
		// This can happen when the player resumes the game and
		// the event was deleted by the game editor.
		return -1, nil
	}
	for i := len(event.Pages()) - 1; i >= 0; i-- {
		page := event.Pages()[i]
		m, err := m.meetsPageCondition(gameState, page, event.ID())
		if err != nil {
			return 0, err
		}
		if m {
			return i, nil
		}
	}
	return -1, nil
}

func (m *Map) currentPage(event *character.Character) (*data.Page, int) {
	i := m.eventPageIndices[event.EventID()]
	if i == -1 {
		return nil, 0
	}
	for _, e := range m.CurrentRoom().Events {
		if e.ID() == event.EventID() {
			return e.Pages()[i], i
		}
	}
	panic(fmt.Sprintf("gamescene: no valid page was found"))
}

var GoToTitle = errors.New("go to title")

func (m *Map) removeNonPageRoutes(eventID int) {
	ids := []int{}
	for _, i := range m.interpreters {
		if i.eventID != eventID {
			continue
		}
		if i.route && !i.pageRoute {
			ids = append(ids, i.id)
		}
	}
	for _, id := range ids {
		delete(m.interpreters, id)
	}
}

func (m *Map) removeRoutes(eventID int) {
	ids := []int{}
	for id, i := range m.interpreters {
		if i.route && i.eventID == eventID {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		delete(m.interpreters, id)
	}
}

func (m *Map) FindImageName(imageID int) string {
	imageName := ""
	for _, tileSetItem := range m.gameData.TileSets {
		if tileSetItem.ID == imageID {
			imageName = tileSetItem.Name
			break
		}
	}
	return imageName
}

func (m *Map) FindImage(imageID int) *ebiten.Image {
	if imageName := m.FindImageName(imageID); imageName != "" {
		return assets.GetImage(imageName + ".png")
	}
	return nil
}

func (m *Map) Update(sceneManager *scene.Manager, gameState *Game) error {
	// TODO: This is a temporary hack for TileSet and currentMap.
	// Remove this if possible.
	m.gameData = sceneManager.Game()
	if m.player == nil {
		x, y, roomID := 0, 0, 1
		// sceneManager.Game().System.Title can be nil for old games.
		if m.isTitle && sceneManager.Game().System.Title != nil {
			m.player = character.NewPlayer(0, 0)
			roomID = sceneManager.Game().System.Title.RoomID
		} else {
			state := sceneManager.Game().System.InitialPlayerState
			if state != nil {
				x, y, roomID = state.X, state.Y, state.RoomID
			}
			m.player = character.NewPlayer(x, y)
			m.player.SetImage(state.ImageType, state.Image)
		}
		m.setRoomID(gameState, roomID, nil)
	}

	if m.itemInterpreter == nil {
		is := []*Interpreter{}

		adoptedRoutes := map[int]*Interpreter{}
		for _, i := range m.interpreters {
			if i.route {
				oldInt, ok := adoptedRoutes[i.eventID]
				// Prefer non-page-route.
				if !ok || (oldInt.pageRoute && !i.pageRoute) {
					adoptedRoutes[i.eventID] = i
				}
			}
			is = append(is, i)
		}
		sort.Slice(is, func(i, j int) bool {
			return is[i].id < is[j].id
		})
		for _, i := range is {
			if i.route {
				if i.pageRoute && m.executingEventIDByUserInput == i.eventID {
					continue
				}
				if adoptedRoutes[i.eventID] != i {
					continue
				}
			}
			if err := i.Update(sceneManager, gameState); err != nil {
				return err
			}
			if !i.IsExecuting() {
				if i.id == m.playerInterpreterID {
					m.executingEventIDByUserInput = 0
				}
				delete(m.interpreters, i.id)
			}
		}
	} else {
		if err := m.itemInterpreter.Update(sceneManager, gameState); err != nil {
			return err
		}
		if !m.itemInterpreter.IsExecuting() {
			m.itemInterpreter = nil
		}
	}
	m.player.Update()
	if err := m.refreshEvents(gameState); err != nil {
		return err
	}
	for _, e := range m.events {
		e.Update()
	}
	m.tryRunParallelEvent(gameState)
	if m.IsPlayerMovingByUserInput() {
		return nil
	}
	m.tryRunAutoEvent(gameState)

	m.pressedMapX = -1
	m.pressedMapY = -1

	return nil
}

func (m *Map) refreshEvents(gameState *Game) error {
	for _, e := range m.events {
		index, err := m.calcPageIndex(gameState, e)
		if err != nil {
			return err
		}
		p := m.eventPageIndices[e.EventID()]
		if p == index {
			continue
		}
		m.removeRoutes(e.EventID())
		m.eventPageIndices[e.EventID()] = index
		page, pageIndex := m.currentPage(e)
		e.UpdateWithPage(page)
		if page == nil {
			continue
		}
		route := page.Route
		if route == nil {
			continue
		}
		commands := []*data.Command{
			{
				Name: data.CommandNameSetRoute,
				Args: &data.CommandArgsSetRoute{
					EventID:  0,
					Repeat:   route.Repeat,
					Skip:     route.Skip,
					Wait:     route.Wait,
					Commands: route.Commands,
				},
			},
		}
		interpreter := NewInterpreter(gameState, m.mapID, m.roomID, e.EventID(), pageIndex, commands)
		interpreter.route = true
		interpreter.pageRoute = true
		m.addInterpreter(interpreter)
	}
	return nil
}

func (m *Map) eventsAt(x, y int) []*character.Character {
	es := []*character.Character{}
	for _, e := range m.events {
		ex, ey := e.Position()
		if ex == x && ey == y {
			es = append(es, e)
		}
	}
	return es
}

func (m *Map) tryRunParallelEvent(gameState *Game) {
	for _, e := range m.events {
		page, pageIndex := m.currentPage(e)

		// If there is already an executing interpreter, check the condition.
		// 1) If all the conditions are same, the new interpreter won't start and the existing interpreter
		//    continues.
		// 2) If all the conditions but the page index are same, the new interpreter starts and the existing
		//    interpreter stops.
		// 3) Otherwise, the new interpreter starts and the existing interpreter continues.
		id := e.EventID()
		interpreterToRemove := -1
		alreadyExecuting := false
		for _, i := range m.interpreters {
			if !i.parallel {
				continue
			}
			if i.mapID == m.mapID && i.roomID == m.roomID && i.eventID == id {
				if page != nil && i.pageIndex == pageIndex {
					alreadyExecuting = true
					break
				}
				if page == nil || i.pageIndex != pageIndex {
					interpreterToRemove = i.id
					break
				}
			}
		}
		if interpreterToRemove != -1 {
			delete(m.interpreters, interpreterToRemove)
		}
		if alreadyExecuting {
			continue
		}

		if page == nil {
			continue
		}
		if page.Trigger != data.TriggerParallel {
			continue
		}

		i := NewInterpreter(gameState, m.mapID, m.roomID, e.EventID(), pageIndex, page.Commands)
		i.parallel = true
		m.addInterpreter(i)
	}
}

func (m *Map) tryRunAutoEvent(gameState *Game) {
	if m.IsBlockingEventExecuting() {
		return
	}
	for _, e := range m.events {
		page, pageIndex := m.currentPage(e)
		if page == nil {
			continue
		}
		if page.Trigger != data.TriggerAuto {
			continue
		}
		// The event is not executed here since IsBlockingEventExecuting returns false.
		i := NewInterpreter(gameState, m.mapID, m.roomID, e.EventID(), pageIndex, page.Commands)
		m.addInterpreter(i)
		return
	}
}

func (m *Map) transferPlayerImmediately(gameState *Game, roomID, x, y int, interpreter *Interpreter) {
	m.player.TransferImmediately(x, y)
	m.setRoomID(gameState, roomID, interpreter)
}

func (m *Map) currentMap() *data.Map {
	for _, d := range m.gameData.Maps {
		if d.ID() == m.mapID {
			return d
		}
	}
	return nil
}

func (m *Map) CurrentRoom() *data.Room {
	for _, r := range m.currentMap().Rooms() {
		if r.ID == m.roomID {
			return r
		}
	}
	return nil
}

func (m *Map) IsPlayerMovingByUserInput() bool {
	return m.isPlayerMovingByUserInput
}

func (m *Map) FinishPlayerMovingByUserInput() {
	m.isPlayerMovingByUserInput = false
}

func (m *Map) FocusingCharacter() *character.Character {
	return m.player
}

func (m *Map) passableTile(x, y int) bool {
	tileIndex := tileset.TileIndex(x, y)
	passageTypeOverrides := m.CurrentRoom().PassageTypeOverrides
	if passageTypeOverrides != nil && passageTypeOverrides[tileIndex] == data.PassageTypeBlock {
		return false
	}

	for layer := 0; layer < 4; layer++ {
		tile := m.CurrentRoom().Tiles[layer][tileIndex]
		if tile == 0 {
			continue
		}
		imageID := tileset.ExtractImageID(tile)
		imageName := m.FindImageName(imageID)
		index := 0
		if !tileset.IsAutoTile(imageName) {
			x, y := tileset.DecodeTile(tile)
			index = tileset.TileIndex(x, y)
		}
		passageType := tileset.PassageType(imageName, index)
		if passageType == data.PassageTypeBlock {
			return false
		}
	}

	return true
}

func (m *Map) Passable(through bool, x, y int, ignoreCharacters bool) bool {
	if x < 0 {
		return false
	}
	if y < 0 {
		return false
	}
	if consts.TileXNum <= x {
		return false
	}
	if consts.TileYNum <= y {
		return false
	}
	if through {
		return true
	}
	if !m.passableTile(x, y) {
		return false
	}
	if ignoreCharacters {
		return true
	}
	es := m.eventsAt(x, y)
	for _, e := range es {
		if e.Through() {
			continue
		}
		if page, _ := m.currentPage(e); page != nil && page.Priority == data.PriorityMiddle {
			return false
		}
	}
	px, py := m.player.Position()
	if x == px && y == py {
		return false
	}
	return true
}

func (m *Map) SetPressedPosition(x, y int) {
	m.pressedMapX = x
	m.pressedMapY = y
}

func (m *Map) GetPressedPosition() (int, int) {
	return m.pressedMapX, m.pressedMapY
}

func (m *Map) abortPlayerInterpreter(gameState *Game) {
	if _, ok := m.interpreters[m.playerInterpreterID]; ok {
		delete(m.interpreters, m.playerInterpreterID)
		m.FinishPlayerMovingByUserInput()
		// TODO: Use m.player
		ch := gameState.Character(m.mapID, m.roomID, character.PlayerEventID)
		ch.SetSpeed(m.origSpeed)
	}
}

func (m *Map) TryRunDirectEvent(gameState *Game, x, y int) bool {
	if m.IsBlockingEventExecuting() {
		return false
	}

	// If there is an executable event at the current player's position, don't fire the direct event in order
	// not to skip the executable event.
	if _, ok := m.interpreters[m.playerInterpreterID]; ok {
		if m.executableEventAt(m.player.Position()) != nil {
			return false
		}
	}

	es := m.eventsAt(x, y)
	for _, e := range es {
		page, pageIndex := m.currentPage(e)
		if page == nil {
			continue
		}
		if len(page.Commands) == 0 {
			continue
		}
		if page.Trigger != data.TriggerDirect {
			continue
		}
		m.abortPlayerInterpreter(gameState)
		i := NewInterpreter(gameState, m.mapID, m.roomID, e.EventID(), pageIndex, page.Commands)
		m.addInterpreter(i)
		return true
	}
	return false
}

func (m *Map) executableEventAt(x, y int) *character.Character {
	// When the player is through-mode, any event should not be triggered (#710).
	if m.player.Through() {
		return nil
	}
	for _, e := range m.eventsAt(x, y) {
		// When an event is through-mode, the event should not be triggered (#710).
		if e.Through() {
			continue
		}
		page, _ := m.currentPage(e)
		if page == nil {
			continue
		}
		if len(page.Commands) == 0 {
			continue
		}
		if page.Trigger != data.TriggerPlayer {
			continue
		}
		return e
	}
	return nil
}

func (m *Map) TryMovePlayerByUserInput(sceneManager *scene.Manager, gameState *Game, x, y int) bool {
	if !gameState.IsPlayerControlEnabled() {
		return false
	}
	if m.IsBlockingEventExecuting() {
		return false
	}
	event := m.executableEventAt(x, y)
	if !m.Passable(m.player.Through(), x, y, false) && event == nil {
		return false
	}
	px, py := m.player.Position()
	path, lastPlayerX, lastPlayerY := pathpkg.Calc(&passableOnMap{
		through: m.player.Through(),
		m:       m,
	}, px, py, x, y, false)
	if len(path) == 0 {
		return false
	}

	// The player can move. Let's save the state here just before starting moving.
	if gameState.IsAutoSaveEnabled() && !m.IsPlayerMovingByUserInput() {
		gameState.RequestSave(0, sceneManager)
	}

	checkBottomEvent := false
	if _, ok := m.interpreters[m.playerInterpreterID]; ok {
		m.abortPlayerInterpreter(gameState)

		// As the last existing interpreter is aborted without checking the event existence,
		// checking events is necessary before moving the player.
		checkBottomEvent = true
	}

	// The player's speed is never changed by another events during the player walks
	// by user input.
	m.origSpeed = m.player.Speed()

	// Set this true before executing other interpreters, or
	// the calculated path can be invalidated.
	m.isPlayerMovingByUserInput = true

	commands := []*data.Command{
		{
			Name: data.CommandNameSetRoute,
			Args: &data.CommandArgsSetRoute{
				EventID:  character.PlayerEventID,
				Repeat:   false,
				Skip:     false,
				Wait:     true,
				Internal: true,
				Commands: []*data.Command{
					{
						Name: data.CommandNameSetCharacterProperty,
						Args: &data.CommandArgsSetCharacterProperty{
							Type:  data.SetCharacterPropertyTypeSpeed,
							Value: gameState.PlayerSpeed(),
						},
					},
				},
			},
		},
	}
	cs := pathpkg.RouteCommandsToEventCommands(path)

	const (
		oldXID = iota + variables.ReservedID
		oldYID
		newXID
		newYID
		moveCompleted
	)
	const labelPlayerMoved = "playerMoved"

	if checkBottomEvent {
		commands = append(commands,
			&data.Command{
				Name: data.CommandNameIf,
				Args: &data.CommandArgsIf{
					Conditions: []*data.Condition{
						{
							Type:  data.ConditionTypeSpecial,
							Value: specialConditionEventExistsAtPlayer,
						},
					},
				},
				Branches: [][]*data.Command{
					{
						{
							Name: data.CommandNameGoto,
							Args: &data.CommandArgsGoto{
								Label: labelPlayerMoved,
							},
						},
					},
				},
			})
	}

	commands = append(commands,
		&data.Command{
			Name: data.CommandNameSetVariable,
			Args: &data.CommandArgsSetVariable{
				ID:        moveCompleted,
				Op:        data.SetVariableOpAssign,
				ValueType: data.SetVariableValueTypeConstant,
				Value:     0,
				Internal:  true,
			},
		})
	for i, c := range cs {
		commands = append(commands,
			// Get the original position.
			&data.Command{
				Name: data.CommandNameSetVariable,
				Args: &data.CommandArgsSetVariable{
					ID:        oldXID,
					Op:        data.SetVariableOpAssign,
					ValueType: data.SetVariableValueTypeCharacter,
					Value: &data.SetVariableCharacterArgs{
						Type: data.SetVariableCharacterTypeRoomX,
					},
					Internal: true,
				},
			},
			&data.Command{
				Name: data.CommandNameSetVariable,
				Args: &data.CommandArgsSetVariable{
					ID:        oldYID,
					Op:        data.SetVariableOpAssign,
					ValueType: data.SetVariableValueTypeCharacter,
					Value: &data.SetVariableCharacterArgs{
						Type: data.SetVariableCharacterTypeRoomY,
					},
					Internal: true,
				},
			},
			// Try to move.
			&data.Command{
				Name: data.CommandNameSetRoute,
				Args: &data.CommandArgsSetRoute{
					EventID:  character.PlayerEventID,
					Repeat:   false,
					Skip:     true,
					Wait:     true,
					Internal: true,
					Commands: []*data.Command{c},
				},
			},
			// Get the current position.
			&data.Command{
				Name: data.CommandNameSetVariable,
				Args: &data.CommandArgsSetVariable{
					ID:        newXID,
					Op:        data.SetVariableOpAssign,
					ValueType: data.SetVariableValueTypeCharacter,
					Value: &data.SetVariableCharacterArgs{
						Type: data.SetVariableCharacterTypeRoomX,
					},
					Internal: true,
				},
			},
			&data.Command{
				Name: data.CommandNameSetVariable,
				Args: &data.CommandArgsSetVariable{
					ID:        newYID,
					Op:        data.SetVariableOpAssign,
					ValueType: data.SetVariableValueTypeCharacter,
					Value: &data.SetVariableCharacterArgs{
						Type: data.SetVariableCharacterTypeRoomY,
					},
					Internal: true,
				},
			},
			&data.Command{
				Name: data.CommandNameIf,
				Args: &data.CommandArgsIf{
					Conditions: []*data.Condition{
						{
							Type:  data.ConditionTypeSpecial,
							Value: specialConditionEventExistsAtPlayer,
						},
					},
				},
				Branches: [][]*data.Command{
					{
						{
							Name: data.CommandNameGoto,
							Args: &data.CommandArgsGoto{
								Label: labelPlayerMoved,
							},
						},
					},
				},
			},
		)
		if i < len(cs)-1 {
			c1 := &data.Command{
				Name: data.CommandNameIf,
				Args: &data.CommandArgsIf{
					Conditions: []*data.Condition{
						{
							Type:      data.ConditionTypeVariable,
							ID:        oldXID,
							Comp:      data.ConditionCompEqualTo,
							ValueType: data.ConditionValueTypeVariable,
							Value:     float64(newXID),
						},
					},
				},
				Branches: [][]*data.Command{
					{},
				},
			}
			c2 := &data.Command{
				Name: data.CommandNameIf,
				Args: &data.CommandArgsIf{
					Conditions: []*data.Condition{
						{
							Type:      data.ConditionTypeVariable,
							ID:        oldYID,
							Comp:      data.ConditionCompEqualTo,
							ValueType: data.ConditionValueTypeVariable,
							Value:     float64(newYID),
						},
					},
				},
				Branches: [][]*data.Command{
					{
						{
							Name: data.CommandNameGoto,
							Args: &data.CommandArgsGoto{
								Label: labelPlayerMoved,
							},
						},
					},
				},
			}
			c1.Branches[0] = append(c1.Branches[0], c2)
			commands = append(commands, c1)
		}
	}
	commands = append(commands,
		// Complete all the moving.
		&data.Command{
			Name: data.CommandNameSetVariable,
			Args: &data.CommandArgsSetVariable{
				ID:        moveCompleted,
				Op:        data.SetVariableOpAssign,
				ValueType: data.SetVariableValueTypeConstant,
				Value:     1,
				Internal:  true,
			},
		},
		&data.Command{
			Name: data.CommandNameLabel,
			Args: &data.CommandArgsLabel{
				Name: labelPlayerMoved,
			},
		},
		&data.Command{
			Name: data.CommandNameSetRoute,
			Args: &data.CommandArgsSetRoute{
				EventID:  character.PlayerEventID,
				Repeat:   false,
				Skip:     false,
				Wait:     true,
				Internal: true,
				Commands: []*data.Command{
					{
						Name: data.CommandNameSetCharacterProperty,
						Args: &data.CommandArgsSetCharacterProperty{
							Type:  data.SetCharacterPropertyTypeSpeed,
							Value: m.origSpeed,
						},
					},
				},
			},
		},
		&data.Command{
			Name: data.CommandNameFinishPlayerMovingByUserInput,
		},
	)

	// Execute an event if there is an event on the way of the route.
	// After executing the event, this interpreter terminates without executing
	// the targeted event.
	commands = append(commands,
		&data.Command{
			Name: data.CommandNameIf,
			Args: &data.CommandArgsIf{
				Conditions: []*data.Condition{
					{
						Type:  data.ConditionTypeSpecial,
						Value: specialConditionEventExistsAtPlayer,
					},
				},
			},
			Branches: [][]*data.Command{
				{
					{
						Name: data.CommandNameExecEventHere,
					},
					{
						Name: data.CommandNameReturn,
					},
				},
			},
		},
	)

	commands = append(commands,
		&data.Command{
			Name: data.CommandNameIf,
			Args: &data.CommandArgsIf{
				Conditions: []*data.Condition{
					{
						Type:      data.ConditionTypeVariable,
						ID:        moveCompleted,
						Comp:      data.ConditionCompEqualTo,
						ValueType: data.ConditionValueTypeConstant,
						Value:     0,
					},
				},
			},
			Branches: [][]*data.Command{
				{
					{
						Name: data.CommandNameReturn,
					},
				},
			},
		},
	)

	if event != nil {
		page, _ := m.currentPage(event)
		if page != nil && page.Trigger == data.TriggerPlayer {
			origDir := event.Dir()
			var dir data.Dir
			ex, ey := event.Position()
			px, py := lastPlayerX, lastPlayerY
			switch {
			case ex == px && ey == py:
				// The player and the event are at the same position.
				dir = event.Dir()
			case ex > px && ey == py:
				dir = data.DirLeft
			case ex < px && ey == py:
				dir = data.DirRight
			case ex == px && ey > py:
				dir = data.DirUp
			case ex == px && ey < py:
				dir = data.DirDown
			default:
				panic(fmt.Sprintf("gamestate: invalid positions: event: (%d, %d), player: (%d, %d)", ex, ey, px, py))
			}
			if !event.DirFix() {
				commands = append(commands,
					&data.Command{
						Name: data.CommandNameSetRoute,
						Args: &data.CommandArgsSetRoute{
							EventID:  event.EventID(),
							Repeat:   false,
							Skip:     false,
							Wait:     true,
							Internal: true,
							Commands: []*data.Command{
								{
									Name: data.CommandNameTurnCharacter,
									Args: &data.CommandArgsTurnCharacter{
										Dir: dir,
									},
								},
							},
						},
					})
			}
			commands = append(commands,
				&data.Command{
					Name: data.CommandNameCallEvent,
					Args: &data.CommandArgsCallEvent{
						EventID:   event.EventID(),
						PageIndex: m.eventPageIndices[event.EventID()],
					},
				})
			// TODO: DirFix state can be different when executing the event (#278).
			if !event.DirFix() {
				commands = append(commands,
					&data.Command{
						Name: data.CommandNameSetRoute,
						Args: &data.CommandArgsSetRoute{
							EventID:  event.EventID(),
							Repeat:   false,
							Skip:     false,
							Wait:     true,
							Internal: true,
							Commands: []*data.Command{
								{
									Name: data.CommandNameTurnCharacter,
									Args: &data.CommandArgsTurnCharacter{
										Dir: origDir,
									},
								},
							},
						},
					})
			}
		}
	}
	i := NewInterpreter(gameState, m.mapID, m.roomID, character.PlayerEventID, 0, commands)
	m.addInterpreter(i)
	m.playerInterpreterID = i.id
	if event != nil {
		m.executingEventIDByUserInput = event.EventID()
	}
	return true
}

func (m *Map) DrawCharacters(screen *ebiten.Image, priority data.Priority, offsetX, offsetY int) {
	chars := []*character.Character{}
	for _, e := range m.events {
		page, _ := m.currentPage(e)
		if page == nil {
			continue
		}
		if page.Priority != priority {
			continue
		}
		chars = append(chars, e)
	}
	if priority == data.PriorityMiddle {
		chars = append(chars, m.player)
	}
	sort.Slice(chars, func(i, j int) bool {
		_, yi := chars[i].DrawFootPosition()
		_, yj := chars[j].DrawFootPosition()
		if yi == yj {
			return chars[j].EventID() < chars[i].EventID()
		}
		return yi < yj
	})
	for _, c := range chars {
		c.Draw(screen, offsetX, offsetY)
	}
}

func (m *Map) StartItemCommands(gameState *Game, itemID int) {
	if m.itemInterpreter != nil {
		return
	}
	if itemID == 0 {
		return
	}
	var item *data.Item
	for _, i := range m.gameData.Items {
		if i.ID == itemID {
			item = i
			break
		}
	}
	if item.Commands == nil {
		return
	}
	m.itemInterpreter = NewInterpreter(gameState, m.mapID, m.roomID, 0, 0, item.Commands)
}

func (m *Map) StartCombineCommands(gameState *Game, combine *data.Combine) {
	if m.itemInterpreter != nil {
		return
	}
	if combine == nil {
		return
	}
	m.itemInterpreter = NewInterpreter(gameState, m.mapID, m.roomID, 0, 0, combine.Commands)
}

func (m *Map) Background(gameState *Game) string {
	if img, ok := gameState.Background(m.mapID, m.roomID); ok {
		return img
	}
	return m.CurrentRoom().Background.Name
}

func (m *Map) Foreground(gameState *Game) string {
	if img, ok := gameState.Foreground(m.mapID, m.roomID); ok {
		return img
	}
	return m.CurrentRoom().Foreground.Name
}
