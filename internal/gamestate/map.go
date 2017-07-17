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
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	pathpkg "github.com/hajimehoshi/rpgsnack-runtime/internal/path"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/sort"
)

type passableOnMap struct {
	through          bool
	ignoreCharacters bool
	m                *Map
}

func (p *passableOnMap) At(x, y int) bool {
	return p.m.passable(p.through, x, y, p.ignoreCharacters)
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

	// Fields that are not dumped
	game     *Game
	gameData *data.Game
}

func NewMap(game *Game) *Map {
	m := &Map{
		game:         game,
		mapID:        1,
		interpreters: map[int]*Interpreter{},
	}
	return m
}

type tmpMap struct {
	Player                      *character.Character   `json:"player"`
	MapID                       int                    `json:"mapId"`
	RoomID                      int                    `json:"roomId"`
	Events                      []*character.Character `json:"events"`
	EventPageIndices            map[int]int            `json:"eventPageIndices"`
	ExecutingEventIDByUserInput int                    `json:"executingEventIdByUserInput"`
	Interpreters                map[int]*Interpreter   `json:"interpreters"`
	PlayerInterpreterID         int                    `json:"playerInterpreterId"`
}

func (m *Map) MarshalJSON() ([]uint8, error) {
	tmp := &tmpMap{
		Player:                      m.player,
		MapID:                       m.mapID,
		RoomID:                      m.roomID,
		Events:                      m.events,
		EventPageIndices:            m.eventPageIndices,
		ExecutingEventIDByUserInput: m.executingEventIDByUserInput,
		Interpreters:                m.interpreters,
		PlayerInterpreterID:         m.playerInterpreterID,
	}
	return json.Marshal(tmp)
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

	e.EndMap()
	return e.Flush()
}

func (m *Map) UnmarshalJSON(jsonData []uint8) error {
	var tmp *tmpMap
	if err := json.Unmarshal(jsonData, &tmp); err != nil {
		return err
	}
	m.player = tmp.Player
	m.mapID = tmp.MapID
	m.roomID = tmp.RoomID
	m.events = tmp.Events
	m.eventPageIndices = tmp.EventPageIndices
	m.executingEventIDByUserInput = tmp.ExecutingEventIDByUserInput
	m.interpreters = tmp.Interpreters
	m.playerInterpreterID = tmp.PlayerInterpreterID
	return nil
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

// setGame sets the current game. This is called only when unmarshalzing.
func (m *Map) setGame(game *Game) {
	m.game = game
	for _, i := range m.interpreters {
		i.setGame(game)
	}
}

func (m *Map) addInterpreter(interpreter *Interpreter) {
	m.interpreters[interpreter.id] = interpreter
}

func (m *Map) waitingRequestResponse() bool {
	for _, i := range m.interpreters {
		if i.waitingRequestResponse() {
			return true
		}
	}
	return false
}

func (m *Map) TileSet(layer int) *data.TileSet {
	tileSetName := m.CurrentRoom().TileSets[layer]
	for _, t := range m.gameData.TileSets {
		if t.Name == tileSetName {
			return t
		}
	}

	return nil
}

func (m *Map) setRoomID(id int, interpreter *Interpreter) error {
	m.roomID = id
	m.events = nil
	m.eventPageIndices = map[int]int{}
	for _, e := range m.CurrentRoom().Events {
		event := character.NewEvent(e.ID, e.X, e.Y)
		m.events = append(m.events, event)
		m.eventPageIndices[event.EventID()] = character.PlayerEventID
	}
	sort.Slice(m.events, func(i, j int) bool {
		return m.events[i].EventID() < m.events[j].EventID()
	})
	m.interpreters = map[int]*Interpreter{}
	if interpreter != nil {
		m.addInterpreter(interpreter)
	}
	return nil
}

func (m *Map) IsEventExecuting() bool {
	for _, i := range m.interpreters {
		if i.route {
			continue
		}
		if i.IsExecuting() {
			return true
		}
	}
	return false
}

func (m *Map) meetsPageCondition(page *data.Page, eventID int) (bool, error) {
	for _, cond := range page.Conditions {
		m, err := m.game.meetsCondition(cond, eventID)
		if err != nil {
			return false, err
		}
		if !m {
			return false, nil
		}
	}
	return true, nil
}

func (m *Map) calcPageIndex(ch *character.Character) (int, error) {
	if ch.Erased() {
		return -1, nil
	}
	var event *data.Event
	for _, e := range m.CurrentRoom().Events {
		if e.ID == ch.EventID() {
			event = e
			break
		}
	}
	if event == nil {
		panic("not reach")
	}
	for i := len(event.Pages) - 1; i >= 0; i-- {
		page := event.Pages[i]
		m, err := m.meetsPageCondition(page, event.ID)
		if err != nil {
			return 0, err
		}
		if m {
			return i, nil
		}
	}
	return -1, nil
}

func (m *Map) currentPage(event *character.Character) *data.Page {
	i := m.eventPageIndices[event.EventID()]
	if i == -1 {
		return nil
	}
	for _, e := range m.CurrentRoom().Events {
		if e.ID == event.EventID() {
			return e.Pages[i]
		}
	}
	panic("not reached")
}

var GoToTitle = errors.New("go to title")

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

func (m *Map) Update(sceneManager *scene.Manager) error {
	// TODO: This is a temporary hack for TileSet and currentMap.
	// Remove this if possible.
	m.gameData = sceneManager.Game()
	if m.player == nil {
		pos := sceneManager.Game().System.InitialPosition
		x, y, roomID := 0, 0, 1
		if pos != nil {
			x, y, roomID = pos.X, pos.Y, pos.RoomID
		}
		m.player = character.NewPlayer(x, y)
		m.setRoomID(roomID, nil)
	}
	is := []*Interpreter{}
	for _, i := range m.interpreters {
		is = append(is, i)
	}
	sort.Slice(is, func(i, j int) bool {
		return is[i].id < is[j].id
	})
	for _, i := range is {
		if m.IsPlayerMovingByUserInput() && i.id != m.playerInterpreterID {
			continue
		}
		if i.route && m.executingEventIDByUserInput == i.eventID {
			continue
		}
		if err := i.Update(sceneManager); err != nil {
			return err
		}
		if !i.IsExecuting() {
			if i.id == m.playerInterpreterID {
				m.executingEventIDByUserInput = 0
			}
			delete(m.interpreters, i.id)
		}
	}
	if err := m.player.Update(); err != nil {
		return err
	}
	if m.IsPlayerMovingByUserInput() {
		return nil
	}
	for _, e := range m.events {
		index, err := m.calcPageIndex(e)
		if err != nil {
			return err
		}
		p := m.eventPageIndices[e.EventID()]
		if p == index {
			continue
		}
		m.removeRoutes(e.EventID())
		m.eventPageIndices[e.EventID()] = index
		page := m.currentPage(e)
		if err := e.UpdateWithPage(page); err != nil {
			return err
		}
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
		interpreter := NewInterpreter(m.game, m.mapID, m.roomID, e.EventID(), commands)
		interpreter.route = true
		m.addInterpreter(interpreter)
	}
	for _, e := range m.events {
		if err := e.Update(); err != nil {
			return err
		}
	}
	m.tryRunAutoEvent()
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

func (m *Map) tryRunAutoEvent() {
	if m.IsEventExecuting() {
		return
	}
	for _, e := range m.events {
		page := m.currentPage(e)
		if page == nil {
			continue
		}
		if page.Trigger != data.TriggerAuto {
			continue
		}
		m.addInterpreter(NewInterpreter(m.game, m.mapID, m.roomID, e.EventID(), page.Commands))
		break
	}
}

func (m *Map) transferPlayerImmediately(roomID, x, y int, interpreter *Interpreter) {
	m.player.TransferImmediately(x, y)
	m.setRoomID(roomID, interpreter)
}

func (m *Map) currentMap() *data.Map {
	for _, d := range m.gameData.Maps {
		if d.ID == m.mapID {
			return d
		}
	}
	return nil
}

func (m *Map) CurrentRoom() *data.Room {
	for _, r := range m.currentMap().Rooms {
		if r.ID == m.roomID {
			return r
		}
	}
	return nil
}

func (m *Map) IsPlayerMovingByUserInput() bool {
	return m.game.variables.InnerVariableValue("is_player_moving_by_user_input") != 0
}

func (m *Map) passableTile(x, y int) bool {
	tileIndex := y*consts.TileXNum + x
	passageTypeOverrides := m.CurrentRoom().PassageTypeOverrides
	if passageTypeOverrides != nil {
		switch passageTypeOverrides[tileIndex] {
		case data.PassageOverrideTypePassable:
			return true
		case data.PassageOverrideTypeBlock:
			return false
		}
	}

	layer := 1
	tileSetTop := m.TileSet(layer)
	if tileSetTop != nil {
		tile := m.CurrentRoom().Tiles[layer][tileIndex]
		switch tileSetTop.PassageTypes[tile] {
		case data.PassageTypeBlock:
			return false
		case data.PassageTypePassable:
			return true
		case data.PassageTypeOver:
		default:
			panic("not reach")
		}
	}

	layer = 0
	tileSetBottom := m.TileSet(layer)
	if tileSetBottom != nil {
		tile := m.CurrentRoom().Tiles[layer][tileIndex]
		if tileSetBottom.PassageTypes[tile] == data.PassageTypeBlock {
			return false
		}
	}
	return true
}

func (m *Map) passable(through bool, x, y int, ignoreCharacters bool) bool {
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
		if page := m.currentPage(e); page != nil && page.Priority == data.PrioritySameAsCharacters {
			return false
		}
	}
	px, py := m.player.Position()
	if x == px && y == py {
		return false
	}
	return true
}

func (m *Map) TryRunDirectEvent(x, y int) bool {
	if m.IsEventExecuting() {
		return false
	}
	es := m.eventsAt(x, y)
	for _, e := range es {
		page := m.currentPage(e)
		if page == nil {
			continue
		}
		if len(page.Commands) == 0 {
			continue
		}
		if page.Trigger != data.TriggerDirect {
			continue
		}
		i := NewInterpreter(m.game, m.mapID, m.roomID, e.EventID(), page.Commands)
		m.addInterpreter(i)
		return true
	}
	return false
}

func (m *Map) executableEventAt(x, y int) *character.Character {
	for _, e := range m.eventsAt(x, y) {
		page := m.currentPage(e)
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

func (m *Map) TryMovePlayerByUserInput(sceneManager *scene.Manager, x, y int) bool {
	if !m.game.IsPlayerControlEnabled() {
		return false
	}
	if m.IsEventExecuting() {
		return false
	}
	event := m.executableEventAt(x, y)
	if !m.passable(m.player.Through(), x, y, false) && event == nil {
		return false
	}
	px, py := m.player.Position()
	path, lastPlayerX, lastPlayerY := pathpkg.Calc(&passableOnMap{
		through: m.player.Through(),
		m:       m,
	}, px, py, x, y)
	if len(path) == 0 {
		return false
	}
	// The player can move. Let's save the state here just before starting moving.
	if m.game.IsAutoSaveEnabled() {
		m.game.RequestSave(sceneManager)
	}

	// The player's speed is never changed by another events during the player walks
	// by user input.
	origSpeed := m.player.Speed()
	commands := []*data.Command{
		{
			Name: data.CommandNameSetInnerVariable,
			Args: &data.CommandArgsSetInnerVariable{
				Name:  "is_player_moving_by_user_input",
				Value: 1,
			},
		},
		{
			Name: data.CommandNameSetRoute,
			Args: &data.CommandArgsSetRoute{
				EventID: character.PlayerEventID,
				Repeat:  false,
				Skip:    false,
				Wait:    true,
				Commands: []*data.Command{
					{
						Name: data.CommandNameSetCharacterProperty,
						Args: &data.CommandArgsSetCharacterProperty{
							Type:  data.SetCharacterPropertyTypeSpeed,
							Value: data.Speed5,
						},
					},
				},
			},
		},
		{
			Name: data.CommandNameSetRoute,
			Args: &data.CommandArgsSetRoute{
				EventID:  character.PlayerEventID,
				Repeat:   false,
				Skip:     false,
				Wait:     true,
				Commands: pathpkg.RouteCommandsToEventCommands(path),
			},
		},
		{
			Name: data.CommandNameSetRoute,
			Args: &data.CommandArgsSetRoute{
				EventID: character.PlayerEventID,
				Repeat:  false,
				Skip:    false,
				Wait:    true,
				Commands: []*data.Command{
					{
						Name: data.CommandNameSetCharacterProperty,
						Args: &data.CommandArgsSetCharacterProperty{
							Type:  data.SetCharacterPropertyTypeSpeed,
							Value: origSpeed,
						},
					},
				},
			},
		},
		{
			Name: data.CommandNameSetInnerVariable,
			Args: &data.CommandArgsSetInnerVariable{
				Name:  "is_player_moving_by_user_input",
				Value: 0,
			},
		},
	}
	if event != nil {
		page := m.currentPage(event)
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
				panic("not reach")
			}
			if !event.DirFix() {
				commands = append(commands,
					&data.Command{
						Name: data.CommandNameSetRoute,
						Args: &data.CommandArgsSetRoute{
							EventID: event.EventID(),
							Repeat:  false,
							Skip:    false,
							Wait:    true,
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
			// TODO: DirFix state can be different when executing the event.
			if !event.DirFix() {
				commands = append(commands,
					&data.Command{
						Name: data.CommandNameSetRoute,
						Args: &data.CommandArgsSetRoute{
							EventID: event.EventID(),
							Repeat:  false,
							Skip:    false,
							Wait:    true,
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
	i := NewInterpreter(m.game, m.mapID, m.roomID, character.PlayerEventID, commands)
	m.addInterpreter(i)
	m.playerInterpreterID = i.id
	if event != nil {
		m.executingEventIDByUserInput = event.EventID()
	}
	return true
}

func (m *Map) DrawCharacters(screen *ebiten.Image) {
	chars := []*character.Character{m.player}
	for _, e := range m.events {
		chars = append(chars, e)
	}
	sort.Slice(chars, func(i, j int) bool {
		_, yi := chars[i].Position()
		_, yj := chars[j].Position()
		return yi < yj
	})
	for _, c := range chars {
		c.Draw(screen)
	}
}
