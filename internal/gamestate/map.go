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
	"sort"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type passableOnMap struct {
	self             *character.Character
	ignoreCharacters bool
	m                *Map
}

func (p *passableOnMap) At(x, y int) (bool, error) {
	return p.m.passable(p.self, x, y, p.ignoreCharacters)
}

type Map struct {
	game                        *Game
	player                      *character.Character
	mapID                       int
	roomID                      int
	events                      []*character.Character
	eventPageIndices            map[*character.Character]int
	eventData                   map[*character.Character]*data.Event
	executingEventIDByUserInput int
	interpreters                map[int]*Interpreter
	playerInterpreterID         int
}

func NewMap(game *Game) (*Map, error) {
	pos := data.Current().System.InitialPosition
	x, y, roomID := 0, 0, 1
	if pos != nil {
		x, y, roomID = pos.X, pos.Y, pos.RoomID
	}
	player := character.NewPlayer(x, y)
	m := &Map{
		game:         game,
		player:       player,
		mapID:        1,
		interpreters: map[int]*Interpreter{},
	}
	m.setRoomID(roomID, nil)
	return m, nil
}

func (m *Map) addInterpreter(interpreter *Interpreter) {
	m.interpreters[interpreter.id] = interpreter
}

func (m *Map) TileSet() (*data.TileSet, error) {
	id := m.currentMap().TileSetID
	for _, t := range data.Current().TileSets {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("mapscene: tile set not found: %d", id)
}

func (m *Map) setRoomID(id int, interpreter *Interpreter) error {
	m.roomID = id
	m.events = nil
	m.eventPageIndices = map[*character.Character]int{}
	m.eventData = map[*character.Character]*data.Event{}
	for _, e := range m.CurrentRoom().Events {
		event := character.NewEvent(e.ID, e.X, e.Y)
		m.events = append(m.events, event)
		m.eventPageIndices[event] = -1
		m.eventData[event] = e
	}
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
		m, err := m.game.MeetsCondition(cond, eventID)
		if err != nil {
			return false, err
		}
		if !m {
			return false, nil
		}
	}
	return true, nil
}

func (m *Map) calcPageIndex(eventID int) (int, error) {
	var event *data.Event
	for _, e := range m.CurrentRoom().Events {
		if e.ID == eventID {
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
	i := m.eventPageIndices[event]
	if i == -1 {
		return nil
	}
	return m.eventData[event].Pages[i]
}

type interpretersByID []*Interpreter

func (i interpretersByID) Len() int           { return len(i) }
func (i interpretersByID) Less(a, b int) bool { return i[a].id < i[b].id }
func (i interpretersByID) Swap(a, b int)      { i[a], i[b] = i[b], i[a] }

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

func (m *Map) Update() error {
	is := []*Interpreter{}
	for _, i := range m.interpreters {
		is = append(is, i)
	}
	sort.Sort(interpretersByID(is))
	for _, i := range is {
		if m.IsPlayerMovingByUserInput() && i.id != m.playerInterpreterID {
			continue
		}
		if i.route && m.executingEventIDByUserInput == i.eventID {
			continue
		}
		if err := i.Update(); err != nil {
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
		index, err := m.calcPageIndex(e.ID())
		if err != nil {
			return err
		}
		p := m.eventPageIndices[e]
		if p == index {
			continue
		}
		m.removeRoutes(e.ID())
		m.eventPageIndices[e] = index
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
		interpreter := NewInterpreter(m.game, m.mapID, m.roomID, e.ID(), commands)
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

func (m *Map) eventAt(x, y int) *character.Character {
	for _, e := range m.events {
		ex, ey := e.Position()
		if ex == x && ey == y {
			return e
		}
	}
	return nil
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
		m.addInterpreter(NewInterpreter(m.game, m.mapID, m.roomID, e.ID(), page.Commands))
		break
	}
}

func (m *Map) transferPlayerImmediately(roomID, x, y int, interpreter *Interpreter) {
	m.player.TransferImmediately(x, y)
	m.setRoomID(roomID, interpreter)
}

func (m *Map) currentMap() *data.Map {
	for _, d := range data.Current().Maps {
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

func (m *Map) passableTile(x, y int) (bool, error) {
	tileSet, err := m.TileSet()
	if err != nil {
		return false, err
	}
	layer := 1
	tile := m.CurrentRoom().Tiles[layer][y*scene.TileXNum+x]
	switch tileSet.PassageTypes[layer][tile] {
	case data.PassageTypeBlock:
		return false, nil
	case data.PassageTypePassable:
		return true, nil
	case data.PassageTypeWall:
		panic("not implemented")
	case data.PassageTypeOver:
	default:
		panic("not reach")
	}
	layer = 0
	tile = m.CurrentRoom().Tiles[layer][y*scene.TileXNum+x]
	if tileSet.PassageTypes[layer][tile] == data.PassageTypePassable {
		return true, nil
	}
	return false, nil
}

func (m *Map) passable(self *character.Character, x, y int, ignoreCharacters bool) (bool, error) {
	if x < 0 {
		return false, nil
	}
	if y < 0 {
		return false, nil
	}
	if scene.TileXNum <= x {
		return false, nil
	}
	if scene.TileYNum <= y {
		return false, nil
	}
	if self.Through() {
		return true, nil
	}
	p, err := m.passableTile(x, y)
	if err != nil {
		return false, err
	}
	if !p {
		return false, nil
	}
	if ignoreCharacters {
		return true, nil
	}
	e := m.eventAt(x, y)
	if e != nil && !e.Through() {
		if page := m.currentPage(e); page != nil && page.Priority == data.PrioritySameAsCharacters {
			return false, nil
		}
	}
	px, py := m.player.Position()
	if x == px && y == py {
		return false, nil
	}
	return true, nil
}

func (m *Map) TryRunDirectEvent(x, y int) (bool, error) {
	if m.IsEventExecuting() {
		return false, nil
	}
	event := m.eventAt(x, y)
	if event == nil {
		return false, nil
	}
	page := m.currentPage(event)
	if page == nil {
		return false, nil
	}
	if len(page.Commands) == 0 {
		return false, nil
	}
	if page.Trigger != data.TriggerDirect {
		return false, nil
	}
	i := NewInterpreter(m.game, m.mapID, m.roomID, event.ID(), page.Commands)
	m.addInterpreter(i)
	return true, nil
}

func (m *Map) TryMovePlayerByUserInput(x, y int) (bool, error) {
	if m.IsEventExecuting() {
		return false, nil
	}
	event := m.eventAt(x, y)
	p, err := m.passable(m.player, x, y, false)
	if err != nil {
		return false, err
	}
	if !p {
		if event == nil {
			return false, nil
		}
		if page := m.currentPage(event); page != nil {
			if len(page.Commands) == 0 {
				return false, nil
			}
			if page.Trigger != data.TriggerPlayer {
				return false, nil
			}
		}
	}
	px, py := m.player.Position()
	path, lastPlayerX, lastPlayerY, err := calcPath(&passableOnMap{
		self: m.player,
		m:    m,
	}, px, py, x, y)
	if err != nil {
		return false, err
	}
	if len(path) == 0 {
		return false, nil
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
				EventID: -1,
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
				EventID:  -1,
				Repeat:   false,
				Skip:     false,
				Wait:     true,
				Commands: routeCommandsToEventCommands(path),
			},
		},
		{
			Name: data.CommandNameSetRoute,
			Args: &data.CommandArgsSetRoute{
				EventID: -1,
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
							EventID: event.ID(),
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
						EventID:   event.ID(),
						PageIndex: m.eventPageIndices[event],
					},
				})
			if !event.DirFix() {
				commands = append(commands,
					&data.Command{
						Name: data.CommandNameSetRoute,
						Args: &data.CommandArgsSetRoute{
							EventID: event.ID(),
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
	i := NewInterpreter(m.game, m.mapID, m.roomID, -1, commands)
	m.addInterpreter(i)
	m.playerInterpreterID = i.id
	if event != nil {
		m.executingEventIDByUserInput = event.ID()
	}
	return true, nil
}

type charactersByY []*character.Character

func (c charactersByY) Len() int { return len(c) }
func (c charactersByY) Less(i, j int) bool {
	_, yi := c[i].Position()
	_, yj := c[j].Position()
	return yi < yj
}
func (c charactersByY) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (m *Map) DrawCharacters(screen *ebiten.Image) error {
	chars := []*character.Character{m.player}
	for _, e := range m.events {
		chars = append(chars, e)
	}
	sort.Sort(charactersByY(chars))
	for _, c := range chars {
		if err := c.Draw(screen); err != nil {
			return err
		}
	}
	return nil
}
